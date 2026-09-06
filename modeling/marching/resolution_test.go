package marching_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reportErrors pulls the errors a node captured while producing an output.
func reportErrors(port nodes.Output[modeling.Mesh]) []string {
	observable, ok := port.(nodes.ObservableExecution)
	if !ok {
		return nil
	}
	return observable.ExecutionReport().Errors
}

func marchSphere(t *testing.T, resolution float64, domainSize float64) *nodes.Struct[marching.MarchNode] {
	t.Helper()

	field := nodes.GetNodeOutputPort[sample.Vec3ToFloat](&nodes.Struct[sdf.SphereNode]{
		Data: sdf.SphereNode{
			Position: nodes.ConstOutput[vector3.Float64]{Val: vector3.Zero[float64]()},
			Radius:   nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Field")

	return &nodes.Struct[marching.MarchNode]{
		Data: marching.MarchNode{
			Field:      field,
			Resolution: nodes.ConstOutput[float64]{Val: resolution},
			Domain: nodes.ConstOutput[geometry.AABB]{
				Val: geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(domainSize)),
			},
		},
	}
}

func TestMarchResolutionDirection(t *testing.T) {
	coarse := nodes.GetNodeOutputPort[modeling.Mesh](marchSphere(t, 5, 4), "Mesh").Value()
	fine := nodes.GetNodeOutputPort[modeling.Mesh](marchSphere(t, 20, 4), "Mesh").Value()

	require.Greater(t, coarse.PrimitiveCount(), 0, "resolution 5 over a 4 unit domain should produce geometry")
	assert.Greater(t, fine.PrimitiveCount(), coarse.PrimitiveCount(),
		"raising Resolution must give more triangles, not fewer")
}

func TestMarchReportsEmptyMeshFromTooLowResolution(t *testing.T) {
	// Resolution 0.05 over a ~4 unit domain: voxels 20 units across.
	port := nodes.GetNodeOutputPort[modeling.Mesh](marchSphere(t, 0.05, 4), "Mesh")

	mesh := port.Value()
	require.Equal(t, 0, mesh.PrimitiveCount(), "sanity: this configuration really does produce nothing")

	errs := reportErrors(port)
	require.NotEmpty(t, errs, "an empty march should report why, not succeed silently")
	assert.Contains(t, errs[0], "empty mesh")
	assert.Contains(t, errs[0], "Resolution", "the message should name the input that caused it")
}

func TestMarchReportsEmptyMeshFromMissedDomain(t *testing.T) {
	field := nodes.GetNodeOutputPort[sample.Vec3ToFloat](&nodes.Struct[sdf.SphereNode]{
		Data: sdf.SphereNode{
			Position: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(100., 100., 100.)},
			Radius:   nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Field")

	port := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[marching.MarchNode]{
		Data: marching.MarchNode{
			Field:      field,
			Resolution: nodes.ConstOutput[float64]{Val: 10},
			Domain: nodes.ConstOutput[geometry.AABB]{
				Val: geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(2.)),
			},
		},
	}, "Mesh")

	require.Equal(t, 0, port.Value().PrimitiveCount())

	errs := reportErrors(port)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0], "domain")
}

func TestMarchHealthyResultReportsNoError(t *testing.T) {
	port := nodes.GetNodeOutputPort[modeling.Mesh](marchSphere(t, 20, 4), "Mesh")
	require.Greater(t, port.Value().PrimitiveCount(), 0)

	assert.Empty(t, reportErrors(port),
		"a march that produced geometry should report nothing")
}
