package ply

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type BinaryReader struct {
	elements []Element
	order    binary.ByteOrder
	reader   *bufio.Reader
}

type (
	floatParser func([]byte) float64
	byteParser  func([]byte) byte
)

func newFloatParser(order binary.ByteOrder, scalarType ScalarPropertyType, offset int) floatParser {
	if scalarType == Float {
		return func(b []byte) float64 {
			return float64(math.Float32frombits(order.Uint32(b[offset:])))
		}
	}

	if scalarType == Double {
		return func(b []byte) float64 {
			return math.Float64frombits(order.Uint64(b[offset:]))
		}
	}

	panic(fmt.Errorf("can not create float parser from scalar data type: %s", scalarType))
}

func newByteParser(offset int) byteParser {
	return func(b []byte) byte {
		return b[offset]
	}
}

func parseVector3FromByteContents(xIndex, yIndex, zIndex floatParser) func(contents []byte) (vector3.Float64, error) {
	return func(contents []byte) (vector3.Float64, error) {
		return vector3.New(xIndex(contents), yIndex(contents), zIndex(contents)), nil
	}
}

func (le *BinaryReader) readVertexData(element Element) (map[string][]vector3.Float64, error) {
	attributeReaders := make(map[string]func(contents []byte) (vector3.Float64, error))
	attributeData := make(map[string][]vector3.Float64)

	var xParser floatParser = nil
	var yParser floatParser = nil
	var zParser floatParser = nil

	var nxParser floatParser = nil
	var nyParser floatParser = nil
	var nzParser floatParser = nil

	var redParser byteParser = nil
	var greenParser byteParser = nil
	var blueParser byteParser = nil

	offset := 0
	nextOffset := 0
	for _, prop := range element.properties {
		scalarProp, ok := prop.(ScalarProperty)
		if !ok {
			return nil, fmt.Errorf("encountered non-scalar property type in vertex data: %s", prop.Name())
		}
		offset = nextOffset
		nextOffset = offset + scalarProp.Size()

		if prop.Name() == "x" && isScalarPropWithType(prop, Float, Double) {
			xParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "y" && isScalarPropWithType(prop, Float, Double) {
			yParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "z" && isScalarPropWithType(prop, Float, Double) {
			zParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "nx" && isScalarPropWithType(prop, Float, Double) {
			nxParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "ny" && isScalarPropWithType(prop, Float, Double) {
			nyParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "nz" && isScalarPropWithType(prop, Float, Double) {
			nzParser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "red" && isScalarPropWithType(prop, UChar) {
			redParser = newByteParser(offset)
			continue
		}

		if prop.Name() == "green" && isScalarPropWithType(prop, UChar) {
			greenParser = newByteParser(offset)
			continue
		}

		if prop.Name() == "blue" && isScalarPropWithType(prop, UChar) {
			blueParser = newByteParser(offset)
			continue
		}
	}

	if xParser != nil && yParser != nil && zParser != nil {
		attributeReaders[modeling.PositionAttribute] = parseVector3FromByteContents(xParser, yParser, zParser)
	}

	if nxParser != nil && nyParser != nil && nzParser != nil {
		attributeReaders[modeling.NormalAttribute] = parseVector3FromByteContents(nxParser, nyParser, nzParser)
	}

	if redParser != nil && greenParser != nil && blueParser != nil {
		attributeReaders[modeling.ColorAttribute] = func(contents []byte) (vector3.Float64, error) {
			return vector3.New(
				float64(redParser(contents))/255.,
				float64(greenParser(contents))/255.,
				float64(blueParser(contents))/255.,
			), nil
		}
	}

	i := 0
	buf := make([]byte, nextOffset)
	for i < element.count {
		_, err := io.ReadFull(le.reader, buf)
		if err != nil {
			return nil, err
		}

		for attribute, reader := range attributeReaders {
			v, err := reader(buf)
			if err != nil {
				return nil, err
			}
			attributeData[attribute] = append(attributeData[attribute], v)
		}
		i++
	}

	return attributeData, nil
}

func (le *BinaryReader) readFaceData(element Element) ([]int, error) {
	if len(element.properties) != 1 {
		return nil, fmt.Errorf("unimplemented case where face data contains %d properties", len(element.properties))
	}

	name := element.properties[0].Name() // vertex_indices
	if name != "vertex_index" && name != "vertex_indices" {
		return nil, fmt.Errorf("unexpected face data property: %s", name)
	}

	listProp, ok := element.properties[0].(ListProperty)
	if !ok {
		return nil, fmt.Errorf("encountered non-list property type for face data: %s", element.properties[0].Name())
	}

	var listCountReader func(*bufio.Reader) (int, error)
	switch listProp.countType {
	case UChar:
		listCountReader = func(r *bufio.Reader) (int, error) {
			b, err := r.ReadByte()
			return int(b), err
		}

	default:
		return nil, fmt.Errorf("unimplemented list count scalar-type: %s", listProp.countType)
	}

	var listDataReader func(*bufio.Reader) (int32, error)
	switch listProp.listType {
	case Int:
		listDataReader = func(r *bufio.Reader) (int32, error) {
			buf := make([]byte, 4)
			_, err := io.ReadFull(r, buf)
			return int32(le.order.Uint32(buf)), err
		}

	default:
		return nil, fmt.Errorf("unimplemented list element scalar-type: %s", listProp.listType)
	}

	triData := make([]int, 0, element.count*3)

	for i := 0; i < element.count; i++ {
		listSize, err := listCountReader(le.reader)
		if err != nil {
			return nil, fmt.Errorf("unable to parse list size: %w", err)
		}

		if listSize < 3 || listSize > 4 {
			return nil, fmt.Errorf("unimplemented tesselation scenario where face vertex data is of size: %d", listSize)
		}

		v1, err := listDataReader(le.reader)
		if err != nil {
			return nil, fmt.Errorf("unable to parse index: %w", err)
		}

		v2, err := listDataReader(le.reader)
		if err != nil {
			return nil, fmt.Errorf("unable to parse index: %w", err)
		}

		v3, err := listDataReader(le.reader)
		if err != nil {
			return nil, fmt.Errorf("unable to parse index: %w", err)
		}
		triData = append(triData, int(v1), int(v2), int(v3))
		if listSize == 4 {
			v4, err := listDataReader(le.reader)
			if err != nil {
				return nil, fmt.Errorf("unable to parse index: %w", err)
			}
			triData = append(triData, int(v1), int(v3), int(v4))
		}
	}

	return triData, nil
}

func (le *BinaryReader) ReadMesh() (*modeling.Mesh, error) {
	var vertexData map[string][]vector3.Float64
	var triData []int
	for _, element := range le.elements {
		if element.name == "vertex" {
			data, err := le.readVertexData(element)
			if err != nil {
				return nil, err
			}
			vertexData = data
		}

		if element.name == "face" {
			data, err := le.readFaceData(element)
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
