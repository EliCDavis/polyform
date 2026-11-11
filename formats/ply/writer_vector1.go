package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
)

type Vector1PropertyWriter struct {
	ModelAttribute string
	PlyProperty    string
	Type           ScalarPropertyType
}

func (vpw Vector1PropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return mesh.HasFloat1Attribute(vpw.ModelAttribute)
}

func (vpw Vector1PropertyWriter) Properties(modeling.Mesh) []Property {
	return []Property{
		ScalarProperty{PropertyName: vpw.PlyProperty, Type: vpw.Type},
	}
}

func (vpw Vector1PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	column := vector1Column{arr: mesh.Float1Attribute(vpw.ModelAttribute)}
	switch format {
	case ASCII:
		switch vpw.Type {
		case UChar:
			return asciiUCharVector1(column)
		case Int, UInt, Short, UShort:
			return asciiIntVector1(column)
		case Float, Double:
			return asciiFloatVector1(column)
		}

	case BinaryBigEndian:
		switch vpw.Type {
		case UChar:
			return ucharVector1(column)
		case Int:
			return bigIntVector1(column)
		case Float:
			return bigFloatVector1(column)
		case Double:
			return bigDoubleVector1(column)
		}

	default:
		switch vpw.Type {
		case UChar:
			return ucharVector1(column)
		case Int:
			return littleIntVector1(column)
		case Float:
			return littleFloatVector1(column)
		case Double:
			return littleDoubleVector1(column)
		}
	}
	panic(fmt.Errorf("unimplemented %s", vpw.Type))
}

type vector1Column struct {
	arr *iter.ArrayIterator[float64]
}

type (
	asciiUCharVector1   vector1Column
	asciiIntVector1     vector1Column
	asciiFloatVector1   vector1Column
	ucharVector1        vector1Column
	littleIntVector1    vector1Column
	bigIntVector1       vector1Column
	littleFloatVector1  vector1Column
	bigFloatVector1     vector1Column
	littleDoubleVector1 vector1Column
	bigDoubleVector1    vector1Column
)

func (w asciiUCharVector1) Append(dst []byte, i int) []byte {
	return strconv.AppendInt(dst, int64(math.Round(w.arr.At(i)*255)), 10)
}

func (w asciiIntVector1) Append(dst []byte, i int) []byte {
	return strconv.AppendInt(dst, int64(w.arr.At(i)), 10)
}

func (w asciiFloatVector1) Append(dst []byte, i int) []byte {
	return strconv.AppendFloat(dst, w.arr.At(i), 'f', -1, 64)
}

func (w ucharVector1) Append(dst []byte, i int) []byte {
	return append(dst, byte(math.Round(w.arr.At(i)*255)))
}

func (w littleIntVector1) Append(dst []byte, i int) []byte {
	return binary.LittleEndian.AppendUint32(dst, uint32(w.arr.At(i)))
}

func (w bigIntVector1) Append(dst []byte, i int) []byte {
	return binary.BigEndian.AppendUint32(dst, uint32(w.arr.At(i)))
}

func (w littleFloatVector1) Append(dst []byte, i int) []byte {
	return binary.LittleEndian.AppendUint32(dst, math.Float32bits(float32(w.arr.At(i))))
}

func (w bigFloatVector1) Append(dst []byte, i int) []byte {
	return binary.BigEndian.AppendUint32(dst, math.Float32bits(float32(w.arr.At(i))))
}

func (w littleDoubleVector1) Append(dst []byte, i int) []byte {
	return binary.LittleEndian.AppendUint64(dst, math.Float64bits(w.arr.At(i)))
}

func (w bigDoubleVector1) Append(dst []byte, i int) []byte {
	return binary.BigEndian.AppendUint64(dst, math.Float64bits(w.arr.At(i)))
}
