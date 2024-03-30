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

func TestRequireAttribute(t *testing.T) {
	mesh := modeling.NewMesh(modeling.TriangleTopology, []int{0}).
		SetFloat1Attribute("f1", []float64{0}).
		SetFloat2Attribute("f2", []vector2.Float64{vector2.Zero[float64]()}).
		SetFloat3Attribute("f3", []vector3.Float64{vector3.Zero[float64]()}).
		SetFloat4Attribute("f4", []vector4.Float64{vector4.Zero[float64]()})

	assert.NoError(t, meshops.RequireV1Attribute(mesh, "f1"))
	assert.NoError(t, meshops.RequireV2Attribute(mesh, "f2"))
	assert.NoError(t, meshops.RequireV3Attribute(mesh, "f3"))
	assert.NoError(t, meshops.RequireV4Attribute(mesh, "f4"))

	assert.EqualError(t, meshops.RequireV1Attribute(mesh, "v1"), "mesh is required to have the vector1 attribute: 'v1'")
	assert.EqualError(t, meshops.RequireV2Attribute(mesh, "v2"), "mesh is required to have the vector2 attribute: 'v2'")
	assert.EqualError(t, meshops.RequireV3Attribute(mesh, "v3"), "mesh is required to have the vector3 attribute: 'v3'")
	assert.EqualError(t, meshops.RequireV4Attribute(mesh, "v4"), "mesh is required to have the vector4 attribute: 'v4'")
}

func TestRequireTopology(t *testing.T) {
	assert.NoError(t, meshops.RequireTopology(modeling.NewMesh(modeling.TriangleTopology, nil), modeling.TriangleTopology))
	assert.ErrorIs(
		t,
		meshops.RequireTopology(modeling.NewMesh(modeling.PointTopology, nil), modeling.TriangleTopology),
		meshops.ErrRequireTriangleTopology,
	)

	assert.ErrorIs(
		t,
		meshops.RequireTopology(modeling.NewMesh(modeling.PointTopology, nil), modeling.LineTopology),
		meshops.ErrRequireLineTopology,
	)

	assert.ErrorIs(
		t,
		meshops.RequireTopology(modeling.NewMesh(modeling.PointTopology, nil), modeling.LineTopology),
		meshops.ErrRequireLineTopology,
	)

	assert.ErrorIs(
		t,
		meshops.RequireTopology(modeling.NewMesh(modeling.TriangleTopology, nil), modeling.PointTopology),
		meshops.ErrRequirePointTopology,
	)

	assert.ErrorIs(
		t,
		meshops.RequireTopology(modeling.NewMesh(modeling.TriangleTopology, nil), modeling.LineLoopTopology),
		meshops.ErrRequireDifferentTopology,
	)
}
