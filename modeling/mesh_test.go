package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestSetFloat3Attribute_EmptyArr_Clears(t *testing.T) {
	// ARRANGE ================================================================

	m := modeling.NewTriangleMesh([]int{0, 0, 0}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{vector3.New(0., 0., 0.)})

	// ACT ====================================================================
	newMesh := m.SetFloat3Attribute(modeling.PositionAttribute, nil)

	// ASSERT =================================================================
	assert.True(t, m.HasFloat3Attribute(modeling.PositionAttribute))
	assert.False(t, newMesh.HasFloat3Attribute(modeling.PositionAttribute))
}

func TestCopyFloat4FromMesh(t *testing.T) {
	// ARRANGE ================================================================

	dest := modeling.NewTriangleMesh([]int{0, 0, 0})
	src := modeling.NewTriangleMesh([]int{0, 0, 0}).
		SetFloat4Attribute(modeling.JointAttribute, []vector4.Float64{
			vector4.New(1., 2., 3., -1.),
			vector4.New(4., 5., 6., -1.),
			vector4.New(7., 8., 9., -1.),
		})

	// ACT ====================================================================
	newMesh := dest.CopyFloat4Attribute(src, modeling.JointAttribute)

	// ASSERT =================================================================
	assert.True(t, newMesh.HasFloat4Attribute(modeling.JointAttribute))
	assert.False(t, dest.HasFloat4Attribute(modeling.JointAttribute))
}

func TestCopyFloat3FromMesh(t *testing.T) {
	// ARRANGE ================================================================

	dest := modeling.NewTriangleMesh([]int{0, 0, 0})
	src := modeling.NewTriangleMesh([]int{0, 0, 0}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(1., 2., 3.),
			vector3.New(4., 5., 6.),
			vector3.New(7., 8., 9.),
		})

	// ACT ====================================================================
	newMesh := dest.CopyFloat3Attribute(src, modeling.PositionAttribute)

	// ASSERT =================================================================
	assert.True(t, newMesh.HasFloat3Attribute(modeling.PositionAttribute))
	assert.False(t, dest.HasFloat3Attribute(modeling.PositionAttribute))
}

func TestCopyFloat2FromMesh(t *testing.T) {
	// ARRANGE ================================================================

	dest := modeling.NewTriangleMesh([]int{0, 0, 0})
	src := modeling.NewTriangleMesh([]int{0, 0, 0}).
		SetFloat2Attribute(modeling.TexCoordAttribute, []vector2.Float64{
			vector2.New(1., 2.),
			vector2.New(4., 5.),
			vector2.New(7., 8.),
		})

	// ACT ====================================================================
	newMesh := dest.CopyFloat2Attribute(src, modeling.TexCoordAttribute)

	// ASSERT =================================================================
	assert.True(t, newMesh.HasFloat2Attribute(modeling.TexCoordAttribute))
	assert.False(t, dest.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestCopyFloat1FromMesh(t *testing.T) {
	// ARRANGE ================================================================

	dest := modeling.NewTriangleMesh([]int{0, 0, 0})
	src := modeling.NewTriangleMesh([]int{0, 0, 0}).
		SetFloat1Attribute(modeling.TexCoordAttribute, []float64{1, 2, 3})

	// ACT ====================================================================
	newMesh := dest.CopyFloat1Attribute(src, modeling.TexCoordAttribute)

	// ASSERT =================================================================
	assert.True(t, newMesh.HasFloat1Attribute(modeling.TexCoordAttribute))
	assert.False(t, dest.HasFloat1Attribute(modeling.TexCoordAttribute))
}

func TestScanFloat3AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 10000
	values := make([]vector3.Float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = vector3.New[float64](float64(i), float64(i), float64(i))
	}
	mesh := modeling.NewPointCloud(
		map[string][]vector3.Float64{
			attribute: values,
		},
		nil,
		nil,
		nil,
	)

	readValues := make([]vector3.Float64, count)

	// ACT ====================================================================
	mesh.ScanFloat3AttributeParallel(attribute, func(i int, v vector3.Float64) {
		readValues[i] = v
	})

	// ASSERT =================================================================

	for i := 0; i < count; i++ {
		assert.Equal(t, values[i], readValues[i])
	}
}

func TestScanFloat2AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 10000
	values := make([]vector2.Float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = vector2.New[float64](float64(i), float64(i))
	}
	mesh := modeling.NewPointCloud(
		nil,
		map[string][]vector2.Float64{
			attribute: values,
		},
		nil,
		nil,
	)

	readValues := make([]vector2.Float64, count)

	// ACT ====================================================================
	mesh.ScanFloat2AttributeParallel(attribute, func(i int, v vector2.Float64) {
		readValues[i] = v
	})

	// ASSERT =================================================================

	for i := 0; i < count; i++ {
		assert.Equal(t, values[i], readValues[i])
	}
}

