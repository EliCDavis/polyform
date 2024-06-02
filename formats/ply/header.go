package ply

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Header struct {
	Format      Format    `json:"format"`
	Elements    []Element `json:"elements"`
	TextureFile *string   `json:"texture,omitempty"`
}

func (h Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	h.Write(buf)
	return buf.Bytes()
}

func (h Header) Write(out io.Writer) (err error) {
	switch h.Format {
	case ASCII:
		_, err = out.Write([]byte("ply\nformat ascii 1.0\n"))

	case BinaryLittleEndian:
		_, err = out.Write([]byte("ply\nformat binary_little_endian 1.0\n"))

	case BinaryBigEndian:
		_, err = out.Write([]byte("ply\nformat binary_big_endian 1.0\n"))
	}

	if err != nil {
		return
	}

	if h.TextureFile != nil {
		_, err = fmt.Fprintf(out, "comment TextureFile %s\n", *h.TextureFile)
		if err != nil {
			return
		}
	}

	_, err = out.Write([]byte("comment Created with github.com/EliCDavis/polyform\n"))
	if err != nil {
		return
	}

	for _, element := range h.Elements {
		err = element.Write(out)
		if err != nil {
			return
		}
	}

	_, err = out.Write([]byte("end_header\n"))
	return
}

func (h Header) BuildReader(in io.Reader) BodyReader {
	switch h.Format {
	case ASCII:
		return &AsciiReader{elements: h.Elements, scanner: bufio.NewScanner(in)}

	case BinaryLittleEndian:
		return &BinaryReader{
			elements: h.Elements,
			order:    binary.LittleEndian,
			reader:   in,
		}

	case BinaryBigEndian:
		return &BinaryReader{
			elements: h.Elements,
			order:    binary.BigEndian,
			reader:   in,
		}

	default:
		panic(fmt.Errorf("unimplemented ply format: %s", h.Format))
	}
}
