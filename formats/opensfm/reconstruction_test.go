package opensfm_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/opensfm"
	"github.com/EliCDavis/polyform/modeling"
	opensfmFormat "github.com/EliCDavis/sfm/opensfm"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestReconstructionToPointcloud(t *testing.T) {
	// ARRANGE ================================================================
	reconstruction := opensfmFormat.ReconstructionSchema{
		Points: map[string]opensfmFormat.PointSchema{
			"1": {
				Color:       []float64{255, 0, 128},
				Coordinates: []float64{1, 2, 3},
			},
		},
	}

	// ACT ====================================================================
	pointcloud := opensfm.ReconstructionToPointcloud(reconstruction)

	// ASSERT =================================================================
	assert.Equal(t, modeling.PointTopology, pointcloud.Topology())

	indexData := pointcloud.Indices()
	assert.Equal(t, 1, indexData.Len())
	assert.Equal(t, 0, indexData.At(0))

	assert.True(t, pointcloud.HasFloat3Attribute(modeling.PositionAttribute))
	assert.True(t, pointcloud.HasFloat3Attribute(modeling.ColorAttribute))

	positionData := pointcloud.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 1, positionData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), positionData.At(0))

	colorData := pointcloud.Float3Attribute(modeling.ColorAttribute)
	assert.Equal(t, 1, colorData.Len())
	assert.Equal(t, 1., colorData.At(0).X())
	assert.Equal(t, 0., colorData.At(0).Y())
	assert.InDelta(t, .5, colorData.At(0).Z(), 0.01)
}
