package ply

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
)

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

// https://bsky.app/profile/elicdavis.bsky.social/post/3lcxkpsvgbs24
var defaultReader MeshReader = MeshReader{
	AttributeElement: VertexElementName,
	Properties: []PropertyReader{
		&Vector3PropertyReader{
			ModelAttribute: modeling.PositionAttribute,
			PlyPropertyX:   "x",
			PlyPropertyY:   "y",
			PlyPropertyZ:   "z",
		},
		&Vector3PropertyReader{
			ModelAttribute: modeling.PositionAttribute,
			PlyPropertyX:   "px",
			PlyPropertyY:   "py",
			PlyPropertyZ:   "pz",
		},
		&Vector3PropertyReader{
			ModelAttribute: modeling.PositionAttribute,
			PlyPropertyX:   "posx",
			PlyPropertyY:   "posy",
			PlyPropertyZ:   "posz",
		},
		&Vector3PropertyReader{
			ModelAttribute: modeling.NormalAttribute,
			PlyPropertyX:   "nx",
			PlyPropertyY:   "ny",
			PlyPropertyZ:   "nz",
		},
		&Vector3PropertyReader{
			ModelAttribute: modeling.NormalAttribute,
			PlyPropertyX:   "normalx",
			PlyPropertyY:   "normaly",
			PlyPropertyZ:   "normalz",
		},
		&Vector4PropertyReader{
			ModelAttribute: modeling.ColorAttribute,
			PlyPropertyX:   "red",
			PlyPropertyY:   "green",
			PlyPropertyZ:   "blue",
			PlyPropertyW:   "alpha",
		},
		&Vector4PropertyReader{
			ModelAttribute: modeling.ColorAttribute,
			PlyPropertyX:   "r",
			PlyPropertyY:   "g",
			PlyPropertyZ:   "b",
			PlyPropertyW:   "a",
		},
		&Vector4PropertyReader{
			ModelAttribute: modeling.ColorAttribute,
			PlyPropertyX:   "diffuse_red",
			PlyPropertyY:   "diffuse_green",
			PlyPropertyZ:   "diffuse_blue",
			PlyPropertyW:   "diffuse_alpha",
		},
		&Vector2PropertyReader{
			ModelAttribute: modeling.TexCoordAttribute,
			PlyPropertyX:   "s",
			PlyPropertyY:   "t",
		},

		// Gaussian Splatting >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		&Vector3PropertyReader{
			ModelAttribute: modeling.FDCAttribute,
			PlyPropertyX:   "f_dc_0",
			PlyPropertyY:   "f_dc_1",
			PlyPropertyZ:   "f_dc_2",
		},
		&Vector1PropertyReader{
			ModelAttribute: modeling.OpacityAttribute,
			PlyProperty:    "opacity",
		},
		&Vector3PropertyReader{
			ModelAttribute: modeling.ScaleAttribute,
			PlyPropertyX:   "scale_0",
			PlyPropertyY:   "scale_1",
			PlyPropertyZ:   "scale_2",
		},
		&Vector4PropertyReader{
			ModelAttribute: modeling.RotationAttribute,
			PlyPropertyX:   "rot_0",
			PlyPropertyY:   "rot_1",
			PlyPropertyZ:   "rot_2",
			PlyPropertyW:   "rot_3",
		},
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	},
}

func ReadMesh(reader io.Reader) (*modeling.Mesh, error) {
	return defaultReader.Read(reader)
}

type PropertyReader interface {
	buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader
	buildAscii(element Element) asciiPropertyReader
}

type builtPropertyReader interface {
	UpdateMesh(m modeling.Mesh) modeling.Mesh
}

type asciiPropertyReader interface {
	builtPropertyReader
	Read(buf []string, i int64) error
}

type binaryPropertyReader interface {
	builtPropertyReader
	Read(buf []byte, i int64)
}

// Builds a modeling.Mesh from PLY data
type MeshReader struct {
	// PLY Element containing the mesh attribute data on a per vertex basis
	//
	// example: "vertex"
	AttributeElement string

	// All defined translations from PLY data to mesh attributes
	Properties []PropertyReader
}

func (mr MeshReader) Load(file string) (*modeling.Mesh, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return mr.Read(bufio.NewReader(f))
}

