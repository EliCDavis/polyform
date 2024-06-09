package ply

import (
	"errors"
)

// Format of the data contained within the body of the PLY file
type Format string

// Human readable string representing the format value
func (f Format) String() string {
	switch f {
	case ASCII:
		return "ASCII"

	case BinaryBigEndian:
		return "Binary Big Endian"

	case BinaryLittleEndian:
		return "Binary Little Endian"

	}
	panic(errors.New("unrecognized format " + string(f)))
}

const (
	ASCII              Format = "ascii"
	BinaryBigEndian    Format = "binary_little_endian"
	BinaryLittleEndian Format = "binary_big_endian"
)

const VertexElementName = "vertex"
