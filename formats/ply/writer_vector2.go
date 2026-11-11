package ply

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
)

type Vector2PropertyWriter struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	Type           ScalarPropertyType
}

func (v2pw Vector2PropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return mesh.HasFloat2Attribute(v2pw.ModelAttribute)
}

func (v2pw Vector2PropertyWriter) Properties(modeling.Mesh) []Property {
	return []Property{
		ScalarProperty{PropertyName: v2pw.PlyPropertyX, Type: v2pw.Type},
		ScalarProperty{PropertyName: v2pw.PlyPropertyY, Type: v2pw.Type},
	}
}

func (v2pw Vector2PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	column := vector2Column{arr: mesh.Float2Attribute(v2pw.ModelAttribute)}
	switch format {
	case ASCII:
		switch v2pw.Type {
		case UChar:
			return asciiUCharVector2(column)
		case Int, UInt, Short, UShort:
			return asciiIntVector2(column)
		case Float, Double:
			return asciiFloatVector2(column)
		}

	case BinaryBigEndian:
		switch v2pw.Type {
		case UChar:
			return ucharVector2(column)
		case Int:
			return bigIntVector2(column)
		case Float:
			return bigFloatVector2(column)
		case Double:
			return bigDoubleVector2(column)
		}

	default:
		switch v2pw.Type {
		case UChar:
			return ucharVector2(column)
		case Int:
			return littleIntVector2(column)
		case Float:
			return littleFloatVector2(column)
		case Double:
			return littleDoubleVector2(column)
		}
	}
	panic(fmt.Errorf("unimplemented %s", v2pw.Type))
}

type vector2Column struct {
	arr *iter.ArrayIterator[vector2.Float64]
}

type (
	asciiUCharVector2   vector2Column
	asciiIntVector2     vector2Column
	asciiFloatVector2   vector2Column
	ucharVector2        vector2Column
	littleIntVector2    vector2Column
	bigIntVector2       vector2Column
	littleFloatVector2  vector2Column
	bigFloatVector2     vector2Column
	littleDoubleVector2 vector2Column
	bigDoubleVector2    vector2Column
)

func (w asciiUCharVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Clamp(0, 1).Scale(255).Round()
	return appendAsciiInts(dst, v.X(), v.Y())
}

func (w asciiIntVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiInts(dst, v.X(), v.Y())
}

func (w asciiFloatVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiFloats(dst, v.X(), v.Y())
}

func (w ucharVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Scale(255).RoundToInt()
	return append(dst, byte(v.X()), byte(v.Y()))
}

func (w littleIntVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.X()))
	return binary.LittleEndian.AppendUint32(dst, uint32(v.Y()))
}

func (w bigIntVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.X()))
	return binary.BigEndian.AppendUint32(dst, uint32(v.Y()))
}

func (w littleFloatVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.X()))
	return binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.Y()))
}

func (w bigFloatVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.X()))
	return binary.BigEndian.AppendUint32(dst, math.Float32bits(v.Y()))
}

func (w littleDoubleVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.X()))
	return binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.Y()))
}

func (w bigDoubleVector2) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.X()))
	return binary.BigEndian.AppendUint64(dst, math.Float64bits(v.Y()))
}
