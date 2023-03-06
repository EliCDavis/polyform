package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSetFloat3Attribute_EmptyArr_Clears(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 0, 0,
		},
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0., 0.),
			},
		},
		nil,
		nil,
		nil,
	)

	// ACT ====================================================================
	newMesh := m.SetFloat3Attribute(modeling.PositionAttribute, nil)

	// ASSERT =================================================================
	assert.True(t, m.HasFloat3Attribute(modeling.PositionAttribute))
	assert.False(t, newMesh.HasFloat3Attribute(modeling.PositionAttribute))
}

func Test_SplitOnUniqueMaterials_Simple(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
			3, 4, 5,
		},
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New[float64](0., 0., 0.),
				vector3.New[float64](0., 1., 0.),
				vector3.New[float64](1., 1., 0.),

				vector3.New[float64](0., 0., 0.),
				vector3.New[float64](1., 1., 0.),
				vector3.New[float64](1., 0., 0.),
			},
			modeling.NormalAttribute: {
				vector3.New[float64](0., 0., 0.),
				vector3.New[float64](0., 1., 0.),
				vector3.New[float64](1., 1., 0.),

				vector3.New[float64](0., 0., 0.),
				vector3.New[float64](1., 1., 0.),
				vector3.New[float64](1., 0., 0.),
			},
		},
		map[string][]vector2.Float64{
			modeling.TexCoordAttribute: {
				vector2.New[float64](0., 0.),
				vector2.New[float64](0., 1.),
				vector2.New[float64](1., 1.),

				vector2.New[float64](0., 0.),
				vector2.New[float64](1., 1.),
				vector2.New[float64](1., 0.),
			},
		},
		nil,
		[]modeling.MeshMaterial{
			{
				PrimitiveCount: 1,
				Material: &modeling.Material{
					Name: "red",
				},
			},
			{
				PrimitiveCount: 1,
				Material: &modeling.Material{
					Name: "blue",
				},
			},
		},
	)

	// ACT ====================================================================
	meshes := m.SplitOnUniqueMaterials()

	// ASSERT =================================================================
	if !assert.Len(t, meshes, 2) {
		return
	}

	v1 := meshes[0].View()
	if assert.Len(t, v1.Indices, 3) {
		assert.Equal(t, 1, v1.Indices[1])
		assert.Equal(t, 0, v1.Indices[0])
		assert.Equal(t, 2, v1.Indices[2])
	}

	v1Verts := v1.Float3Data[modeling.PositionAttribute]
	if assert.Len(t, v1Verts, 3) {
		assert.Equal(t, vector3.New[float64](0, 0, 0), v1Verts[0])
		assert.Equal(t, vector3.New[float64](0, 1, 0), v1Verts[1])
		assert.Equal(t, vector3.New[float64](1, 1, 0), v1Verts[2])
	}

	v1UVs := v1.Float2Data[modeling.TexCoordAttribute]
	if assert.Len(t, v1UVs, 3) {
		assert.Equal(t, vector2.New[float64](0, 0), v1UVs[0])
		assert.Equal(t, vector2.New[float64](0, 1), v1UVs[1])
		assert.Equal(t, vector2.New[float64](1, 1), v1UVs[2])
	}

	v2 := meshes[1].View()
	if assert.Len(t, v2.Indices, 3) {
		assert.Equal(t, 0, v2.Indices[0])
		assert.Equal(t, 1, v2.Indices[1])
		assert.Equal(t, 2, v2.Indices[2])
	}

	v2Verts := v2.Float3Data[modeling.PositionAttribute]
	if assert.Len(t, v2Verts, 3) {
		assert.Equal(t, vector3.New[float64](0, 0, 0), v2Verts[0])
		assert.Equal(t, vector3.New[float64](1, 1, 0), v2Verts[1])
		assert.Equal(t, vector3.New[float64](1, 0, 0), v2Verts[2])
	}

	v2UVs := v2.Float2Data[modeling.TexCoordAttribute]
	if assert.Len(t, v2UVs, 3) {
		assert.Equal(t, vector2.New[float64](0, 0), v2UVs[0])
		assert.Equal(t, vector2.New[float64](1, 1), v2UVs[1])
		assert.Equal(t, vector2.New[float64](1, 0), v2UVs[2])
	}
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
