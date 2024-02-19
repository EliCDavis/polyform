package splat_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestWrite_ErrorOnNonPointcloud(t *testing.T) {
	err := splat.Write(nil, modeling.EmptyMesh(modeling.TriangleTopology))
	assert.EqualError(t, err, "mesh must be point topology, was instead triangle")
}

func TestWrite_ErrorOnMissingAttributes(t *testing.T) {
	err := splat.Write(nil, modeling.EmptyMesh(modeling.PointTopology))
	assert.EqualError(t, err, "required attribute not present on mesh: Position")
}

func vector3InDelta(t *testing.T, a, b vector3.Float64, delta float64) {
	assert.InDelta(t, a.X(), b.X(), delta)
	assert.InDelta(t, a.Y(), b.Y(), delta)
	assert.InDelta(t, a.Z(), b.Z(), delta)
}

func vector4InDelta(t *testing.T, a, b vector4.Float64, delta float64) {
	assert.InDelta(t, a.X(), b.X(), delta)
	assert.InDelta(t, a.Y(), b.Y(), delta)
	assert.InDelta(t, a.Z(), b.Z(), delta)
	assert.InDelta(t, a.W(), b.W(), delta)
}

func TestReadWrite(t *testing.T) {
	in := modeling.NewPointCloud(
		map[string][]vector4.Vector[float64]{
			modeling.RotationAttribute: []vector4.Float64{
				vector4.New(0., .2, .4, .8),
			},
		},
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: []vector3.Float64{
				vector3.New(0., 1., 2.),
			},
			modeling.ScaleAttribute: []vector3.Float64{
				vector3.New(0., 1., 2.),
			},
			modeling.FDCAttribute: []vector3.Float64{
				vector3.New(0., 0.5, 1),
			},
		},
		nil,
		map[string][]float64{
			modeling.OpacityAttribute: {
				.5,
			},
		},
		nil,
	)

	buf := &bytes.Buffer{}
	err := splat.Write(buf, in)
	assert.NoError(t, err)

	out, err := splat.Read(bytes.NewReader(buf.Bytes()))
	assert.NoError(t, err)
	vector4InDelta(
		t,
		in.Float4Attribute(modeling.RotationAttribute).At(0),
		out.Float4Attribute(modeling.RotationAttribute).At(0),
		.01,
	)

	vector3InDelta(
		t,
		in.Float3Attribute(modeling.PositionAttribute).At(0),
		out.Float3Attribute(modeling.PositionAttribute).At(0),
		.01,
	)

	vector3InDelta(
		t,
		in.Float3Attribute(modeling.ScaleAttribute).At(0),
		out.Float3Attribute(modeling.ScaleAttribute).At(0),
		.001,
	)

	vector3InDelta(
		t,
		in.Float3Attribute(modeling.FDCAttribute).At(0),
		out.Float3Attribute(modeling.FDCAttribute).At(0),
		.01,
	)

	assert.InDelta(
		t,
		in.Float1Attribute(modeling.OpacityAttribute).At(0),
		out.Float1Attribute(modeling.OpacityAttribute).At(0),
		.02,
	)
}
