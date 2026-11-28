package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestPlaceOnGroundLiftsTheLowestPointToZero(t *testing.T) {
	sunk := meshops.TranslateAttribute3D(primitives.Cube{Width: 1, Height: 2, Depth: 1}.UnweldedQuads(), modeling.PositionAttribute, vector3.New(3., -5., 1.))
	box := meshops.PlaceOnGround(sunk).BoundingBox(modeling.PositionAttribute)
	assert.InDelta(t, 0, box.Min().Distance(vector3.New(2.5, 0., 0.5)), 1e-9)
	assert.InDelta(t, 0, box.Max().Distance(vector3.New(3.5, 2., 1.5)), 1e-9)

	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	assert.Equal(t, empty, meshops.PlaceOnGround(empty))
}
