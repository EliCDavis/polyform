package ply

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
)

type BodyReader interface {
	ReadMesh(vertexAttributes map[string]bool) (*modeling.Mesh, error)
}

var GuassianSplatVertexAttributes map[string]bool = map[string]bool{
	"x":       true,
	"y":       true,
	"z":       true,
	"scale_0": true,
	"scale_1": true,
	"scale_2": true,
	"rot_0":   true,
	"rot_1":   true,
	"rot_2":   true,
	"rot_3":   true,
	"f_dc_0":  true,
	"f_dc_1":  true,
	"f_dc_2":  true,
	"opacity": true,
}

func readLine(in io.Reader) (string, error) {
	data := make([]byte, 0)

	buf := make([]byte, 1)
	var err error = nil
	for {
		_, err = io.ReadFull(in, buf)
		if err != nil {
			return "", err
		}

		b := buf[0]
		if b == '\n' {
			return string(data), nil
		}

		data = append(data, b)
	}
}

func scanToNextNonEmptyLine(reader io.Reader) (string, error) {
	for {
		text, err := readLine(reader)
		if err != nil {
			return "", err
		}

		if strings.TrimSpace(text) != "" {
			return text, nil
		}
	}
}

func readPlyHeaderFormat(reader io.Reader) (Format, error) {
	line, err := scanToNextNonEmptyLine(reader)
	if err != nil {
		return -1, err
	}

	contents := strings.Fields(line)

	if len(contents) != 3 {
		return -1, fmt.Errorf("unrecognized format line")
	}

	if contents[0] != "format" {
		return -1, fmt.Errorf("expected format line, received %s", contents[0])
	}

	if contents[2] != "1.0" {
		return -1, fmt.Errorf("unrecognized version format: %s", contents[2])
	}

	switch contents[1] {
	case "ascii":
		return ASCII, nil

	case "binary_little_endian":
		return BinaryLittleEndian, nil

	case "binary_big_endian":
		return BinaryBigEndian, nil

	default:
		return -1, fmt.Errorf("unrecognized format: %s", contents[1])
	}
}

var scalarPropTypeNameToScalarPropertyType = map[string]ScalarPropertyType{
	"char": Char,
	"int8": Char,

	"uchar": UChar,
	"uint8": UChar,

	"short": Short,
	"int16": Short,

	"ushort": UShort,
	"uint16": UShort,

	"int":   Int,
	"int32": Int,

	"uint":   UInt,
	"uint32": UInt,

	"float":   Float,
	"float32": Float,

	"double":  Double,
	"float64": Double,
}

func readPlyProperty(contents []string) (Property, error) {
	if strings.ToLower(contents[1]) == "list" {
		if len(contents) != 5 {
			return nil, errors.New("ill-formatted list property")
		}
		return ListProperty{
			name:      strings.ToLower(contents[4]),
			CountType: scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[2])],
			ListType:  scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[3])],
		}, nil
	}

	if len(contents) != 3 {
		return nil, errors.New("ill-formatted scalar property")
	}

	return ScalarProperty{
		PropertyName: contents[2],
		Type:         scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[1])],
	}, nil
}

func ReadHeader(in io.Reader) (Header, error) {
	header := Header{
		Elements: make([]Element, 0),
	}

	magicNumber, err := readLine(in)
	if err != nil {
		return header, err
	}

	if magicNumber != "ply" {
		return header, fmt.Errorf("unrecognized magic number: '%s' (expected 'ply')", magicNumber)
	}

	format, err := readPlyHeaderFormat(in)
	if err != nil {
		return header, err
	}
	header.Format = format

	for {
		line, err := readLine(in)
		if err != nil {
			return header, err
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		if line == "end_header" {
			break
		}

		contents := strings.Fields(line)
		if contents[0] == "comment" {
			if strings.ToLower(contents[1]) == "texturefile" {
				name := contents[2]
				header.TextureFile = &name
			}
			continue
		}

		if contents[0] == "element" {
			if len(contents) != 3 {
				return header, errors.New("illegal element line in ply header")
			}

			elementCount, err := strconv.ParseInt(contents[2], 10, 64)
			if err != nil {
				return header, fmt.Errorf("unable to parse element count: %w", err)
			}

			header.Elements = append(header.Elements, Element{
				Name:       strings.ToLower(contents[1]),
				Count:      elementCount,
				Properties: make([]Property, 0),
			})
		}

		if contents[0] == "property" {
			property, err := readPlyProperty(contents)
			if err != nil {
				return header, fmt.Errorf("unable to parse property: %w", err)
			}
			lastEle := header.Elements[len(header.Elements)-1]
			lastEle.Properties = append(lastEle.Properties, property)
			header.Elements[len(header.Elements)-1] = lastEle
		}
	}

	return header, nil
}

func buildReader(in io.Reader) (BodyReader, *modeling.Material, error) {
	header, err := ReadHeader(in)
	if err != nil {
		return nil, nil, err
	}

	var mat *modeling.Material = nil
	if header.TextureFile != nil {
		mat = &modeling.Material{
			Name:            *header.TextureFile,
			DiffuseColor:    color.White,
			ColorTextureURI: header.TextureFile,
		}
	}
	return header.BuildReader(in), mat, nil
}

func ReadMesh(in io.Reader) (*modeling.Mesh, error) {
	reader, mat, err := buildReader(in)
	if err != nil {
		return nil, err
	}

	mesh, err := reader.ReadMesh(nil)
	if err != nil {
		return nil, err
	}

	if mat != nil {
		matmesh := mesh.SetMaterial(*mat)
		mesh = &matmesh
	}

	return mesh, nil
}

func buildVertexElements(attributes []string, size int64) Element {
	vertexElement := Element{
		Name:       VertexElementName,
		Count:      size,
		Properties: make([]Property, 0),
	}

	for _, attribute := range attributes {
		if attribute == modeling.PositionAttribute {
			vertexElement.Properties = append(vertexElement.Properties,
				&ScalarProperty{
					PropertyName: "x",
					Type:         Float,
				},
				&ScalarProperty{
					PropertyName: "y",
					Type:         Float,
				},
				&ScalarProperty{
					PropertyName: "z",
					Type:         Float,
				},
			)
			continue
		}

		if attribute == modeling.NormalAttribute {
			vertexElement.Properties = append(vertexElement.Properties,
				&ScalarProperty{
					PropertyName: "nx",
					Type:         Float,
				},
				&ScalarProperty{
					PropertyName: "ny",
					Type:         Float,
				},
				&ScalarProperty{
					PropertyName: "nz",
					Type:         Float,
				},
			)
			continue
		}

		if attribute == modeling.ColorAttribute {
			vertexElement.Properties = append(vertexElement.Properties,
				&ScalarProperty{
					PropertyName: "red",
					Type:         UChar,
				},
				&ScalarProperty{
					PropertyName: "green",
					Type:         UChar,
				},
				&ScalarProperty{
					PropertyName: "blue",
					Type:         UChar,
				},
			)
			continue
		}
	}

	return vertexElement
}
