package obj

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/mesh"
)

func parseFloatLine(components []string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	if err != nil {
		return 0, fmt.Errorf("unable to parse component %s: %w", components[0], err)
	}
	return f, nil
}

func ReadMaterials(in io.Reader) ([]mesh.Material, error) {
	if in == nil {
		panic(errors.New("cannot build obj materials from nil reader"))
	}

	scanner := bufio.NewScanner(in)

	materials := make([]mesh.Material, 0)

	var workingMaterial *mesh.Material = nil

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		components := strings.Fields(line)

		switch components[0] {
		case "#":
			continue

		case "newmtl":
			if workingMaterial != nil {
				materials = append(materials, *workingMaterial)
			}

			workingMaterial = &mesh.Material{
				Name: strings.Join(components[1:], " "),
			}

		case "map_Kd":
			if workingMaterial == nil {
				return nil, errors.New("received material parameters before newmtl declaration")
			}

			path := strings.Join(components[1:], " ")
			workingMaterial.ColorTextureURI = &path

		case "Ns":
			if workingMaterial == nil {
				return nil, errors.New("received material parameters before newmtl declaration")
			}

			f, err := parseFloatLine(components)
			if err != nil {
				return nil, err
			}

			workingMaterial.SpecularHighlight = f
		}

	}

	if workingMaterial != nil {
		materials = append(materials, *workingMaterial)
	}

	return materials, nil
}
