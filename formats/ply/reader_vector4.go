package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector4"
)

type Vector4PropertyReader struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	PlyPropertyZ   string
	PlyPropertyW   string
	IgnorableW     bool
}

func (v4pr Vector4PropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	totalSize := 0
	xOffset := -1
	yOffset := -1
	zOffset := -1
	wOffset := -1
	var scalarType ScalarPropertyType
	for _, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v4pr.PlyPropertyX {
			xOffset = totalSize
			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				xOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyY {
			yOffset = totalSize

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				yOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyZ {
			zOffset = totalSize

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				zOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyW {
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

	if xOffset > -1 && yOffset > -1 && zOffset > -1 && wOffset > -1 {
		return &builtVector4PropertyReader{
			arr:            make([]vector4.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			zOffset:        zOffset,
			wOffset:        wOffset,
			plyPropertyX:   v4pr.PlyPropertyX,
			plyPropertyY:   v4pr.PlyPropertyY,
			plyPropertyZ:   v4pr.PlyPropertyZ,
			plyPropertyW:   v4pr.PlyPropertyW,
			modelAttribute: v4pr.ModelAttribute,
			scalarType:     scalarType,
			endian:         endian,
		}
	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 && v4pr.IgnorableW {
		return Vector3PropertyReader{
			ModelAttribute: v4pr.ModelAttribute,
			PlyPropertyX:   v4pr.PlyPropertyX,
			PlyPropertyY:   v4pr.PlyPropertyY,
			PlyPropertyZ:   v4pr.PlyPropertyZ,
		}.buildBinary(element, endian)
	}

	return nil
}

func (v4pr Vector4PropertyReader) buildAscii(element Element) asciiPropertyReader {
	xOffset := -1
	yOffset := -1
	zOffset := -1
	wOffset := -1
	var scalarType ScalarPropertyType
	for i, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v4pr.PlyPropertyX {
			xOffset = i
			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				xOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyY {
			yOffset = i

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				yOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyZ {
			zOffset = i

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				zOffset = -1
			}
		}

		if scalar.PropertyName == v4pr.PlyPropertyW {
			wOffset = i

			if string(scalarType) == "" {
				scalarType = scalar.Type
			}

			// At the moment, there's no support for mix/matching type
			if scalarType != scalar.Type {
				wOffset = -1
			}
		}
	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 && wOffset > -1 {
		return &builtAsciiVector4PropertyReader{
			arr:            make([]vector4.Float64, element.Count),
			xOffset:        xOffset,
			yOffset:        yOffset,
			zOffset:        zOffset,
			wOffset:        wOffset,
			plyPropertyX:   v4pr.PlyPropertyX,
			plyPropertyY:   v4pr.PlyPropertyY,
			plyPropertyZ:   v4pr.PlyPropertyZ,
			plyPropertyW:   v4pr.PlyPropertyW,
			modelAttribute: v4pr.ModelAttribute,
			scalarType:     scalarType,
		}
	}

	if xOffset > -1 && yOffset > -1 && zOffset > -1 && v4pr.IgnorableW {
		return Vector3PropertyReader{
			ModelAttribute: v4pr.ModelAttribute,
			PlyPropertyX:   v4pr.PlyPropertyX,
			PlyPropertyY:   v4pr.PlyPropertyY,
			PlyPropertyZ:   v4pr.PlyPropertyZ,
		}.buildAscii(element)
	}

	return nil
}

type builtAsciiVector4PropertyReader struct {
	arr            []vector4.Float64
	scalarType     ScalarPropertyType
	modelAttribute string
	xOffset        int
	yOffset        int
	zOffset        int
	wOffset        int
	plyPropertyX   string
	plyPropertyY   string
	plyPropertyZ   string
	plyPropertyW   string
}

func (bav3pr builtAsciiVector4PropertyReader) ClaimsProperty(prop Property) bool {
	if prop.Name() == bav3pr.plyPropertyX {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyY {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyZ {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyW {
		return true
	}

	return false
}

func (bav3pr builtAsciiVector4PropertyReader) Read(buf []string, i int64) error {
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

	wParsed, err := strconv.ParseFloat(buf[bav3pr.wOffset], 32)
	if err != nil {
		return err
	}

	v := vector4.New(xParsed, yParsed, zParsed, wParsed)
	if bav3pr.scalarType == UChar {
		v = v.DivByConstant(255.)
	}

	bav3pr.arr[i] = v
	return nil
}

func (bv3pr *builtAsciiVector4PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat4Attribute(bv3pr.modelAttribute, bv3pr.arr)
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
	plyPropertyX   string
	plyPropertyY   string
	plyPropertyZ   string
	plyPropertyW   string
}

func (bav3pr builtVector4PropertyReader) ClaimsProperty(prop Property) bool {
	if prop.Name() == bav3pr.plyPropertyX {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyY {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyZ {
		return true
	}

	if prop.Name() == bav3pr.plyPropertyW {
		return true
	}

	return false
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
