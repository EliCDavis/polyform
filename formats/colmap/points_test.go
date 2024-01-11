package colmap_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/colmap"
	"github.com/EliCDavis/polyform/modeling"
	colmapFormat "github.com/EliCDavis/sfm/colmap"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestReconstructionToPointcloud(t *testing.T) {
	// ARRANGE ================================================================
	reconstruction := []colmapFormat.Point3D{
		{
			Position: vector3.New(1., 2., 3.),
			Color:    color.RGBA{R: 255, G: 0, B: 128},
		},
	}

	// ACT ====================================================================
	pointcloud := colmap.PointDataToPointCloud(reconstruction)

	// ASSERT =================================================================
	assert.Equal(t, modeling.PointTopology, pointcloud.Topology())

	indexData := pointcloud.Indices()
	assert.Equal(t, 1, indexData.Len())
	assert.Equal(t, 0, indexData.At(0))

	assert.True(t, pointcloud.HasFloat3Attribute(modeling.PositionAttribute))
	positionData := pointcloud.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 1, positionData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), positionData.At(0))

	assert.True(t, pointcloud.HasFloat3Attribute(modeling.ColorAttribute))
	colorData := pointcloud.Float3Attribute(modeling.ColorAttribute)
	assert.Equal(t, 1, colorData.Len())
	assert.Equal(t, 1., colorData.At(0).X())
	assert.Equal(t, 0., colorData.At(0).Y())
	assert.InDelta(t, .5, colorData.At(0).Z(), 0.01)
}
