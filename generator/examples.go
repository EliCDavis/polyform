package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
)

//go:embed examples/*
var examplesFs embed.FS

func allExamples() []string {
	entries := make([]string, 0)
	fs.WalkDir(examplesFs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		entries = append(entries, filepath.Base(path))
		return nil
	})
	return entries
}

func loadExample(example string) []byte {
	var contents []byte
	found := false
	fs.WalkDir(examplesFs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Base(path) != example {
			return nil
		}

		contents, err = examplesFs.ReadFile(path)
		found = true

		return err
	})

	if !found {
		panic(fmt.Errorf("example: %q doesn't exist", example))
	}

	if len(contents) == 0 {
		panic("example loaded is empty")
	}

	return contents
}
