package ply

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Header struct {
	Format      Format    `json:"format"`
	Elements    []Element `json:"elements"`
	TextureFile *string   `json:"texture,omitempty"`
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
		panic(fmt.Errorf("unimplemented ply format: %d", h.Format))
	}
}
