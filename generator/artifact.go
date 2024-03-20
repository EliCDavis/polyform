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

type ImageArtifactNode struct {
	nodes.StructData[Artifact]
	In nodes.NodeOutput[image.Image]
}

func (pn *ImageArtifactNode) Out() nodes.NodeOutput[Artifact] {
	return &nodes.StructNodeOutput[Artifact]{Definition: pn}
}

func (pn ImageArtifactNode) Process() (Artifact, error) {
	return ImageArtifact{Image: pn.In.Data()}, nil
}

func NewImageArtifactNode(imageNode nodes.NodeOutput[image.Image]) nodes.NodeOutput[Artifact] {
	return (&ImageArtifactNode{In: imageNode}).Out()
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

type BinaryArtifactNode struct {
	nodes.StructData[Artifact]
	In nodes.NodeOutput[[]byte]
}

func (pn *BinaryArtifactNode) Out() nodes.NodeOutput[Artifact] {
	return &nodes.StructNodeOutput[Artifact]{Definition: pn}
}

func (pn BinaryArtifactNode) Process() (Artifact, error) {
	return BinaryArtifact{Data: pn.In.Data()}, nil
}

func NewBinaryArtifactNode(bytesNode nodes.NodeOutput[[]byte]) nodes.NodeOutput[Artifact] {
	return (&BinaryArtifactNode{In: bytesNode}).Out()

}

// ============================================================================

type TextArtifact struct {
	Data string
}

func (ga TextArtifact) Write(w io.Writer) error {
	_, err := w.Write([]byte(ga.Data))
	return err
}

type TextArtifactNode struct {
	nodes.StructData[Artifact]
	In nodes.NodeOutput[string]
}

func (pn *TextArtifactNode) Out() nodes.NodeOutput[Artifact] {
	return &nodes.StructNodeOutput[Artifact]{Definition: pn}
}

func (pn TextArtifactNode) Process() (Artifact, error) {
	return TextArtifact{Data: pn.In.Data()}, nil
}

func NewTextArtifactNode(textNode nodes.NodeOutput[string]) nodes.NodeOutput[Artifact] {
	return (&TextArtifactNode{In: textNode}).Out()
}

// ============================================================================

type SplatArtifact struct {
	Mesh modeling.Mesh
}

func (sa SplatArtifact) Write(w io.Writer) error {
	return splat.Write(w, sa.Mesh)
}

type SplatArtifactNode struct {
	nodes.StructData[Artifact]
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn *SplatArtifactNode) Out() nodes.NodeOutput[Artifact] {
	return &nodes.StructNodeOutput[Artifact]{Definition: pn}
}

func (pn SplatArtifactNode) Process() (Artifact, error) {
	return SplatArtifact{Mesh: pn.In.Data()}, nil
}

func NewSplatArtifactNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[Artifact] {
	return (&SplatArtifactNode{In: meshNode}).Out()
}

// ============================================================================

type IOArtifact struct {
	Reader io.Reader
}

func (ga IOArtifact) Write(w io.Writer) error {
	_, err := io.Copy(w, ga.Reader)
	return err
}

type IOArtifactNode struct {
	nodes.StructData[Artifact]
	In nodes.NodeOutput[io.Reader]
}

func (pn *IOArtifactNode) Out() nodes.NodeOutput[Artifact] {
	return &nodes.StructNodeOutput[Artifact]{Definition: pn}
}

func (pn IOArtifactNode) Process() (Artifact, error) {
	return IOArtifact{Reader: pn.In.Data()}, nil
}

func NewIOArtifactNode(readerNode nodes.NodeOutput[io.Reader]) nodes.NodeOutput[Artifact] {
	return (&IOArtifactNode{In: readerNode}).Out()

}
