package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCropFloat3AttributeKeepsContainedPointsInOrder(t *testing.T) {
	m := numberedPointCloud(100)
	box := geometry.NewAABBFromPoints(vector3.New(10., 0, 0), vector3.New(20.5, 100, 100))

	cropped := meshops.CropFloat3Attribute(m, modeling.PositionAttribute, box)

	require.Equal(t, 11, cropped.AttributeLength())
	for k := range 11 {
		f := float64(10 + k)
		assert.Equal(t, vector3.New(f, f+0.1, f+0.2), cropped.Float3Attribute(modeling.PositionAttribute).At(k))
		assert.Equal(t, f, cropped.Float1Attribute("v1").At(k))
		assert.Equal(t, f, cropped.Float4Attribute("v4").At(k).X())
		assert.Equal(t, f, cropped.Float2Attribute("v2").At(k).X())
	}
}

func TestCropFloat3AttributeOutsideEverythingIsEmpty(t *testing.T) {
	box := geometry.NewAABBFromPoints(vector3.New(-10., -10, -10), vector3.New(-5., -5, -5))

	cropped := meshops.CropFloat3Attribute(numberedPointCloud(10), modeling.PositionAttribute, box)

	assert.Equal(t, 0, cropped.AttributeLength())
	assert.Equal(t, modeling.PointTopology, cropped.Topology())
}
