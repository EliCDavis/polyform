package generator

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/cli"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Producers   map[string]nodes.NodeOutput[artifact.Artifact]

	Out io.Writer

	// Runtime data
	nodeIDs       map[nodes.Node]string
	graphMetadata *NestedSyncMap
}

func (a *App) ApplyGraph(jsonPayload []byte) error {

	graph, err := jbtf.Unmarshal[Graph](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}

	decoder, err := jbtf.NewDecoder(jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}

	if graph.Name != "" {
		a.Name = graph.Name
	}

	a.Authors = graph.Authors

	if graph.Version != "" {
		a.Version = graph.Version
	}

	if graph.Description != "" {
		a.Description = graph.Description
	}

	if graph.WebScene != nil {
		a.WebScene = graph.WebScene
	}

	a.nodeIDs = make(map[nodes.Node]string)
	a.graphMetadata = NewNestedSyncMap()
	a.graphMetadata.OverwriteData(graph.Metadata)
	createdNodes := make(map[string]nodes.Node)

	// Create the Nodes
	for nodeID, instanceDetails := range graph.Nodes {
		if nodeID == "" {
			panic("attempting to create a node without an ID")
		}
		newNode := types.New(instanceDetails.Type)
		casted, ok := newNode.(nodes.Node)
		if !ok {
			panic(fmt.Errorf("graph definition contained type that instantiated a non node: %s", instanceDetails.Type))
		}
		createdNodes[nodeID] = casted
		a.nodeIDs[casted] = nodeID
	}

	// Connect the nodes we just created
	for nodeID, instanceDetails := range graph.Nodes {
		node := createdNodes[nodeID]
		for _, dependency := range instanceDetails.Dependencies {

			outNode := createdNodes[dependency.DependencyID]
			outPortVals := refutil.CallFuncValuesOfType(outNode, dependency.DependencyPort)
			ref := outPortVals[0].(nodes.NodeOutputReference)

			node.SetInput(dependency.Name, nodes.Output{
				NodeOutput: ref,
			})
		}
	}

	// Set the Producers
	a.Producers = make(map[string]nodes.NodeOutput[artifact.Artifact])
	for fileName, producerDetails := range graph.Producers {
		producerNode := createdNodes[producerDetails.NodeID]
		outPortVals := refutil.CallFuncValuesOfType(producerNode, producerDetails.Port)
		ref := outPortVals[0].(nodes.NodeOutput[artifact.Artifact])
		if ref == nil {
			panic(fmt.Errorf("REF IS NIL FOR FILE %s (node id: %s) and port %s", fileName, producerDetails.NodeID, producerDetails.Port))
		}
		a.Producers[fileName] = ref
	}

	// Set Parameters
	for nodeID, instanceDetails := range graph.Nodes {
		nodeI := createdNodes[nodeID]
		if p, ok := nodeI.(CustomGraphSerialization); ok {
			err := p.FromJSON(decoder, instanceDetails.Data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) FindNodeByID(nodeId string) nodes.Node {
	var node nodes.Node
	for n, id := range a.nodeIDs {
		if id == nodeId {
			node = n
		}
	}
	return node
}

func (a *App) Graph() []byte {
	g := Graph{
		Name:        a.Name,
		Version:     a.Version,
		Description: a.Description,
		Authors:     a.Authors,
		WebScene:    a.WebScene,
		Producers:   make(map[string]schema.Producer),

		// TODO: Is this unsafe?
		Metadata: a.graphMetadata.data,
	}

	nodeInstances := make(map[string]GraphNodeInstance)

	encoder := &jbtf.Encoder{}

	for node := range a.nodeIDs {
		id, ok := a.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := nodeInstances[id]; ok {
			panic(fmt.Errorf("we've arrived to a invalid state. two nodes refer to the same ID. There's a bug somewhere"))
		}

		nodeInstances[id] = a.buildNodeGraphInstanceSchema(node, encoder)
	}

	for key, producer := range a.Producers {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := a.nodeIDs[producer.Node()]
		node := nodeInstances[id]
		nodeInstances[id] = node

		g.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.Port(),
		}
	}

	g.Nodes = nodeInstances

	data, err := encoder.ToPgtf(g)
	if err != nil {
		panic(err)
	}
	return data
}

func writeProducersToZip(path string, producers map[string]nodes.NodeOutput[artifact.Artifact], zw *zip.Writer) error {
	if producers == nil {
		panic("can't write nil producers")
	}

	if zw == nil {
		panic("can't write to nil zip writer")
	}

	for file, producer := range producers {
		filePath := path + file
		f, err := zw.Create(filePath)
		if err != nil {
			return err
		}
		artifact := producer.Value()
		err = artifact.Write(f)
		if err != nil {
			return err
		}
		// log.Printf("wrote %s", filePath)
	}

	return nil
}

func (a *App) nodeFromID(id string) nodes.Node {
	for node, nodeID := range a.nodeIDs {
		if nodeID == id {
			return node
		}
	}
	panic(fmt.Sprintf("no node with id '%s' found", id))
}

func (a *App) addType(v any) {
	if !types.TypeRegistered(v) {
		types.RegisterType(v)
	}
}

func (a App) getParameters() []Parameter {
	if a.Producers == nil {
		return nil
	}

	parameterSet := make(map[Parameter]struct{})
	for _, n := range a.Producers {
		params := recurseDependenciesType[Parameter](n.Node())
		for _, p := range params {
			parameterSet[p] = struct{}{}
		}
	}

	uniqueParams := make([]Parameter, 0, len(parameterSet))
	for p := range parameterSet {
		uniqueParams = append(uniqueParams, p)
	}

	return uniqueParams
}

func recurseDependenciesType[T any](dependent nodes.Dependent) []T {
	allDependencies := make([]T, 0)
	for _, dep := range dependent.Dependencies() {
		subDependent := dep.Dependency()
		subDependencies := recurseDependenciesType[T](subDependent)
		allDependencies = append(allDependencies, subDependencies...)

		ofT, ok := subDependent.(T)
		if ok {
			allDependencies = append(allDependencies, ofT)
		}
	}

	return allDependencies
}

func (a App) initialize(set *flag.FlagSet) {
	for _, p := range a.getParameters() {
		p.InitializeForCLI(set)
	}
}

func (a App) WriteZip(out io.Writer) error {
	z := zip.NewWriter(out)

	err := writeProducersToZip("", a.Producers, z)
	if err != nil {
		return err
	}

	return z.Close()
}

//go:embed cli.tmpl
var cliTemplate string

type appCLI struct {
	Name        string
	Version     string
	Description string
	Authors     []Author
	Commands    []*cli.Command
}

func BuildNodeTypeSchema(node nodes.Node) schema.NodeType {

	typeSchema := schema.NodeType{
		DisplayName: "Untyped",
		Outputs:     make([]schema.NodeOutput, 0),
		Inputs:      make(map[string]schema.NodeInput),
	}

	outputs := node.Outputs()
	for _, o := range outputs {
		typeSchema.Outputs = append(typeSchema.Outputs, schema.NodeOutput{
			Name: o.NodeOutput.Port(),
			Type: o.Type,
		})
	}

	inputs := node.Inputs()
	for _, o := range inputs {
		typeSchema.Inputs[o.Name] = schema.NodeInput{
			Type:    o.Type,
			IsArray: o.Array,
		}
	}

	if param, ok := node.(Parameter); ok {
		typeSchema.Parameter = param.Schema()
	}

	if typed, ok := node.(nodes.Typed); ok {
		typeSchema.DisplayName = typed.Type()
	} else {
		typeSchema.DisplayName = refutil.GetTypeName(node)
	}

	if pathed, ok := node.(nodes.Pathed); ok {
		typeSchema.Path = pathed.Path()
	} else {
		packagePath := refutil.GetPackagePath(node)
		if strings.Contains(packagePath, "/") {
			path := strings.Split(packagePath, "/")
			path = path[1:]
			if path[0] == "EliCDavis" {
				path = path[1:]
			}

			if path[0] == "polyform" {
				path = path[1:]
			}
			typeSchema.Path = strings.Join(path, "/")
		} else {
			typeSchema.Path = packagePath
		}
	}

	if described, ok := node.(nodes.Describable); ok {
		typeSchema.Info = described.Description()
	}

	return typeSchema
}

func (a *App) recursivelyRegisterNodeTypes(node nodes.Node) {
	a.addType(node)
	for _, subDependency := range node.Dependencies() {
		a.recursivelyRegisterNodeTypes(subDependency.Dependency())
	}
}

func (a App) buildNodeGraphInstanceSchema(node nodes.Node, encoder *jbtf.Encoder) GraphNodeInstance {

	nodeInstance := GraphNodeInstance{
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]schema.NodeDependency, 0),
	}

	for _, subDependency := range node.Dependencies() {
		nodeInstance.Dependencies = append(nodeInstance.Dependencies, schema.NodeDependency{
			DependencyID:   a.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	sort.Slice(nodeInstance.Dependencies, func(i, j int) bool {
		return strings.ToLower(nodeInstance.Dependencies[i].Name) < strings.ToLower(nodeInstance.Dependencies[j].Name)
	})

	if param, ok := node.(CustomGraphSerialization); ok {
		data, err := param.ToJSON(encoder)
		if err != nil {
			panic(err)
		}
		nodeInstance.Data = data
	}

	return nodeInstance
}

func (a App) buildNodeInstanceSchema(node nodes.Node) schema.NodeInstance {

	var metadata map[string]any
	metadataPath := "nodes." + a.nodeIDs[node]

	if a.graphMetadata.PathExists(metadataPath) {
		if data := a.graphMetadata.Get(metadataPath); data != nil {
			metadata = data.(map[string]any)
		}
	}

	nodeInstance := schema.NodeInstance{
		Name:         "Unamed",
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]schema.NodeDependency, 0),
		Version:      node.Version(),
		Metadata:     metadata,
	}

	for _, subDependency := range node.Dependencies() {
		nodeInstance.Dependencies = append(nodeInstance.Dependencies, schema.NodeDependency{
			DependencyID:   a.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	if param, ok := node.(Parameter); ok {
		nodeInstance.Name = param.DisplayName()
		nodeInstance.Parameter = param.Schema()
	} else {
		named, ok := node.(nodes.Named)
		if ok {
			nodeInstance.Name = named.Name()
		}
	}

	return nodeInstance
}

func (a *App) buildIDsForNode(node nodes.Node) {
	if a.nodeIDs == nil {
		a.nodeIDs = make(map[nodes.Node]string)
		a.graphMetadata = NewNestedSyncMap()
	}

	// IDs for this node has already been built.
	if _, ok := a.nodeIDs[node]; ok {
		return
	}

	a.addType(node)

	for _, subDependency := range node.Dependencies() {
		a.buildIDsForNode(subDependency.Dependency())
	}

	// TODO: UGLY UGLY UGLY UGLY
	i := len(a.nodeIDs)
	for {
		id := fmt.Sprintf("Node-%d", i)

		taken := false
		for _, usedId := range a.nodeIDs {
			if usedId == id {
				taken = true
			}
			if taken {
				break
			}
		}
		if !taken {
			a.nodeIDs[node] = id
			break
		}
		i++
	}

}

func (a *App) GetParameter(nodeId string) Parameter {
	node := a.FindNodeByID(nodeId)

	if node == nil {
		panic(fmt.Errorf("no node exists with id %q", nodeId))
	}

	param, ok := node.(Parameter)
	if !ok {
		panic(fmt.Errorf("node %q is not a parameter", nodeId))
	}

	return param
}

func (a *App) Schema() schema.App {
	a.SetupProducers()

	var noteMetadata map[string]any
	if notes := a.graphMetadata.Get("notes"); notes != nil {
		casted, ok := notes.(map[string]any)
		if ok {
			noteMetadata = casted
		}
	}

	appSchema := schema.App{
		Producers: make(map[string]schema.Producer),
		Notes:     noteMetadata,
	}

	appNodeSchema := make(map[string]schema.NodeInstance)

	for node := range a.nodeIDs {
		id, ok := a.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := appNodeSchema[id]; ok {
			panic("not sure how this happened")
		}

		appNodeSchema[id] = a.buildNodeInstanceSchema(node)
	}

	for key, producer := range a.Producers {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := a.nodeIDs[producer.Node()]
		node := appNodeSchema[id]
		node.Name = key
		appNodeSchema[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.Port(),
		}
	}

	appSchema.Nodes = appNodeSchema

	registeredTypes := types.Types()
	nodeTypes := make([]schema.NodeType, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		nodeInstance, ok := types.New(registeredType).(nodes.Node)
		if !ok {
			panic("Registered type is somehow not a node. Not sure how we got here :/")
		}
		if nodeInstance == nil {
			panic("New registered type")
		}
		// log.Printf("%T: %+v\n", nodeInstance, nodeInstance)
		b := BuildNodeTypeSchema(nodeInstance)
		b.Type = registeredType
		nodeTypes = append(nodeTypes, b)
	}
	appSchema.Types = nodeTypes

	return appSchema
}

func (a App) Generate(outputPath string) error {
	for name, p := range a.Producers {
		fp := path.Join(outputPath, name)

		// Producer names are paths which can contain subfolders, so be sure
		// the subfolders exist before creating the file
		err := os.MkdirAll(filepath.Dir(fp), os.ModeDir)
		if err != nil {
			return err
		}

		// Create the File
		f, err := os.Create(fp)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write data to file
		arifact := p.Value()
		err = arifact.Write(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) SetupProducers() {
	for _, p := range a.Producers {
		a.recursivelyRegisterNodeTypes(p.Node())
	}

	if a.nodeIDs == nil {
		for _, producer := range a.Producers {
			a.buildIDsForNode(producer.Node())
		}
	}
}

func (a *App) Run(args []string) error {
	// if a.Producers == nil || len(a.Producers) == 0 {
	// 	return errors.New("application has not defined any producers")
	// }

	os_setup(a)
	a.SetupProducers()

	configFile := ""
	var commands []*cli.Command
	commands = []*cli.Command{
		{
			Name:        "New",
			Description: "Create a new graph",
			Aliases:     []string{"new"},
			Run: func(state *cli.RunState) error {
				newCmd := flag.NewFlagSet("new", flag.ExitOnError)
				a.initialize(newCmd)
				nameFlag := newCmd.String("name", "Graph", "name of the program")
				versionFlag := newCmd.String("version", "v0.0.1", "version of the program")
				descriptionFlag := newCmd.String("description", "", "description of the program")
				authorFlag := newCmd.String("author", "", "author of the program")
				// authorFlag := newCmd.String("author", "", "author of the program")
				outFlag := newCmd.String("out", "", "Optional path to file to write content to")

				if err := newCmd.Parse(state.Args); err != nil {
					return err
				}

				graph := Graph{}

				if nameFlag != nil {
					graph.Name = *nameFlag
				}

				if versionFlag != nil {
					graph.Version = *versionFlag
				}

				if descriptionFlag != nil {
					graph.Description = *descriptionFlag
				}

				if authorFlag != nil && *authorFlag != "" {
					graph.Authors = append(graph.Authors, Author{
						Name: *authorFlag,
					})
				}

				data, err := json.MarshalIndent(graph, "", "\t")
				if err != nil {
					return err
				}

				var out io.Writer = state.Out
				if outFlag != nil && *outFlag != "" {
					f, err := os.Create(*outFlag)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}
				_, err = out.Write(data)
				return err
			},
		},
		{
			Name:        "Generate",
			Description: "Runs all producers the graph has defined and saves it to the file system",
			Aliases:     []string{"generate", "gen"},
			Run: func(appState *cli.RunState) error {
				generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
				a.initialize(generateCmd)
				folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
				if err := generateCmd.Parse(appState.Args); err != nil {
					return err
				}
				return a.Generate(*folderFlag)
			},
		},
		{
			Name:        "Edit",
			Description: "Starts an http server and hosts a webplayer for editing the execution graph",
			Aliases:     []string{"edit"},
			Run: func(appState *cli.RunState) error {
				editCmd := flag.NewFlagSet("edit", flag.ExitOnError)
				a.initialize(editCmd)
				hostFlag := editCmd.String("host", "localhost", "interface to bind to")
				portFlag := editCmd.String("port", "8080", "port to serve over")

				autoSave := editCmd.Bool("autosave", false, "Whether or not to save changes back to the graph loaded")
				launchWebBrowser := editCmd.Bool("launch-browser", true, "Whether or not to open the web page in the web browser")

				sslFlag := editCmd.Bool("ssl", false, "Whether or not to use SSL")
				certFlag := editCmd.String("ssl.cert", "cert.pem", "Path to cert file")
				keyFlag := editCmd.String("ssl.key", "key.pem", "Path to key file")

				// Websocket
				maxMessageSizeFlag := editCmd.Int64(
					"max-message-size",
					1024*2,
					"Maximum message size allowed from peer over websocketed connection",
				)

				pingPeriodFlag := editCmd.Duration(
					"ping-period",
					time.Second*54,
					"Send pings to peer with this period over websocketed connection. Must be less than pongWait.",
				)

				pongWaitFlag := editCmd.Duration(
					"pong-wait",
					time.Second*60,
					"Time allowed to read the next pong message from the peer over a websocketed connection.",
				)

				writeWaitFlag := editCmd.Duration(
					"write-wait",
					time.Second*10,
					"Time allowed to write a message to the peer over a websocketed connection.",
				)

				if err := editCmd.Parse(appState.Args); err != nil {
					return err
				}

				server := AppServer{
					app:              a,
					host:             *hostFlag,
					port:             *portFlag,
					webscene:         a.WebScene,
					launchWebbrowser: *launchWebBrowser,

					autosave:   *autoSave,
					configPath: configFile,

					tls:      *sslFlag,
					certPath: *certFlag,
					keyPath:  *keyFlag,

					clientConfig: &room.ClientConfig{
						MaxMessageSize: *maxMessageSizeFlag,
						PingPeriod:     *pingPeriodFlag,
						PongWait:       *pongWaitFlag,
						WriteWait:      *writeWaitFlag,
					},
				}
				return server.Serve()
			},
		},
		{
			Name:        "Outline",
			Description: "Enumerates all parameters and producers in a heirarchial fashion formatted in JSON",
			Aliases:     []string{"outline"},
			Run: func(appState *cli.RunState) error {
				outlineCmd := flag.NewFlagSet("outline", flag.ExitOnError)
				a.initialize(outlineCmd)

				if err := outlineCmd.Parse(appState.Args); err != nil {
					return err
				}

				schema := a.Schema()

				usedTypes := make(map[string]struct{})
				for _, n := range schema.Nodes {
					usedTypes[n.Type] = struct{}{}
				}

				for i := len(schema.Types) - 1; i >= 0; i-- {
					if _, ok := usedTypes[schema.Types[i].Type]; !ok {
						schema.Types = append(schema.Types[:i], schema.Types[i+1:]...)
					}
				}

				data, err := json.MarshalIndent(schema, "", "    ")
				if err != nil {
					return err
				}

				_, err = appState.Out.Write(data)
				return err
			},
		},
		{
			Name:        "Zip",
			Description: "Runs all producers defined and writes it to a zip file",
			Aliases:     []string{"zip", "z"},
			Run: func(appState *cli.RunState) error {
				zipCmd := flag.NewFlagSet("zip", flag.ExitOnError)
				a.initialize(zipCmd)
				fileFlag := zipCmd.String("out", "", "file to write the contents of the zip too")

				if err := zipCmd.Parse(appState.Args); err != nil {
					return err
				}

				var out io.Writer = appState.Out

				if fileFlag != nil && *fileFlag != "" {
					f, err := os.Create(*fileFlag)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}

				return a.WriteZip(out)
			},
		},
		{
			Name:        "Mermaid",
			Description: "Create a mermaid flow chart for a specific producer",
			Aliases:     []string{"mermaid"},
			Run: func(appState *cli.RunState) error {
				mermaidCmd := flag.NewFlagSet("mermaid", flag.ExitOnError)
				a.initialize(mermaidCmd)
				fileFlag := mermaidCmd.String("out", "", "Optional path to file to write content to")

				if err := mermaidCmd.Parse(appState.Args); err != nil {
					return err
				}

				var out io.Writer = os.Stdout

				if fileFlag != nil && *fileFlag != "" {
					f, err := os.Create(*fileFlag)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}

				return WriteMermaid(*a, out)
			},
		},
		{
			Name:        "Swagger",
			Description: "Create a swagger 2.0 file",
			Aliases:     []string{"swagger"},
			Run: func(appState *cli.RunState) error {
				swaggerCmd := flag.NewFlagSet("swagger", flag.ExitOnError)
				a.initialize(swaggerCmd)
				fileFlag := swaggerCmd.String("out", "", "Optional path to file to write content to")

				if err := swaggerCmd.Parse(appState.Args); err != nil {
					return err
				}

				var out io.Writer = appState.Out

				if fileFlag != nil && *fileFlag != "" {
					f, err := os.Create(*fileFlag)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}

				return a.WriteSwagger(out)
			},
		},
		{
			Name:        "Help",
			Description: "",
			Aliases:     []string{"help", "h"},
			Run: func(appState *cli.RunState) error {
				cliDetails := appCLI{
					Name:        a.Name,
					Version:     a.Version,
					Commands:    commands,
					Authors:     a.Authors,
					Description: a.Description,
				}

				if cliDetails.Version == "" {
					cliDetails.Version = "(no version)"
				}

				tmpl, err := template.New("CLI App").Parse(cliTemplate)
				if err != nil {
					return err
				}
				return tmpl.Execute(appState.Out, cliDetails)
			},
		},
	}

	cliApp := cli.App{
		Commands: commands,
		Out:      a.Out,
		ConfigProvided: func(config string) error {
			fileData, err := os.ReadFile(config)
			if err != nil {
				return err
			}

			err = a.ApplyGraph(fileData)
			if err != nil {
				return err
			}

			configFile = config
			a.SetupProducers()
			return nil
		},
	}

	return cliApp.Run(args)
}
