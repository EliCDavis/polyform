package obj

import (
	"bytes"
	"io"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ManifestNode](factory)
	refutil.RegisterType[ReadNode](factory)
	refutil.RegisterType[SceneNode](factory)
	refutil.RegisterType[ObjectNode](factory)
	refutil.RegisterType[EntryNode](factory)
	refutil.RegisterType[MaterialNode](factory)

	generator.RegisterTypes(factory)
}

type Artifact struct {
	Scene        Scene
	MaterialFile string
}

func (sa Artifact) Write(w io.Writer) error {
	return Write(sa.Scene, sa.MaterialFile, w)
}

func (Artifact) Mime() string {
	return "model/obj"
}

type ManifestNode = nodes.Struct[ManifestNodeData]

type ManifestNodeData struct {
	Scene        nodes.Output[Scene]
	MaterialFile nodes.Output[string]
}

func (pn ManifestNodeData) Out() nodes.StructOutput[manifest.Manifest] {
	out := nodes.StructOutput[manifest.Manifest]{}
	entry := manifest.Entry{Artifact: Artifact{
		MaterialFile: nodes.TryGetOutputValue(&out, pn.MaterialFile, ""),
		Scene:        nodes.TryGetOutputValue(&out, pn.Scene, Scene{}),
	}}
	out.Set(manifest.SingleEntryManifest("model.obj", entry))
	return out
}

type SceneNode = nodes.Struct[SceneNodeData]

type SceneNodeData struct {
	Objects []nodes.Output[Object]
}

func (pn SceneNodeData) Out() nodes.StructOutput[Scene] {
	out := nodes.StructOutput[Scene]{}
	out.Set(Scene{Objects: nodes.GetOutputValues(&out, pn.Objects)})
	return out
}

type ObjectNode = nodes.Struct[ObjectNodeData]

type ObjectNodeData struct {
	Name   nodes.Output[string]
	Entrys []nodes.Output[Entry]
}

func (pn ObjectNodeData) Out() nodes.StructOutput[Object] {
	out := nodes.StructOutput[Object]{}
	out.Set(Object{
		Name:    nodes.TryGetOutputValue(&out, pn.Name, ""),
		Entries: nodes.GetOutputValues(&out, pn.Entrys),
	})
	return out
}

type EntryNode = nodes.Struct[EntryNodeData]

type EntryNodeData struct {
	Mesh     nodes.Output[modeling.Mesh]
	Material nodes.Output[Material]
}

func (pn EntryNodeData) Out() nodes.StructOutput[Entry] {
	out := nodes.StructOutput[Entry]{}
	out.Set(Entry{
		Mesh:     nodes.TryGetOutputValue(&out, pn.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		Material: nodes.TryGetOutputReference(&out, pn.Material, nil),
	})
	return out
}

type MaterialNode = nodes.Struct[MaterialNodeData]

type MaterialNodeData struct {
	Name              nodes.Output[string]
	AmbientColor      nodes.Output[coloring.WebColor]
	DiffuseColor      nodes.Output[coloring.WebColor]
	SpecularColor     nodes.Output[coloring.WebColor]
	SpecularHighlight nodes.Output[float64]
	OpticalDensity    nodes.Output[float64]
	Transparency      nodes.Output[float64]
}

func (pn MaterialNodeData) Out() nodes.StructOutput[Material] {
	out := nodes.StructOutput[Material]{}
	out.Set(Material{
		Name:              nodes.TryGetOutputValue(&out, pn.Name, ""),
		SpecularHighlight: nodes.TryGetOutputValue(&out, pn.SpecularHighlight, 100),
		OpticalDensity:    nodes.TryGetOutputValue(&out, pn.OpticalDensity, 1),
		Transparency:      nodes.TryGetOutputValue(&out, pn.Transparency, 0),
		AmbientColor:      nodes.TryGetOutputReference(&out, pn.AmbientColor, nil),
		DiffuseColor:      nodes.TryGetOutputReference(&out, pn.DiffuseColor, nil),
		SpecularColor:     nodes.TryGetOutputReference(&out, pn.SpecularColor, nil),
	})
	return out
}

type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	out := nodes.StructOutput[modeling.Mesh]{}
	data := nodes.GetOutputValue(&out, pn.In)

	scene, _, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		out.CaptureError(err)
		return out
	}
	out.Set(scene.ToMesh())
	return out
}
