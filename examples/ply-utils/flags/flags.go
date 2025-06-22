package flags

import (
	"os"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/urfave/cli/v2"
)

var PlyFile = &cli.StringFlag{
	Name:        "in",
	Required:    true,
	Aliases:     []string{"i", "f", "file"},
	Destination: &inFilePath,
}

var inFilePath string

func OpenPlyFile() (*os.File, error) {
	return os.Open(inFilePath)
}

func GetPlyFile() (*modeling.Mesh, error) {
	return ply.Load(inFilePath)
}
