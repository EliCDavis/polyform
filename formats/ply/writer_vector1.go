package ply

import (
	"encoding/binary"
	"fmt"
	"io"
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

func (vpw Vector1PropertyWriter) Properties() []Property {
	return []Property{
		ScalarProperty{PropertyName: vpw.PlyProperty, Type: vpw.Type},
	}
}

func (vpw Vector1PropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	if format == ASCII {
		return &asciiVector1PropertyWriter{
			arr:    mesh.Float1Attribute(vpw.ModelAttribute),
			format: vpw.Type,
			buf:    make([]byte, 0),
		}
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}
	return builtVector1PropertyWriter{
		arr:    mesh.Float1Attribute(vpw.ModelAttribute),
		format: vpw.Type,
		buf:    make([]byte, vpw.Type.Size()),
		endian: endian,
	}
}

type builtVector1PropertyWriter struct {
	arr    *iter.ArrayIterator[float64]
	format ScalarPropertyType
	endian binary.ByteOrder
	buf    []byte
}

func (bvpw builtVector1PropertyWriter) Write(out io.Writer, i int) (err error) {
	v := bvpw.arr.At(i)

	switch bvpw.format {
	case UChar:
		bvpw.buf[0] = byte(math.Round(bvpw.arr.At(i) * 255))

	case Int:
		bvpw.endian.PutUint32(bvpw.buf, uint32(v))

	case Float:
		bvpw.endian.PutUint32(bvpw.buf, math.Float32bits(float32(v)))

	case Double:
		bvpw.endian.PutUint64(bvpw.buf, math.Float64bits(v))

	default:
		panic(fmt.Errorf("unimplemented %s", bvpw.format))
	}

	_, err = out.Write(bvpw.buf)
	return
}

type asciiVector1PropertyWriter struct {
	arr    *iter.ArrayIterator[float64]
	format ScalarPropertyType
	buf    []byte
}

func (av4pw *asciiVector1PropertyWriter) Write(out io.Writer, i int) (err error) {
	v := av4pw.arr.At(i)

	switch av4pw.format {
	case UChar:
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(math.Round(v*255)), 10)

	case Int, UInt, Short, UShort:
		av4pw.buf = strconv.AppendInt(av4pw.buf, int64(v), 10)

	case Double, Float:
		av4pw.buf = strconv.AppendFloat(av4pw.buf, v, 'f', -1, 64)

	default:
		panic(fmt.Errorf("unimplemented %s", av4pw.format))
	}

	_, err = out.Write(av4pw.buf)
	av4pw.buf = av4pw.buf[:0]
	return
}
