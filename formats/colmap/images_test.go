package colmap_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/colmap"
	"github.com/EliCDavis/polyform/modeling"
	colmapFormat "github.com/EliCDavis/sfm/colmap"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestImagesToPointcloud(t *testing.T) {
	// ARRANGE ================================================================
	reconstruction := []colmapFormat.Image{
		{
			Id:          1,
			CameraId:    2,
			Name:        "Yeet",
			Translation: vector3.New(1., 2., 3.),
			Rotation:    vector4.New(1., 2., 3., 4.),
			Points:      make([]colmapFormat.ImagePoint, 10),
		},
	}

	// ACT ====================================================================
	pointcloud := colmap.ImageDataToPointCloud(reconstruction)

	// ASSERT =================================================================
	assert.Equal(t, modeling.PointTopology, pointcloud.Topology())

	indexData := pointcloud.Indices()
	assert.Equal(t, 1, indexData.Len())
	assert.Equal(t, 0, indexData.At(0))

	assert.True(t, pointcloud.HasFloat1Attribute("id"))
	idData := pointcloud.Float1Attribute("id")
	assert.Equal(t, 1, idData.Len())
	assert.Equal(t, 1., idData.At(0))

	assert.True(t, pointcloud.HasFloat1Attribute("point count"))
	pointCountData := pointcloud.Float1Attribute("point count")
	assert.Equal(t, 1, pointCountData.Len())
	assert.Equal(t, 10., pointCountData.At(0))

	assert.True(t, pointcloud.HasFloat1Attribute("camera id"))
	cameraIdData := pointcloud.Float1Attribute("camera id")
	assert.Equal(t, 1, cameraIdData.Len())
	assert.Equal(t, 2., cameraIdData.At(0))

	assert.True(t, pointcloud.HasFloat3Attribute(modeling.PositionAttribute))
	positionData := pointcloud.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 1, positionData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), positionData.At(0))

	assert.True(t, pointcloud.HasFloat4Attribute(modeling.RotationAttribute))
	rotData := pointcloud.Float4Attribute(modeling.RotationAttribute)
	assert.Equal(t, 1, rotData.Len())
	assert.Equal(t, vector4.New(1., 2., 3., 4.), rotData.At(0))
}
