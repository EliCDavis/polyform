package generator

import (
	"image"
	"image/png"
	"io"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/nodes"
)

type Artifact interface {
	Write(io.Writer) error
}

type PolyformArtifact[T any] interface {
	Artifact
	Value() T
}

// Image Artifact =============================================================

type ImageArtifact struct {
	Image image.Image
}

func (ia ImageArtifact) Write(w io.Writer) error {
	return png.Encode(w, ia.Image)
}

func ImageArtifactNode(imageNode nodes.Node[image.Image]) nodes.Node[Artifact] {
	return nodes.Transformer("Image Artifact", imageNode, func(i nodes.Node[image.Image]) (Artifact, error) {
		return &ImageArtifact{Image: i.Data()}, nil
	})
}

type GltfArtifact struct {
	Scene gltf.PolyformScene
}

func (ga GltfArtifact) Write(w io.Writer) error {
	return gltf.WriteBinary(ga.Scene, w)
}
