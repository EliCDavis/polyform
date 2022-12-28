package ply

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

type AsciiReader struct {
	elements []Element
	scanner  *bufio.Scanner
}

func (ar *AsciiReader) readVertexData(element Element) ([]vector.Vector3, error) {
	xIndex := -1
	yIndex := -1
	zIndex := -1

	for propIndex, prop := range element.properties {
		if prop.Name() == "x" {
			xIndex = propIndex
		}

		if prop.Name() == "y" {
			yIndex = propIndex
		}

		if prop.Name() == "z" {
			zIndex = propIndex
		}
	}

	vertices := make([]vector.Vector3, element.count)
	i := 0
	for i < element.count {
		ar.scanner.Scan()

		text := ar.scanner.Text()
		if text == "" {
			continue
		}

		contents := strings.Fields(text)

		xParsed, err := strconv.ParseFloat(contents[xIndex], 32)
		if err != nil {
			return nil, fmt.Errorf("unable to parse x component: %w", err)
		}

		yParsed, err := strconv.ParseFloat(contents[yIndex], 32)
		if err != nil {
			return nil, fmt.Errorf("unable to parse y component: %w", err)
		}

		zParsed, err := strconv.ParseFloat(contents[zIndex], 32)
		if err != nil {
			return nil, fmt.Errorf("unable to parse z component: %w", err)
		}

		vertices[i] = vector.NewVector3(xParsed, yParsed, zParsed)
		i++
	}

	return vertices, nil
}

func (ar *AsciiReader) readFaceData(element Element) ([]int, error) {
	if len(element.properties) != 1 {
		return nil, fmt.Errorf("unimplemented case where face data contains %d properties", len(element.properties))
	}

	name := element.properties[0].Name() // vertex_indices
	if name != "vertex_index" && name != "vertex_indices" {
		return nil, fmt.Errorf("unexpected face data property: %s", name)
	}

	triData := make([]int, 0)

	for i := 0; i < element.count; i++ {
		ar.scanner.Scan()
		contents := strings.Fields(ar.scanner.Text())
		listSize, err := strconv.Atoi(contents[0])
		if err != nil {
			return nil, fmt.Errorf("unable to parse list size: %w", err)
		}

		if listSize < 3 || listSize > 4 {
			return nil, fmt.Errorf("unimplemented tesselation scenario where face vertex data is of size: %d", listSize)
		}
		v1, err := strconv.Atoi(contents[1])
		if err != nil {
			return nil, fmt.Errorf("unable to index: %w", err)
		}

		v2, err := strconv.Atoi(contents[2])
		if err != nil {
			return nil, fmt.Errorf("unable to vert index: %w", err)
		}

		v3, err := strconv.Atoi(contents[3])
		if err != nil {
			return nil, fmt.Errorf("unable to index: %w", err)
		}
		triData = append(triData, v1, v2, v3)
		if listSize == 4 {
			v4, err := strconv.Atoi(contents[4])
			if err != nil {
				return nil, fmt.Errorf("unable to index: %w", err)
			}
			triData = append(triData, v1, v3, v4)
		}

	}

	return triData, nil
}

func (ar *AsciiReader) ReadMesh() (*modeling.Mesh, error) {
	var vertexData []vector.Vector3
	var triData []int
	for _, element := range ar.elements {
		if element.name == "vertex" {
			data, err := ar.readVertexData(element)
			if err != nil {
				return nil, err
			}
			vertexData = data
		}

		if element.name == "face" {
			data, err := ar.readFaceData(element)
			if err != nil {
				return nil, err
			}
			triData = data
		}
	}

	finalMesh := modeling.NewMesh(
		triData,
		map[string][]vector.Vector3{
			modeling.PositionAttribute: vertexData,
		},
		nil,
		nil,
		nil,
	)

	return &finalMesh, nil
}
