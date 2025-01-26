package generator

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
	"github.com/urfave/cli/v2"
)

func (a *App) cli() error {
	app := &cli.App{
		Name:    a.Name,
		Usage:   a.Description,
		Version: a.Version,
		Action: func(ctx *cli.Context) error {
			log.Printf("I hav ebeen called")
			return nil
		},
		Before: func(ctx *cli.Context) error {
			log.Printf("before called")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Usage:   "Runs all producers the app has defined and saves it to the file system",
				Action: func(ctx *cli.Context) error {
					generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
					a.initialize(generateCmd)
					folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
					if err := generateCmd.Parse(ctx.Args().Slice()); err != nil {
						return err
					}
					return a.Generate(*folderFlag)
				},
			},
			{
				Name:  "edit",
				Usage: "Starts an http server and hosts a webplayer for editing the execution graph",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Usage: "interface to bind to",
						Value: "localhost",
					},
					&cli.StringFlag{
						Name:  "port",
						Usage: "port to serve over",
						Value: "8080",
					},

					// SSL
					&cli.BoolFlag{
						Name:     "ssl",
						Category: "SSL",
						Value:    false,
						Usage:    "Whether or not to use SSL",
					},
					&cli.PathFlag{
						Name:     "ssl.cert",
						Category: "SSL",
						Usage:    "path to cert file",
						Value:    "cert.pem",
					},
					&cli.PathFlag{
						Name:     "ssl.key",
						Category: "SSL",
						Usage:    "path to key file",
						Value:    "key.pem",
					},

					// Websocket
					&cli.Int64Flag{
						Name:     "ws.max-message-size",
						Category: "Websocket",
						Usage:    "Maximum message size allowed from peer over websocketed connection",
						Value:    1024 * 2,
					},
					&cli.DurationFlag{
						Name:     "ws.ping-period",
						Category: "Websocket",
						Usage:    "Send pings to peer with this period over websocketed connection. Must be less than pong-wait.",
						Value:    time.Second * 54,
					},
					&cli.DurationFlag{
						Name:     "ws.pong-wait",
						Category: "Websocket",
						Usage:    "Time allowed to read the next pong message from the peer over a websocketed connection.",
						Value:    time.Second * 60,
					},
					&cli.DurationFlag{
						Name:     "ws.write-wait",
						Category: "Websocket",
						Usage:    "Time allowed to write a message to the peer over a websocketed connection.",
						Value:    time.Second * 10,
					},
				},
				Action: func(ctx *cli.Context) error {
					editCmd := flag.NewFlagSet("edit", flag.ExitOnError)
					a.initialize(editCmd)
					if err := editCmd.Parse(ctx.Args().Slice()); err != nil {
						return err
					}

					server := AppServer{
						app:      a,
						host:     ctx.String("host"),
						port:     ctx.String("port"),
						webscene: a.WebScene,

						tls:      ctx.Bool("ssl"),
						certPath: ctx.Path("ssl.cert"),
						keyPath:  ctx.Path("ssl.key"),

						clientConfig: &room.ClientConfig{
							MaxMessageSize: ctx.Int64("ws.max-message-size"),
							PingPeriod:     ctx.Duration("ws.ping-period"),
							PongWait:       ctx.Duration("ws.pong-wait"),
							WriteWait:      ctx.Duration("ws.write-wait"),
						},
					}
					return server.Serve()
				},
			},
			{
				Name:  "outline",
				Usage: "Enumerates all generators, parameters, and producers in a heirarchial fashion formatted in JSON",
				Action: func(ctx *cli.Context) error {
					outlineCmd := flag.NewFlagSet("outline", flag.ExitOnError)
					a.initialize(outlineCmd)

					if err := outlineCmd.Parse(ctx.Args().Slice()); err != nil {
						return err
					}

					data, err := json.MarshalIndent(a.graphInstance.Schema(), "", "    ")
					if err != nil {
						return err
					}

					_, err = os.Stdout.Write(data)
					return err
				},
			},
			{
				Name:  "zip",
				Usage: "Runs all producers defined and writes it to a zip file",
				Action: func(ctx *cli.Context) error {
					zipCmd := flag.NewFlagSet("zip", flag.ExitOnError)
					a.initialize(zipCmd)
					fileFlag := zipCmd.String("file-name", "out.zip", "file to write the contents of the zip too")

					if err := zipCmd.Parse(ctx.Args().Slice()); err != nil {
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
				Name:        "mermaid",
				Description: "Create a mermaid flow chart for a specific producer",
				Action: func(ctx *cli.Context) error {
					mermaidCmd := flag.NewFlagSet("mermaid", flag.ExitOnError)
					a.initialize(mermaidCmd)
					fileFlag := mermaidCmd.String("file-name", "", "Optional path to file to write content to")

					if err := mermaidCmd.Parse(ctx.Args().Slice()); err != nil {
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
		},
	}

	return app.Run(os.Args)
}
