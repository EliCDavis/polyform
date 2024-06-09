package ply

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type Header struct {
	Format   Format    `json:"format"`
	Elements []Element `json:"elements"`
	Comments []string  `json:"comments"`
	// TextureFile *string   `json:"texture,omitempty"`

	// Object information (arbitrary text)
	ObjInfo []string `json:"objInfo"`
}

func (h Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	h.Write(buf)
	return buf.Bytes()
}

func (h Header) TextureFiles() []string {
	textures := make([]string, 0)
	for _, c := range h.Comments {
		contents := strings.Fields(c)
		if len(contents) == 0 {
			continue
		}

		if len(contents) < 2 {
			continue
		}

		if strings.ToLower(contents[0]) != "texturefile" {
			continue
		}

		start := strings.Index(strings.ToLower(c), "texturefile")

		// len("texturefile") == 11

		textures = append(textures, strings.TrimSpace(c[start+11:]))
	}
	return textures
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

	for _, info := range h.Comments {
		_, err = fmt.Fprintf(out, "comment %s\n", info)
		if err != nil {
			return
		}
	}

	for _, info := range h.ObjInfo {
		fmt.Fprintf(out, "obj_info %s\n", info)
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
