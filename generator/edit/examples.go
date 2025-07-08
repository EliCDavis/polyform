package edit

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

func loadExample(example string) ([]byte, error) {
	var contents []byte
	found := false
	err := fs.WalkDir(examplesFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

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

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("example: %q doesn't exist", example)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("example: %q loaded empty", example)
	}

	return contents, nil
}
