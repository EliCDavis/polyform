package ply

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/polyform/modeling"
)

func BuildHeaderFromModel(model modeling.Mesh, format Format) Header {
	if model.Topology() != modeling.PointTopology && model.Topology() != modeling.TriangleTopology {
		panic(fmt.Errorf("unimplemented ply topology export: %s", model.Topology().String()))
	}

	header := Header{
		Format: format,
		Elements: []Element{
			buildVertexElements(model.Float3Attributes(), int64(model.AttributeLength())),
		},
		Comments: []string{},
	}

	// Pull a texture file if relevant.
	if len(model.Materials()) > 0 && model.Materials()[0].Material != nil {
		mat := model.Materials()[0].Material
		if mat.ColorTextureURI != nil {
			header.Comments = append(header.Comments, fmt.Sprintf("TextureFile %s", *mat.ColorTextureURI))
		}
	}

	header.Comments = append(header.Comments, "Created with github.com/EliCDavis/polyform")

	// Optionally build face element
	if model.Topology() == modeling.TriangleTopology {
		faceProperties := []Property{
			ListProperty{
				PropertyName: "vertex_indices",
				CountType:    UChar,
				ListType:     Int,
			},
		}

		if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
			faceProperties = append(faceProperties, ListProperty{
				PropertyName: "texcoord",
				CountType:    UChar,
				ListType:     Float,
			})
		}

		header.Elements = append(header.Elements, Element{
			Name:       "face",
			Count:      int64(model.PrimitiveCount()),
			Properties: faceProperties,
		})
	}

	return header
}

func Write(out io.Writer, model modeling.Mesh, format Format) error {
	switch format {
	case ASCII:
		return writeASCII(out, model)

	case BinaryLittleEndian, BinaryBigEndian:
		return writeBinary(out, model, format)

	default:
		panic(fmt.Errorf("unrecognized format %s", format))
	}
}

func writeASCII(out io.Writer, model modeling.Mesh) error {

	header := BuildHeaderFromModel(model, ASCII)
	err := header.Write(out)
	if err != nil {
		return err
	}

	attributes := model.Float3Attributes()
	vertexCount := model.AttributeLength()

	for i := 0; i < vertexCount; i++ {
		for atrI, atr := range attributes {

			v := model.Float3Attribute(atr).At(i)

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

func writeBinary(out io.Writer, model modeling.Mesh, format Format) error {

	header := BuildHeaderFromModel(model, format)
	err := header.Write(out)
	if err != nil {
		return err
	}

	attributes := model.Float3Attributes()
	vertexCount := model.AttributeLength()

	var endian binary.ByteOrder = binary.LittleEndian
	if format == BinaryBigEndian {
		endian = binary.BigEndian
	}

	writer := bitlib.NewWriter(out, endian)

	for i := 0; i < vertexCount; i++ {
		for _, atr := range attributes {

			v := model.Float3Attribute(atr).At(i)

			if atr == modeling.ColorAttribute {
				writer.Byte(byte(v.X() * 255))
				writer.Byte(byte(v.Y() * 255))
				writer.Byte(byte(v.Z() * 255))
			} else {
				writer.Float32(float32(v.X()))
				writer.Float32(float32(v.Y()))
				writer.Float32(float32(v.Z()))
			}
		}
	}

	if model.Topology() == modeling.TriangleTopology {
		if model.HasFloat2Attribute(modeling.TexCoordAttribute) {
			indices := model.Indices()
			texData := model.Float2Attribute(modeling.TexCoordAttribute)
			for i := 0; i < indices.Len(); i += 3 {
				writer.Byte(3)
				writer.Int32(int32(indices.At(i)))
				writer.Int32(int32(indices.At(i + 1)))
				writer.Int32(int32(indices.At(i + 2)))
				writer.Byte(6)

				p1 := texData.At(i)
				writer.Float32(float32(p1.X()))
				writer.Float32(float32(p1.Y()))

				p2 := texData.At(i + 1)
				writer.Float32(float32(p2.X()))
				writer.Float32(float32(p2.Y()))

				p3 := texData.At(i + 2)
				writer.Float32(float32(p3.X()))
				writer.Float32(float32(p3.Y()))
			}
		} else {
			indices := model.Indices()
			for i := 0; i < indices.Len(); i += 3 {
				writer.Byte(3)
				writer.Int32(int32(indices.At(i)))
				writer.Int32(int32(indices.At(i + 1)))
				writer.Int32(int32(indices.At(i + 2)))
			}
		}
	}

	return writer.Error()
}
