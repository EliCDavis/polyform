package gausops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bigSplat(n int) modeling.Mesh {
	v3 := map[string][]vector3.Float64{
		modeling.PositionAttribute: make([]vector3.Float64, n),
		modeling.ScaleAttribute:    make([]vector3.Float64, n),
		modeling.FDCAttribute:      make([]vector3.Float64, n),
	}
	for j := range 15 {
		v3[modeling.SHAttribute(j)] = make([]vector3.Float64, n)
	}
	rotation := make([]vector4.Float64, n)
	opacity := make([]float64, n)
	for i := range n {
		f := float64(i)
		v3[modeling.PositionAttribute][i] = vector3.New(math.Sin(f), math.Cos(f*0.7), f*1e-5)
		v3[modeling.ScaleAttribute][i] = vector3.New(-5+math.Sin(f), -5+math.Cos(f), -5)
		v3[modeling.FDCAttribute][i] = vector3.New(f*1e-6, 0.5, -0.5)
		for j := range 15 {
			v3[modeling.SHAttribute(j)][i] = vector3.New(math.Sin(f+float64(j)), float64(j)*0.01, -f*1e-7)
		}
		rotation[i] = vector4.New(1., 0.1, 0.2, 0.3).Normalized()
		opacity[i] = -5 + 10*float64(i%97)/97
	}
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{modeling.RotationAttribute: rotation},
		v3,
		nil,
		map[string][]float64{modeling.OpacityAttribute: opacity},
	)
}

func TestFilterByOpacityKeepsTheRangeInOrder(t *testing.T) {
	m := modeling.NewPointCloud(
		nil,
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {vector3.New(0., 0, 0), vector3.New(1., 0, 0), vector3.New(2., 0, 0), vector3.New(3., 0, 0)},
			modeling.ScaleAttribute:    {vector3.Zero[float64](), vector3.Zero[float64](), vector3.Zero[float64](), vector3.Zero[float64]()},
		},
		nil,
		map[string][]float64{modeling.OpacityAttribute: {-3, 0.5, 2, -1}},
	)

	filtered := gausops.Filter(m, -2, 1, 0, 0)

	require.Equal(t, 2, filtered.AttributeLength())
	assert.Equal(t, 1., filtered.Float3Attribute(modeling.PositionAttribute).At(0).X())
	assert.Equal(t, 3., filtered.Float3Attribute(modeling.PositionAttribute).At(1).X())
	assert.Equal(t, []int{0, 1}, []int{filtered.Indices().At(0), filtered.Indices().At(1)})
}

func TestFilterByVolumeUsesTheProductOfExpScales(t *testing.T) {
	m := modeling.NewPointCloud(
		nil,
		map[string][]vector3.Float64{
			modeling.ScaleAttribute: {vector3.Zero[float64](), vector3.Fill(math.Log(2)), vector3.Fill(-10.)},
		},
		nil,
		map[string][]float64{modeling.OpacityAttribute: {0, 0, 0}},
	)

	filtered := gausops.Filter(m, 0, 0, 0.5, 7.9)

	require.Equal(t, 1, filtered.AttributeLength())
	assert.Equal(t, vector3.Zero[float64](), filtered.Float3Attribute(modeling.ScaleAttribute).At(0))
}

func TestFilterWithNoRangesKeepsEverything(t *testing.T) {
	m := bigSplat(1000)

	filtered := gausops.Filter(m, 0, 0, 1, 1)

	assert.Equal(t, 1000, filtered.AttributeLength())
	assert.Equal(t, m.Float3Attribute(modeling.SHAttribute(14)).At(999), filtered.Float3Attribute(modeling.SHAttribute(14)).At(999))
}

func BenchmarkFilter(b *testing.B) {
	m := bigSplat(1_000_000)
	b.ResetTimer()
	for range b.N {
		gausops.Filter(m, -2.944, 20, 0, 1e-5)
	}
}

func BenchmarkRotate(b *testing.B) {
	m := bigSplat(1_000_000)
	q := quaternion.New(vector3.New(0.93, 0, -0.11), 0.35)
	b.ResetTimer()
	for range b.N {
		gausops.Rotate(m, q)
	}
}
