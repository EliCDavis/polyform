package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Vector3PropertyReader struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	PlyPropertyZ   string
}

func (v3pr Vector3PropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	totalSize := 0
	xOffset := -1
	yOffset := -1
	zOffset := -1
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

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				zOffset = -1
			}
		}

		totalSize += scalar.Size()
	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 {
		return &builtBinaryVector3PropertyReader{
			arr:            make([]vector3.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			zOffset:        zOffset,
			plyPropertyX:   v3pr.PlyPropertyX,
			plyPropertyY:   v3pr.PlyPropertyY,
			plyPropertyZ:   v3pr.PlyPropertyZ,
			modelAttribute: v3pr.ModelAttribute,
			scalarType:     scalarType,
			endian:         endian,
		}
	}

	return nil
}

func (v3pr Vector3PropertyReader) buildAscii(element Element) asciiPropertyReader {
	xOffset := -1
	yOffset := -1
	zOffset := -1
	var scalarType ScalarPropertyType
	for i, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v3pr.PlyPropertyX {
			xOffset = i
			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				xOffset = -1
			}
		}

		if scalar.PropertyName == v3pr.PlyPropertyY {
			yOffset = i

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				yOffset = -1
			}
		}

		if scalar.PropertyName == v3pr.PlyPropertyZ {
			zOffset = i

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				zOffset = -1
			}
		}

	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 {
		return &builtAsciiVector3PropertyReader{
			arr:            make([]vector3.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			zOffset:        zOffset,
			plyPropertyX:   v3pr.PlyPropertyX,
			plyPropertyY:   v3pr.PlyPropertyY,
			plyPropertyZ:   v3pr.PlyPropertyZ,
			modelAttribute: v3pr.ModelAttribute,
			scalarType:     scalarType,
		}
	}

	return nil
}

type builtAsciiVector3PropertyReader struct {
	arr            []vector3.Float64
	scalarType     ScalarPropertyType
	modelAttribute string
	xOffset        int
	yOffset        int
	zOffset        int
	plyPropertyX   string
	plyPropertyY   string
	plyPropertyZ   string
}

func (bav3pr builtAsciiVector3PropertyReader) ClaimsProperty(prop Property) bool {
	if prop.Name() == bav3pr.plyPropertyX {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyY {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyZ {
		return true
	}

	return false
}

func (bav3pr builtAsciiVector3PropertyReader) Read(buf []string, i int64) error {
	xParsed, err := strconv.ParseFloat(buf[bav3pr.xOffset], 32)
	if err != nil {
		return err
	}

	yParsed, err := strconv.ParseFloat(buf[bav3pr.yOffset], 32)
	if err != nil {
		return err
	}

	zParsed, err := strconv.ParseFloat(buf[bav3pr.zOffset], 32)
	if err != nil {
		return err
	}

	v := vector3.New(xParsed, yParsed, zParsed)
	if bav3pr.scalarType == UChar {
		v = v.DivByConstant(255.)
	}

	bav3pr.arr[i] = v
	return nil
}

func (bv3pr *builtAsciiVector3PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat3Attribute(bv3pr.modelAttribute, bv3pr.arr)
}

type builtBinaryVector3PropertyReader struct {
	arr            []vector3.Float64
	scalarType     ScalarPropertyType
	endian         binary.ByteOrder
	modelAttribute string
	xOffset        int
	yOffset        int
	zOffset        int
	plyPropertyX   string
	plyPropertyY   string
	plyPropertyZ   string
}

func (bav3pr builtBinaryVector3PropertyReader) ClaimsProperty(prop Property) bool {
	if prop.Name() == bav3pr.plyPropertyX {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyY {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyZ {
		return true
	}

	return false
}

func (bv3pr *builtBinaryVector3PropertyReader) Read(buf []byte, i int64) {

	var v vector3.Float64
	switch bv3pr.scalarType {
	case UChar:
		v = vector3.New(
			float64(buf[bv3pr.xOffset]),
			float64(buf[bv3pr.yOffset]),
			float64(buf[bv3pr.zOffset]),
		).DivByConstant(255)

	case Int:
		v = vector3.New(
			int32(bv3pr.endian.Uint32(buf[bv3pr.xOffset:])),
			int32(bv3pr.endian.Uint32(buf[bv3pr.yOffset:])),
			int32(bv3pr.endian.Uint32(buf[bv3pr.zOffset:])),
		).ToFloat64()

	case Float:
		v = vector3.New(
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.xOffset:])),
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.yOffset:])),
			math.Float32frombits(bv3pr.endian.Uint32(buf[bv3pr.zOffset:])),
		).ToFloat64()

	case Double:
		v = vector3.New(
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.xOffset:])),
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.yOffset:])),
			math.Float64frombits(bv3pr.endian.Uint64(buf[bv3pr.zOffset:])),
		)

	default:
		panic(fmt.Errorf("unimplemented %s", bv3pr.scalarType))
	}

	bv3pr.arr[i] = v
}

func (bv3pr *builtBinaryVector3PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat3Attribute(bv3pr.modelAttribute, bv3pr.arr)
}
