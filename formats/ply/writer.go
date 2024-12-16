package ply

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/polyform/formats/txt"
	"github.com/EliCDavis/polyform/modeling"
)

// ============================================================================

type PropertyWriter interface {
	MeshQualifies(mesh modeling.Mesh) bool
	Properties() []Property
	build(mesh modeling.Mesh, endian Format) builtPropertyWriter
}

type builtPropertyWriter interface {
	Write(out io.Writer, i int) error
}

// ============================================================================

type MeshWriter struct {
	Format     Format
	Properties []PropertyWriter
}

func (mw MeshWriter) Write(mesh modeling.Mesh, writer io.Writer) error {

	properties := make([]Property, 0)
	filteredProps := make([]builtPropertyWriter, 0)

	for _, prop := range mw.Properties {
		if !prop.MeshQualifies(mesh) {
			continue
		}
		filteredProps = append(filteredProps, prop.build(mesh, mw.Format))
		properties = append(properties, prop.Properties()...)
	}

	attributeLength := mesh.AttributeLength()

	header := Header{
		Format: mw.Format,
		Elements: []Element{
			{
				Name:       VertexElementName,
				Count:      int64(attributeLength),
				Properties: properties,
			},
		},
		Comments: []string{},
	}

	if len(mesh.Materials()) > 0 && mesh.Materials()[0].Material != nil {
		mat := mesh.Materials()[0].Material
		if mat.ColorTextureURI != nil {
			header.Comments = append(header.Comments, fmt.Sprintf("TextureFile %s", *mat.ColorTextureURI))
		}
	}

	header.Comments = append(header.Comments, "Created with github.com/EliCDavis/polyform")

	if mesh.Topology() == modeling.TriangleTopology {

		faceProps := []Property{
			ListProperty{
				PropertyName: "vertex_indices",
				CountType:    UChar,
				ListType:     Int,
			},
		}

		if mesh.HasFloat2Attribute(modeling.TexCoordAttribute) {
			faceProps = append(faceProps, ListProperty{
				PropertyName: "texcoord",
				CountType:    UChar,
				ListType:     Float,
			})
		}

		header.Elements = append(header.Elements, Element{
			Name:       "face",
			Count:      int64(mesh.PrimitiveCount()),
			Properties: faceProps,
		})
	}

	err := header.Write(writer)
	if err != nil {
		return err
	}

	spaceByte := []byte{' '}
	newLineByte := []byte{'\n'}

	for i := 0; i < attributeLength; i++ {
		for propI, prop := range filteredProps {
			err = prop.Write(writer, i)
			if err != nil {
				return err
			}

			if mw.Format == ASCII {
				if propI < len(filteredProps)-1 {
					writer.Write(spaceByte)
				} else {
					writer.Write(newLineByte)
				}
			}

		}
	}

	if mesh.Topology() != modeling.TriangleTopology {
		return nil
	}

	switch mw.Format {
	case ASCII:
		return writeAsciiTriTopo(writer, mesh)

	case BinaryLittleEndian, BinaryBigEndian:
		return writeBinaryTriTopo(writer, mesh, mw.Format)
	}

	return nil
}

func writeBinaryTriTopo(out io.Writer, model modeling.Mesh, format Format) error {
	if model.Topology() != modeling.TriangleTopology {
		return nil
	}

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}

	indices := model.Indices()

	if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
		texData := model.Float2Attribute(modeling.TexCoordAttribute)
		buf := make([]byte, 1+(3*4)+(2*4*3))
		buf[0] = 3
		buf[13] = 6

		for i := 0; i < indices.Len(); i += 3 {
			endian.PutUint32(buf[1:], uint32(indices.At(i)))
			endian.PutUint32(buf[5:], uint32(indices.At(i+1)))
			endian.PutUint32(buf[9:], uint32(indices.At(i+2)))

			p1 := texData.At(i).ToFloat32()
			endian.PutUint32(buf[14:], math.Float32bits(p1.X()))
			endian.PutUint32(buf[18:], math.Float32bits(p1.Y()))

			p2 := texData.At(i + 1).ToFloat32()
			endian.PutUint32(buf[22:], math.Float32bits(p2.X()))
			endian.PutUint32(buf[26:], math.Float32bits(p2.Y()))

			p3 := texData.At(i + 2).ToFloat32()
			endian.PutUint32(buf[30:], math.Float32bits(p3.X()))
			endian.PutUint32(buf[34:], math.Float32bits(p3.Y()))

			_, err := out.Write(buf)
			if err != nil {
				return err
			}
		}
		return nil
	}

	buf := make([]byte, 1+(3*4))
	buf[0] = 3
	for i := 0; i < indices.Len(); i += 3 {
		endian.PutUint32(buf[1:], uint32(indices.At(i)))
		endian.PutUint32(buf[5:], uint32(indices.At(i+1)))
		endian.PutUint32(buf[9:], uint32(indices.At(i+2)))
		_, err := out.Write(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeAsciiTriTopo(out io.Writer, model modeling.Mesh) error {
	writer := txt.NewWriter(out)

	if model.Topology() != modeling.TriangleTopology {
		return nil
	}

	if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
		for i := 0; i < model.PrimitiveCount(); i++ {
			writer.StartEntry()
			writer.String("3 ")

			tri := model.Tri(i)

			writer.Int(tri.P1())
			writer.Space()
			writer.Int(tri.P2())
			writer.Space()
			writer.Int(tri.P3())

			writer.String(" 6 ")

			writer.Float64(tri.P1Vec2Attr(modeling.TexCoordAttribute).X())
			writer.Space()
			writer.Float64(tri.P1Vec2Attr(modeling.TexCoordAttribute).Y())
			writer.Space()
			writer.Float64(tri.P2Vec2Attr(modeling.TexCoordAttribute).X())
			writer.Space()
			writer.Float64(tri.P2Vec2Attr(modeling.TexCoordAttribute).Y())
			writer.Space()
			writer.Float64(tri.P3Vec2Attr(modeling.TexCoordAttribute).X())
			writer.Space()
			writer.Float64(tri.P3Vec2Attr(modeling.TexCoordAttribute).Y())
			writer.NewLine()
			writer.FinishEntry()
		}
		return writer.Error()
	}

	for i := 0; i < model.PrimitiveCount(); i++ {
		writer.StartEntry()
		writer.String("3 ")

		tri := model.Tri(i)

		writer.Int(tri.P1())
		writer.Space()
		writer.Int(tri.P2())
		writer.Space()
		writer.Int(tri.P3())
		writer.NewLine()

		writer.FinishEntry()
	}
	return writer.Error()
}
