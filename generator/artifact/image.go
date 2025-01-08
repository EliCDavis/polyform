package artifact

import (
	"image"
	"image/png"
	"io"

	"github.com/EliCDavis/polyform/nodes"
)

type Image struct {
	Image image.Image
}

func (im Image) Mime() string {
	return "iamge/png"
}

func (ia Image) Write(w io.Writer) error {
	return png.Encode(w, ia.Image)
}

type ImageNode = nodes.Struct[Artifact, ImageNodeData]

type ImageNodeData struct {
	In nodes.NodeOutput[image.Image]
}

func (pn ImageNodeData) Process() (Artifact, error) {
	return Image{Image: pn.In.Value()}, nil
}

func NewImageNode(imageNode nodes.NodeOutput[image.Image]) nodes.NodeOutput[Artifact] {
	return (&ImageNode{
		Data: ImageNodeData{
			In: imageNode,
		},
	}).Out()
}
