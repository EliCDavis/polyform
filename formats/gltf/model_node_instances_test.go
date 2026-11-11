package gltf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func modelWithInstances(instances nodes.Output[[]trs.TRS]) *gltf.PolyformModel {
	node := &nodes.Struct[gltf.ModelNode]{Data: gltf.ModelNode{
		Name:         nodes.ConstOutput[string]{Val: "part"},
		Mesh:         nodes.ConstOutput[modeling.Mesh]{Val: primitives.UnitCube()},
		GpuInstances: instances,
	}}
	return nodes.GetNodeOutputPort[*gltf.PolyformModel](node, "Out").Value()
}

func meshNodesIn(t *testing.T, model *gltf.PolyformModel) int {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, gltf.WriteText(gltf.PolyformScene{Models: []*gltf.PolyformModel{model}}, &buf, nil))
	doc, err := gltf.ParseGLTF(&buf)
	require.NoError(t, err)
	drawn := 0
	for _, node := range doc.Nodes {
		if node.Mesh != nil {
			drawn++
		}
	}
	return drawn
}

func TestModelNodeWiredToAnEmptyInstanceListDrawsNothing(t *testing.T) {
	for name, instances := range map[string][]trs.TRS{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			model := modelWithInstances(nodes.ConstOutput[[]trs.TRS]{Val: instances})
			assert.NotNil(t, model.GpuInstances, "a wired list is never nil, so the writer can tell it from an unwired port")
			assert.Zero(t, meshNodesIn(t, model))
		})
	}
}

func TestModelNodeWithNoInstanceListDrawsOnce(t *testing.T) {
	assert.Equal(t, 1, meshNodesIn(t, modelWithInstances(nil)))
}
