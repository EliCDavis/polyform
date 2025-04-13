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
	In nodes.Output[image.Image]
}

func (pn ImageNodeData) Out() nodes.StructOutput[manifest.Artifact] {
	return nodes.NewStructOutput[manifest.Artifact](Image{Image: pn.In.Value()})
}
