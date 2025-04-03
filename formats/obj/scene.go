package obj

import (
	"fmt"

	"github.com/EliCDavis/polyform/formats/txt"
	"github.com/EliCDavis/polyform/modeling"
)

type Entry struct {
	Mesh     modeling.Mesh
	Material *Material
}

var vTxt = []byte("v ")
var vtTxt = []byte("vt ")
var vnTxt = []byte("vn ")

func (e Entry) writeVertexData(writer *txt.Writer) error {
	m := e.Mesh
	if m.HasFloat3Attribute(modeling.PositionAttribute) {
		posData := m.Float3Attribute(modeling.PositionAttribute)
		for i := 0; i < posData.Len(); i++ {
			v := posData.At(i)
			writer.StartEntry()
			writer.Append(vTxt)
			writer.Float64(v.X())
			writer.Space()
			writer.Float64(v.Y())
			writer.Space()
			writer.Float64(v.Z())
			writer.NewLine()
			writer.FinishEntry()
		}

		if err := writer.Error(); err != nil {
			return fmt.Errorf("failed to write position attr: %w", err)
		}
	}

	if m.HasFloat2Attribute(modeling.TexCoordAttribute) {
		uvData := m.Float2Attribute(modeling.TexCoordAttribute)
		for i := 0; i < uvData.Len(); i++ {
			uv := uvData.At(i)
			writer.StartEntry()
			writer.Append(vtTxt)
			writer.Float64(uv.X())
			writer.Space()
			writer.Float64(uv.Y())
			writer.NewLine()
			writer.FinishEntry()
		}
		if err := writer.Error(); err != nil {
			return fmt.Errorf("failed to write UV attr: %w", err)
		}
	}

	if m.HasFloat3Attribute(modeling.NormalAttribute) {
		normalData := m.Float3Attribute(modeling.NormalAttribute)
		for i := range normalData.Len() {
			n := normalData.At(i)
			writer.StartEntry()
			writer.Append(vnTxt)
			writer.Float64(n.X())
			writer.Space()
			writer.Float64(n.Y())
			writer.Space()
			writer.Float64(n.Z())
			writer.NewLine()
			writer.FinishEntry()
		}

		if err := writer.Error(); err != nil {
			return fmt.Errorf("failed to write UV normal attr: %w", err)
		}
	}

	return nil
}

type Object struct {
	Name    string
	Entries []Entry
}

func (o Object) writeVertexData(writer *txt.Writer) error {
	for _, e := range o.Entries {
		if err := e.writeVertexData(writer); err != nil {
			return err
		}
	}
	return nil
}

func (o Object) ToMesh() modeling.Mesh {
	mesh := modeling.EmptyMesh(modeling.TriangleTopology)
	for _, e := range o.Entries {
		mesh = mesh.Append(e.Mesh)
	}
	return mesh
}

type Scene struct {
	Objects []Object
}

func (s Scene) containsMaterials() bool {
	for _, o := range s.Objects {
		for _, e := range o.Entries {
			if e.Material != nil {
				return true
			}
		}
	}
	return false
}

func (s Scene) writeVertexData(writer *txt.Writer) error {
	for _, o := range s.Objects {
		if err := o.writeVertexData(writer); err != nil {
			return err
		}
	}
	return nil
}

func (s Scene) ToMesh() modeling.Mesh {
	mesh := modeling.EmptyMesh(modeling.TriangleTopology)
	for _, o := range s.Objects {
		mesh = mesh.Append(o.ToMesh())
	}
	return mesh
}
