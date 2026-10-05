package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTRSKeepsEveryCopyWhenOneIsScaledToNothing(t *testing.T) {
	parents := []trs.TRS{trs.Position(vector3.New(10., 0., 0.))}
	slots := []trs.TRS{
		trs.Position(vector3.New(1., 0., 0.)),
		trs.New(vector3.New(2., 0., 0.), quaternion.Identity(), vector3.Zero[float64]()),
		trs.Position(vector3.New(3., 0., 0.)),
	}

	result, err := repeat.TRS(parents, slots)
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, vector3.New(11., 0., 0.), result[0].Position())
	assert.Equal(t, vector3.New(12., 0., 0.), result[1].Position())
	assert.Equal(t, vector3.Zero[float64](), result[1].Scale(), "the hidden slot stays hidden")
	assert.Equal(t, vector3.New(13., 0., 0.), result[2].Position())
}