func TestScanFloat1AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 10000
	values := make([]float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = float64(i)
	}
	mesh := modeling.NewPointCloud(
		nil,
		nil,
		map[string][]float64{
			attribute: values,
		},
		nil,
	)

	readValues := make([]float64, count)

	// ACT ====================================================================
	mesh.ScanFloat1AttributeParallel(attribute, func(i int, v float64) {
		readValues[i] = v
	})

	// ASSERT =================================================================

	for i := 0; i < count; i++ {
		assert.Equal(t, values[i], readValues[i])
	}
}

func TestModifyFloat3AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 1000
	values := make([]vector3.Float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = vector3.New(float64(i), float64(i), float64(i))
	}
	mesh := modeling.NewPointCloud(
		map[string][]vector3.Float64{
			attribute: values,
		},
		nil,
		nil,
		nil,
	)

	readValues := make([]vector3.Float64, count)

	// ACT ====================================================================
	mesh.
		ModifyFloat3AttributeParallel(attribute, func(i int, v vector3.Float64) vector3.Float64 {
			return v.Add(vector3.New(float64(i), float64(i), float64(i)))
		}).
		ScanFloat3AttributeParallel(attribute, func(i int, v vector3.Float64) {
			readValues[i] = v
		})

	// ASSERT =================================================================
	for i := 0; i < count; i++ {
		assert.Equal(
			t,
			values[i].Add(vector3.New(float64(i), float64(i), float64(i))),
			readValues[i],
		)
	}
}

func TestModifyFloat2AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 1000
	values := make([]vector2.Float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = vector2.New[float64](float64(i), float64(i))
	}
	mesh := modeling.NewPointCloud(
		nil,
		map[string][]vector2.Float64{
			attribute: values,
		},
		nil,
		nil,
	)

	readValues := make([]vector2.Float64, count)

	// ACT ====================================================================
	mesh.
		ModifyFloat2AttributeParallel(attribute, func(i int, v vector2.Float64) vector2.Float64 {
			return v.Add(vector2.New[float64](float64(i), float64(i)))
		}).
		ScanFloat2AttributeParallel(attribute, func(i int, v vector2.Float64) {
			readValues[i] = v
		})

	// ASSERT =================================================================
	for i := 0; i < count; i++ {
		assert.Equal(
			t,
			values[i].Add(vector2.New[float64](float64(i), float64(i))),
			readValues[i],
		)
	}
}

func TestModifyFloat1AttributeParallel(t *testing.T) {
	// ARRANGE ================================================================
	count := 1000
	values := make([]float64, count)
	attribute := "random-atr"
	for i := 0; i < count; i++ {
		values[i] = float64(i)
	}
	mesh := modeling.NewPointCloud(
		nil,
		nil,
		map[string][]float64{
			attribute: values,
		},
		nil,
	)

	readValues := make([]float64, count)

	// ACT ====================================================================
	mesh.
		ModifyFloat1AttributeParallel(attribute, func(i int, v float64) float64 {
			return v + float64(i)
		}).
		ScanFloat1AttributeParallel(attribute, func(i int, v float64) {
			readValues[i] = v
		})

	// ASSERT =================================================================
	for i := 0; i < count; i++ {
		assert.Equal(
			t,
			values[i]+float64(i),
			readValues[i],
		)
	}
}

