package mesh_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestWriteObj_EmptyMesh(t *testing.T) {
	// ARRANGE ================================================================
	m := mesh.MeshFromView(mesh.MeshView{
		Vertices:  []vector.Vector3{},
		Triangles: []int{},
	})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := m.WriteObj(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t, ``, buf.String())
}

func TestWriteObj_NoNormalsOrUVs(t *testing.T) {
	// ARRANGE ================================================================
	m := mesh.MeshFromView(mesh.MeshView{
		Vertices: []vector.Vector3{
			vector.NewVector3(1, 2, 3),
			vector.NewVector3(4, 5, 6),
			vector.NewVector3(7, 8, 9),
		},
		Triangles: []int{
			0, 1, 2,
		},
	})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := m.WriteObj(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t,
		`v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
f 1 2 3
`, buf.String())
}

func TestWriteObj_NoUVs(t *testing.T) {
	// ARRANGE ================================================================
	m := mesh.MeshFromView(mesh.MeshView{
		Vertices: []vector.Vector3{
			vector.NewVector3(1, 2, 3),
			vector.NewVector3(4, 5, 6),
			vector.NewVector3(7, 8, 9),
		},
		Triangles: []int{
			0, 1, 2,
		},
		Normals: []vector.Vector3{
			vector.NewVector3(0, 1, 0),
			vector.NewVector3(0, 0, 1),
			vector.NewVector3(1, 0, 0),
		},
	})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := m.WriteObj(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t,
		`v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
vn 0.000000 1.000000 0.000000
vn 0.000000 0.000000 1.000000
vn 1.000000 0.000000 0.000000
f 1//1 2//2 3//3
`, buf.String())
}

func TestWriteObj_NoNormals(t *testing.T) {
	// ARRANGE ================================================================
	m := mesh.MeshFromView(mesh.MeshView{
		Vertices: []vector.Vector3{
			vector.NewVector3(1, 2, 3),
			vector.NewVector3(4, 5, 6),
			vector.NewVector3(7, 8, 9),
		},
		Triangles: []int{
			0, 1, 2,
		},
		UVs: [][]vector.Vector2{
			{
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
	})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := m.WriteObj(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t,
		`v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
vt 1.000000 0.500000
vt 0.500000 1.000000
vt 0.000000 0.000000
f 1/1 2/2 3/3
`, buf.String())
}

func TestWriteObj(t *testing.T) {
	// ARRANGE ================================================================
	m := mesh.MeshFromView(mesh.MeshView{
		Vertices: []vector.Vector3{
			vector.NewVector3(1, 2, 3),
			vector.NewVector3(4, 5, 6),
			vector.NewVector3(7, 8, 9),
		},
		Triangles: []int{
			0, 1, 2,
		},
		Normals: []vector.Vector3{
			vector.NewVector3(0, 1, 0),
			vector.NewVector3(0, 0, 1),
			vector.NewVector3(1, 0, 0),
		},
		UVs: [][]vector.Vector2{
			{
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
	})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := m.WriteObj(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t,
		`v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
vt 1.000000 0.500000
vt 0.500000 1.000000
vt 0.000000 0.000000
vn 0.000000 1.000000 0.000000
vn 0.000000 0.000000 1.000000
vn 1.000000 0.000000 0.000000
f 1/1/1 2/2/2 3/3/3
`, buf.String())
}
