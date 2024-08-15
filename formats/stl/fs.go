package stl

import (
	"bufio"
	"os"

	"github.com/EliCDavis/polyform/modeling"
)

func Save(fp string, m modeling.Mesh) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	if err := WriteMesh(writer, m); err != nil {
		return err
	}

	return writer.Flush()
}

func Load(fp string) (*modeling.Mesh, error) {
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadMesh(bufio.NewReader(f))
}
