package mcp

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPreviewWarnsAboutAndShowsZeroNormals(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).SetFloat3Data(map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.)},
		modeling.NormalAttribute:   {vector3.New(0., 0., 1.), vector3.New(0., 0., 0.), vector3.New(0., 0., 1.)},
	})
	glass := &gltf.PolyformModel{Name: "Glazing", Mesh: &mesh}

	scene, err := buildScene([]*gltf.PolyformModel{{Name: "Helicopter", Children: []*gltf.PolyformModel{glass}}}, modelRefs{}, modelRefs{}, "Node-1", 0)
	require.NoError(t, err)

	assert.Equal(t, []string{"Helicopter/Glazing: 1 of 3 vertex normals are zero or not finite"}, scene.warnings)
	require.Len(t, scene.triangles, 1)
	assert.Equal(t, fauxgl.Vector{}, scene.triangles[0].V2.Normal, "a viewer has no direction to shade it with either")
}
