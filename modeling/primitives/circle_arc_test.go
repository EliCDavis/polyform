package primitives_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircleWithNoAnglesIsAWholeDisc(t *testing.T) {
	disc := primitives.Circle{Sides: 8, Radius: 1}.ToMesh()

	assert.Equal(t, 8, disc.PrimitiveCount(), "one triangle per side, wrapping closed")
	assert.Equal(t, 9, disc.Float3Attribute(modeling.PositionAttribute).Len(), "8 on the rim plus the middle")
}

func TestCircleSectorStopsWhereItIsToldTo(t *testing.T) {
	quarter := primitives.Circle{
		Sides: 4, Radius: 1,
		StartAngle: 0, EndAngle: math.Pi / 2,
	}.ToMesh()

	assert.Equal(t, 4, quarter.PrimitiveCount(), "one triangle per segment, and no closing one")

	positions := quarter.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, 6, positions.Len(), "5 along the arc plus the middle")

	first := positions.At(0)
	assert.InDelta(t, 1, first.X(), 1e-9)
	assert.InDelta(t, 0, first.Z(), 1e-9)

	last := positions.At(4)
	assert.InDelta(t, 0, last.X(), 1e-9, "a quarter turn lands on +Z")
	assert.InDelta(t, 1, last.Z(), 1e-9)
}

func TestCircleSectorStartsWhereItIsToldTo(t *testing.T) {
	arc := primitives.Circle{
		Sides: 2, Radius: 1,
		StartAngle: math.Pi, EndAngle: math.Pi * 1.5,
	}.ToMesh()

	first := arc.Float3Attribute(modeling.PositionAttribute).At(0)
	assert.InDelta(t, -1, first.X(), 1e-9)
	assert.InDelta(t, 0, first.Z(), 1e-9)
}

func TestCircleAFullTurnApartIsStillAWholeDisc(t *testing.T) {
	disc := primitives.Circle{
		Sides: 8, Radius: 1,
		StartAngle: 0, EndAngle: 2 * math.Pi,
	}.ToMesh()

	assert.Equal(t, 8, disc.PrimitiveCount())
	assert.Equal(t, 9, disc.Float3Attribute(modeling.PositionAttribute).Len())
}
