package main

import (
	// "bufio"

	// "compress/gzip"
	"embed"
	_ "embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

//go:embed index.html
var htmlContents []byte

//go:embed sw.js
var swJsContents []byte

//go:embed icons/*
var iconFs embed.FS

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
		},
		Action: func(ctx *cli.Context) error {
			outFolder := ctx.String("out")
			err := os.MkdirAll(outFolder, os.ModeDir)
			if err != nil {
				return err
			}

			err = Copy(ctx.String("wasm"), filepath.Join(outFolder, "main.wasm"))
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(outFolder, "index.html"), htmlContents, 0666)
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(outFolder, "sw.js"), swJsContents, 0666)
			if err != nil {
				return err
			}

			return fs.WalkDir(iconFs, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				content, err := iconFs.ReadFile(path)
				if err != nil {
					return err
				}

				return os.WriteFile(filepath.Join(outFolder, filepath.Base(path)), content, 0666)
			})
		},
	}
}
