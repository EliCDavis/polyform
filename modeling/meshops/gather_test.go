package meshops_test

import (
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func numberedPointCloud(n int) modeling.Mesh {
	v4 := make([]vector4.Float64, n)
	v3 := make([]vector3.Float64, n)
	v2 := make([]vector2.Float64, n)
	v1 := make([]float64, n)
	for i := range n {
		f := float64(i)
		v4[i] = vector4.New(f, f+0.1, f+0.2, f+0.3)
		v3[i] = vector3.New(f, f+0.1, f+0.2)
		v2[i] = vector2.New(f, f+0.1)
		v1[i] = f
	}
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{"v4": v4},
		map[string][]vector3.Float64{modeling.PositionAttribute: v3},
		map[string][]vector2.Float64{"v2": v2},
		map[string][]float64{"v1": v1},
	)
}

func TestGatherPointsKeepsOrderAcrossEveryAttributeKind(t *testing.T) {
	m := numberedPointCloud(100)

	gathered := meshops.GatherPoints(m, []int{42, 3, 99, 3})

	require.Equal(t, 4, gathered.AttributeLength())
	assert.Equal(t, 4, gathered.Indices().Len())
	for k, i := range []int{42, 3, 99, 3} {
		f := float64(i)
		assert.Equal(t, vector4.New(f, f+0.1, f+0.2, f+0.3), gathered.Float4Attribute("v4").At(k))
		assert.Equal(t, vector3.New(f, f+0.1, f+0.2), gathered.Float3Attribute(modeling.PositionAttribute).At(k))
		assert.Equal(t, vector2.New(f, f+0.1), gathered.Float2Attribute("v2").At(k))
		assert.Equal(t, f, gathered.Float1Attribute("v1").At(k))
		assert.Equal(t, k, gathered.Indices().At(k))
	}
}

func TestGatherPointsWithNothingKeptHasNoAttributes(t *testing.T) {
	gathered := meshops.GatherPoints(numberedPointCloud(10), nil)

	assert.Equal(t, 0, gathered.AttributeLength())
	assert.Empty(t, gathered.Float3Attributes())
	assert.Equal(t, modeling.PointTopology, gathered.Topology())
}

func TestGatherPointsRequiresPointTopology(t *testing.T) {
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{{}, {}, {}})

	assert.PanicsWithError(t, meshops.ErrRequirePointTopology.Error(), func() {
		meshops.GatherPoints(m, []int{0})
	})
}

func TestGatherPointsMatchesSequentialGatherOnLargeInput(t *testing.T) {
	m := numberedPointCloud(300_000)
	indices := meshops.SelectIndices(m.AttributeLength(), func(i int) bool { return i%7 != 0 })

	gathered := meshops.GatherPoints(m, indices)

	positions := gathered.Float3Attribute(modeling.PositionAttribute)
	v1 := gathered.Float1Attribute("v1")
	require.Equal(t, len(indices), positions.Len())
	for k, i := range indices {
		f := float64(i)
		if positions.At(k) != vector3.New(f, f+0.1, f+0.2) || v1.At(k) != f {
			t.Fatalf("index %d gathered wrong point", k)
		}
	}
}

func TestSelectIndicesReturnsAscendingMatches(t *testing.T) {
	assert.Equal(t, []int{1, 3, 5}, meshops.SelectIndices(6, func(i int) bool { return i%2 == 1 }))
	assert.Equal(t, []int{}, meshops.SelectIndices(6, func(int) bool { return false }))
	assert.Equal(t, []int{}, meshops.SelectIndices(0, func(int) bool { return true }))

	many := meshops.SelectIndices(250_000, func(i int) bool { return i%3 == 0 })
	require.Len(t, many, 83_334)
	for k, i := range many {
		if i != k*3 {
			t.Fatalf("entry %d is %d", k, i)
		}
	}
}

func benchmarkPointCloud(n int) modeling.Mesh {
	v3 := make(map[string][]vector3.Float64)
	for a := range 19 {
		data := make([]vector3.Float64, n)
		for i := range data {
			data[i] = vector3.New(float64(i%1000)*0.01, float64(a), float64(i))
		}
		v3[fmt.Sprintf("v3_%d", a)] = data
	}
	v4 := make([]vector4.Float64, n)
	v1 := make([]float64, n)
	for i := range n {
		v4[i] = vector4.New(1., 0, 0, 0)
		v1[i] = float64(i)
	}
	v3[modeling.PositionAttribute] = v3["v3_0"]
	delete(v3, "v3_0")
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{"v4": v4},
		v3,
		nil,
		map[string][]float64{"v1": v1},
	)
}

func BenchmarkCropFloat3Attribute(b *testing.B) {
	m := benchmarkPointCloud(1_000_000)
	bounds := m.BoundingBox(modeling.PositionAttribute)
	box := geometry.NewAABB(bounds.Center(), bounds.Size().Scale(0.8))
	b.ResetTimer()
	for range b.N {
		meshops.CropFloat3Attribute(m, modeling.PositionAttribute, box)
	}
}

func BenchmarkRemovedUnreferencedVertices(b *testing.B) {
	m := benchmarkPointCloud(1_000_000)
	kept := meshops.SelectIndices(m.AttributeLength(), func(i int) bool { return i%5 != 0 })
	m = m.SetIndices(kept)
	b.ResetTimer()
	for range b.N {
		meshops.RemovedUnreferencedVertices(m)
	}
}
