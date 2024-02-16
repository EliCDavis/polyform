package generator

import (
	"image"
	"image/png"
	"io"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/modeling"
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

func ImageArtifactNode(imageNode nodes.NodeOutput[image.Image]) nodes.NodeOutput[Artifact] {
	return nodes.Transformer("Image Artifact", imageNode, func(i nodes.NodeOutput[image.Image]) (Artifact, error) {
		return &ImageArtifact{Image: i.Data()}, nil
	})
}

// ============================================================================

type GltfArtifact struct {
	Scene gltf.PolyformScene
}

func (ga GltfArtifact) Write(w io.Writer) error {
	return gltf.WriteBinary(ga.Scene, w)
}

// ============================================================================

type BinaryArtifact struct {
	Data []byte
}

func (ga BinaryArtifact) Write(w io.Writer) error {
	_, err := w.Write(ga.Data)
	return err
}

func BinaryArtifactNode(imageNode nodes.NodeOutput[[]byte]) nodes.NodeOutput[Artifact] {
	return nodes.Transformer("Binary Artifact", imageNode, func(i nodes.NodeOutput[[]byte]) (Artifact, error) {
		return &BinaryArtifact{Data: i.Data()}, nil
	})
}

// ============================================================================

type SplatArtifact struct {
	Mesh modeling.Mesh
}

func (sa SplatArtifact) Write(w io.Writer) error {
	return splat.Write(w, sa.Mesh)
}

func SplatArtifactNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[Artifact] {
	return nodes.Transformer("Splat Artifact", meshNode, func(i nodes.NodeOutput[modeling.Mesh]) (Artifact, error) {
		return &SplatArtifact{Mesh: i.Data()}, nil
	})
}

// ============================================================================

type IOArtifact struct {
	Reader io.Reader
}

func (ga IOArtifact) Write(w io.Writer) error {
	_, err := io.Copy(w, ga.Reader)
	return err
}

func IOArtifactNode(imageNode nodes.NodeOutput[io.Reader]) nodes.NodeOutput[Artifact] {
	return nodes.Transformer("IO Artifact", imageNode, func(i nodes.NodeOutput[io.Reader]) (Artifact, error) {
		return &IOArtifact{Reader: i.Data()}, nil
	})
}
