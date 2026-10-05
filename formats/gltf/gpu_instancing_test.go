package gltf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func threeInstances() []trs.TRS {
	return []trs.TRS{
		trs.Position(vector3.New(0., 0., 0.)),
		trs.Position(vector3.New(3., 0., 0.)),
		trs.Position(vector3.New(6., 0., 0.)),
	}
}

func TestWrite_GpuInstancesRepeatTheMeshAndNotTheChildren(t *testing.T) {
	cube := primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads()
	lead := primitives.Cube{Width: .1, Height: 1, Depth: .1}.UnweldedQuads()
	body := &gltf.PolyformModel{
		Name:         "body",
		Mesh:         &cube,
		GpuInstances: threeInstances(),
		Children:     []*gltf.PolyformModel{{Name: "lead", Mesh: &lead}},
	}

	var buf bytes.Buffer
	require.NoError(t, gltf.WriteText(gltf.PolyformScene{Models: []*gltf.PolyformModel{body}}, &buf, nil))
	doc, err := gltf.ParseGLTF(&buf)
	require.NoError(t, err)

	require.Len(t, doc.Nodes, 2, "the lead is written once")
	for _, node := range doc.Nodes {
		if node.Name == "body" {
			assert.Contains(t, node.Extensions, "EXT_mesh_gpu_instancing")
		} else {
			assert.NotContains(t, node.Extensions, "EXT_mesh_gpu_instancing")
		}
	}
}

func TestWrite_GpuInstancesOnAGroupWithNoMeshAreRefused(t *testing.T) {
	cube := primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads()
	group := &gltf.PolyformModel{
		Name:         "resistor",
		GpuInstances: threeInstances(),
		Children:     []*gltf.PolyformModel{{Name: "body", Mesh: &cube}},
	}

	var buf bytes.Buffer
	err := gltf.WriteText(gltf.PolyformScene{Models: []*gltf.PolyformModel{group}}, &buf, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"resistor"`)
	assert.Contains(t, err.Error(), "has GPU instances but no mesh")
}
