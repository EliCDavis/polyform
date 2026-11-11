package ply

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector4"
)

type Vector4PropertyWriter struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	PlyPropertyZ   string
	PlyPropertyW   string
	Type           ScalarPropertyType
}

func (v4pw Vector4PropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return mesh.HasFloat4Attribute(v4pw.ModelAttribute)
}

func (v4pw Vector4PropertyWriter) Properties(modeling.Mesh) []Property {
	return []Property{
		ScalarProperty{PropertyName: v4pw.PlyPropertyX, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyY, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyZ, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyW, Type: v4pw.Type},
	}
}

func (v4pw Vector4PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	column := vector4Column{arr: mesh.Float4Attribute(v4pw.ModelAttribute)}
	switch format {
	case ASCII:
		switch v4pw.Type {
		case UChar:
			return asciiUCharVector4(column)
		case Int, UInt, Short, UShort:
			return asciiIntVector4(column)
		case Float, Double:
			return asciiFloatVector4(column)
		}

	case BinaryBigEndian:
		switch v4pw.Type {
		case UChar:
			return ucharVector4(column)
		case Int:
			return bigIntVector4(column)
		case Float:
			return bigFloatVector4(column)
		case Double:
			return bigDoubleVector4(column)
		}

	default:
		switch v4pw.Type {
		case UChar:
			return ucharVector4(column)
		case Int:
			return littleIntVector4(column)
		case Float:
			return littleFloatVector4(column)
		case Double:
			return littleDoubleVector4(column)
		}
	}
	panic(fmt.Errorf("unimplemented %s", v4pw.Type))
}

type vector4Column struct {
	arr *iter.ArrayIterator[vector4.Float64]
}

type (
	asciiUCharVector4   vector4Column
	asciiIntVector4     vector4Column
	asciiFloatVector4   vector4Column
	ucharVector4        vector4Column
	littleIntVector4    vector4Column
	bigIntVector4       vector4Column
	littleFloatVector4  vector4Column
	bigFloatVector4     vector4Column
	littleDoubleVector4 vector4Column
	bigDoubleVector4    vector4Column
)

func (w asciiUCharVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Clamp(0, 1).Scale(255).Round()
	return appendAsciiInts(dst, v.X(), v.Y(), v.Z(), v.W())
}

func (w asciiIntVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiInts(dst, v.X(), v.Y(), v.Z(), v.W())
}

func (w asciiFloatVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiFloats(dst, v.X(), v.Y(), v.Z(), v.W())
}

func (w ucharVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Scale(255).RoundToInt()
	return append(dst, byte(v.X()), byte(v.Y()), byte(v.Z()), byte(v.W()))
}

func (w littleIntVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.X()))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.Y()))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.Z()))
	return binary.LittleEndian.AppendUint32(dst, uint32(v.W()))
}

func (w bigIntVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.X()))
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.Y()))
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.Z()))
	return binary.BigEndian.AppendUint32(dst, uint32(v.W()))
}

func (w littleFloatVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.X()))
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.Y()))
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.Z()))
	return binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.W()))
}

func (w bigFloatVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.X()))
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.Y()))
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.Z()))
	return binary.BigEndian.AppendUint32(dst, math.Float32bits(v.W()))
}

func (w littleDoubleVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.X()))
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.Y()))
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.Z()))
	return binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.W()))
}

func (w bigDoubleVector4) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.X()))
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.Y()))
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.Z()))
	return binary.BigEndian.AppendUint64(dst, math.Float64bits(v.W()))
}
