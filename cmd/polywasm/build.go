package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

//go:embed index.html
var htmlContents []byte

//go:embed wasm.js
var wasmJsContents []byte

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

func buildCommand() *cli.Command {
	return &cli.Command{
		Name:        "build",
		Description: "Compiles a polyform app using tinygo into a wasm binary and wraps it in a website",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "app-path",
				Usage:    "Path to polyform application to compile",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "out",
				Usage:    "folder to write all contents to",
				Required: false,
				Value:    "html",
			},
			&cli.BoolFlag{
				Name:  "no-debug",
				Usage: "whether or not to strip debug symbols",
				Value: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			outFolder := ctx.String("out")
			err := os.MkdirAll(outFolder, os.ModeDir)
			if err != nil {
				return err
			}

			args := []string{
				"build",
				"-o", filepath.Join(outFolder, "wasm.wasm"),
				"-target", "wasm",
			}

			if ctx.Bool("no-debug") {
				args = append(args, "-no-debug")
			}

			args = append(args, ctx.String("app-path"))

			cmd := exec.Command("tinygo", args...)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()

			if len(outb.Bytes()) > 0 {
				fmt.Fprintf(ctx.App.Writer, "Tinygo Build Output:\n%s", outb.String())
			}

			if len(errb.Bytes()) > 0 {
				fmt.Fprintf(ctx.App.ErrWriter, "Tinygo Build Error:\n%s", errb.String())
			}

			if err != nil {
				fmt.Fprintln(ctx.App.Writer, args)
				return err
			}

			tinygoRoot, err := TinygoRoot()
			if err != nil {
				return err
			}

			err = Copy(filepath.Join(tinygoRoot, "targets", "wasm_exec.js"), filepath.Join(outFolder, "wasm_exec.js"))
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(outFolder, "index.html"), htmlContents, 0666)
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(outFolder, "wasm.js"), wasmJsContents, 0666)
			if err != nil {
				return err
			}

			return err
		},
	}
}
