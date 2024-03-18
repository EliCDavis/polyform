package generator

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/nodes"
)

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Producers   map[string]nodes.NodeOutput[Artifact]

	// Runtime data
	nodeIDs map[nodes.Node]string
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
		artifact := producer.Data()
		err = artifact.Write(f)
		if err != nil {
			return err
		}
		// log.Printf("wrote %s", filePath)
	}

	return nil
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

func (a App) buildSchemaForNode(dependency nodes.Node, currentSchema map[string]NodeSchema) {
	id, ok := a.nodeIDs[dependency]
	if !ok {
		panic(fmt.Errorf("node %v has not had an ID generated for it", dependency))
	}

	if _, ok := currentSchema[id]; ok {
		return
	}

	schema := NodeSchema{
		Name:         "Unamed",
		Dependencies: make([]NodeDependencySchema, 0),
		Outputs:      make([]NodeOutput, 0),
		Version:      dependency.Version(),
	}

	for _, subDependency := range dependency.Dependencies() {
		a.buildSchemaForNode(subDependency.Dependency(), currentSchema)
		schema.Dependencies = append(schema.Dependencies, NodeDependencySchema{
			DependencyID: a.nodeIDs[subDependency.Dependency()],
			Name:         subDependency.Name(),
		})
	}

	outputs := dependency.Outputs()
	for _, o := range outputs {
		schema.Outputs = append(schema.Outputs, NodeOutput{
			Name: o.Name,
			Type: o.Type,
		})
	}

	inputs := dependency.Inputs()
	for _, o := range inputs {
		schema.Inputs = append(schema.Inputs, NodeInput{
			Name: o.Name,
			Type: o.Type,
		})
	}

	if param, ok := dependency.(Parameter); ok {
		schema.Name = param.DisplayName()
		schema.parameter = param
		schema.Parameter = param.Schema()
	} else {
		named, ok := dependency.(nodes.Named)
		if ok {
			schema.Name = named.Name()
		}
	}

	currentSchema[id] = schema
}

func (a *App) buildIDsForNode(dep nodes.Node) {
	if a.nodeIDs == nil {
		a.nodeIDs = make(map[nodes.Node]string)
	}

	// IDs for this node has already been built.
	if _, ok := a.nodeIDs[dep]; ok {
		return
	}

	for _, subDependency := range dep.Dependencies() {
		a.buildIDsForNode(subDependency.Dependency())
	}

	id := fmt.Sprintf("Node-%d", len(a.nodeIDs))
	a.nodeIDs[dep] = id
}

func (a *App) Schema() AppSchema {
	if a.nodeIDs == nil {
		for _, producer := range a.Producers {
			a.buildIDsForNode(producer.Node())
		}
	}

	schema := AppSchema{
		Producers: make([]string, 0, len(a.Producers)),
	}

	appNodeSchema := make(map[string]NodeSchema)

	for key, producer := range a.Producers {
		schema.Producers = append(schema.Producers, key)
		a.buildSchemaForNode(producer.Node(), appNodeSchema)

		id := a.nodeIDs[producer.Node()]
		node := appNodeSchema[id]
		node.Name = key
		appNodeSchema[id] = node
	}

	schema.Nodes = appNodeSchema

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
		arifact := p.Data()
		err = arifact.Write(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Run() error {
	if a.Producers == nil || len(a.Producers) == 0 {
		return errors.New("application has not defined any producers")
	}

	os_setup(a)

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
