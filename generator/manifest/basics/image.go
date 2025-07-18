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
	entry := manifest.Entry{Artifact: Image{Image: pn.Image.Value()}}
	name := nodes.TryGetOutputValue(pn.Name, "image.png")
	return nodes.NewStructOutput(manifest.SingleEntryManifest(name, entry))
}
