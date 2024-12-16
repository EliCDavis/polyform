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

var GuassianSplatVertexAttributesNoHarmonics map[string]bool = map[string]bool{
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
	var err error
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
		return "", err
	}

	contents := strings.Fields(line)

	if len(contents) != 3 {
		return "", fmt.Errorf("unrecognized format line")
	}

	if contents[0] != "format" {
		return "", fmt.Errorf("expected format line, received %s", contents[0])
	}

	if contents[2] != "1.0" {
		return "", fmt.Errorf("unrecognized version format: %s", contents[2])
	}

	switch contents[1] {
	case "ascii":
		return ASCII, nil

	case "binary_little_endian":
		return BinaryLittleEndian, nil

	case "binary_big_endian":
		return BinaryBigEndian, nil

	default:
		return "", fmt.Errorf("unrecognized format: %s", contents[1])
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

// Attempts to interpret the string as some scalar property type, and panics
// if it can't.
func ParseScalarPropertyType(str string) ScalarPropertyType {
	cleaned := strings.ToLower(strings.TrimSpace(str))
	if t, ok := scalarPropTypeNameToScalarPropertyType[cleaned]; ok {
		return t
	}
	panic(fmt.Errorf("unrecognized type %s", str))
}

func readPlyProperty(contents []string) (Property, error) {
	if strings.ToLower(contents[1]) == "list" {
		if len(contents) != 5 {
			return nil, errors.New("ill-formatted list property")
		}
		return ListProperty{
			PropertyName: strings.ToLower(contents[4]),
			CountType:    ParseScalarPropertyType(contents[2]),
			ListType:     ParseScalarPropertyType(contents[3]),
		}, nil
	}

	if len(contents) != 3 {
		return nil, errors.New("ill-formatted scalar property")
	}

	return ScalarProperty{
		PropertyName: contents[2],
		Type:         ParseScalarPropertyType(contents[1]),
	}, nil
}

// Builds a header from the contents of the reader passed in. Reading from the
// reader passed in stops once we recieve the "end_header" token
func ReadHeader(in io.Reader) (Header, error) {
	header := Header{
		Elements: make([]Element, 0),
		Comments: make([]string, 0),
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
			start := strings.Index(line, "comment")
			header.Comments = append(header.Comments, strings.TrimSpace(line[7+start:]))
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
	textures := header.TextureFiles()
	if len(textures) > 0 {
		tex := textures[0]
		mat = &modeling.Material{
			Name:            tex,
			DiffuseColor:    color.White,
			ColorTextureURI: &tex,
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
