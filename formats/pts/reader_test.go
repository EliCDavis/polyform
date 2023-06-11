package pts_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/pts"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestPtsOnlyPositionData(t *testing.T) {
	ptsData := `4
0 0 0
1 0 0
0 1 0
0 0 1
`
	pointCloud, err := pts.ReadPointCloud(strings.NewReader(ptsData))

	assert.NoError(t, err)
	assert.NotNil(t, pointCloud)
	assert.Equal(t, modeling.PointTopology, pointCloud.Topology())

	posData := pointCloud.Float3Attribute(modeling.PositionAttribute)
	if assert.Equal(t, posData.Len(), 4) {
		assert.Equal(t, vector3.New[float64](0, 0, 0), posData.At(0))
		assert.Equal(t, vector3.New[float64](1, 0, 0), posData.At(1))
		assert.Equal(t, vector3.New[float64](0, 1, 0), posData.At(2))
		assert.Equal(t, vector3.New[float64](0, 0, 1), posData.At(3))
	}
}
