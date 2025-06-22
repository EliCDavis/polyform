package properties

import (
	"fmt"
	"os"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

func fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func dataSizeFormat(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d b", size)
	}
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%d kb", size)
	}
	return fmt.Sprintf("%d mb", size/1024)
}

func getOutpath(ctx *cli.Context) (string, error) {
	outPath := ctx.String("out")
	if fileExists(outPath) {
		if ctx.Bool("force") {
			if err := os.Remove(outPath); err != nil {
				return outPath, err
			}
		} else {
			return outPath, fmt.Errorf("file %s already exists, use the --force flag to overwrite", outPath)
		}
	}
	return outPath, nil
}

func calculateTotalPropertySize(properties []ply.Property) int {
	total := 0
	for _, p := range properties {
		scalar, ok := p.(ply.ScalarProperty)
		if !ok {
			panic(fmt.Errorf("can not calculate total size: property %s has no fixed size", p.Name()))
		}
		total += scalar.Size()
	}
	return total
}

func InSlice[T comparable](a T, list []T) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
