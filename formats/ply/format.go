package ply

import "fmt"

type Format int64

func (f Format) String() string {
	switch f {
	case ASCII:
		return "ASCII"

	case BinaryBigEndian:
		return "Binary Big Endian"

	case BinaryLittleEndian:
		return "Binary Little Endian"

	}
	panic(fmt.Errorf("unrecognized format %d", f))
}

const (
	ASCII Format = iota
	BinaryBigEndian
	BinaryLittleEndian
)

const VertexElementName = "vertex"
