package ply

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
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

func binReaderByteReader(r *bufio.Reader) (int, error) {
	b, err := r.ReadByte()
	return int(b), err
}

func binReaderIntReader(order binary.ByteOrder) func(r *bufio.Reader) (int32, error) {
	return func(r *bufio.Reader) (int32, error) {
		buf := make([]byte, 4)
		_, err := io.ReadFull(r, buf)
		return int32(order.Uint32(buf)), err
	}
}

func binReaderFloat32Reader(order binary.ByteOrder) func(r *bufio.Reader) (float32, error) {
	return func(r *bufio.Reader) (float32, error) {
		buf := make([]byte, 4)
		_, err := io.ReadFull(r, buf)
		return math.Float32frombits(order.Uint32(buf)), err
	}
}

func (le *BinaryReader) listCountReader(listProp ListProperty) (func(r *bufio.Reader) (int, error), error) {
	switch listProp.countType {
	case UChar:
		return binReaderByteReader, nil

	default:
		return nil, fmt.Errorf("unimplemented list count scalar-type: %s", listProp.countType)
	}
}

type binReaderListReader func(r *bufio.Reader) error

func (le *BinaryReader) readFaceData(element Element) ([]int, []vector2.Float64, error) {
	readers := make([]binReaderListReader, len(element.properties))

	triData := make([]int, 0, element.count*3)
	uvCords := make([]vector2.Float64, 0)

	for i, prop := range element.properties {
		listProp, ok := prop.(ListProperty)
		if !ok {
			return nil, nil, fmt.Errorf("encountered non-list property type for face data: %s", prop.Name())
		}

		listCountReader, err := le.listCountReader(listProp)
		if err != nil {
			return nil, nil, err
		}

		if prop.Name() == "vertex_index" || prop.Name() == "vertex_indices" {
			var listDataReader func(*bufio.Reader) (int32, error)
			switch listProp.listType {
			case Int, UInt:
				listDataReader = binReaderIntReader(le.order)

			default:
				return nil, nil, fmt.Errorf("unimplemented list element scalar-type: %s", listProp.listType)
			}

			readers[i] = func(r *bufio.Reader) error {
				listSize, err := listCountReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse list size: %w", err)
				}

				if listSize < 3 || listSize > 4 {
					return fmt.Errorf("unimplemented tesselation scenario where face vertex data is of size: %d", listSize)
				}

				v1, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse index: %w", err)
				}

				v2, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse index: %w", err)
				}

				v3, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse index: %w", err)
				}
				triData = append(triData, int(v1), int(v2), int(v3))
				if listSize == 4 {
					v4, err := listDataReader(le.reader)
					if err != nil {
						return fmt.Errorf("unable to parse index: %w", err)
					}
					triData = append(triData, int(v1), int(v3), int(v4))
				}
				return nil
			}
			continue
		}

		if prop.Name() == "texcoord" {
			var listDataReader func(*bufio.Reader) (float32, error)
			switch listProp.listType {
			case Float:
				listDataReader = binReaderFloat32Reader(le.order)

			default:
				return nil, nil, fmt.Errorf("unimplemented list element scalar-type: %s", listProp.listType)
			}

			readers[i] = func(r *bufio.Reader) error {
				listSize, err := listCountReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse list size: %w", err)
				}

				if listSize < 6 || listSize > 8 {
					return fmt.Errorf("unimplemented tesselation scenario where face tex data is of size: %d", listSize)
				}

				v1X, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				v1Y, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v1X, v1Y).ToFloat64())

				v2X, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				v2Y, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v2X, v2Y).ToFloat64())

				v3X, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				v3Y, err := listDataReader(le.reader)
				if err != nil {
					return fmt.Errorf("unable to parse texcord: %w", err)
				}
				uvCords = append(uvCords, vector2.New(v3X, v3Y).ToFloat64())

				if listSize == 4 {
					v4X, err := listDataReader(le.reader)
					if err != nil {
						return fmt.Errorf("unable to parse texcord: %w", err)
					}
					v4Y, err := listDataReader(le.reader)
					if err != nil {
						return fmt.Errorf("unable to parse texcord: %w", err)
					}
					uvCords = append(uvCords, vector2.New(v1X, v1Y).ToFloat64())
					uvCords = append(uvCords, vector2.New(v3X, v3Y).ToFloat64())
					uvCords = append(uvCords, vector2.New(v4X, v4Y).ToFloat64())
				}
				return nil
			}
			continue
		}
	}

	for i := 0; i < element.count; i++ {
		for _, reader := range readers {
			reader(le.reader)
		}
	}

	return triData, uvCords, nil
}

func (le *BinaryReader) ReadMesh() (*modeling.Mesh, error) {
	var vertexData map[string][]vector3.Float64
	var triData []int
	var uvData []vector2.Float64
	for _, element := range le.elements {
		if element.name == "vertex" {
			data, err := le.readVertexData(element)
			if err != nil {
				return nil, err
			}
			vertexData = data
		}

		if element.name == "face" {
			data, uvs, err := le.readFaceData(element)
			if err != nil {
				return nil, err
			}
			uvData = uvs
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
		if len(uvData) == len(triData) {
			finalMesh = finalMesh.
				Unweld().
				SetFloat2Attribute(modeling.TexCoordAttribute, uvData)
		}
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
