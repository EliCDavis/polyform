package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

//go:embed index.html
var htmlContents []byte

//go:embed sw.js
var swJsContents []byte

func Copy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer r.Close() // ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		// Report the error from Close, if any,
		// but do so only if there isn't already
		// an outgoing error.
		if c := w.Close(); c != nil && err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func TinygoRoot() (string, error) {
	cmd := exec.Command(
		"tinygo",
		"env",
		"TINYGOROOT",
	)
	var outb bytes.Buffer
	cmd.Stdout = &outb
	err := cmd.Run()
	return strings.TrimSpace(outb.String()), err
}

func RegularGo(out, programToBuild string, stripDebug bool) (string, string, error) {
	args := []string{
		"build",
	}

	if stripDebug {
		// args = append(args, "-ldflags", "-s -w")
	}

	args = append(
		args,
		"-o", out,
		programToBuild,
	)

	log.Print(args)
	cmd := exec.Command("go", args...)
	cmd.Env = []string{
		"GOOS=js",
		"GOARCH=wasm",
	}
	var outb bytes.Buffer
	var outErr bytes.Buffer

	cmd.Stdout = &outb
	cmd.Stderr = &outErr
	err := cmd.Run()

	return strings.TrimSpace(outb.String()), strings.TrimSpace(outErr.String()), err
}

func buildCommand() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Compiles a polyform app using tinygo into a wasm binary and wraps it in a website",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "wasm",
				Usage:    "Path to wasm to bundle",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "folder to write all contents to",
				Value:   "html",
			},
			&cli.BoolFlag{
				Name:  "no-debug",
				Usage: "whether or not to strip debug symbols (reduces file size)",
				Value: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			outFolder := ctx.String("out")
			err := os.MkdirAll(outFolder, os.ModeDir)
			if err != nil {
				return err
			}

			wasmFile, err := os.Open(ctx.String("wasm"))
			if err != nil {
				return err
			}
			defer wasmFile.Close()

			gzippedWasmFile, err := os.Create(filepath.Join(outFolder, "main.wasm.gz"))
			if err != nil {
				return err
			}
			defer gzippedWasmFile.Close()

			gzipWriter := gzip.NewWriter(gzippedWasmFile)
			chunkSize := 1024 // Adjust as needed
			reader := bufio.NewReader(wasmFile)

			for {
				buf := make([]byte, chunkSize)
				n, err := reader.Read(buf)
				if err != nil {
					if err.Error() == "EOF" {
						break // Reached end of file
					}
					return err
				}
				_, err = gzipWriter.Write(buf[:n])
				if err != nil {
					return err
				}
			}

			gzipWriter.Close()

			err = os.WriteFile(filepath.Join(outFolder, "index.html"), htmlContents, 0666)
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(outFolder, "sw.js"), swJsContents, 0666)
			if err != nil {
				return err
			}

			return err
		},
	}
}
