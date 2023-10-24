package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
)

func serverCommand() *cli.Command {
	return &cli.Command{
		Name: "serve",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
			},
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
			},
			&cli.StringFlag{
				Name:  "dir",
				Value: "./html",
			},
		},
		Action: func(ctx *cli.Context) error {
			dir := ctx.String("dir")

			host := ctx.String("host")
			port := ctx.String("port")
			url := fmt.Sprintf("%s:%s", host, port)

			fs := http.FileServer(http.Dir(dir))
			log.Print("Serving " + dir + " on http://" + url)
			return http.ListenAndServe(url, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				resp.Header().Add("Cache-Control", "no-cache")
				if strings.HasSuffix(req.URL.Path, ".wasm") {
					resp.Header().Set("content-type", "application/wasm")
				}
				fs.ServeHTTP(resp, req)
			}))
		},
	}

}
