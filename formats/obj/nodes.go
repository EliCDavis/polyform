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

	refutil.RegisterType[nodes.Struct[ManifestNode]](factory)
	refutil.RegisterType[nodes.Struct[ReadNode]](factory)
	refutil.RegisterType[nodes.Struct[SceneNode]](factory)
	refutil.RegisterType[nodes.Struct[ObjectNode]](factory)
	refutil.RegisterType[nodes.Struct[EntryNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialNode]](factory)

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

type ManifestNode struct {
	Scene        nodes.Output[Scene]
	MaterialFile nodes.Output[string]
}

func (pn ManifestNode) Out(out *nodes.StructOutput[manifest.Manifest]) {
	entry := manifest.Entry{Artifact: Artifact{
		MaterialFile: nodes.TryGetOutputValue(out, pn.MaterialFile, ""),
		Scene:        nodes.TryGetOutputValue(out, pn.Scene, Scene{}),
	}}
	out.Set(manifest.SingleEntryManifest("model.obj", entry))
}

type SceneNode struct {
	Objects []nodes.Output[Object]
}

func (pn SceneNode) Out(out *nodes.StructOutput[Scene]) {
	out.Set(Scene{Objects: nodes.GetOutputValues(out, pn.Objects)})
}

type ObjectNode struct {
	Name   nodes.Output[string]
	Entrys []nodes.Output[Entry]
}

func (pn ObjectNode) Out(out *nodes.StructOutput[Object]) {
	out.Set(Object{
		Name:    nodes.TryGetOutputValue(out, pn.Name, ""),
		Entries: nodes.GetOutputValues(out, pn.Entrys),
	})
}

type EntryNode struct {
	Mesh     nodes.Output[modeling.Mesh]
	Material nodes.Output[Material]
}

func (pn EntryNode) Out(out *nodes.StructOutput[Entry]) {
	out.Set(Entry{
		Mesh:     nodes.TryGetOutputValue(out, pn.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		Material: nodes.TryGetOutputReference(out, pn.Material, nil),
	})
}

type MaterialNode struct {
	Name              nodes.Output[string]
	AmbientColor      nodes.Output[coloring.Color]
	DiffuseColor      nodes.Output[coloring.Color]
	SpecularColor     nodes.Output[coloring.Color]
	SpecularHighlight nodes.Output[float64]
	OpticalDensity    nodes.Output[float64]
	Transparency      nodes.Output[float64]
}

func (pn MaterialNode) Out(out *nodes.StructOutput[Material]) {
	out.Set(Material{
		Name:              nodes.TryGetOutputValue(out, pn.Name, ""),
		SpecularHighlight: nodes.TryGetOutputValue(out, pn.SpecularHighlight, 100),
		OpticalDensity:    nodes.TryGetOutputValue(out, pn.OpticalDensity, 1),
		Transparency:      nodes.TryGetOutputValue(out, pn.Transparency, 0),
		AmbientColor:      nodes.TryGetOutputReference(out, pn.AmbientColor, nil),
		DiffuseColor:      nodes.TryGetOutputReference(out, pn.DiffuseColor, nil),
		SpecularColor:     nodes.TryGetOutputReference(out, pn.SpecularColor, nil),
	})
}

type ReadNode struct {
	In nodes.Output[[]byte]
}

func (pn ReadNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if pn.In == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	data := nodes.GetOutputValue(out, pn.In)

	scene, _, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		out.CaptureError(err)
		return
	}
	out.Set(scene.ToMesh())
}
