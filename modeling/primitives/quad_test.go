package primitives_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuadNodeWithNoUVsConnected(t *testing.T) {
	node := &nodes.Struct[primitives.QuadNode]{}

	var mesh modeling.Mesh
	require.NotPanics(t, func() {
		mesh = nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()
	})
	assert.Equal(t, 2, mesh.PrimitiveCount())
	assert.False(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute))
}
