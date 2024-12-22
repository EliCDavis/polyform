package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/EliCDavis/polyform/modeling"
)

type Vector1PropertyReader struct {
	ModelAttribute string
	PlyProperty    string
}

func (v1pr Vector1PropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	totalSize := 0
	for _, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v1pr.PlyProperty {
			return &builtVector1PropertyReader{
				arr:            make([]float64, element.Count),
				offset:         totalSize,
				modelAttribute: v1pr.ModelAttribute,
				scalarType:     scalar.Type,
				endian:         endian,
				plyProperty:    v1pr.PlyProperty,
			}
		}

		totalSize += scalar.Size()
	}

	return nil
}

func (v1pr Vector1PropertyReader) buildAscii(element Element) asciiPropertyReader {
	var scalarType ScalarPropertyType
	for i, prop := range element.Properties {
		scalar := prop.(ScalarProperty)

		if scalar.PropertyName == v1pr.PlyProperty {
			return &builtAsciiVector1PropertyReader{
				arr:            make([]float64, element.Count),
				offset:         i,
				modelAttribute: v1pr.ModelAttribute,
				scalarType:     scalarType,
				plyProperty:    v1pr.PlyProperty,
			}
		}

	}

	return nil
}

type builtAsciiVector1PropertyReader struct {
	arr            []float64
	scalarType     ScalarPropertyType
	modelAttribute string
	offset         int
	plyProperty    string
}

func (bav3pr builtAsciiVector1PropertyReader) Read(buf []string, i int64) error {
	v, err := strconv.ParseFloat(buf[bav3pr.offset], 32)
	if err != nil {
		return err
	}

	if bav3pr.scalarType == UChar {
		v /= 255.
	}

	bav3pr.arr[i] = v
	return nil
}

func (bv3pr *builtAsciiVector1PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat1Attribute(bv3pr.modelAttribute, bv3pr.arr)
}

func (bav3pr builtAsciiVector1PropertyReader) ClaimsProperty(prop Property) bool {
	return prop.Name() == bav3pr.plyProperty
}

type builtVector1PropertyReader struct {
	arr            []float64
	scalarType     ScalarPropertyType
	endian         binary.ByteOrder
	modelAttribute string
	offset         int
	plyProperty    string
}

func (bv1pr *builtVector1PropertyReader) Read(buf []byte, i int64) {

	var v float64
	switch bv1pr.scalarType {
	case UChar:
		v = float64(buf[bv1pr.offset]) / 255.

	case Int:
		v = float64(int32(bv1pr.endian.Uint32(buf[bv1pr.offset:])))

	case Float:
		v = float64(math.Float32frombits(bv1pr.endian.Uint32(buf[bv1pr.offset:])))

	case Double:
		v = math.Float64frombits(bv1pr.endian.Uint64(buf[bv1pr.offset:]))

	default:
		panic(fmt.Errorf("unimplemented %s", bv1pr.scalarType))
	}

	bv1pr.arr[i] = v
}

func (bv1pr *builtVector1PropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	return m.SetFloat1Attribute(bv1pr.modelAttribute, bv1pr.arr)
}

func (bav3pr builtVector1PropertyReader) ClaimsProperty(prop Property) bool {
	return prop.Name() == bav3pr.plyProperty
}
