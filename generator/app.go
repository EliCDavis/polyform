package generator

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/cli"
	"github.com/EliCDavis/polyform/generator/edit"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
)

type App struct {
	Out   io.Writer
	Graph *graph.Instance
}

func (a *App) ApplySchema(jsonPayload []byte) error {
	return a.Graph.ApplyAppSchema(jsonPayload)
}

func (a *App) Schema() []byte {
	a.initGraphInstance()

	data, err := a.Graph.EncodeToAppSchema()

	if err != nil {
		panic(err)
	}

	return data
}

func (a App) initialize(set *flag.FlagSet) {
	a.Graph.InitializeFromCLI(set)
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

func (a *App) initGraphInstance() {
	if a.Graph != nil {
		return
	}
	a.Graph = graph.New(graph.Config{
		TypeFactory: types,
	})
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
				return graph.WriteToFolder(a.Graph, *folderFlag)
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

				server := edit.Server{
					Host:             *hostFlag,
					Port:             *portFlag,
					LaunchWebbrowser: *launchWebBrowser,

					Autosave:   *autoSave,
					ConfigPath: configFile,

					Tls:      *sslFlag,
					CertPath: *certFlag,
					KeyPath:  *keyFlag,

					ClientConfig: &room.ClientConfig{
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

				z := zip.NewWriter(out)

				if err := graph.WriteToZip(a.Graph, z); err != nil {
					return err
				}

				return z.Close()
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

				return graph.WriteMermaid(a.Graph, out)
			},
		},
		{
			Name:        "Documentation",
			Description: "Create a document describing all available nodes",
			Aliases:     []string{"documentation", "docs"},
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
					Title:       a.Graph.GetName(),
					Description: a.Graph.GetDescription(),
					Version:     a.Graph.GetVersion(),
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

				return graph.WriteSwagger(a.Graph, out)
			},
		},
		{
			Name:        "Help",
			Description: "",
			Aliases:     []string{"help", "h"},
			Run: func(appState *cli.RunState) error {
				cliDetails := appCLI{
					Name:        a.Graph.GetName(),
					Version:     a.Graph.GetVersion(),
					Authors:     a.Graph.GetAuthors(),
					Description: a.Graph.GetDescription(),
					Commands:    commands,
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
