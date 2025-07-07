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
	Name        string
	Version     string
	Description string
	Authors     []schema.Author

	Out   io.Writer
	Graph *graph.Instance
}

func (a *App) ApplySchema(jsonPayload []byte) error {
	a.initGraphInstance()
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

func (a *App) loadGraphFromDisk(config string) error {
	if config == "" {
		return nil
	}

	fileData, err := os.ReadFile(config)
	if err != nil {
		return err
	}

	err = a.ApplySchema(fileData)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Run(args []string) error {
	os_setup(a)
	// a.initGraphInstance()

	graphFlagName := "graph"
	graphFlag := &cli.StringFlag{
		Name:        graphFlagName,
		Description: "graph to load",
		Action: func(s string) error {
			if s == "" && a.Graph == nil {
				return fmt.Errorf("graph flag is not provided and app has no embedded graph")
			}
			return a.loadGraphFromDisk(s)
		},
	}

	var commands []*cli.Command
	commands = []*cli.Command{
		{
			Name:        "New",
			Description: "Create a new graph",
			Aliases:     []string{"new"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "name",
					Value:       "Graph",
					Description: "name of the graph",
				},
				&cli.StringFlag{
					Name:        "version",
					Value:       "v0.0.1",
					Description: "version of the graph",
				},
				&cli.StringFlag{
					Name:        "description",
					Description: "description of the graph",
				},
				&cli.StringFlag{
					Name:        "author",
					Description: "author of the graph",
				},
				&cli.StringFlag{
					Name:        "out",
					Description: "Optional path to file to write content to",
				},
			},
			Run: func(state *cli.RunState) error {
				graph := schema.App{
					Name:        state.String("name"),
					Version:     state.String("version"),
					Description: state.String("description"),
				}

				authorFlag := state.String("author")
				if authorFlag != "" {
					graph.Authors = append(graph.Authors, schema.Author{Name: authorFlag})
				}

				data, err := json.MarshalIndent(graph, "", "\t")
				if err != nil {
					return err
				}

				var out io.Writer = state.Out
				outFlag := state.String("out")
				if outFlag != "" {
					f, err := os.Create(outFlag)
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "folder",
					Value:       ".",
					Description: "folder to save generated contents to",
				},
				graphFlag,
			},
			Run: func(appState *cli.RunState) error {

				return graph.WriteToFolder(a.Graph, appState.String("folder"))
			},
		},
		{
			Name:        "Edit",
			Description: "Starts an http server and hosts a webplayer for editing the execution graph",
			Aliases:     []string{"edit"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "host",
					Value:       "localhost",
					Description: "interface to bind to",
				},
				&cli.StringFlag{
					Name:        "port",
					Value:       "8080",
					Description: "port to serve over",
				},
				&cli.BoolFlag{
					Name:        "autosave",
					Description: "Whether or not to save changes back to the loaded graph",
				},
				&cli.BoolFlag{
					Name:        "launch-browser",
					Description: "Whether or not to open the web page in the web browser",
					Value:       true,
				},
				&cli.BoolFlag{
					Name:        "ssl",
					Description: "Whether or not to use SSL",
				},
				&cli.StringFlag{
					Name:        "ssl.cert",
					Value:       "cert.pem",
					Description: "Path to cert file",
				},
				&cli.StringFlag{
					Name:        "ssl.key",
					Value:       "key.pem",
					Description: "Path to key file",
				},
				&cli.DurationFlag{
					Name:        "ping-period",
					Value:       time.Second * 54,
					Description: "Send pings to peer with this period over websocketed connection. Must be less than pongWait",
				},
				&cli.DurationFlag{
					Name:        "pong-wait",
					Value:       time.Second * 60,
					Description: "Time allowed to read the next pong message from the peer over a websocketed connection",
				},
				&cli.DurationFlag{
					Name:        "write-wait",
					Value:       time.Second * 10,
					Description: "Time allowed to write a message to the peer over a websocketed connection",
				},
				&cli.Int64Flag{
					Name:        "max-message-size",
					Value:       1024 * 2,
					Description: "Maximum message size allowed from peer over websocketed connection",
				},
				&cli.StringFlag{
					Name:        graphFlagName,
					Description: "graph to load",
					Action: func(s string) error {
						if s == "" {
							return nil
						}
						return a.loadGraphFromDisk(s)
					},
				},
			},
			Run: func(appState *cli.RunState) error {
				server := edit.Server{
					Graph:            a.Graph,
					Host:             appState.String("host"),
					Port:             appState.String("port"),
					LaunchWebbrowser: appState.Bool("launch-browser"),

					Autosave:   appState.Bool("autosave"),
					ConfigPath: appState.String(graphFlagName),

					Tls:      appState.Bool("ssl"),
					CertPath: appState.String("ssl.cert"),
					KeyPath:  appState.String("ssl.key"),

					ClientConfig: &room.ClientConfig{
						MaxMessageSize: appState.Int64("max-message-size"),
						PingPeriod:     appState.Duration("ping-period"),
						PongWait:       appState.Duration("pong-wait"),
						WriteWait:      appState.Duration("write-wait"),
					},
				}
				return server.Serve()
			},
		},
		{
			Name:        "Zip",
			Description: "Runs all producers defined and writes it to a zip file",
			Aliases:     []string{"zip", "z"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "out",
					Description: "file to write the contents of the zip too",
				},
				graphFlag,
			},
			Run: func(appState *cli.RunState) error {
				if err := a.loadGraphFromDisk(appState.String(graphFlagName)); err != nil {
					return err
				}

				fileFlag := appState.String("out")

				var out io.Writer = appState.Out
				if fileFlag != "" {
					f, err := os.Create(fileFlag)
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "out",
					Description: "Optional path to file to write content to",
				},
				graphFlag,
			},
			Run: func(appState *cli.RunState) error {
				var out io.Writer = os.Stdout
				fileFlag := appState.String("out")
				if fileFlag != "" {
					f, err := os.Create(fileFlag)
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "out",
					Description: "Optional path to file to write content to",
				},
				&cli.StringFlag{
					Name:        "format",
					Value:       "markdown",
					Description: "How to write documentation [markdown, html]",
				},
			},
			Run: func(appState *cli.RunState) error {
				format := strings.ToLower(strings.TrimSpace(appState.String("format")))
				if format != "html" && format != "markdown" {
					return fmt.Errorf("unrecognized format %q", format)
				}

				var out io.Writer = os.Stdout

				fileFlag := appState.String("out")
				if fileFlag != "" {
					f, err := os.Create(fileFlag)
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "out",
					Description: "Optional path to file to write content to",
				},
				graphFlag,
			},
			Run: func(appState *cli.RunState) error {
				var out io.Writer = appState.Out
				fileFlag := appState.String("out")
				if fileFlag != "" {
					f, err := os.Create(fileFlag)
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
					Name:        a.Name,
					Version:     a.Version,
					Authors:     a.Authors,
					Description: a.Description,
					Commands:    commands,
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
	}

	if isWasm() {
		return cliApp.Run([]string{".", "edit"})
	}

	return cliApp.Run(args)
}
