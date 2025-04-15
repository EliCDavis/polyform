package ply

import (
	"bytes"
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

// ============================================================================

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ManifestNode](factory)
	refutil.RegisterType[ReadNode](factory)

	generator.RegisterTypes(factory)
}

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

	return writer.Write(sa.Mesh, "", w)
}

func (SplatPly) Mime() string {
	return "application/octet-stream"
}

// ============================================================================

type ArtifactNode = nodes.Struct[ArtifactNodeData]

type ArtifactNodeData struct {
	In nodes.Output[modeling.Mesh]
}

func (pn ArtifactNodeData) Out() nodes.StructOutput[manifest.Artifact] {
	if pn.In == nil {
		return nodes.NewStructOutput[manifest.Artifact](SplatPly{Mesh: modeling.EmptyPointcloud()})
	}
	return nodes.NewStructOutput[manifest.Artifact](SplatPly{Mesh: pn.In.Value()})
}

// ============================================================================

type ManifestNode = nodes.Struct[ManifestNodeData]

type ManifestNodeData struct {
	Name nodes.Output[string] `description:"Name of the main file in the manifest, defaults to 'model.ply'"`
	Mesh nodes.Output[modeling.Mesh]
}

func (pn ManifestNodeData) Out() nodes.StructOutput[manifest.Manifest] {
	name := nodes.TryGetOutputValue(pn.Name, "model.ply")
	if pn.Mesh == nil {
		entry := manifest.Entry{Artifact: SplatPly{Mesh: modeling.EmptyPointcloud()}}
		return nodes.NewStructOutput(manifest.SingleEntryManifest(name, entry))
	}

	entry := manifest.Entry{Artifact: SplatPly{Mesh: pn.Mesh.Value()}}
	return nodes.NewStructOutput(manifest.SingleEntryManifest(name, entry))
}

// ============================================================================
type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}

	data := pn.In.Value()

	mesh, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}
	return nodes.NewStructOutput(*mesh)
}
