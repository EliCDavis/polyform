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
	artifact := Artifact{
		MaterialFile: nodes.TryGetOutputValue(pn.MaterialFile, ""),
	}

	if pn.Scene != nil {
		artifact.Scene = pn.Scene.Value()
	}

	entry := manifest.Entry{Artifact: artifact}
	return nodes.NewStructOutput(manifest.SingleEntryManifest("model.obj", entry))
}

type SceneNode = nodes.Struct[SceneNodeData]

type SceneNodeData struct {
	Objects []nodes.Output[Object]
}

func (pn SceneNodeData) Out() nodes.StructOutput[Scene] {
	objects := make([]Object, 0)
	for _, o := range pn.Objects {
		if o == nil {
			continue
		}
		objects = append(objects, o.Value())
	}
	return nodes.NewStructOutput(Scene{Objects: objects})
}

type ObjectNode = nodes.Struct[ObjectNodeData]

type ObjectNodeData struct {
	Name   nodes.Output[string]
	Entrys []nodes.Output[Entry]
}

func (pn ObjectNodeData) Out() nodes.StructOutput[Object] {
	entries := make([]Entry, 0)
	for _, o := range pn.Entrys {
		if o == nil {
			continue
		}
		entries = append(entries, o.Value())
	}
	return nodes.NewStructOutput(Object{Name: nodes.TryGetOutputValue(pn.Name, ""), Entries: entries})
}

type EntryNode = nodes.Struct[EntryNodeData]

type EntryNodeData struct {
	Mesh     nodes.Output[modeling.Mesh]
	Material nodes.Output[Material]
}

func (pn EntryNodeData) Out() nodes.StructOutput[Entry] {
	var mat *Material
	if pn.Material != nil {
		m := pn.Material.Value()
		mat = &m
	}

	return nodes.NewStructOutput(Entry{
		Mesh:     nodes.TryGetOutputValue(pn.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		Material: mat,
	})
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
	mat := Material{
		Name:              nodes.TryGetOutputValue(pn.Name, ""),
		SpecularHighlight: nodes.TryGetOutputValue(pn.SpecularHighlight, 100),
		OpticalDensity:    nodes.TryGetOutputValue(pn.OpticalDensity, 1),
		Transparency:      nodes.TryGetOutputValue(pn.Transparency, 0),
	}

	if pn.AmbientColor != nil {
		mat.AmbientColor = pn.AmbientColor.Value()
	}

	if pn.DiffuseColor != nil {
		mat.DiffuseColor = pn.DiffuseColor.Value()
	}

	if pn.SpecularColor != nil {
		mat.SpecularColor = pn.SpecularColor.Value()
	}

	return nodes.NewStructOutput(mat)
}

type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	data := pn.In.Value()

	scene, _, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		output := nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
		output.LogError(err)
		return output
	}
	return nodes.NewStructOutput(scene.ToMesh())
}
