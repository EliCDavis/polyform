package modeling_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestApplyTRSNonUniformScaleKeepsNormalsOutward(t *testing.T) {
	squashed := primitives.UVSphere(1, 12, 16).ApplyTRS(trs.Scale(vector3.New(1., 0.2, 1.)))

	positions := squashed.Float3Attribute(modeling.PositionAttribute)
	normals := squashed.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < positions.Len(); i++ {
		n := normals.At(i)
		assert.InDeltaf(t, 1, n.Length(), 1e-9, "normal %d is not unit length", i)
		assert.Greaterf(t, n.Dot(positions.At(i)), 0., "normal %d points inward", i)
	}
}

func TestApplyTRSNonUniformScaleTiltsNormalsTowardTheSquashedAxis(t *testing.T) {
	squashed := primitives.UVSphere(1, 12, 16).ApplyTRS(trs.Scale(vector3.New(1., 0.2, 1.)))

	positions := squashed.Float3Attribute(modeling.PositionAttribute)
	normals := squashed.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < positions.Len(); i++ {
		p := positions.At(i)
		if math.Abs(p.Y()) < 1e-9 || p.XZ().Length() < 1e-9 {
			continue
		}
		radial := vector3.New(p.X(), 0., p.Z()).Normalized()
		assert.Lessf(t, normals.At(i).Dot(radial), p.Normalized().Dot(radial),
			"normal %d should lean further toward Y than the position does", i)
	}
}
