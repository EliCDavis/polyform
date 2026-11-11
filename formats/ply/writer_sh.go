package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type SHPropertyWriter struct {
	Type ScalarPropertyType
}

func (w SHPropertyWriter) MeshQualifies(mesh modeling.Mesh) bool {
	return modeling.SHCoefficientCount(mesh) > 0
}

func (w SHPropertyWriter) Properties(mesh modeling.Mesh) []Property {
	total := modeling.SHCoefficientCount(mesh) * 3
	props := make([]Property, total)
	for i := range props {
		props[i] = ScalarProperty{PropertyName: shPropertyName(i), Type: w.Type}
	}
	return props
}

func (w SHPropertyWriter) build(mesh modeling.Mesh, format Format) builtPropertyWriter {
	columns := shColumns{arrays: make([]*iter.ArrayIterator[vector3.Float64], modeling.SHCoefficientCount(mesh))}
	for j := range columns.arrays {
		columns.arrays[j] = mesh.Float3Attribute(modeling.SHAttribute(j))
	}

	switch format {
	case ASCII:
		switch w.Type {
		case Float, Double:
			return asciiFloatSH(columns)
		}

	case BinaryBigEndian:
		switch w.Type {
		case Float:
			return bigFloatSH(columns)
		case Double:
			return bigDoubleSH(columns)
		}

	default:
		switch w.Type {
		case Float:
			return littleFloatSH(columns)
		case Double:
			return littleDoubleSH(columns)
		}
	}
	panic(fmt.Errorf("unimplemented %s", w.Type))
}

// A row holds every coefficient's red, then every green, then every blue.
type shColumns struct {
	arrays []*iter.ArrayIterator[vector3.Float64]
}

func (c shColumns) grow(dst []byte, scalarSize int) ([]byte, []byte) {
	start := len(dst)
	dst = slices.Grow(dst, len(c.arrays)*3*scalarSize)[:start+len(c.arrays)*3*scalarSize]
	return dst, dst[start:]
}

type (
	asciiFloatSH   shColumns
	littleFloatSH  shColumns
	bigFloatSH     shColumns
	littleDoubleSH shColumns
	bigDoubleSH    shColumns
)

func (w asciiFloatSH) Append(dst []byte, i int) []byte {
	for channel := 0; channel < 3; channel++ {
		for j, arr := range w.arrays {
			if channel > 0 || j > 0 {
				dst = append(dst, ' ')
			}
			dst = strconv.AppendFloat(dst, arr.At(i).Component(channel), 'f', -1, 64)
		}
	}
	return dst
}

func (w littleFloatSH) Append(dst []byte, i int) []byte {
	dst, row := shColumns(w).grow(dst, 4)
	n := len(w.arrays)
	for j, arr := range w.arrays {
		v := arr.At(i).ToFloat32()
		binary.LittleEndian.PutUint32(row[j*4:], math.Float32bits(v.X()))
		binary.LittleEndian.PutUint32(row[(n+j)*4:], math.Float32bits(v.Y()))
		binary.LittleEndian.PutUint32(row[(2*n+j)*4:], math.Float32bits(v.Z()))
	}
	return dst
}

func (w bigFloatSH) Append(dst []byte, i int) []byte {
	dst, row := shColumns(w).grow(dst, 4)
	n := len(w.arrays)
	for j, arr := range w.arrays {
		v := arr.At(i).ToFloat32()
		binary.BigEndian.PutUint32(row[j*4:], math.Float32bits(v.X()))
		binary.BigEndian.PutUint32(row[(n+j)*4:], math.Float32bits(v.Y()))
		binary.BigEndian.PutUint32(row[(2*n+j)*4:], math.Float32bits(v.Z()))
	}
	return dst
}

func (w littleDoubleSH) Append(dst []byte, i int) []byte {
	dst, row := shColumns(w).grow(dst, 8)
	n := len(w.arrays)
	for j, arr := range w.arrays {
		v := arr.At(i)
		binary.LittleEndian.PutUint64(row[j*8:], math.Float64bits(v.X()))
		binary.LittleEndian.PutUint64(row[(n+j)*8:], math.Float64bits(v.Y()))
		binary.LittleEndian.PutUint64(row[(2*n+j)*8:], math.Float64bits(v.Z()))
	}
	return dst
}

func (w bigDoubleSH) Append(dst []byte, i int) []byte {
	dst, row := shColumns(w).grow(dst, 8)
	n := len(w.arrays)
	for j, arr := range w.arrays {
		v := arr.At(i)
		binary.BigEndian.PutUint64(row[j*8:], math.Float64bits(v.X()))
		binary.BigEndian.PutUint64(row[(n+j)*8:], math.Float64bits(v.Y()))
		binary.BigEndian.PutUint64(row[(2*n+j)*8:], math.Float64bits(v.Z()))
	}
	return dst
}