func (mr MeshReader) Read(reader io.Reader) (*modeling.Mesh, error) {
	header, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	var vertexElement *Element
	var facesElement *Element
	for i, element := range header.Elements {
		if element.Name == "face" {
			facesElement = &header.Elements[i]
		}

		if element.Name != mr.AttributeElement {
			continue
		}
		vertexElement = &header.Elements[i]
	}

	if vertexElement == nil {
		return nil, fmt.Errorf("ply missing '%s' element", mr.AttributeElement)
	}

	totalSize := 0
	for _, prop := range vertexElement.Properties {
		scalar, ok := prop.(ScalarProperty)
		if !ok {
			return nil, fmt.Errorf("unimplemented scenario: '%s.%s' is an array property type", mr.AttributeElement, prop.Name())
		}
		totalSize += scalar.Size()
	}

	builtReaders := make([]builtPropertyReader, 0)

	var indices []int
	var uvs []vector2.Float64
	var topo modeling.Topology

	if facesElement == nil {
		topo = modeling.PointTopology
		indices = make([]int, vertexElement.Count)
		for i := 0; i < int(vertexElement.Count); i++ {
			indices[i] = i
		}
	}

	if header.Format == ASCII {
		// Build all readers
		asciiReaders := make([]asciiPropertyReader, 0)
		for _, reader := range mr.Properties {
			builtReader := reader.buildAscii(*vertexElement)
			if builtReader == nil {
				continue
			}
			builtReaders = append(builtReaders, builtReader)
			asciiReaders = append(asciiReaders, builtReader)
		}

		// Read data
		scanner := bufio.NewScanner(reader)
		for i := int64(0); i < vertexElement.Count; i++ {
			scanner.Scan()

			text := scanner.Text()
			if text == "" {
				continue
			}

			contents := strings.Fields(text)

			for _, reader := range asciiReaders {
				err = reader.Read(contents, i)
				if err != nil {
					return nil, err
				}
			}

		}

		// Read face data if present
		if facesElement != nil {
			topo = modeling.TriangleTopology
			indices, uvs, err = readAsciiFaceElement(*facesElement, scanner)
			if err != nil {
				return nil, err
			}
		}
	} else {
		var endian binary.ByteOrder = binary.LittleEndian
		if header.Format == BinaryBigEndian {
			endian = binary.BigEndian
		}

		// Build all readers
		binReaders := make([]binaryPropertyReader, 0)
		for _, reader := range mr.Properties {
			builtReader := reader.buildBinary(*vertexElement, endian)
			if builtReader == nil {
				continue
			}
			builtReaders = append(builtReaders, builtReader)
			binReaders = append(binReaders, builtReader)
		}

		// Read vertex buffers
		vertexBuf := make([]byte, totalSize)
		for i := int64(0); i < vertexElement.Count; i++ {
			_, err := io.ReadFull(reader, vertexBuf)
			if err != nil {
				return nil, fmt.Errorf("can't read %q element %w", mr.AttributeElement, err)
			}

			for _, reader := range binReaders {
				reader.Read(vertexBuf, i)
			}
		}

		// Read face data if present
		if facesElement != nil {
			topo = modeling.TriangleTopology
			indices, uvs, err = readBinaryFaceElement(*facesElement, endian, reader)
			if err != nil {
				return nil, err
			}
		}
	}

	mesh := modeling.NewMesh(topo, indices)
	for _, reader := range builtReaders {
		mesh = reader.UpdateMesh(mesh)
	}

	if len(uvs) == len(indices) {
		mesh = mesh.
			Transform(meshops.UnweldTransformer{}).
			SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
	}

	return &mesh, nil
}

