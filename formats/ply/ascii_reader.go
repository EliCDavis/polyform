package ply

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func isScalarPropWithType(prop Property, scalarType ScalarPropertyType) bool {
	v, ok := prop.(ScalarProperty)
	if !ok {
		return false
	}
	return v.Type == scalarType
}

func parseVector3FromContents(xIndex, yIndex, zIndex int) func(contents []string) (vector.Vector3, error) {
	return func(contents []string) (vector.Vector3, error) {
		xParsed, err := strconv.ParseFloat(contents[xIndex], 32)
		if err != nil {
			return vector.Vector3Zero(), fmt.Errorf("unable to parse x component: %w", err)
		}

		yParsed, err := strconv.ParseFloat(contents[yIndex], 32)
		if err != nil {
			return vector.Vector3Zero(), fmt.Errorf("unable to parse y component: %w", err)
		}

		zParsed, err := strconv.ParseFloat(contents[zIndex], 32)
		if err != nil {
			return vector.Vector3Zero(), fmt.Errorf("unable to parse z component: %w", err)
		}
		return vector.NewVector3(xParsed, yParsed, zParsed), nil
	}
}

type AsciiReader struct {
	elements []Element
	scanner  *bufio.Scanner
}

func (ar *AsciiReader) readVertexData(element Element) (map[string][]vector.Vector3, error) {
	attributeReaders := make(map[string]func(contents []string) (vector.Vector3, error))
	attributeData := make(map[string][]vector.Vector3)

	xIndex := -1
	yIndex := -1
	zIndex := -1

	nxIndex := -1
	nyIndex := -1
	nzIndex := -1

	redIndex := -1
	greenIndex := -1
	blueIndex := -1

	for propIndex, prop := range element.properties {
		if prop.Name() == "x" && isScalarPropWithType(prop, Float) {
			xIndex = propIndex
			continue
		}

		if prop.Name() == "y" && isScalarPropWithType(prop, Float) {
			yIndex = propIndex
			continue
		}

		if prop.Name() == "z" && isScalarPropWithType(prop, Float) {
			zIndex = propIndex
			continue
		}

		if prop.Name() == "nx" && isScalarPropWithType(prop, Float) {
			nxIndex = propIndex
			continue
		}

		if prop.Name() == "ny" && isScalarPropWithType(prop, Float) {
			nyIndex = propIndex
			continue
		}

		if prop.Name() == "nz" && isScalarPropWithType(prop, Float) {
			nzIndex = propIndex
			continue
		}

		if prop.Name() == "red" && isScalarPropWithType(prop, UChar) {
			redIndex = propIndex
			continue
		}

		if prop.Name() == "green" && isScalarPropWithType(prop, UChar) {
			greenIndex = propIndex
			continue
		}

		if prop.Name() == "blue" && isScalarPropWithType(prop, UChar) {
			blueIndex = propIndex
			continue
		}
	}

	if xIndex != -1 && yIndex != -1 && zIndex != -1 {
		attributeReaders[modeling.PositionAttribute] = parseVector3FromContents(xIndex, yIndex, zIndex)
	}

	if nxIndex != -1 && nyIndex != -1 && nzIndex != -1 {
		attributeReaders[modeling.NormalAttribute] = parseVector3FromContents(nxIndex, nyIndex, nzIndex)
	}

	if redIndex != -1 && greenIndex != -1 && blueIndex != -1 {
		attributeReaders[modeling.ColorAttribute] = func(contents []string) (vector.Vector3, error) {
			xParsed, err := strconv.ParseInt(contents[redIndex], 10, 64)
			if err != nil {
				return vector.Vector3Zero(), fmt.Errorf("unable to parse r component: %w", err)
			}

			yParsed, err := strconv.ParseInt(contents[greenIndex], 10, 64)
			if err != nil {
				return vector.Vector3Zero(), fmt.Errorf("unable to parse g component: %w", err)
			}

			zParsed, err := strconv.ParseInt(contents[blueIndex], 10, 64)
			if err != nil {
				return vector.Vector3Zero(), fmt.Errorf("unable to parse b component: %w", err)
			}

			return vector.NewVector3(
				float64(xParsed)/255.,
				float64(yParsed)/255.,
				float64(zParsed)/255.,
			), nil
		}
	}

	i := 0
	for i < element.count {
		ar.scanner.Scan()

		text := ar.scanner.Text()
		if text == "" {
			continue
		}

		contents := strings.Fields(text)

		for attribute, reader := range attributeReaders {
			v, err := reader(contents)
			if err != nil {
				return nil, err
			}
			attributeData[attribute] = append(attributeData[attribute], v)
		}

		i++
	}

	return attributeData, nil
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
	var vertexData map[string][]vector.Vector3
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

	var finalMesh modeling.Mesh

	if len(triData) > 0 {
		finalMesh = modeling.NewMesh(
			triData,
			vertexData,
			nil,
			nil,
			nil,
		)
	} else {
		finalMesh = modeling.NewPointCloud(
			vertexData,
			nil,
			nil,
			nil,
		)
	}

	return &finalMesh, nil
}
