package ply

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
)

type Format int64

const (
	ASCII Format = iota
	Binary
)

func scanToNextNonEmptyLine(scanner *bufio.Scanner) string {
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			return text
		}
	}
	panic("end of scanner")
}

func readPlyHeaderFormat(scanner *bufio.Scanner) (Format, error) {
	line := scanToNextNonEmptyLine(scanner)
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

	"ushort": Ushort,
	"uint16": Ushort,

	"int":   Int,
	"int32": Int,

	"uint":   Uint,
	"uint32": Uint,

	"float":   Float,
	"float32": Float,

	"double":  Double,
	"float64": Double,
}

func readPlyProperty(contents []string) (Property, error) {
	if contents[1] == "list" {
		if len(contents) != 5 {
			return nil, errors.New("ill-formatted list property")
		}
		return ListProperty{
			name:      contents[4],
			countType: ScalarPropertyType(contents[2]),
			listType:  ScalarPropertyType(contents[3]),
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

func readPlyHeader(in io.Reader) (reader, error) {
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	magicNumber := scanner.Text()
	if magicNumber != "ply" {
		return nil, fmt.Errorf("unrecognized magic number: '%s' (expected 'ply')", magicNumber)
	}

	format, err := readPlyHeaderFormat(scanner)
	if err != nil {
		return nil, err
	}

	if format != ASCII {
		return nil, fmt.Errorf("unimplemented format type: %d", format)
	}

	elements := make([]Element, 0)

	for {
		scanner.Scan()
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if line == "end_header" {
			break
		}

		contents := strings.Fields(line)
		if contents[0] == "comment" {
			continue
		}

		if contents[0] == "element" {
			if len(contents) != 3 {
				return nil, errors.New("illegal element line in ply header")
			}

			elementCount, err := strconv.Atoi(contents[2])
			if err != nil {
				return nil, fmt.Errorf("unable to parse element count: %w", err)
			}

			elements = append(elements, Element{
				name:       contents[1],
				count:      elementCount,
				properties: make([]Property, 0),
			})
		}

		if contents[0] == "property" {
			property, err := readPlyProperty(contents)
			if err != nil {
				return nil, fmt.Errorf("unable to parse property: %w", err)
			}
			elements[len(elements)-1].properties = append(elements[len(elements)-1].properties, property)
		}
	}

	return &AsciiReader{elements: elements, scanner: scanner}, nil
}

func ToMesh(in io.Reader) (*modeling.Mesh, error) {
	reader, err := readPlyHeader(in)
	if err != nil {
		return nil, err
	}

	return reader.ReadMesh()
}

func Load(filepath string) (*modeling.Mesh, error) {
	in, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	return ToMesh(in)
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
	fmt.Fprintln(out, "comment created by github.com/EliCDavis/polyform")

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
		for i := 0; i < model.PrimitiveCount(); i++ {
			tri := model.Tri(i)
			fmt.Fprintf(out, "3 %d %d %d\n", tri.P1(), tri.P2(), tri.P3())
		}
	}

	return nil
}
