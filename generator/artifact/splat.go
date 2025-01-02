package artifact

import (
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/generator"
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

type SplatNode = nodes.Struct[generator.Artifact, SplatNodeData]

type SplatNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatNodeData) Process() (generator.Artifact, error) {
	return Splat{Mesh: pn.In.Value()}, nil
}

func NewSplatNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[generator.Artifact] {
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

	writers := []ply.PropertyWriter{
		ply.Vector3PropertyWriter{
			ModelAttribute: modeling.PositionAttribute,
			Type:           ply.Float,
			PlyPropertyX:   "x",
			PlyPropertyY:   "y",
			PlyPropertyZ:   "z",
		},
		ply.Vector3PropertyWriter{
			ModelAttribute: modeling.NormalAttribute,
			Type:           ply.Float,
			PlyPropertyX:   "nx",
			PlyPropertyY:   "ny",
			PlyPropertyZ:   "nz",
		},
		ply.Vector3PropertyWriter{
			ModelAttribute: modeling.FDCAttribute,
			Type:           ply.Float,
			PlyPropertyX:   "f_dc_0",
			PlyPropertyY:   "f_dc_1",
			PlyPropertyZ:   "f_dc_2",
		},
		ply.Vector3PropertyWriter{
			ModelAttribute: modeling.ScaleAttribute,
			Type:           ply.Float,
			PlyPropertyX:   "scale_0",
			PlyPropertyY:   "scale_1",
			PlyPropertyZ:   "scale_2",
		},
		ply.Vector4PropertyWriter{
			ModelAttribute: modeling.RotationAttribute,
			Type:           ply.Float,
			PlyPropertyX:   "rot_0",
			PlyPropertyY:   "rot_1",
			PlyPropertyZ:   "rot_2",
			PlyPropertyW:   "rot_3",
		},
		ply.Vector1PropertyWriter{
			ModelAttribute: modeling.OpacityAttribute,
			PlyProperty:    "opacity",
			Type:           ply.Float,
		},
	}

	harmonics := 45
	for i := 0; i < harmonics; i++ {
		writers = append(writers, ply.Vector1PropertyWriter{
			ModelAttribute: fmt.Sprintf("f_rest_%d", i),
			PlyProperty:    fmt.Sprintf("f_rest_%d", i),
			Type:           ply.Float,
		})
	}

	writer := ply.MeshWriter{
		Format:     ply.BinaryLittleEndian,
		Properties: writers,
	}

	return writer.Write(sa.Mesh, w)
}

func (SplatPly) Mime() string {
	return "application/octet-stream"
}

type SplatPlyNode = nodes.Struct[generator.Artifact, SplatPlyNodeData]

type SplatPlyNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatPlyNodeData) Process() (generator.Artifact, error) {
	return SplatPly{Mesh: pn.In.Value()}, nil
}

func NewSplatPlyNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[generator.Artifact] {
	return (&SplatPlyNode{
		Data: SplatPlyNodeData{
			In: meshNode,
		},
	}).Out()
}
