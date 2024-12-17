package ply

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

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

func (v3pw Vector3PropertyWriter) Properties() []Property {
	return []Property{
		ScalarProperty{PropertyName: v3pw.PlyPropertyX, Type: v3pw.Type},
		ScalarProperty{PropertyName: v3pw.PlyPropertyY, Type: v3pw.Type},
		ScalarProperty{PropertyName: v3pw.PlyPropertyZ, Type: v3pw.Type},
	}
}

func (v3pw Vector3PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {

	if format == ASCII {
		return &asciiVector3PropertyWriter{
			arr:    mesh.Float3Attribute(v3pw.ModelAttribute),
			format: v3pw.Type,
			buf:    make([]byte, 0),
		}
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}

	return builtVector3PropertyWriter{
		arr:    mesh.Float3Attribute(v3pw.ModelAttribute),
		format: v3pw.Type,
		buf:    make([]byte, v3pw.Type.Size()*3),
		endian: endian,
	}
}

type builtVector3PropertyWriter struct {
	arr    *iter.ArrayIterator[vector3.Float64]
	format ScalarPropertyType
	endian binary.ByteOrder
	buf    []byte
}

func (bv3pw builtVector3PropertyWriter) Write(out io.Writer, i int) (err error) {

	switch bv3pw.format {
	case UChar:
		v3 := bv3pw.arr.At(i).Scale(255).RoundToInt()
		bv3pw.buf[0] = byte(v3.X())
		bv3pw.buf[1] = byte(v3.Y())
		bv3pw.buf[2] = byte(v3.Z())

	case Int:
		v3 := bv3pw.arr.At(i)
		bv3pw.endian.PutUint32(bv3pw.buf, uint32(v3.X()))
		bv3pw.endian.PutUint32(bv3pw.buf[4:], uint32(v3.Y()))
		bv3pw.endian.PutUint32(bv3pw.buf[8:], uint32(v3.Z()))

	case Float:
		v3 := bv3pw.arr.At(i).ToFloat32()
		bv3pw.endian.PutUint32(bv3pw.buf, math.Float32bits(v3.X()))
		bv3pw.endian.PutUint32(bv3pw.buf[4:], math.Float32bits(v3.Y()))
		bv3pw.endian.PutUint32(bv3pw.buf[8:], math.Float32bits(v3.Z()))

	case Double:
		v3 := bv3pw.arr.At(i)
		bv3pw.endian.PutUint64(bv3pw.buf, math.Float64bits(v3.X()))
		bv3pw.endian.PutUint64(bv3pw.buf[8:], math.Float64bits(v3.Y()))
		bv3pw.endian.PutUint64(bv3pw.buf[16:], math.Float64bits(v3.Z()))

	default:
		panic(fmt.Errorf("unimplemented %s", bv3pw.format))
	}

	_, err = out.Write(bv3pw.buf)
	return
}

type asciiVector3PropertyWriter struct {
	arr    *iter.ArrayIterator[vector3.Float64]
	format ScalarPropertyType
	buf    []byte
}

func (av4pw *asciiVector3PropertyWriter) Write(out io.Writer, i int) (err error) {

	v3 := av4pw.arr.At(i)

	switch av4pw.format {
	case UChar:
		v3 = av4pw.arr.At(i).Clamp(0, 1).Scale(255).Round()
		fallthrough

	case Int, UInt, Short, UShort:
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v3.X()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v3.Y()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v3.Z()), 10)

	case Double, Float:
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v3.X(), 'f', -1, 64)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v3.Y(), 'f', -1, 64)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v3.Z(), 'f', -1, 64)

	default:
		panic(fmt.Errorf("unimplemented %s", av4pw.format))
	}

	_, err = out.Write(av4pw.buf)
	av4pw.buf = av4pw.buf[:0]
	return
}
