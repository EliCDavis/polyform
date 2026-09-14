package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tetrahedron() modeling.Mesh {
	return modeling.NewTriangleMesh([]int{
		0, 2, 1,
		0, 1, 3,
		1, 2, 3,
		2, 0, 3,
	})
}

func TestBoundaryEdgesIsEmptyForAClosedSurface(t *testing.T) {
	assert.Empty(t, tetrahedron().BoundaryEdges())
	assert.Empty(t, tetrahedron().BoundaryLoops())
}

func TestBoundaryEdgesRimAHoleTheWayItsFacesWind(t *testing.T) {
	nicked := modeling.NewTriangleMesh([]int{
		0, 1, 3,
		1, 2, 3,
		2, 0, 3,
	})

	assert.Equal(t, [][2]int{{0, 1}, {1, 2}, {2, 0}}, nicked.BoundaryEdges())

	loops := nicked.BoundaryLoops()
	require.Len(t, loops, 1)
	assert.Equal(t, []int{0, 1, 2}, loops[0])
}

func TestBoundaryEdgesSeeWhatLoopsCannot(t *testing.T) {
	hinged := modeling.NewTriangleMesh([]int{0, 1, 2, 0, 1, 3, 0, 1, 4})

	assert.Empty(t, hinged.BoundaryLoops())
	assert.NotEmpty(t, hinged.BoundaryEdges())
}

func TestBoundaryLoopsSkipsAPinchedRim(t *testing.T) {
	// Two open triangles sharing only vertex 0: two rims through one vertex.
	pinched := modeling.NewTriangleMesh([]int{0, 1, 2, 0, 3, 4})

	assert.Empty(t, pinched.BoundaryLoops())
	assert.Len(t, pinched.BoundaryEdges(), 6)
}

func TestBoundaryEdgesIgnoreOtherTopologies(t *testing.T) {
	lines := modeling.NewMesh(modeling.LineTopology, []int{0, 1, 1, 2})

	assert.Nil(t, lines.BoundaryEdges())
	assert.Nil(t, lines.BoundaryLoops())
}