func readAsciiFaceElement(element Element, scanner *bufio.Scanner) ([]int, []vector2.Float64, error) {
	indicesProp := -1
	texCordProp := -1

	readers := make([]*listAsciiPropertyReader, 0)

	for i, prop := range element.Properties {
		arrayProp, ok := prop.(ListProperty)
		if !ok {
			return nil, nil, fmt.Errorf("unimplemented scenario: %q element contains non list property %q", element.Name, prop.Name())
		}

		if prop.Name() == "vertex_index" || prop.Name() == "vertex_indices" {
			indicesProp = i
		}

		if prop.Name() == "texcoord" {
			texCordProp = i
		}

		readers = append(readers, &listAsciiPropertyReader{
			property:         arrayProp,
			lastReadListSize: -1,
		})
	}

	if indicesProp == -1 {
		return nil, nil, fmt.Errorf("%q did not contain indices property", element.Name)
	}

	indices := make([]int, 0)
	uvs := make([]vector2.Float64, 0)

	indicesBuf := make([]int, 4)
	texBuf := make([]float64, 8)

	var i int
	for i < int(element.Count) {
		scanner.Scan()
		line := scanner.Text()

		if line == "" {
			continue
		}

		contents := strings.Fields(line)

		// Read everything
		currentOffset := 0
		for readerIndex, reader := range readers {
			off, err := reader.Read(contents[currentOffset:])
			if err != nil {
				return nil, nil, err
			}
			currentOffset += off

			if readerIndex == indicesProp {
				err = reader.Int(indicesBuf)
				if err != nil {
					return nil, nil, err
				}
			}

			if readerIndex == texCordProp {
				err = reader.Float64(texBuf)
				if err != nil {
					return nil, nil, err
				}
			}
		}

		// Interpret read data ================================================
		points := readers[indicesProp].lastReadListSize

		if points < 3 || points > 4 {
			return nil, nil, fmt.Errorf("face contained indices entry of size %d", points)
		}

		indices = append(indices, indicesBuf[:3]...)

		// Tesselate the quad
		if points == 4 {
			indices = append(indices, indicesBuf[0], indicesBuf[2], indicesBuf[3])
		}

		if texCordProp > -1 {
			uvs = append(
				uvs,
				vector2.New(texBuf[0], texBuf[1]),
				vector2.New(texBuf[2], texBuf[3]),
				vector2.New(texBuf[4], texBuf[5]),
			)

			// Tesselate the quad
			if points == 4 {
				uvs = append(
					uvs,
					vector2.New(texBuf[0], texBuf[1]),
					vector2.New(texBuf[4], texBuf[5]),
					vector2.New(texBuf[6], texBuf[7]),
				)
			}
		}

		i++
	}

	return indices, uvs, nil
}

func readBinaryFaceElement(element Element, endian binary.ByteOrder, in io.Reader) ([]int, []vector2.Float64, error) {
	indicesProp := -1
	texCordProp := -1

	readers := make([]*listBinaryPropertyReader, 0)

	for i, prop := range element.Properties {
		arrayProp, ok := prop.(ListProperty)
		if !ok {
			return nil, nil, fmt.Errorf("unimplemented scenario: %q element contains non list property %q", element.Name, prop.Name())
		}

		if prop.Name() == "vertex_index" || prop.Name() == "vertex_indices" {
			indicesProp = i
		}

		if prop.Name() == "texcoord" {
			texCordProp = i
		}

		readers = append(readers, &listBinaryPropertyReader{
			property:         arrayProp,
			endian:           endian,
			buf:              make([]byte, arrayProp.CountType.Size()),
			lastReadListSize: -1,
		})
	}

	if indicesProp == -1 {
		return nil, nil, fmt.Errorf("%q did not contain indices property", element.Name)
	}

	indices := make([]int, 0)
	uvs := make([]vector2.Float64, 0)

	indicesBuf := make([]int, 4)
	texBuf := make([]float64, 8)

	for i := 0; i < int(element.Count); i++ {
		// Read everything
		for readerIndex, reader := range readers {
			err := reader.Read(in)
			if err != nil {
				return nil, nil, err
			}
			if readerIndex == indicesProp {
				reader.Int(indicesBuf)
			}

			if readerIndex == texCordProp {
				reader.Float64(texBuf)
			}
		}

		// Interpret read data ================================================
		points := readers[indicesProp].lastReadListSize

		if points < 3 || points > 4 {
			return nil, nil, fmt.Errorf("face contained indices entry of size %d", points)
		}

		indices = append(indices, indicesBuf[:3]...)

		// Tesselate the quad
		if points == 4 {
			indices = append(indices, indicesBuf[0], indicesBuf[2], indicesBuf[3])
		}

		if texCordProp > -1 {
			uvs = append(
				uvs,
				vector2.New(texBuf[0], texBuf[1]),
				vector2.New(texBuf[2], texBuf[3]),
				vector2.New(texBuf[4], texBuf[5]),
			)

			// Tesselate the quad
			if points == 4 {
				uvs = append(
					uvs,
					vector2.New(texBuf[0], texBuf[1]),
					vector2.New(texBuf[4], texBuf[5]),
					vector2.New(texBuf[6], texBuf[7]),
				)
			}
		}
	}

	return indices, uvs, nil
}
