package ply

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type listBinaryPropertyReader struct {
	property         ListProperty
	endian           binary.ByteOrder
	buf              []byte
	lastReadListSize int32
}

func (lpr listBinaryPropertyReader) Count(in io.Reader) (int32, error) {
	switch lpr.property.CountType {
	case UChar:
		_, err := io.ReadFull(in, lpr.buf[:1])
		return int32(lpr.buf[0]), err

	case UInt, Int:
		_, err := io.ReadFull(in, lpr.buf[:4])
		return int32(lpr.endian.Uint32(lpr.buf)), err
	}
	return -1, fmt.Errorf("unimplemented list property count type: %s", lpr.property.CountType)
}

func (lpr *listBinaryPropertyReader) Read(in io.Reader) (err error) {
	lpr.lastReadListSize, err = lpr.Count(in)
	if err != nil {
		return err
	}

	payloadSize := int(lpr.lastReadListSize) * lpr.property.ListType.Size()
	if len(lpr.buf) < payloadSize {
		lpr.buf = make([]byte, payloadSize)
	}

	_, err = io.ReadFull(in, lpr.buf[:payloadSize])
	return err
}

func (lpr listBinaryPropertyReader) Int(out []int) error {
	if int32(len(out)) < lpr.lastReadListSize {
		return errors.New("can't fit property reader data in provided out slice")
	}

	for i := 0; i < int(lpr.lastReadListSize); i++ {
		switch lpr.property.ListType {

		case UInt, Int:
			out[i] = int(int32(lpr.endian.Uint32(lpr.buf[i*4:])))

		default:
			return fmt.Errorf("unimplemented list property list type: '%s' for int deserialization", lpr.property.ListType)
		}
	}

	return nil
}

func (lpr listBinaryPropertyReader) Float64(out []float64) error {
	if int32(len(out)) < lpr.lastReadListSize {
		return errors.New("can't fit property reader data in provided out slice")
	}

	for i := 0; i < int(lpr.lastReadListSize); i++ {
		switch lpr.property.ListType {
		case Float:
			out[i] = float64(math.Float32frombits(lpr.endian.Uint32(lpr.buf[i*4:])))

		case Double:
			out[i] = math.Float64frombits(lpr.endian.Uint64(lpr.buf[i*8:]))

		default:
			return fmt.Errorf("unimplemented list property list type: '%s' for float deserialization", lpr.property.ListType)
		}
	}

	return nil
}
