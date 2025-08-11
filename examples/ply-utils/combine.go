package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

// FindFilesByRegex searches for files matching the given regex pattern
// The regex is matched against the full relative path from the current directory
func FindFilesByRegex(pattern string) ([]string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	var matches []string

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Test the full path against the regex
		if regex.MatchString(path) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	return matches, nil
}

func addToHeader(fp string, logging io.Writer, currentHeader *ply.Header) (bool, error) {
	f, err := os.Open(fp)
	if err != nil {
		return false, fmt.Errorf("unable to open %s: %w", fp, err)
	}
	defer f.Close()

	fh, err := ply.ReadHeader(f)
	if err != nil {
		return false, fmt.Errorf("unable to interpret %q ply header: %w", fp, err)
	}

	if len(fh.Elements) != 1 {
		fmt.Fprintf(logging, "skippping %q, has %d elements instead of 1\n", fp, len(fh.Elements))
		return false, nil
	}

	ele := fh.Elements[0]
	if !ele.DeterministicPointSize() {
		fmt.Fprintf(logging, "skippping %q, does not have deterministic point size\n", fp)
		return false, nil
	}

	if len(currentHeader.Elements) == 0 {
		*currentHeader = fh
		return true, nil
	}

	accEle := currentHeader.Elements[0]
	if len(accEle.Properties) != len(ele.Properties) {
		fmt.Fprintf(logging, "Skipping %s, mismatch prop count (%d != %d)\n", fp, len(accEle.Properties), len(ele.Properties))
		return false, nil
	}

	// Verify all properties match
	for i, p := range ele.Properties {
		if accEle.Properties[i].Name() != p.Name() {
			fmt.Fprintf(logging, "Skipping %s, mismatch prop[%d] (%s != %s)\n", fp, i, accEle.Properties[i].Name(), p.Name())
			return false, nil
		}

		accProp := accEle.Properties[i].(ply.ScalarProperty)
		prop := p.(ply.ScalarProperty)

		if accProp.Type != prop.Type {
			fmt.Fprintf(logging, "Skipping %s, mismatch prop[%d] (%s != %s)\n", fp, i, accProp.Type, prop.Type)
			return false, nil
		}
	}

	accEle.Count += ele.Count
	currentHeader.Elements[0] = accEle
	return true, nil
}

func determineHeader(fp []string, logging io.Writer) (*ply.Header, []string, error) {

	plysToCombine := make([]string, 0)
	var header *ply.Header = new(ply.Header)
	for _, p := range fp {
		kept, err := addToHeader(p, logging, header)
		if err != nil {
			return nil, nil, err
		}

		if kept {
			plysToCombine = append(plysToCombine, p)
		}
	}

	return header, plysToCombine, nil
}

var CombineCommand = &cli.Command{
	Name:  "combine",
	Usage: "Combines PLYs into single file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "in",
			Usage: "regex matching files we're intrested in combining",
			Value: "*.ply",
		},
		&cli.StringFlag{
			Name:    "out",
			Usage:   "The path to the file to write our output to",
			Aliases: []string{"o"},
			Value:   "combined.ply",
		},
	},
	Action: func(ctx *cli.Context) error {
		start := time.Now()
		filesToCombine, err := FindFilesByRegex(ctx.String("in"))
		if err != nil {
			return fmt.Errorf("unable to determine PLYs to process: %w", err)
		}

		if len(filesToCombine) == 0 {
			return fmt.Errorf("found no files matching %q to combine", ctx.String("in"))
		}

		if len(filesToCombine) == 1 {
			return fmt.Errorf("found single file %q to combine, skipping", filesToCombine)
		}

		header, vettedPlys, err := determineHeader(filesToCombine, ctx.App.Writer)
		if err != nil {
			return fmt.Errorf("unable to determine valid plys to combine: %w", err)
		}

		switch len(vettedPlys) {
		case 0:
			return fmt.Errorf("found no PLYs that can be combined (this tool only supports pointclouds atm)")

		case 1:
			return fmt.Errorf("found only one PLY to combine %q, skipping", vettedPlys[0])
		}

		out, err := os.Create(ctx.String("out"))
		if err != nil {
			return err
		}
		defer out.Close()

		writer := bufio.NewWriter(out)
		defer writer.Flush()

		err = header.Write(writer)
		if err != nil {
			return fmt.Errorf("unable to write aggregate PLY header: %w", err)
		}

		for _, p := range vettedPlys {
			plyStart := time.Now()
			f, err := os.Open(p)
			if err != nil {
				return fmt.Errorf("unable to open %q: %w", p, err)
			}
			defer f.Close()

			reader := bufio.NewReader(f)

			header, err := ply.ReadHeader(reader)
			if err != nil {
				return fmt.Errorf("unable to interpret header for %q: %w", p, err)
			}

			err = header.Elements[0].Scan(reader, func(buf []byte) error {
				_, err = writer.Write(buf)
				return err
			})

			if err != nil {
				return fmt.Errorf("failed scanning for %q: %w", p, err)
			}
			fmt.Fprintf(ctx.App.Writer, "Coppied %q (%d points) in %s\n", p, header.Elements[0].Count, time.Since(plyStart))
		}

		fmt.Fprintf(ctx.App.Writer, "Completed copying %d points in %s\n", header.Elements[0].Count, time.Since(start))
		return nil
	},
}
