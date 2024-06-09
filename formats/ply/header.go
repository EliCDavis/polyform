package ply

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// A PLY Header dictates how to interpret the rest of the file's contents, as
// well as containing any extra information stored in the comments and obj info
type Header struct {
	Format   Format    `json:"format"`
	Elements []Element `json:"elements"`
	Comments []string  `json:"comments"` // Provide informal descriptive and contextual metadata/information
	ObjInfo  []string  `json:"objInfo"`  // Object information (arbitrary text)
}

// Builds a byte array containing the header information in PLY format.
func (h Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	h.Write(buf)
	return buf.Bytes()
}

// All texture files found within the comments section of the header
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

// Writes the contents of the header out in PLY format to the writer provided.
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

// Builds a reader to interpret the contents of the body of the PLY format,
// based on the elements and format of this header.
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
