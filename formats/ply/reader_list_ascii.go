package ply

import (
	"errors"
	"strconv"
)

type listAsciiPropertyReader struct {
	property         ListProperty
	lastReadListSize int32
	buf              []string
}

func (lpr *listAsciiPropertyReader) Read(line []string) (offset int, err error) {
	v, err := strconv.ParseInt(line[0], 10, 32)
	if err != nil {
		return -1, err
	}
	lpr.lastReadListSize = int32(v)

	// Resize to fit contents
	if len(lpr.buf) < int(lpr.lastReadListSize) {
		lpr.buf = make([]string, lpr.lastReadListSize)
	}

	copy(lpr.buf, line[1:lpr.lastReadListSize+1])
	return int(lpr.lastReadListSize) + 1, err
}

func (lpr listAsciiPropertyReader) Float64(out []float64) (err error) {
	if len(out) < int(lpr.lastReadListSize) {
		return errors.New("can't fit property reader data in provided out slice")
	}

	for i := 0; i < int(lpr.lastReadListSize); i++ {
		v, err := strconv.ParseFloat(lpr.buf[i], 64)
		if err != nil {
			return err
		}
		out[i] = v
	}

	return nil
}

func (lpr listAsciiPropertyReader) Int(out []int) error {
	if len(out) < int(lpr.lastReadListSize) {
		return errors.New("can't fit property reader data in provided out slice")
	}

	for i := 0; i < int(lpr.lastReadListSize); i++ {
		v, err := strconv.ParseInt(lpr.buf[i], 10, 32)
		if err != nil {
			return err
		}
		out[i] = int(v)
	}

	return nil
}
