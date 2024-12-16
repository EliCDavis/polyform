package ply

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

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

func (v4pw Vector4PropertyWriter) Properties() []Property {
	return []Property{
		ScalarProperty{PropertyName: v4pw.PlyPropertyX, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyY, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyZ, Type: v4pw.Type},
		ScalarProperty{PropertyName: v4pw.PlyPropertyW, Type: v4pw.Type},
	}
}

func (v4pw Vector4PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {

	if format == ASCII {
		return &asciiVector4PropertyWriter{
			arr:    mesh.Float4Attribute(v4pw.ModelAttribute),
			format: v4pw.Type,
			buf:    make([]byte, 0),
		}
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}
	return binaryVector4PropertyWriter{
		arr:    mesh.Float4Attribute(v4pw.ModelAttribute),
		format: v4pw.Type,
		buf:    make([]byte, v4pw.Type.Size()*4),
		endian: endian,
	}
}

type binaryVector4PropertyWriter struct {
	arr    *iter.ArrayIterator[vector4.Float64]
	format ScalarPropertyType
	endian binary.ByteOrder
	buf    []byte
}

func (bv4pw binaryVector4PropertyWriter) Write(out io.Writer, i int) (err error) {

	switch bv4pw.format {
	case UChar:
		v4 := bv4pw.arr.At(i).Scale(255).RoundToInt()
		bv4pw.buf[0] = byte(v4.X())
		bv4pw.buf[1] = byte(v4.Y())
		bv4pw.buf[2] = byte(v4.Z())
		bv4pw.buf[3] = byte(v4.W())

	case Int:
		v4 := bv4pw.arr.At(i)
		bv4pw.endian.PutUint32(bv4pw.buf, uint32(v4.X()))
		bv4pw.endian.PutUint32(bv4pw.buf[4:], uint32(v4.Y()))
		bv4pw.endian.PutUint32(bv4pw.buf[8:], uint32(v4.Z()))
		bv4pw.endian.PutUint32(bv4pw.buf[12:], uint32(v4.W()))

	case Float:
		v4 := bv4pw.arr.At(i).ToFloat32()
		bv4pw.endian.PutUint32(bv4pw.buf, math.Float32bits(v4.X()))
		bv4pw.endian.PutUint32(bv4pw.buf[4:], math.Float32bits(v4.Y()))
		bv4pw.endian.PutUint32(bv4pw.buf[8:], math.Float32bits(v4.Z()))
		bv4pw.endian.PutUint32(bv4pw.buf[12:], math.Float32bits(v4.W()))

	case Double:
		v4 := bv4pw.arr.At(i)
		bv4pw.endian.PutUint64(bv4pw.buf, math.Float64bits(v4.X()))
		bv4pw.endian.PutUint64(bv4pw.buf[8:], math.Float64bits(v4.Y()))
		bv4pw.endian.PutUint64(bv4pw.buf[16:], math.Float64bits(v4.Z()))
		bv4pw.endian.PutUint64(bv4pw.buf[24:], math.Float64bits(v4.W()))

	default:
		panic(fmt.Errorf("unimplemented %s", bv4pw.format))
	}

	_, err = out.Write(bv4pw.buf)
	return
}

type asciiVector4PropertyWriter struct {
	arr    *iter.ArrayIterator[vector4.Float64]
	format ScalarPropertyType
	buf    []byte
}

func (av4pw *asciiVector4PropertyWriter) Write(out io.Writer, i int) (err error) {

	switch av4pw.format {
	case UChar:
		v4 := av4pw.arr.At(i).Scale(255).RoundToInt()
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.X()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.Y()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.Z()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.W()), 10)

	case Int, UInt, Short, UShort:
		v4 := av4pw.arr.At(i)
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.X()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.Y()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.Z()), 10)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v4.W()), 10)

	case Double, Float:
		v4 := av4pw.arr.At(i)
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v4.X(), 'f', -1, 64)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v4.Y(), 'f', -1, 64)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v4.Z(), 'f', -1, 64)
		av4pw.buf = append(av4pw.buf, ' ')
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v4.W(), 'f', -1, 64)

	default:
		panic(fmt.Errorf("unimplemented %s", av4pw.format))
	}

	_, err = out.Write(av4pw.buf)
	av4pw.buf = av4pw.buf[:0]
	return
}
