package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

func writeHeaderAsPlaintext(header ply.Header, out io.Writer) error {
	fmt.Fprintf(out, "Format: %s\n", header.Format.String())

	textures := header.TextureFiles()
	fmt.Fprintf(out, "Texture Files: %d\n", len(textures))
	for _, tex := range textures {
		fmt.Fprintf(out, "%20s\n", tex)
	}

	for _, ele := range header.Elements {
		fmt.Fprintf(out, "%s %d entries\n", ele.Name+":", ele.Count)
		for _, prop := range ele.Properties {
			if scalar, ok := prop.(ply.ScalarProperty); ok {
				fmt.Fprintf(out, "\t%-14s (%s)\n", prop.Name(), scalar.Type)
			} else if arr, ok := prop.(ply.ListProperty); ok {
				fmt.Fprintf(out, "\t%-14s (count type: %s, list type: %s)\n", prop.Name(), arr.CountType, arr.ListType)
			}
		}
	}
	return nil
}

func writeHeaderAsJSON(header ply.Header, out io.Writer) error {
	data, err := json.MarshalIndent(header, "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}

var HeaderCommand = &cli.Command{
	Name: "header",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "json",
			Usage: "whether or not to print the header information out in JSON format",
			Value: false,
		},
		&cli.StringFlag{
			Name:    "out",
			Usage:   "if defined, the path to the file to write our output to",
			Aliases: []string{"o"},
		},
	},
	Action: func(ctx *cli.Context) error {
		f, err := openPlyFile()
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := ply.ReadHeader(f)
		if err != nil {
			return err
		}

		var out io.Writer = ctx.App.Writer
		outPath := ctx.String("out")
		if strings.TrimSpace(outPath) != "" {
			out, err = os.Create(outPath)
			if err != nil {
				return err
			}
		}

		if ctx.Bool("json") {
			return writeHeaderAsJSON(header, out)
		}

		return writeHeaderAsPlaintext(header, out)
	},
}
