package ply

import (
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type Splat struct {
	Mesh modeling.Mesh
}

func (sa Splat) Write(w io.Writer) error {
	return splat.Write(w, sa.Mesh)
}

func (Splat) Mime() string {
	return "application/octet-stream"
}

type SplatNode = nodes.Struct[artifact.Artifact, SplatNodeData]

type SplatNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatNodeData) Process() (artifact.Artifact, error) {
	return Splat{Mesh: pn.In.Value()}, nil
}

func NewSplatNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[artifact.Artifact] {
	return (&SplatNode{
		Data: SplatNodeData{
			In: meshNode,
		},
	}).Out()
}

// ============================================================================

type SplatPly struct {
	Mesh modeling.Mesh
}

func (sa SplatPly) Write(w io.Writer) error {

	writers := []PropertyWriter{
		Vector3PropertyWriter{
			ModelAttribute: modeling.PositionAttribute,
			Type:           Float,
			PlyPropertyX:   "x",
			PlyPropertyY:   "y",
			PlyPropertyZ:   "z",
		},
		Vector3PropertyWriter{
			ModelAttribute: modeling.NormalAttribute,
			Type:           Float,
			PlyPropertyX:   "nx",
			PlyPropertyY:   "ny",
			PlyPropertyZ:   "nz",
		},
		Vector3PropertyWriter{
			ModelAttribute: modeling.FDCAttribute,
			Type:           Float,
			PlyPropertyX:   "f_dc_0",
			PlyPropertyY:   "f_dc_1",
			PlyPropertyZ:   "f_dc_2",
		},
		Vector3PropertyWriter{
			ModelAttribute: modeling.ScaleAttribute,
			Type:           Float,
			PlyPropertyX:   "scale_0",
			PlyPropertyY:   "scale_1",
			PlyPropertyZ:   "scale_2",
		},
		Vector4PropertyWriter{
			ModelAttribute: modeling.RotationAttribute,
			Type:           Float,
			PlyPropertyX:   "rot_0",
			PlyPropertyY:   "rot_1",
			PlyPropertyZ:   "rot_2",
			PlyPropertyW:   "rot_3",
		},
		Vector1PropertyWriter{
			ModelAttribute: modeling.OpacityAttribute,
			PlyProperty:    "opacity",
			Type:           Float,
		},
	}

	harmonics := 45
	for i := 0; i < harmonics; i++ {
		writers = append(writers, Vector1PropertyWriter{
			ModelAttribute: fmt.Sprintf("f_rest_%d", i),
			PlyProperty:    fmt.Sprintf("f_rest_%d", i),
			Type:           Float,
		})
	}

	writer := MeshWriter{
		Format:     BinaryLittleEndian,
		Properties: writers,
	}

	return writer.Write(sa.Mesh, w)
}

func (SplatPly) Mime() string {
	return "application/octet-stream"
}

type SplatPlyNode = nodes.Struct[artifact.Artifact, SplatPlyNodeData]

type SplatPlyNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatPlyNodeData) Process() (artifact.Artifact, error) {
	return SplatPly{Mesh: pn.In.Value()}, nil
}

func NewSplatPlyNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[artifact.Artifact] {
	return (&SplatPlyNode{
		Data: SplatPlyNodeData{
			In: meshNode,
		},
	}).Out()
}
