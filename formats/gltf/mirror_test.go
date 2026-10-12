package gltf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nodeTransform(node gltf.Node) trs.TRS {
	x := trs.Identity()
	if node.Translation != nil {
		x = x.SetTranslation(vector3.FromArray(node.Translation[:]))
	}
	if node.Scale != nil {
		x = x.SetScale(vector3.FromArray(node.Scale[:]))
	}
	if node.Matrix != nil {
		decomposed, err := trs.FromMatrix(mat.FromColArray(*node.Matrix))
		if err == nil {
			x = decomposed
		}
	}
	return x
}

func TestWrite_MirroredModelOverInstancedChildrenKeepsItsNegativeScale(t *testing.T) {
	hull := primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads()
	greeble := primitives.Cube{Width: .1, Height: .1, Depth: .1}.UnweldedQuads()
	mirror := trs.Scale(vector3.New(-1., 1., 1.))

	for name, strategy := range map[string]gltf.WriterInstancingStrategy{
		"default":  gltf.WriterInstancingStrategy_Default,
		"collapse": gltf.WriterInstancingStrategy_Collapse,
		"expand":   gltf.WriterInstancingStrategy_Expand,
	} {
		t.Run(name, func(t *testing.T) {
			scene := gltf.PolyformScene{Models: []*gltf.PolyformModel{{
				Name: "mirror",
				Mesh: &hull,
				TRS:  &mirror,
				Children: []*gltf.PolyformModel{{
					Name:         "greebles",
					Mesh:         &greeble,
					GpuInstances: threeInstances(),
				}},
			}}}

			var buf bytes.Buffer
			require.NoError(t, gltf.WriteText(scene, &buf, &gltf.WriterOptions{GpuInstancingStrategy: strategy}))
			doc, err := gltf.ParseGLTF(&buf)
			require.NoError(t, err)

			found := false
			for _, node := range doc.Nodes {
				if node.Name != "mirror" {
					continue
				}
				found = true
				assert.Less(t, nodeTransform(node).Matrix().Determinant(), 0.)
			}
			assert.True(t, found, "mirror node was not written")
		})
	}
}

func TestWrite_MirroredInstancesExpandToMirroredNodes(t *testing.T) {
	greeble := primitives.Cube{Width: .1, Height: .1, Depth: .1}.UnweldedQuads()
	mirror := trs.Scale(vector3.New(-1., 1., 1.))
	scene := gltf.PolyformScene{Models: []*gltf.PolyformModel{{
		Name: "root",
		Children: []*gltf.PolyformModel{{
			Name:         "greebles",
			Mesh:         &greeble,
			TRS:          &mirror,
			GpuInstances: threeInstances(),
		}},
	}}}

	var buf bytes.Buffer
	require.NoError(t, gltf.WriteText(scene, &buf, &gltf.WriterOptions{GpuInstancingStrategy: gltf.WriterInstancingStrategy_Expand}))
	doc, err := gltf.ParseGLTF(&buf)
	require.NoError(t, err)

	expanded := 0
	for _, node := range doc.Nodes {
		if node.Name != "greebles" {
			continue
		}
		expanded++
		assert.Less(t, nodeTransform(node).Matrix().Determinant(), 0.)
	}
	assert.Equal(t, 3, expanded)
}
