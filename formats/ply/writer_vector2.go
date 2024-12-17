package ply

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

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

func (v3pw Vector2PropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return mesh.HasFloat2Attribute(v3pw.ModelAttribute)
}

func (v3pw Vector2PropertyWriter) Properties() []Property {
	return []Property{
		ScalarProperty{PropertyName: v3pw.PlyPropertyX, Type: v3pw.Type},
		ScalarProperty{PropertyName: v3pw.PlyPropertyY, Type: v3pw.Type},
	}
}

func (v3pw Vector2PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	if format == ASCII {
		return &asciiVector2PropertyWriter{
			arr:    mesh.Float2Attribute(v3pw.ModelAttribute),
			format: v3pw.Type,
			buf:    make([]byte, 0),
		}
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}
	return builtVector2PropertyWriter{
		arr:    mesh.Float2Attribute(v3pw.ModelAttribute),
		format: v3pw.Type,
		buf:    make([]byte, v3pw.Type.Size()*2),
		endian: endian,
	}
}

type builtVector2PropertyWriter struct {
	arr    *iter.ArrayIterator[vector2.Float64]
	format ScalarPropertyType
	endian binary.ByteOrder
	buf    []byte
}

func (bv3pw builtVector2PropertyWriter) Write(out io.Writer, i int) (err error) {

	switch bv3pw.format {
	case UChar:
		v3 := bv3pw.arr.At(i).Scale(255).RoundToInt()
		bv3pw.buf[0] = byte(v3.X())
		bv3pw.buf[1] = byte(v3.Y())

	case Int:
		v3 := bv3pw.arr.At(i)
		bv3pw.endian.PutUint32(bv3pw.buf, uint32(v3.X()))
		bv3pw.endian.PutUint32(bv3pw.buf[4:], uint32(v3.Y()))

	case Float:
		v3 := bv3pw.arr.At(i).ToFloat32()
		bv3pw.endian.PutUint32(bv3pw.buf, math.Float32bits(v3.X()))
		bv3pw.endian.PutUint32(bv3pw.buf[4:], math.Float32bits(v3.Y()))

	case Double:
		v3 := bv3pw.arr.At(i)
		bv3pw.endian.PutUint64(bv3pw.buf, math.Float64bits(v3.X()))
		bv3pw.endian.PutUint64(bv3pw.buf[8:], math.Float64bits(v3.Y()))

	default:
		panic(fmt.Errorf("unimplemented %s", bv3pw.format))
	}

	_, err = out.Write(bv3pw.buf)
	return
}

type asciiVector2PropertyWriter struct {
	arr    *iter.ArrayIterator[vector2.Float64]
	format ScalarPropertyType
	buf    []byte
}

func (av2pw *asciiVector2PropertyWriter) Write(out io.Writer, i int) (err error) {
	v2 := av2pw.arr.At(i)

	switch av2pw.format {
	case UChar:
		v2 = av2pw.arr.At(i).Clamp(0, 1).Scale(255).Round()
		fallthrough

	case Int, UInt, Short, UShort:
		av2pw.buf = strconv.AppendInt(av2pw.buf, int64(v2.X()), 10)
		av2pw.buf = append(av2pw.buf, ' ')
		av2pw.buf = strconv.AppendInt(av2pw.buf, int64(v2.Y()), 10)

	case Double, Float:
		av2pw.buf = strconv.AppendFloat(av2pw.buf, v2.X(), 'f', -1, 64)
		av2pw.buf = append(av2pw.buf, ' ')
		av2pw.buf = strconv.AppendFloat(av2pw.buf, v2.Y(), 'f', -1, 64)

	default:
		panic(fmt.Errorf("unimplemented %s", av2pw.format))
	}

	_, err = out.Write(av2pw.buf)
	av2pw.buf = av2pw.buf[:0]
	return
}
