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

type ImageArtifactNode = nodes.StructNode[Artifact, ImageArtifactNodeData]

type ImageArtifact struct {
	Image image.Image
}

func (ia ImageArtifact) Write(w io.Writer) error {
	return png.Encode(w, ia.Image)
}

type ImageArtifactNodeData struct {
	In nodes.NodeOutput[image.Image]
}

func (pn ImageArtifactNodeData) Process() (Artifact, error) {
	return ImageArtifact{Image: pn.In.Value()}, nil
}

func NewImageArtifactNode(imageNode nodes.NodeOutput[image.Image]) nodes.NodeOutput[Artifact] {
	return (&ImageArtifactNode{
		Data: ImageArtifactNodeData{
			In: imageNode,
		},
	}).Out()
}

// ============================================================================

type GltfArtifact struct {
	Scene gltf.PolyformScene
}

func (ga GltfArtifact) Write(w io.Writer) error {
	return gltf.WriteBinary(ga.Scene, w)
}

// ============================================================================

type BinaryArtifactNode = nodes.StructNode[Artifact, BinaryArtifactNodeData]

type BinaryArtifact struct {
	Data []byte
}

func (ga BinaryArtifact) Write(w io.Writer) error {
	_, err := w.Write(ga.Data)
	return err
}

type BinaryArtifactNodeData struct {
	In nodes.NodeOutput[[]byte]
}

func (pn BinaryArtifactNodeData) Process() (Artifact, error) {
	return BinaryArtifact{Data: pn.In.Value()}, nil
}

func NewBinaryArtifactNode(bytesNode nodes.NodeOutput[[]byte]) nodes.NodeOutput[Artifact] {
	return (&BinaryArtifactNode{
		Data: BinaryArtifactNodeData{
			In: bytesNode,
		},
	}).Out()
}

// ============================================================================

type TextArtifactNode = nodes.StructNode[Artifact, TextArtifactNodeData]

type TextArtifact struct {
	Data string
}

func (ga TextArtifact) Write(w io.Writer) error {
	_, err := w.Write([]byte(ga.Data))
	return err
}

type TextArtifactNodeData struct {
	In nodes.NodeOutput[string]
}

func (pn TextArtifactNodeData) Process() (Artifact, error) {
	return TextArtifact{Data: pn.In.Value()}, nil
}

func NewTextArtifactNode(textNode nodes.NodeOutput[string]) nodes.NodeOutput[Artifact] {
	return (&TextArtifactNode{
		Data: TextArtifactNodeData{
			In: textNode,
		},
	}).Out()
}

// ============================================================================

type SplatArtifactNode = nodes.StructNode[Artifact, SplatArtifactNodeData]

type SplatArtifact struct {
	Mesh modeling.Mesh
}

func (sa SplatArtifact) Write(w io.Writer) error {
	return splat.Write(w, sa.Mesh)
}

type SplatArtifactNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatArtifactNodeData) Process() (Artifact, error) {
	return SplatArtifact{Mesh: pn.In.Value()}, nil
}

func NewSplatArtifactNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[Artifact] {
	return (&SplatArtifactNode{
		Data: SplatArtifactNodeData{
			In: meshNode,
		},
	}).Out()
}

// ============================================================================

type IOArtifactNode = nodes.StructNode[Artifact, IOArtifactNodeData]

type IOArtifact struct {
	Reader io.Reader
}

func (ga IOArtifact) Write(w io.Writer) error {
	_, err := io.Copy(w, ga.Reader)
	return err
}

type IOArtifactNodeData struct {
	In nodes.NodeOutput[io.Reader]
}

func (pn IOArtifactNodeData) Process() (Artifact, error) {
	return IOArtifact{Reader: pn.In.Value()}, nil
}

func NewIOArtifactNode(readerNode nodes.NodeOutput[io.Reader]) nodes.NodeOutput[Artifact] {
	return (&IOArtifactNode{
		Data: IOArtifactNodeData{
			In: readerNode,
		},
	}).Out()
}
