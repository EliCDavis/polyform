package ply

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Vector3PropertyWriter struct {
	ModelAttribute string
	PlyPropertyX   string
	PlyPropertyY   string
	PlyPropertyZ   string
	Type           ScalarPropertyType
}

func (v3pw Vector3PropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return mesh.HasFloat3Attribute(v3pw.ModelAttribute)
}

func (v3pw Vector3PropertyWriter) Properties(modeling.Mesh) []Property {
	return []Property{
		ScalarProperty{PropertyName: v3pw.PlyPropertyX, Type: v3pw.Type},
		ScalarProperty{PropertyName: v3pw.PlyPropertyY, Type: v3pw.Type},
		ScalarProperty{PropertyName: v3pw.PlyPropertyZ, Type: v3pw.Type},
	}
}

func (v3pw Vector3PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	column := vector3Column{arr: mesh.Float3Attribute(v3pw.ModelAttribute)}
	switch format {
	case ASCII:
		switch v3pw.Type {
		case UChar:
			return asciiUCharVector3(column)
		case Int, UInt, Short, UShort:
			return asciiIntVector3(column)
		case Float, Double:
			return asciiFloatVector3(column)
		}

	case BinaryBigEndian:
		switch v3pw.Type {
		case UChar:
			return ucharVector3(column)
		case Int:
			return bigIntVector3(column)
		case Float:
			return bigFloatVector3(column)
		case Double:
			return bigDoubleVector3(column)
		}

	default:
		switch v3pw.Type {
		case UChar:
			return ucharVector3(column)
		case Int:
			return littleIntVector3(column)
		case Float:
			return littleFloatVector3(column)
		case Double:
			return littleDoubleVector3(column)
		}
	}
	panic(fmt.Errorf("unimplemented %s", v3pw.Type))
}

type vector3Column struct {
	arr *iter.ArrayIterator[vector3.Float64]
}

type (
	asciiUCharVector3   vector3Column
	asciiIntVector3     vector3Column
	asciiFloatVector3   vector3Column
	ucharVector3        vector3Column
	littleIntVector3    vector3Column
	bigIntVector3       vector3Column
	littleFloatVector3  vector3Column
	bigFloatVector3     vector3Column
	littleDoubleVector3 vector3Column
	bigDoubleVector3    vector3Column
)

func (w asciiUCharVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Clamp(0, 1).Scale(255).Round()
	return appendAsciiInts(dst, v.X(), v.Y(), v.Z())
}

func (w asciiIntVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiInts(dst, v.X(), v.Y(), v.Z())
}

func (w asciiFloatVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	return appendAsciiFloats(dst, v.X(), v.Y(), v.Z())
}

func (w ucharVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).Scale(255).RoundToInt()
	return append(dst, byte(v.X()), byte(v.Y()), byte(v.Z()))
}

func (w littleIntVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.X()))
	dst = binary.LittleEndian.AppendUint32(dst, uint32(v.Y()))
	return binary.LittleEndian.AppendUint32(dst, uint32(v.Z()))
}

func (w bigIntVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.X()))
	dst = binary.BigEndian.AppendUint32(dst, uint32(v.Y()))
	return binary.BigEndian.AppendUint32(dst, uint32(v.Z()))
}

func (w littleFloatVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.X()))
	dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.Y()))
	return binary.LittleEndian.AppendUint32(dst, math.Float32bits(v.Z()))
}

func (w bigFloatVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i).ToFloat32()
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.X()))
	dst = binary.BigEndian.AppendUint32(dst, math.Float32bits(v.Y()))
	return binary.BigEndian.AppendUint32(dst, math.Float32bits(v.Z()))
}

func (w littleDoubleVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.X()))
	dst = binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.Y()))
	return binary.LittleEndian.AppendUint64(dst, math.Float64bits(v.Z()))
}

func (w bigDoubleVector3) Append(dst []byte, i int) []byte {
	v := w.arr.At(i)
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.X()))
	dst = binary.BigEndian.AppendUint64(dst, math.Float64bits(v.Y()))
	return binary.BigEndian.AppendUint64(dst, math.Float64bits(v.Z()))
}
