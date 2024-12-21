package ply

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
)

type Vector2PropertyReader struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
}

func (v2pr Vector2PropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	totalSize := 0
	xOffset := -1
	yOffset := -1
	var scalarType ScalarPropertyType
	for _, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v2pr.PlyPropertyX {
			xOffset = totalSize
			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				xOffset = -1
			}
		}

		if scalar.PropertyName == v2pr.PlyPropertyY {
			yOffset = totalSize
			scalarType = scalar.Type

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				yOffset = -1
			}
		}

		totalSize += scalar.Size()
	}

	if xOffset > -1 && yOffset > -1 {
		return &builtVector2PropertyReader{
			arr:            make([]vector2.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			modelAttribute: v2pr.ModelAttribute,
			scalarType:     scalarType,
			endian:         endian,
		}
	}

	return nil
}

type builtVector2PropertyReader struct {
	arr            []vector2.Float64
	scalarType     ScalarPropertyType
	endian         binary.ByteOrder
	modelAttribute string
	xOffset        int
	yOffset        int
}

func (bv2pr *builtVector2PropertyReader) Read(buf []byte, i int64) {

	var v vector2.Float64
	switch bv2pr.scalarType {
	case UChar:
		v = vector2.New(
			float64(buf[bv2pr.xOffset]),
			float64(buf[bv2pr.yOffset]),
		).DivByConstant(255)

	case Int:
		v = vector2.New(
			int32(bv2pr.endian.Uint32(buf[bv2pr.xOffset:])),
			int32(bv2pr.endian.Uint32(buf[bv2pr.yOffset:])),
		).ToFloat64()

	case Float:
		v = vector2.New(
			math.Float32frombits(bv2pr.endian.Uint32(buf[bv2pr.xOffset:])),
			math.Float32frombits(bv2pr.endian.Uint32(buf[bv2pr.yOffset:])),
		).ToFloat64()

	case Double:
		v = vector2.New(
			math.Float64frombits(bv2pr.endian.Uint64(buf[bv2pr.xOffset:])),
			math.Float64frombits(bv2pr.endian.Uint64(buf[bv2pr.yOffset:])),
		)

	default:
		panic(fmt.Errorf("unimplemented %s", bv2pr.scalarType))
	}

	bv2pr.arr[i] = v
}

func (bv2pr *builtVector2PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat2Attribute(bv2pr.modelAttribute, bv2pr.arr)
}
