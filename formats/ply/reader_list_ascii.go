package ply

import (
	"errors"
	"strconv"
)

type listAsciiPropertyReader struct {
	property         ListProperty
	lastReadListSize int64
	buf              []string
}

func (lpr *listAsciiPropertyReader) Read(line []string) (offset int, err error) {
	lpr.lastReadListSize, err = strconv.ParseInt(line[0], 10, 64)
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
		v, err := strconv.ParseInt(lpr.buf[i], 10, 64)
		if err != nil {
			return err
		}
		out[i] = int(v)
	}

	return nil
}
