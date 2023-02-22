package ply

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
)

type Format int64

const (
	ASCII Format = iota
	BinaryBigEndian
	BinaryLittleEndian
)

func readLine(in *bufio.Reader) (string, error) {
	data := make([]byte, 0)

	for {
		b, err := in.ReadByte()
		if err != nil {
			return "", err
		}

		if b == '\n' {
			return string(data), nil
		}

		data = append(data, b)
	}
}

func scanToNextNonEmptyLine(reader *bufio.Reader) (string, error) {
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

func readPlyHeaderFormat(reader *bufio.Reader) (Format, error) {
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
			countType: scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[2])],
			listType:  scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[3])],
		}, nil
	}

	if len(contents) != 3 {
		return nil, errors.New("ill-formatted scalar property")
	}

	return ScalarProperty{
		name: contents[2],
		Type: scalarPropTypeNameToScalarPropertyType[strings.ToLower(contents[1])],
	}, nil
}

func readPlyHeader(in io.Reader) (reader, *modeling.Material, error) {
	reader := bufio.NewReader(in)
	magicNumber, err := readLine(reader)
	if err != nil {
		return nil, nil, err
	}

	if magicNumber != "ply" {
		return nil, nil, fmt.Errorf("unrecognized magic number: '%s' (expected 'ply')", magicNumber)
	}

	format, err := readPlyHeaderFormat(reader)
	if err != nil {
		return nil, nil, err
	}

	elements := make([]Element, 0)

	var mats *modeling.Material

	for {
		line, err := readLine(reader)
		if err != nil {
			return nil, nil, err
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
				mats = &modeling.Material{
					Name:            name,
					DiffuseColor:    color.White,
					ColorTextureURI: &name,
				}
			}
			continue
		}

		if contents[0] == "element" {
			if len(contents) != 3 {
				return nil, nil, errors.New("illegal element line in ply header")
			}

			elementCount, err := strconv.Atoi(contents[2])
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse element count: %w", err)
			}

			elements = append(elements, Element{
				name:       strings.ToLower(contents[1]),
				count:      elementCount,
				properties: make([]Property, 0),
			})
		}

		if contents[0] == "property" {
			property, err := readPlyProperty(contents)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse property: %w", err)
			}
			elements[len(elements)-1].properties = append(elements[len(elements)-1].properties, property)
		}
	}

	switch format {
	case ASCII:
		return &AsciiReader{elements: elements, scanner: bufio.NewScanner(reader)}, mats, nil

	case BinaryLittleEndian:
		return &BinaryReader{
			elements: elements,
			order:    binary.LittleEndian,
			reader:   reader,
		}, mats, nil

	case BinaryBigEndian:
		return &BinaryReader{
			elements: elements,
			order:    binary.BigEndian,
			reader:   reader,
		}, mats, nil

	default:
		return nil, mats, fmt.Errorf("unimplemented ply format: %d", format)
	}
}

func ReadMesh(in io.Reader) (*modeling.Mesh, error) {
	reader, mat, err := readPlyHeader(in)
	if err != nil {
		return nil, err
	}

	mesh, err := reader.ReadMesh()
	if err != nil {
		return nil, err
	}

	if mat != nil {
		matmesh := mesh.SetMaterial(*mat)
		mesh = &matmesh
	}

	return mesh, nil
}

func Load(filepath string) (*modeling.Mesh, error) {
	in, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	return ReadMesh(in)
}

func buildVertexElements(attributes []string, size int) Element {
	vertexElement := Element{
		name:       "vertex",
		count:      size,
		properties: make([]Property, 0),
	}

	for _, attribute := range attributes {
		if attribute == modeling.PositionAttribute {
			vertexElement.properties = append(vertexElement.properties,
				&ScalarProperty{
					name: "x",
					Type: Float,
				},
				&ScalarProperty{
					name: "y",
					Type: Float,
				},
				&ScalarProperty{
					name: "z",
					Type: Float,
				},
			)
			continue
		}

		if attribute == modeling.NormalAttribute {
			vertexElement.properties = append(vertexElement.properties,
				&ScalarProperty{
					name: "nx",
					Type: Float,
				},
				&ScalarProperty{
					name: "ny",
					Type: Float,
				},
				&ScalarProperty{
					name: "nz",
					Type: Float,
				},
			)
			continue
		}

		if attribute == modeling.ColorAttribute {
			vertexElement.properties = append(vertexElement.properties,
				&ScalarProperty{
					name: "red",
					Type: UChar,
				},
				&ScalarProperty{
					name: "green",
					Type: UChar,
				},
				&ScalarProperty{
					name: "blue",
					Type: UChar,
				},
			)
			continue
		}
	}

	return vertexElement
}

func WriteASCII(out io.Writer, model modeling.Mesh) error {
	fmt.Fprintln(out, "ply")
	fmt.Fprintln(out, "format ascii 1.0")

	if len(model.Materials()) > 0 && model.Materials()[0].Material != nil {
		mat := model.Materials()[0].Material
		if mat.ColorTextureURI != nil {
			fmt.Fprintf(out, "comment TextureFile %s\n", *mat.ColorTextureURI)
		}
	}

	fmt.Fprintln(out, "comment Created with github.com/EliCDavis/polyform")

	attributes := model.Float3Attributes()
	vertexCount := model.AttributeLength()
	vertexElement := buildVertexElements(attributes, vertexCount)
	vertexElement.Write(out)

	if model.Topology() != modeling.PointTopology && model.Topology() != modeling.TriangleTopology {
		panic(fmt.Errorf("unimplemented ply topology export: %s", model.Topology().String()))
	}

	if model.Topology() == modeling.TriangleTopology {
		fmt.Fprintf(out, "element face %d\n", model.PrimitiveCount())
		fmt.Fprintln(out, "property list uchar int vertex_indices")
		if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
			fmt.Fprintln(out, "property list uchar float texcoord")
		}
	}

	fmt.Fprintln(out, "end_header")

	view := model.View()

	for i := 0; i < vertexCount; i++ {
		for atrI, atr := range attributes {

			v := view.Float3Data[atr][i]

			if atr == modeling.ColorAttribute {
				fmt.Fprintf(out, "%d %d %d", int(v.X()*255), int(v.Y()*255), int(v.Z()*255))
			} else {
				fmt.Fprintf(out, "%f %f %f", v.X(), v.Y(), v.Z())
			}
			if atrI < len(attributes)-1 {
				fmt.Fprintf(out, " ")
			}
		}
		fmt.Fprint(out, "\n")
	}

	if model.Topology() == modeling.TriangleTopology {
		if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
			for i := 0; i < model.PrimitiveCount(); i++ {
				tri := model.Tri(i)
				fmt.Fprintf(
					out,
					"3 %d %d %d 6 %f %f %f %f %f %f\n",
					tri.P1(),
					tri.P2(),
					tri.P3(),
					tri.P1Vec2Attr(modeling.TexCoordAttribute).X(),
					tri.P1Vec2Attr(modeling.TexCoordAttribute).Y(),
					tri.P2Vec2Attr(modeling.TexCoordAttribute).X(),
					tri.P2Vec2Attr(modeling.TexCoordAttribute).Y(),
					tri.P3Vec2Attr(modeling.TexCoordAttribute).X(),
					tri.P3Vec2Attr(modeling.TexCoordAttribute).Y(),
				)
			}
		} else {
			for i := 0; i < model.PrimitiveCount(); i++ {
				tri := model.Tri(i)
				fmt.Fprintf(out, "3 %d %d %d\n", tri.P1(), tri.P2(), tri.P3())
			}
		}
	}

	return nil
}
