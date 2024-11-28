package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/urfave/cli/v2"
)

func writeHeaderAsJSON(header *spz.Header, out io.Writer) error {
	data, err := json.MarshalIndent(header, "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}

func writeHeaderAsPlaintext(header *spz.Header, out io.Writer) error {
	fmt.Fprintf(out, "Version:          %d\n", header.Version)
	fmt.Fprintf(out, "Points:           %d\n", header.NumPoints)
	fmt.Fprintf(out, "SH Degree:        %d\n", header.ShDegree)
	fmt.Fprintf(out, "Fractional Bits:  %08b\n", header.FractionalBits)
	_, err := fmt.Fprintf(out, "Flags:            %08b\n", header.Flags)
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
		f, err := openSPZFile()
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := spz.ReadHeader(f)
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
