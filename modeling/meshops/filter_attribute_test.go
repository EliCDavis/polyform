package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestFilterFloat1(t *testing.T) {
	mesh := modeling.NewPointCloud(nil, nil, nil, map[string][]float64{
		"test": {1, 2, 3},
	}, nil)

	transformer := meshops.FilterFloat1Transformer{
		Attribute: "test",
		Filter: func(v float64) bool {
			return v == 2
		},
	}

	transformed, err := transformer.Transform(mesh)
	assert.NoError(t, err)
	assert.Equal(t, 1, transformed.AttributeLength())
	assert.Equal(t, 1, transformed.PrimitiveCount())
	assert.Equal(t, 2., transformed.Float1Attribute("test").At(0))
	assert.Equal(t, 0, transformed.Indices().At(0))
}

func TestFilterFloat2(t *testing.T) {
	mesh := modeling.NewPointCloud(nil, nil, map[string][]vector2.Vector[float64]{
		"test": {
			vector2.Up[float64]().Scale(1),
			vector2.Up[float64]().Scale(2),
			vector2.Up[float64]().Scale(3),
		},
	}, nil, nil)

	transformer := meshops.FilterFloat2Transformer{
		Attribute: "test",
		Filter: func(v vector2.Float64) bool {
			return v.Length() == 2
		},
	}

	transformed, err := transformer.Transform(mesh)
	assert.NoError(t, err)
	assert.Equal(t, 1, transformed.AttributeLength())
	assert.Equal(t, 1, transformed.PrimitiveCount())
	assert.Equal(t, vector2.Up[float64]().Scale(2), transformed.Float2Attribute("test").At(0))
	assert.Equal(t, 0, transformed.Indices().At(0))
}

func TestFilterFloat3(t *testing.T) {
	mesh := modeling.NewPointCloud(nil, map[string][]vector3.Vector[float64]{
		"test": {
			vector3.Up[float64]().Scale(1),
			vector3.Up[float64]().Scale(2),
			vector3.Up[float64]().Scale(3),
		},
	}, nil, nil, nil)

	transformer := meshops.FilterFloat3Transformer{
		Attribute: "test",
		Filter: func(v vector3.Float64) bool {
			return v.Length() == 2
		},
	}

	transformed, err := transformer.Transform(mesh)
	assert.NoError(t, err)
	assert.Equal(t, 1, transformed.AttributeLength())
	assert.Equal(t, 1, transformed.PrimitiveCount())
	assert.Equal(t, vector3.Up[float64]().Scale(2), transformed.Float3Attribute("test").At(0))
	assert.Equal(t, 0, transformed.Indices().At(0))
}

func TestFilterFloat4(t *testing.T) {
	mesh := modeling.NewPointCloud(map[string][]vector4.Vector[float64]{
		"test": {
			vector4.New[float64](1, 0, 0, 0),
			vector4.New[float64](2, 0, 0, 0),
			vector4.New[float64](3, 0, 0, 0),
		},
	}, nil, nil, nil, nil)

	transformer := meshops.FilterFloat4Transformer{
		Attribute: "test",
		Filter: func(v vector4.Float64) bool {
			return v.Length() == 2
		},
	}

	transformed, err := transformer.Transform(mesh)
	assert.NoError(t, err)
	assert.Equal(t, 1, transformed.AttributeLength())
	assert.Equal(t, 1, transformed.PrimitiveCount())
	assert.Equal(t, vector4.New[float64](2, 0, 0, 0), transformed.Float4Attribute("test").At(0))
	assert.Equal(t, 0, transformed.Indices().At(0))
}
