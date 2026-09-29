package meshops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestProjectUVAlongZMapsXToUAndYToV(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 5.),
			vector3.New(0.5, 0., -5.),
			vector3.New(0., 2., 0.),
		})

	uv := meshops.ProjectUV(mesh, vector3.Forward[float64](), vector2.New(0.25, 2.), vector2.Zero[float64](), 0).
		Float2Attribute(modeling.TexCoordAttribute)

	assert.Equal(t, vector2.New(0., 0.), uv.At(0), "depth does not affect a Z projection")
	assert.Equal(t, vector2.New(2., 0.), uv.At(1), "x = 0.5 over a 0.25 repeat")
	assert.Equal(t, vector2.New(0., 1.), uv.At(2), "y = 2 over a 2 repeat")
}

func TestProjectUVLeavesAMeshWithoutPositionsAlone(t *testing.T) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	result := meshops.ProjectUV(empty, vector3.Up[float64](), vector2.One[float64](), vector2.Zero[float64](), 0)
	assert.False(t, result.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestProjectUVRotationSwapsTheAxes(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(0., 1., 0.),
		})

	uv := meshops.ProjectUV(mesh, vector3.Forward[float64](), vector2.One[float64](), vector2.Zero[float64](), math.Pi/2).
		Float2Attribute(modeling.TexCoordAttribute)

	assert.InDelta(t, 0, uv.At(1).X(), 1e-9, "world X no longer drives u")
	assert.InDelta(t, -1, uv.At(1).Y(), 1e-9, "it drives v instead")
	assert.InDelta(t, 1, uv.At(2).X(), 1e-9, "world Y drives u after the turn")
	assert.InDelta(t, 0, uv.At(2).Y(), 1e-9)
}
