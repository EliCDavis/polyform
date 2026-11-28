package texturing

import (
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[PNGNode]](factory)
	registry.Register(factory)
}

type PNGNode struct {
	Texture nodes.Output[Texture[coloring.Color]]
	Name    nodes.Output[string] `default:"texture.png"`
}

func (PNGNode) Description() string {
	return "Saves a color texture as a PNG file."
}

func (n PNGNode) Out(out *nodes.StructOutput[manifest.Manifest]) {
	name := nodes.GetOutputValue(out, n.Name)
	tex := nodes.TryGetOutputValue(out, n.Texture, Texture[coloring.Color]{})
	if tex.width == 0 || tex.height == 0 {
		return
	}
	artifact := Artifact[coloring.Color]{Texture: tex, Conversion: func(c coloring.Color) color.Color { return c }}
	out.Set(manifest.SingleEntryManifest(name, manifest.Entry{Artifact: artifact}))
}
