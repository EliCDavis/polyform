package obj

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"strings"
)

func parseFloatLine(components []string) (f float64, err error) {
	if f, err = strconv.ParseFloat(strings.TrimSpace(components[1]), 32); err != nil {
		return 0, fmt.Errorf("unable to parse component[0] %q: %w", components[0], err)
	}
	return f, nil
}

func parseColorLine(components []string) (color.Color, error) {
	r, err := strconv.ParseFloat(strings.TrimSpace(components[1]), 32)
	g, err := strconv.ParseFloat(strings.TrimSpace(components[2]), 32)
	b, err := strconv.ParseFloat(strings.TrimSpace(components[3]), 32)
	if err != nil {
		return nil, fmt.Errorf("unable to parse component %q: %w", components[0], err)
	}
	return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}, nil
}

func ReadMaterials(in io.Reader) ([]Material, error) {
	if in == nil {
		panic(errors.New("cannot build obj materials from nil reader"))
	}

	scanner := bufio.NewScanner(in)

	materials := make([]Material, 0)

	var workingMaterial *Material = nil

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

			workingMaterial = &Material{
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
				return nil, fmt.Errorf("failed to parse float line: %w", err)
			}

			workingMaterial.SpecularHighlight = f

		case "Kd":
			if workingMaterial == nil {
				return nil, errors.New("received material parameters before newmtl declaration")
			}

			f, err := parseColorLine(components)
			if err != nil {
				return nil, fmt.Errorf("failed to parse color line: %w", err)
			}

			workingMaterial.DiffuseColor = f
		}

	}

	if workingMaterial != nil {
		materials = append(materials, *workingMaterial)
	}

	return materials, nil
}
