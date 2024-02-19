package ply

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type BinaryReader struct {
	elements []Element
	order    binary.ByteOrder
	reader   io.Reader
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

func parseVector4FromByteContents(xIndex, yIndex, zIndex, wIndex floatParser) func(contents []byte) (vector4.Float64, error) {
	return func(contents []byte) (vector4.Float64, error) {
		return vector4.New(xIndex(contents), yIndex(contents), zIndex(contents), wIndex(contents)), nil
	}
}

type vertexData struct {
	F1 map[string][]float64
	F2 map[string][]vector2.Float64
	F3 map[string][]vector3.Float64
	F4 map[string][]vector4.Float64
}

func (le *BinaryReader) readVertexData(element Element, approvedData map[string]bool) (vertexData, error) {
	data := vertexData{
		F1: make(map[string][]float64),
		F2: make(map[string][]vector2.Vector[float64]),
		F3: make(map[string][]vector3.Vector[float64]),
		F4: make(map[string][]vector4.Vector[float64]),
	}

	float1AttributeReaders := make(map[string]func(contents []byte) (float64, error))
	float3AttributeReaders := make(map[string]func(contents []byte) (vector3.Float64, error))
	float4AttributeReaders := make(map[string]func(contents []byte) (vector4.Float64, error))

	var xParser floatParser = nil
	var yParser floatParser = nil
	var zParser floatParser = nil

	var nxParser floatParser = nil
	var nyParser floatParser = nil
	var nzParser floatParser = nil

	var redParser byteParser = nil
	var greenParser byteParser = nil
	var blueParser byteParser = nil

	// Found in guassian splats >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	var scale0Parser floatParser = nil
	var scale1Parser floatParser = nil
	var scale2Parser floatParser = nil

	var fDc0Parser floatParser = nil
	var fDc1Parser floatParser = nil
	var fDc2Parser floatParser = nil

	var rot0Parser floatParser = nil
	var rot1Parser floatParser = nil
	var rot2Parser floatParser = nil
	var rot3Parser floatParser = nil
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	offset := 0
	nextOffset := 0
	for _, prop := range element.Properties {
		scalarProp, ok := prop.(ScalarProperty)
		if !ok {
			return data, fmt.Errorf("encountered non-scalar property type in vertex data: %s", prop.Name())
		}
		offset = nextOffset
		nextOffset = offset + scalarProp.Size()

		if approvedData != nil {
			if !approvedData[prop.Name()] {
				continue
			}
		}

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

		if prop.Name() == "scale_0" && isScalarPropWithType(prop, Float) {
			scale0Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "scale_1" && isScalarPropWithType(prop, Float) {
			scale1Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "scale_2" && isScalarPropWithType(prop, Float) {
			scale2Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "rot_0" && isScalarPropWithType(prop, Float) {
			rot0Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "rot_1" && isScalarPropWithType(prop, Float) {
			rot1Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "rot_2" && isScalarPropWithType(prop, Float) {
			rot2Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "rot_3" && isScalarPropWithType(prop, Float) {
			rot3Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "f_dc_0" && isScalarPropWithType(prop, Float) {
			fDc0Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "f_dc_1" && isScalarPropWithType(prop, Float) {
			fDc1Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if prop.Name() == "f_dc_2" && isScalarPropWithType(prop, Float) {
			fDc2Parser = newFloatParser(le.order, scalarProp.Type, offset)
			continue
		}

		if isScalarPropWithType(prop, Float) {
			p := newFloatParser(le.order, scalarProp.Type, offset)

			name := prop.Name()
			if name == "opacity" {
				name = "Opacity"
			}

			float1AttributeReaders[name] = func(contents []byte) (float64, error) {
				return p(contents), nil
			}
		}
	}

	if xParser != nil && yParser != nil && zParser != nil {
		float3AttributeReaders[modeling.PositionAttribute] = parseVector3FromByteContents(xParser, yParser, zParser)
	}

	if nxParser != nil && nyParser != nil && nzParser != nil {
		float3AttributeReaders[modeling.NormalAttribute] = parseVector3FromByteContents(nxParser, nyParser, nzParser)
	}

	if scale0Parser != nil && scale1Parser != nil && scale2Parser != nil {
		float3AttributeReaders[modeling.ScaleAttribute] = parseVector3FromByteContents(scale0Parser, scale1Parser, scale2Parser)
	}

	if fDc0Parser != nil && fDc1Parser != nil && fDc2Parser != nil {
		float3AttributeReaders[modeling.FDCAttribute] = parseVector3FromByteContents(fDc0Parser, fDc1Parser, fDc2Parser)
	}

	if rot0Parser != nil && rot1Parser != nil && rot2Parser != nil && rot3Parser != nil {
		float4AttributeReaders[modeling.RotationAttribute] = parseVector4FromByteContents(rot0Parser, rot1Parser, rot2Parser, rot3Parser)
	}

	if redParser != nil && greenParser != nil && blueParser != nil {
		float3AttributeReaders[modeling.ColorAttribute] = func(contents []byte) (vector3.Float64, error) {
			return vector3.New(
				float64(redParser(contents))/255.,
				float64(greenParser(contents))/255.,
				float64(blueParser(contents))/255.,
			), nil
		}
	}

	i := 0
	buf := make([]byte, nextOffset)
	for i < element.Count {
		_, err := io.ReadFull(le.reader, buf)
		if err != nil {
			return data, err
		}

		for attribute, reader := range float1AttributeReaders {
			v, err := reader(buf)
			if err != nil {
				return data, err
			}
			data.F1[attribute] = append(data.F1[attribute], v)
		}

		for attribute, reader := range float3AttributeReaders {
			v, err := reader(buf)
			if err != nil {
				return data, err
			}
			data.F3[attribute] = append(data.F3[attribute], v)
		}

		for attribute, reader := range float4AttributeReaders {
			v, err := reader(buf)
			if err != nil {
				return data, err
			}
			data.F4[attribute] = append(data.F4[attribute], v)
		}
		i++
	}

	return data, nil
}

func binReaderByteReader(r io.Reader) (int, error) {
	b, err := readByte(r)
	return int(b), err
}

func binReaderIntReader(order binary.ByteOrder) func(r io.Reader) (int32, error) {
	return func(r io.Reader) (int32, error) {
		buf := make([]byte, 4)
		_, err := io.ReadFull(r, buf)
		return int32(order.Uint32(buf)), err
	}
}

func binReaderFloat32Reader(order binary.ByteOrder) func(r io.Reader) (float32, error) {
	return func(r io.Reader) (float32, error) {
		buf := make([]byte, 4)
		_, err := io.ReadFull(r, buf)
		return math.Float32frombits(order.Uint32(buf)), err
	}
}

func (le *BinaryReader) listCountReader(listProp ListProperty) (func(r io.Reader) (int, error), error) {
	switch listProp.countType {
	case UChar:
		return binReaderByteReader, nil

	default:
		return nil, fmt.Errorf("unimplemented list count scalar-type: %s", listProp.countType)
	}
}

type binReaderListReader func(r io.Reader) error

func (le *BinaryReader) readFaceData(element Element) ([]int, []vector2.Float64, error) {
	readers := make([]binReaderListReader, len(element.Properties))

	triData := make([]int, 0, element.Count*3)
	uvCords := make([]vector2.Float64, 0)

	for i, prop := range element.Properties {
		listProp, ok := prop.(ListProperty)
		if !ok {
			return nil, nil, fmt.Errorf("encountered non-list property type for face data: %s", prop.Name())
		}

		listCountReader, err := le.listCountReader(listProp)
		if err != nil {
			return nil, nil, err
		}

		if prop.Name() == "vertex_index" || prop.Name() == "vertex_indices" {
			var listDataReader func(io.Reader) (int32, error)
			switch listProp.listType {
			case Int, UInt:
				listDataReader = binReaderIntReader(le.order)

			default:
				return nil, nil, fmt.Errorf("unimplemented list element scalar-type: %s", listProp.listType)
			}

			readers[i] = func(r io.Reader) error {
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
			var listDataReader func(io.Reader) (float32, error)
			switch listProp.listType {
			case Float:
				listDataReader = binReaderFloat32Reader(le.order)

			default:
				return nil, nil, fmt.Errorf("unimplemented list element scalar-type: %s", listProp.listType)
			}

			readers[i] = func(r io.Reader) error {
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

	for i := 0; i < element.Count; i++ {
		for _, reader := range readers {
			reader(le.reader)
		}
	}

	return triData, uvCords, nil
}

func (le *BinaryReader) ReadMesh(vertexAttributes map[string]bool) (*modeling.Mesh, error) {
	var vertexData vertexData
	var triData []int
	var uvData []vector2.Float64
	for _, element := range le.elements {
		if element.Name == "vertex" {
			data, err := le.readVertexData(element, vertexAttributes)
			if err != nil {
				return nil, err
			}
			vertexData = data
		}

		if element.Name == "face" {
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
		finalMesh = modeling.
			NewTriangleMesh(triData).
			SetFloat1Data(vertexData.F1).
			SetFloat2Data(vertexData.F2).
			SetFloat3Data(vertexData.F3).
			SetFloat4Data(vertexData.F4)

		if len(uvData) == len(triData) {
			finalMesh = finalMesh.
				Transform(meshops.UnweldTransformer{}).
				SetFloat2Attribute(modeling.TexCoordAttribute, uvData)
		}
	} else {
		finalMesh = modeling.NewPointCloud(
			vertexData.F4,
			vertexData.F3,
			vertexData.F2,
			vertexData.F1,
			nil,
		)
	}

	return &finalMesh, nil
}
