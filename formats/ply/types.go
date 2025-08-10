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

	refutil.RegisterType[nodes.Struct[ManifestNode]](factory)
	refutil.RegisterType[nodes.Struct[ReadNode]](factory)

	generator.RegisterTypes(factory)
}

type Artifact struct {
	Mesh modeling.Mesh
}

func (sa Artifact) Write(w io.Writer) error {
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

func (Artifact) Mime() string {
	return "application/octet-stream"
}

// ============================================================================

type ArtifactNode struct {
	In nodes.Output[modeling.Mesh]
}

func (pn ArtifactNode) Out(out *nodes.StructOutput[manifest.Artifact]) {
	out.Set(Artifact{Mesh: nodes.TryGetOutputValue(out, pn.In, modeling.EmptyPointcloud())})
}

// ============================================================================

type ManifestNode struct {
	Name nodes.Output[string] `description:"Name of the main file in the manifest, defaults to 'model.ply'"`
	Mesh nodes.Output[modeling.Mesh]
}

func (pn ManifestNode) Out(out *nodes.StructOutput[manifest.Manifest]) {
	name := nodes.TryGetOutputValue(out, pn.Name, "model.ply")
	mesh := nodes.TryGetOutputValue(out, pn.Mesh, modeling.EmptyPointcloud())
	metadata := map[string]any{}

	// TODO: Is this really the best way to determine if it's a splat?
	if mesh.HasFloat3Attribute(modeling.FDCAttribute) {
		metadata["gaussianSplat"] = true
	}

	entry := manifest.Entry{Artifact: Artifact{Mesh: mesh}, Metadata: metadata}
	out.Set(manifest.SingleEntryManifest(name, entry))
}

// ============================================================================

type ReadNode struct {
	In nodes.Output[[]byte]
}

func (pn ReadNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if pn.In == nil {
		out.Set(modeling.EmptyMesh(modeling.PointTopology))
		return
	}

	data := nodes.GetOutputValue(out, pn.In)
	mesh, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		out.CaptureError(err)
		return
	}

	out.Set(*mesh)
}
