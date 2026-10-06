package gltf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cube() modeling.Mesh {
	return primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads()
}

func TestWriterSkipsNilModels(t *testing.T) {
	mesh := cube()

	buf := &bytes.Buffer{}
	err := gltf.WriteBinary(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			nil,
			{Name: "real", Mesh: &mesh},
			nil,
		},
	}, buf, nil)

	require.NoError(t, err)
	assert.NotZero(t, buf.Len(), "the surviving model still wrote")
}

func TestManifestNodeDropsNilModels(t *testing.T) {
	mesh := cube()

	node := &nodes.Struct[gltf.ManifestNode]{Data: gltf.ManifestNode{
		Models: []nodes.Output[*gltf.PolyformModel]{
			nodes.ConstOutput[*gltf.PolyformModel]{Val: nil},
			nodes.ConstOutput[*gltf.PolyformModel]{Val: &gltf.PolyformModel{Name: "real", Mesh: &mesh}},
		},
	}}

	out := nodes.GetNodeOutputPort[manifest.Manifest](node, "Out").Value()
	entry, ok := out.Entries["model.glb"]
	require.True(t, ok)

	artifact, ok := entry.Artifact.(*gltf.Artifact)
	require.True(t, ok)
	require.Len(t, artifact.Scene.Models, 1, "the nil never reached the scene")
	assert.Equal(t, "real", artifact.Scene.Models[0].Name)
}

func TestModelNodeDropsNilChildren(t *testing.T) {
	mesh := cube()

	node := &nodes.Struct[gltf.ModelNode]{Data: gltf.ModelNode{
		Children: []nodes.Output[*gltf.PolyformModel]{
			nodes.ConstOutput[*gltf.PolyformModel]{Val: nil},
			nodes.ConstOutput[*gltf.PolyformModel]{Val: &gltf.PolyformModel{Name: "real", Mesh: &mesh}},
		},
	}}

	model := nodes.GetNodeOutputPort[*gltf.PolyformModel](node, "Out").Value()
	require.Len(t, model.Children, 1, "the nil never reached the model")
	assert.Equal(t, "real", model.Children[0].Name)

	buf := &bytes.Buffer{}
	require.NoError(t, gltf.WriteBinary(gltf.PolyformScene{Models: []*gltf.PolyformModel{model}}, buf, nil))
}
