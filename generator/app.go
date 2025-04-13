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
	"strings"
	"text/template"
	"time"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/cli"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *schema.WebScene
	Authors     []schema.Author
	Files       map[string]nodes.Output[manifest.Artifact]

	graphInstance *graph.Instance
	Out           io.Writer
}

func (a *App) ApplySchema(jsonPayload []byte) error {

	graph, err := jbtf.Unmarshal[schema.App](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a jbtf: %w", err)
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

	return a.graphInstance.ApplyAppSchema(jsonPayload)
}

func (a *App) Schema() []byte {
	a.initGraphInstance()
	g := schema.App{
		Name:        a.Name,
		Version:     a.Version,
		Description: a.Description,
		Authors:     a.Authors,
		WebScene:    a.WebScene,
		Producers:   make(map[string]schema.Producer),
	}

	encoder := &jbtf.Encoder{}
	a.graphInstance.EncodeToAppSchema(&g, encoder)

	data, err := encoder.ToPgtf(g)
	if err != nil {
		panic(err)
	}
	return data
}

func writeProducersToZip(path string, graph *graph.Instance, zw *zip.Writer) error {
	if graph == nil {
		panic("can't zip nil graph")
	}

	if zw == nil {
		panic("can't write to nil zip writer")
	}

	for _, file := range graph.ProducerNames() {
		filePath := path + file
		f, err := zw.Create(filePath)
		if err != nil {
			return err
		}
		artifact := graph.Artifact(file)
		err = artifact.Write(f)
		if err != nil {
			return err
		}
		// log.Printf("wrote %s", filePath)
	}

	return nil
}

func (a App) initialize(set *flag.FlagSet) {
	a.graphInstance.InitializeParameters(set)
}

func (a App) WriteZip(out io.Writer) error {
	z := zip.NewWriter(out)

	err := writeProducersToZip("", a.graphInstance, z)
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
	Authors     []schema.Author
	Commands    []*cli.Command
}

func (a App) Generate(outputPath string) error {
	for _, name := range a.graphInstance.ProducerNames() {
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
		arifact := a.graphInstance.Artifact(name)
		err = arifact.Write(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initGraphInstance() {
	if a.graphInstance != nil {
		return
	}
	a.graphInstance = graph.New(types)
	for name, file := range a.Files {
		a.graphInstance.AddProducer(name, file)
	}
}

func (a *App) Run(args []string) error {
	os_setup(a)
	a.initGraphInstance()

	configFile := ""
	var commands []*cli.Command
	commands = []*cli.Command{
		{
			Name:        "New",
			Description: "Create a new graph",
			Aliases:     []string{"new"},
			Run: func(state *cli.RunState) error {
				newCmd := flag.NewFlagSet("new", flag.ExitOnError)
				// newCmd.SetOutput(state.Out)
				a.initialize(newCmd)
				nameFlag := newCmd.String("name", "Graph", "name of the program")
				versionFlag := newCmd.String("version", "v0.0.1", "version of the program")
				descriptionFlag := newCmd.String("description", "", "description of the program")
				authorFlag := newCmd.String("author", "", "author of the program")
				outFlag := newCmd.String("out", "", "Optional path to file to write content to")

				if err := newCmd.Parse(state.Args); err != nil {
					return err
				}

				graph := schema.App{}

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
					graph.Authors = append(graph.Authors, schema.Author{
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

				schema := a.graphInstance.Schema()

				usedTypes := make(map[string]struct{})
				for _, n := range schema.Nodes {
					usedTypes[n.Type] = struct{}{}
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
			Name:        "Documentation",
			Description: "Create a document describing all savailable nodes",
			Aliases:     []string{"documentation"},
			Run: func(appState *cli.RunState) error {
				markdownCmd := flag.NewFlagSet("documentation", flag.ExitOnError)
				a.initialize(markdownCmd)
				fileFlag := markdownCmd.String("out", "", "Optional path to file to write content to")
				formatFlag := markdownCmd.String("format", "markdown", "How to write documentation [markdown, html]")

				if err := markdownCmd.Parse(appState.Args); err != nil {
					return err
				}

				format := strings.ToLower(strings.TrimSpace(*formatFlag))
				if format != "html" && format != "markdown" {
					return fmt.Errorf("unrecognized format %q", format)
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

				doc := DocumentationWriter{
					Title:       a.Name,
					Description: a.Description,
					Version:     a.Version,
					NodeTypes:   types,
				}

				switch format {
				case "markdown":
					return doc.WriteSingleMarkdown(out)

				case "html":
					return doc.WriteSingleHTML(out)

				}
				return nil
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

			err = a.ApplySchema(fileData)
			if err != nil {
				return err
			}

			configFile = config
			return nil
		},
	}

	if isWasm() {
		return cliApp.Run([]string{".", "edit"})
	}

	return cliApp.Run(args)
}
