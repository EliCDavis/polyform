package ply

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func isScalarPropWithType(prop Property, scalarType ...ScalarPropertyType) bool {
	v, ok := prop.(ScalarProperty)
	if !ok {
		return false
	}
	for _, t := range scalarType {
		if v.Type == t {
			return true
		}
	}
	return false
}

func parseVector3FromStringContents(xIndex, yIndex, zIndex int) func(contents []string) (vector3.Float64, error) {
	return func(contents []string) (vector3.Float64, error) {
		xParsed, err := strconv.ParseFloat(contents[xIndex], 32)
		if err != nil {
			return vector3.Zero[float64](), fmt.Errorf("unable to parse x component: %w", err)
		}

		yParsed, err := strconv.ParseFloat(contents[yIndex], 32)
		if err != nil {
			return vector3.Zero[float64](), fmt.Errorf("unable to parse y component: %w", err)
		}

		zParsed, err := strconv.ParseFloat(contents[zIndex], 32)
		if err != nil {
			return vector3.Zero[float64](), fmt.Errorf("unable to parse z component: %w", err)
		}
		return vector3.New(xParsed, yParsed, zParsed), nil
	}
}

type AsciiReader struct {
	elements []Element
	scanner  *bufio.Scanner
}

func (ar *AsciiReader) readVertexData(element Element, approvedData map[string]bool) (map[string][]vector3.Float64, error) {
	attributeReaders := make(map[string]func(contents []string) (vector3.Float64, error))
	attributeData := make(map[string][]vector3.Float64)

	xIndex := -1
	yIndex := -1
	zIndex := -1

	nxIndex := -1
	nyIndex := -1
	nzIndex := -1

	redIndex := -1
	greenIndex := -1
	blueIndex := -1

	for propIndex, prop := range element.Properties {
		if approvedData != nil {
			if !approvedData[prop.Name()] {
				continue
			}
		}

		if prop.Name() == "x" && isScalarPropWithType(prop, Float, Double) {
			xIndex = propIndex
			continue
		}

		if prop.Name() == "y" && isScalarPropWithType(prop, Float, Double) {
			yIndex = propIndex
			continue
		}

		if prop.Name() == "z" && isScalarPropWithType(prop, Float, Double) {
			zIndex = propIndex
			continue
		}

		if prop.Name() == "nx" && isScalarPropWithType(prop, Float, Double) {
			nxIndex = propIndex
			continue
		}

		if prop.Name() == "ny" && isScalarPropWithType(prop, Float, Double) {
			nyIndex = propIndex
			continue
		}

		if prop.Name() == "nz" && isScalarPropWithType(prop, Float, Double) {
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
		attributeReaders[modeling.PositionAttribute] = parseVector3FromStringContents(xIndex, yIndex, zIndex)
	}

	if nxIndex != -1 && nyIndex != -1 && nzIndex != -1 {
		attributeReaders[modeling.NormalAttribute] = parseVector3FromStringContents(nxIndex, nyIndex, nzIndex)
	}

	if redIndex != -1 && greenIndex != -1 && blueIndex != -1 {
		attributeReaders[modeling.ColorAttribute] = func(contents []string) (vector3.Float64, error) {
			xParsed, err := strconv.ParseInt(contents[redIndex], 10, 64)
			if err != nil {
				return vector3.Zero[float64](), fmt.Errorf("unable to parse r component: %w", err)
			}

			yParsed, err := strconv.ParseInt(contents[greenIndex], 10, 64)
			if err != nil {
				return vector3.Zero[float64](), fmt.Errorf("unable to parse g component: %w", err)
			}

			zParsed, err := strconv.ParseInt(contents[blueIndex], 10, 64)
			if err != nil {
				return vector3.Zero[float64](), fmt.Errorf("unable to parse b component: %w", err)
			}

			return vector3.New(
				float64(xParsed)/255.,
				float64(yParsed)/255.,
				float64(zParsed)/255.,
			), nil
		}
	}

	i := 0
	for i < element.Count {
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

type asciiListReader func(data []string, start int) (int, error)

func (ar *AsciiReader) readFaceData(element Element) ([]int, []vector2.Float64, error) {
	listReaders := make([]asciiListReader, len(element.Properties))

	triData := make([]int, 0)
	uvCords := make([]vector2.Float64, 0)

	for i, prop := range element.Properties {
		if prop.Name() == "vertex_index" || prop.Name() == "vertex_indices" {
			listReaders[i] = func(data []string, start int) (int, error) {
				listSize, err := strconv.Atoi(data[start])
				if err != nil {
					return -1, fmt.Errorf("unable to parse list size: %w", err)
				}

				if listSize < 3 || listSize > 4 {
					return -1, fmt.Errorf("unimplemented tesselation scenario where face vertex data is of size: %d", listSize)
				}
				v1, err := strconv.Atoi(data[start+1])
				if err != nil {
					return -1, fmt.Errorf("unable to parse index: %w", err)
				}

				v2, err := strconv.Atoi(data[start+2])
				if err != nil {
					return -1, fmt.Errorf("unable to parse index: %w", err)
				}

				v3, err := strconv.Atoi(data[start+3])
				if err != nil {
					return -1, fmt.Errorf("unable to parse index: %w", err)
				}
				triData = append(triData, v1, v2, v3)
				if listSize == 4 {
					v4, err := strconv.Atoi(data[start+4])
					if err != nil {
						return -1, fmt.Errorf("unable to parse index: %w", err)
					}
					triData = append(triData, v1, v3, v4)
				}
				return listSize + 1, nil
			}
			continue
		}

		if prop.Name() == "texcoord" {
			listReaders[i] = func(data []string, start int) (int, error) {
				listSize, err := strconv.Atoi(data[start])
				if err != nil {
					return -1, fmt.Errorf("unable to parse list size: %w", err)
				}

				if listSize < 6 || listSize > 8 {
					return -1, fmt.Errorf("unimplemented tesselation scenario where face texture data is of size: %d", listSize)
				}

				v1X, err := strconv.ParseFloat(data[start+1], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				v1Y, err := strconv.ParseFloat(data[start+2], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v1X, v1Y))

				v2X, err := strconv.ParseFloat(data[start+3], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				v2Y, err := strconv.ParseFloat(data[start+4], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v2X, v2Y))

				v3X, err := strconv.ParseFloat(data[start+5], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				v3Y, err := strconv.ParseFloat(data[start+6], 64)
				if err != nil {
					return -1, fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v3X, v3Y))

				if listSize == 8 {
					v4X, err := strconv.ParseFloat(data[start+7], 64)
					if err != nil {
						return -1, fmt.Errorf("unable to parse texcord: %w", err)
					}
					v4Y, err := strconv.ParseFloat(data[start+8], 64)
					if err != nil {
						return -1, fmt.Errorf("unable to parse texcord: %w", err)
					}
					uvCords = append(uvCords, vector2.New(v1X, v1Y))
					uvCords = append(uvCords, vector2.New(v3X, v3Y))
					uvCords = append(uvCords, vector2.New(v4X, v4Y))
				}
				return listSize + 1, nil
			}
			continue
		}

		// Just eat the input if we don't know what it is.
		listReaders[i] = func(data []string, start int) (int, error) {
			listSize, err := strconv.Atoi(data[start])
			if err != nil {
				return -1, fmt.Errorf("unable to parse list size: %w", err)
			}
			return listSize + 1, nil
		}
	}

	for i := 0; i < element.Count; i++ {
		ar.scanner.Scan()
		contents := strings.Fields(ar.scanner.Text())
		offset := 0
		for _, reader := range listReaders {
			shift, err := reader(contents, offset)
			if err != nil {
				return nil, nil, err
			}
			offset += shift
		}
	}

	return triData, uvCords, nil
}

func (ar *AsciiReader) ReadMesh(vertexAttributes map[string]bool) (*modeling.Mesh, error) {
	var vertexData map[string][]vector3.Float64
	var triData []int
	var uvData []vector2.Float64
	for _, element := range ar.elements {
		if element.Name == "vertex" {
			data, err := ar.readVertexData(element, vertexAttributes)
			if err != nil {
				return nil, err
			}
			vertexData = data
		}

		if element.Name == "face" {
			data, uvs, err := ar.readFaceData(element)
			if err != nil {
				return nil, err
			}
			uvData = uvs
			triData = data
		}
	}

	var finalMesh modeling.Mesh

	if len(triData) > 0 {
		finalMesh = modeling.NewTriangleMesh(triData).SetFloat3Data(vertexData)
		if len(uvData) == len(triData) {
			finalMesh = finalMesh.
				Transform(meshops.UnweldTransformer{}).
				SetFloat2Attribute(modeling.TexCoordAttribute, uvData)
		}
	} else {
		finalMesh = modeling.NewPointCloud(
			nil,
			vertexData,
			nil,
			nil,
			nil,
		)
	}

	return &finalMesh, nil
}
