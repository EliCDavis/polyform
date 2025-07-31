package basics

import (
	"image"
	"image/png"
	"io"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type Image struct {
	Image image.Image
}

func (im Image) Mime() string {
	return "image/png"
}

func (ia Image) Write(w io.Writer) error {
	return png.Encode(w, ia.Image)
}

type ImageNode = nodes.Struct[ImageNodeData]

type ImageNodeData struct {
	Image nodes.Output[image.Image] `description:"The image to save"`
	Name  nodes.Output[string]      `description:"Name of the image file, defaults to 'image.png'"`
}

func (pn ImageNodeData) Out() nodes.StructOutput[manifest.Manifest] {
	out := nodes.StructOutput[manifest.Manifest]{}
	entry := manifest.Entry{Artifact: Image{Image: nodes.TryGetOutputValue(&out, pn.Image, nil)}}
	name := nodes.TryGetOutputValue(&out, pn.Name, "image.png")
	out.Set(manifest.SingleEntryManifest(name, entry))
	return out
}
