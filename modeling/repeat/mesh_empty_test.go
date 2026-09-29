package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestMeshPassesAnEmptyMeshThrough(t *testing.T) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)

	result := repeat.Mesh(empty, []trs.TRS{trs.Identity(), trs.Position(vector3.New(1., 0., 0.))})

	assert.Zero(t, result.PrimitiveCount())
	assert.Zero(t, empty.Translate(vector3.Up[float64]()).PrimitiveCount())
	assert.Zero(t, empty.ApplyTRS(trs.Identity()).PrimitiveCount())
}
