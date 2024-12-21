package ply

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector4"
)

type Vector4PropertyReader struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	PlyPropertyZ   string
	PlyPropertyW   string
}

func (v3pr Vector4PropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	totalSize := 0
	xOffset := -1
	yOffset := -1
	zOffset := -1
	wOffset := -1
	var scalarType ScalarPropertyType
	for _, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v3pr.PlyPropertyX {
			xOffset = totalSize
			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				xOffset = -1
			}
		}

		if scalar.PropertyName == v3pr.PlyPropertyY {
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

		if scalar.PropertyName == v3pr.PlyPropertyZ {
			zOffset = totalSize
			scalarType = scalar.Type

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				zOffset = -1
			}
		}

		if scalar.PropertyName == v3pr.PlyPropertyW {
			wOffset = totalSize
			scalarType = scalar.Type

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				wOffset = -1
			}
		}

		totalSize += scalar.Size()
	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 {
		return &builtVector4PropertyReader{
			arr:            make([]vector4.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			zOffset:        zOffset,
			wOffset:        wOffset,
			modelAttribute: v3pr.ModelAttribute,
			scalarType:     scalarType,
			endian:         endian,
		}
	}

	return nil
}

type builtVector4PropertyReader struct {
	arr            []vector4.Float64
	scalarType     ScalarPropertyType
	endian         binary.ByteOrder
	modelAttribute string
	xOffset        int
	yOffset        int
	zOffset        int
	wOffset        int
}

func (bv3pr *builtVector4PropertyReader) Read(buf []byte, i int64) {

	var v vector4.Float64
	switch bv3pr.scalarType {
	case UChar:
		v = vector4.New(
			float64(buf[bv3pr.xOffset]),
			float64(buf[bv3pr.yOffset]),
			float64(buf[bv3pr.zOffset]),
			float64(buf[bv3pr.wOffset]),
		).DivByConstant(255)

	case Int:
		v = vector4.New(
			int32(bv3pr.endian.Uint32(buf[bv3pr.xOffset:])),
			int32(bv3pr.endian.Uint32(buf[bv3pr.yOffset:])),
			int32(bv3pr.endian.Uint32(buf[bv3pr.zOffset:])),
			int32(bv3pr.endian.Uint32(buf[bv3pr.wOffset:])),
		).ToFloat64()

	case Float:
		v = vector4.New(
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.xOffset:])),
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.yOffset:])),
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.zOffset:])),
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.wOffset:])),
		).ToFloat64()

	case Double:
		v = vector4.New(
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.xOffset:])),
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.yOffset:])),
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.zOffset:])),
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.wOffset:])),
		)

	default:
		panic(fmt.Errorf("unimplemented %s", bv3pr.scalarType))
	}

	bv3pr.arr[i] = v
}

func (bv3pr *builtVector4PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat4Attribute(bv3pr.modelAttribute, bv3pr.arr)
}
