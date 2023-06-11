package ply

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/polyform/modeling"
)

func writeHeader(out io.Writer, model modeling.Mesh, attributes []string, vertexCount int) error {
	if len(model.Materials()) > 0 && model.Materials()[0].Material != nil {
		mat := model.Materials()[0].Material
		if mat.ColorTextureURI != nil {
			fmt.Fprintf(out, "comment TextureFile %s\n", *mat.ColorTextureURI)
		}
	}

	fmt.Fprintln(out, "comment Created with github.com/EliCDavis/polyform")

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

	_, err := fmt.Fprintln(out, "end_header")
	return err
}

func WriteASCII(out io.Writer, model modeling.Mesh) error {
	fmt.Fprintln(out, "ply")
	fmt.Fprintln(out, "format ascii 1.0")

	attributes := model.Float3Attributes()
	vertexCount := model.AttributeLength()

	if err := writeHeader(out, model, attributes, vertexCount); err != nil {
		return err
	}

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

func WriteBinary(out io.Writer, model modeling.Mesh) error {
	fmt.Fprintln(out, "ply")
	fmt.Fprintln(out, "format binary_little_endian 1.0")

	attributes := model.Float3Attributes()
	vertexCount := model.AttributeLength()

	if err := writeHeader(out, model, attributes, vertexCount); err != nil {
		return err
	}
	writer := bitlib.NewWriter(out, binary.LittleEndian)

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
			for i := 0; i < model.PrimitiveCount(); i++ {
				tri := model.Tri(i)
				writer.Byte(3)
				writer.Int32(int32(tri.P1()))
				writer.Int32(int32(tri.P2()))
				writer.Int32(int32(tri.P3()))
				writer.Byte(6)
				writer.Float32(float32(tri.P1Vec2Attr(modeling.TexCoordAttribute).X()))
				writer.Float32(float32(tri.P1Vec2Attr(modeling.TexCoordAttribute).Y()))
				writer.Float32(float32(tri.P2Vec2Attr(modeling.TexCoordAttribute).X()))
				writer.Float32(float32(tri.P2Vec2Attr(modeling.TexCoordAttribute).Y()))
				writer.Float32(float32(tri.P3Vec2Attr(modeling.TexCoordAttribute).X()))
				writer.Float32(float32(tri.P3Vec2Attr(modeling.TexCoordAttribute).Y()))
			}
		} else {
			for i := 0; i < model.PrimitiveCount(); i++ {
				tri := model.Tri(i)
				writer.Byte(3)
				writer.Int32(int32(tri.P1()))
				writer.Int32(int32(tri.P2()))
				writer.Int32(int32(tri.P3()))
			}
		}
	}

	return writer.Error()
}
