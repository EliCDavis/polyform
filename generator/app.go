package generator

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/formats/pgtf"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type NodeBuilder func() nodes.Node

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Producers   map[string]nodes.NodeOutput[Artifact]

	// Runtime data
	nodeIDs map[nodes.Node]string
	types   *refutil.TypeFactory
}

func (a *App) ApplyGraph(jsonPayload []byte) error {

	graph, err := pgtf.Unmarshal[Graph](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a pgtf: %w", err)
	}

	decoder, err := pgtf.NewDecoder(jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to build a pgtf decoder: %w", err)
	}

	log.Println(graph.Name)

	if graph.Name != "" {
		a.Name = graph.Name
	}

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
	createdNodes := make(map[string]nodes.Node)

	// Create the Nodes
	for nodeID, instanceDetails := range graph.Nodes {
		if nodeID == "" {
			panic("attempting to create a node without an ID")
		}
		newNode := a.types.New(instanceDetails.Type)
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
	for fileName, producerDetails := range graph.Producers {
		producerNode := createdNodes[producerDetails.NodeID]
		outPortVals := refutil.CallFuncValuesOfType(producerNode, producerDetails.Port)
		ref := outPortVals[0].(nodes.NodeOutput[Artifact])
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

func (a *App) Graph() []byte {
	g := Graph{
		Name:        a.Name,
		Version:     a.Version,
		Description: a.Description,
		WebScene:    a.WebScene,
		Producers:   make(map[string]ProducerSchema),
	}

	appNodeSchema := make(map[string]GraphNodeInstance)

	encoder := &pgtf.Encoder{}

	for node := range a.nodeIDs {
		id, ok := a.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := appNodeSchema[id]; ok {
			panic("not sure how this happened")
		}

		appNodeSchema[id] = a.buildNodeGraphInstanceSchema(node, encoder)
	}

	for key, producer := range a.Producers {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := a.nodeIDs[producer.Node()]
		node := appNodeSchema[id]
		appNodeSchema[id] = node

		g.Producers[key] = ProducerSchema{
			NodeID: id,
			Port:   producer.Port(),
		}
	}

	g.Nodes = appNodeSchema

	data, err := encoder.ToPgtf(g)
	if err != nil {
		panic(err)
	}
	return data
}

func writeProducersToZip(path string, producers map[string]nodes.NodeOutput[Artifact], zw *zip.Writer) error {
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
	if a.types == nil {
		a.types = Nodes()
	}
	if !a.types.TypeRegistered(v) {
		a.types.RegisterType(v)
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
		p.initializeForCLI(set)
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
	Commands    []*cliCommand
}

func BuildNodeTypeSchema(node nodes.Node) NodeTypeSchema {

	typeSchema := NodeTypeSchema{
		DisplayName: "Untyped",
		Outputs:     make([]NodeOutput, 0),
		Inputs:      make(map[string]NodeInput),
	}

	outputs := node.Outputs()
	for _, o := range outputs {
		typeSchema.Outputs = append(typeSchema.Outputs, NodeOutput{
			Name: o.NodeOutput.Port(),
			Type: o.Type,
		})
	}

	inputs := node.Inputs()
	for _, o := range inputs {
		typeSchema.Inputs[o.Name] = NodeInput{
			Type: o.Type,
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

	return typeSchema
}

func (a *App) recursivelyRegisterNodeTypes(node nodes.Node) {
	a.addType(node)
	for _, subDependency := range node.Dependencies() {
		a.recursivelyRegisterNodeTypes(subDependency.Dependency())
	}
}

func (a App) buildNodeGraphInstanceSchema(node nodes.Node, encoder *pgtf.Encoder) GraphNodeInstance {

	schema := GraphNodeInstance{
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]NodeDependencySchema, 0),
	}

	for _, subDependency := range node.Dependencies() {
		schema.Dependencies = append(schema.Dependencies, NodeDependencySchema{
			DependencyID:   a.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	if param, ok := node.(CustomGraphSerialization); ok {
		data, err := param.ToJSON(encoder)
		if err != nil {
			panic(err)
		}
		schema.Data = data
	}

	return schema
}

func (a App) buildNodeInstanceSchema(node nodes.Node) NodeInstanceSchema {

	schema := NodeInstanceSchema{
		Name:         "Unamed",
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]NodeDependencySchema, 0),
		Version:      node.Version(),
	}

	for _, subDependency := range node.Dependencies() {
		schema.Dependencies = append(schema.Dependencies, NodeDependencySchema{
			DependencyID:   a.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	if param, ok := node.(Parameter); ok {
		schema.Name = param.DisplayName()
		schema.parameter = param
		schema.Parameter = param.Schema()
	} else {
		named, ok := node.(nodes.Named)
		if ok {
			schema.Name = named.Name()
		}
	}

	return schema
}

func (a *App) buildIDsForNode(dep nodes.Node) {
	if a.nodeIDs == nil {
		a.nodeIDs = make(map[nodes.Node]string)
	}

	// IDs for this node has already been built.
	if _, ok := a.nodeIDs[dep]; ok {
		return
	}

	a.addType(dep)

	for _, subDependency := range dep.Dependencies() {
		a.buildIDsForNode(subDependency.Dependency())
	}

	id := fmt.Sprintf("Node-%d", len(a.nodeIDs))
	a.nodeIDs[dep] = id
}

func (a *App) Schema() AppSchema {
	a.SetupProducers()

	schema := AppSchema{
		Producers: make(map[string]ProducerSchema),
	}

	appNodeSchema := make(map[string]NodeInstanceSchema)

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

		schema.Producers[key] = ProducerSchema{
			NodeID: id,
			Port:   producer.Port(),
		}
	}

	schema.Nodes = appNodeSchema

	registeredTypes := a.types.Types()
	nodeTypes := make([]NodeTypeSchema, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		nodeInstance, ok := a.types.New(registeredType).(nodes.Node)
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
	schema.Types = nodeTypes

	return schema
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

func (a *App) Run() error {
	if a.Producers == nil || len(a.Producers) == 0 {
		return errors.New("application has not defined any producers")
	}

	os_setup(a)

	a.SetupProducers()

	commandMap := make(map[string]*cliCommand)

	var commands []*cliCommand
	commands = []*cliCommand{
		{
			Name:        "Generate",
			Description: "Runs all producers the app has defined and saves it to the file system",
			Aliases:     []string{"generate", "gen"},
			Run: func() error {
				generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
				a.initialize(generateCmd)
				folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
				if err := generateCmd.Parse(os.Args[2:]); err != nil {
					return err
				}
				return a.Generate(*folderFlag)
			},
		},
		{
			Name:        "Serve",
			Description: "Starts an http server and hosts a webplayer for configuring the models produced from this app",
			Aliases:     []string{"serve"},
			Run: func() error {
				serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
				a.initialize(serveCmd)
				hostFlag := serveCmd.String("host", "localhost", "interface to bind to")
				portFlag := serveCmd.String("port", "8080", "port to serve over")

				sslFlag := serveCmd.Bool("ssl", false, "Whether or not to use SSL")
				certFlag := serveCmd.String("ssl.cert", "cert.pem", "Path to cert file")
				keyFlag := serveCmd.String("ssl.key", "key.pem", "Path to key file")

				// Websocket
				maxMessageSizeFlag := serveCmd.Int64(
					"max-message-size",
					1024*2,
					"Maximum message size allowed from peer over websocketed connection",
				)

				pingPeriodFlag := serveCmd.Duration(
					"ping-period",
					time.Second*54,
					"Send pings to peer with this period over websocketed connection. Must be less than pongWait.",
				)

				pongWaitFlag := serveCmd.Duration(
					"pong-wait",
					time.Second*60,
					"Time allowed to read the next pong message from the peer over a websocketed connection.",
				)

				writeWaitFlag := serveCmd.Duration(
					"write-wait",
					time.Second*10,
					"Time allowed to write a message to the peer over a websocketed connection.",
				)

				if err := serveCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				server := AppServer{
					app:      a,
					host:     *hostFlag,
					port:     *portFlag,
					webscene: a.WebScene,

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
			Description: "Enumerates all generators, parameters, and producers in a heirarchial fashion formatted in JSON",
			Aliases:     []string{"outline"},
			Run: func() error {
				outlineCmd := flag.NewFlagSet("outline", flag.ExitOnError)
				a.initialize(outlineCmd)

				if err := outlineCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				data, err := json.MarshalIndent(a.Schema(), "", "    ")
				if err != nil {
					panic(err)
				}
				os.Stdout.Write(data)

				return nil
			},
		},
		{
			Name:        "Zip",
			Description: "Runs all producers defined and writes it to a zip file",
			Aliases:     []string{"zip", "z"},
			Run: func() error {
				zipCmd := flag.NewFlagSet("zip", flag.ExitOnError)
				a.initialize(zipCmd)
				fileFlag := zipCmd.String("file-name", "out.zip", "file to write the contents of the zip too")

				if err := zipCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				f, err := os.Create(*fileFlag)
				if err != nil {
					return err
				}
				defer f.Close()

				return a.WriteZip(f)
			},
		},
		{
			Name:        "Mermaid",
			Description: "Create a mermaid flow chart for a specific producer",
			Aliases:     []string{"mermaid"},
			Run: func() error {
				mermaidCmd := flag.NewFlagSet("mermaid", flag.ExitOnError)
				a.initialize(mermaidCmd)
				fileFlag := mermaidCmd.String("file-name", "", "Optional path to file to write content to")

				if err := mermaidCmd.Parse(os.Args[2:]); err != nil {
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
			Name:        "Help",
			Description: "",
			Aliases:     []string{"help", "h"},
			Run: func() error {
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
				return tmpl.Execute(os.Stdout, cliDetails)
			},
		},
	}

	for _, cmd := range commands {
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		return commandMap["help"].Run()
	}

	if cmd, ok := commandMap[argsWithoutProg[0]]; ok {
		return cmd.Run()
	}

	fmt.Fprintf(os.Stdout, "unrecognized command %s", argsWithoutProg[0])
	return nil
}