func TestClearAttributeData(t *testing.T) {
	originalMesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat1Attribute("attr-1", []float64{1, 1, 1}).
		SetFloat2Attribute("attr-2", []vector2.Float64{vector2.One[float64](), vector2.One[float64](), vector2.One[float64]()}).
		SetFloat3Attribute("attr-3", []vector3.Float64{vector3.One[float64](), vector3.One[float64](), vector3.One[float64]()}).
		SetFloat4Attribute("attr-4", []vector4.Float64{vector4.One[float64](), vector4.One[float64](), vector4.One[float64]()})

	newMesh := originalMesh.ClearAttributeData()

	assert.False(t, newMesh.HasFloat1Attribute("attr-1"))
	assert.False(t, newMesh.HasFloat2Attribute("attr-2"))
	assert.False(t, newMesh.HasFloat3Attribute("attr-3"))
	assert.False(t, newMesh.HasFloat4Attribute("attr-4"))
}

func TestHasAttribute(t *testing.T) {
	v1Mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat1Attribute("attr-1", []float64{1, 1, 1})
	v2Mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat2Attribute("attr-2", []vector2.Float64{vector2.One[float64](), vector2.One[float64](), vector2.One[float64]()})
	v3Mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute("attr-3", []vector3.Float64{vector3.One[float64](), vector3.One[float64](), vector3.One[float64]()})
	v4Mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat4Attribute("attr-4", []vector4.Float64{vector4.One[float64](), vector4.One[float64](), vector4.One[float64]()})

	assert.True(t, v1Mesh.HasVertexAttribute("attr-1"))
	assert.False(t, v1Mesh.HasVertexAttribute("fake"))

	assert.True(t, v2Mesh.HasVertexAttribute("attr-2"))
	assert.False(t, v2Mesh.HasVertexAttribute("fake"))

	assert.True(t, v3Mesh.HasVertexAttribute("attr-3"))
	assert.False(t, v3Mesh.HasVertexAttribute("fake"))

	assert.True(t, v4Mesh.HasVertexAttribute("attr-4"))
	assert.False(t, v4Mesh.HasVertexAttribute("fake"))
}

func contains(m map[int]struct{}, i int) bool {
	_, ok := m[i]
	return ok
}

func TestVertexLUT_Triangle(t *testing.T) {
	lut := modeling.NewTriangleMesh([]int{
		0, 1, 2,
		2, 3, 4,
	}).VertexNeighborTable()

	neighbor0 := lut.Lookup(0)
	assert.True(t, contains(neighbor0, 1))
	assert.True(t, contains(neighbor0, 2))
	assert.False(t, contains(neighbor0, 3))
	assert.False(t, contains(neighbor0, 4))

	neighbor1 := lut.Lookup(1)
	assert.True(t, contains(neighbor1, 0))
	assert.True(t, contains(neighbor1, 2))
	assert.False(t, contains(neighbor1, 3))
	assert.False(t, contains(neighbor1, 4))

	neighbor2 := lut.Lookup(2)
	assert.True(t, contains(neighbor2, 0))
	assert.True(t, contains(neighbor2, 1))
	assert.True(t, contains(neighbor2, 3))
	assert.True(t, contains(neighbor2, 4))

	neighbor3 := lut.Lookup(3)
	assert.False(t, contains(neighbor3, 0))
	assert.False(t, contains(neighbor3, 1))
	assert.True(t, contains(neighbor3, 2))
	assert.True(t, contains(neighbor3, 4))

	neighbor4 := lut.Lookup(4)
	assert.False(t, contains(neighbor4, 0))
	assert.False(t, contains(neighbor4, 1))
	assert.True(t, contains(neighbor4, 2))
	assert.True(t, contains(neighbor4, 3))
}
