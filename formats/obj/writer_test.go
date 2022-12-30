package obj_test

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestWriteObj_EmptyMesh(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.EmptyMesh()
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t, ``, buf.String())
}

func TestWriteObj_NoNormalsOrUVs(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
		},
		nil, nil, nil,
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

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
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: {
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
			modeling.NormalAttribute: {
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(0, 0, 1),
				vector.NewVector3(1, 0, 0),
			},
		},
		nil,
		nil,
		nil,
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

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
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: []vector.Vector2{
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
		nil,
		nil,
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

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
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
			modeling.NormalAttribute: []vector.Vector3{
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(0, 0, 1),
				vector.NewVector3(1, 0, 0),
			},
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: []vector.Vector2{
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
		nil,
		nil,
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

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

func TestWriteObjWithSingleMaterial(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
			modeling.NormalAttribute: []vector.Vector3{
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(0, 0, 1),
				vector.NewVector3(1, 0, 0),
			},
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: {
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
		nil,
		[]modeling.MeshMaterial{
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name: "red",
				},
			},
		},
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

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
usemtl red
f 1/1/1 2/2/2 3/3/3
`, buf.String())
}

func TestWriteObjWithMultipleMaterials(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
			3, 4, 5,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
				vector.NewVector3(1, 2, 3),
				vector.NewVector3(4, 5, 6),
				vector.NewVector3(7, 8, 9),
			},
			modeling.NormalAttribute: []vector.Vector3{
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(0, 0, 1),
				vector.NewVector3(1, 0, 0),
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(0, 0, 1),
				vector.NewVector3(1, 0, 0),
			},
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: []vector.Vector2{
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
				vector.NewVector2(1, 0.5),
				vector.NewVector2(0.5, 1),
				vector.NewVector2(0, 0),
			},
		},
		nil,
		[]modeling.MeshMaterial{
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name: "red",
				},
			},
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name: "blue",
				},
			},
		},
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t,
		`v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
v 1.000000 2.000000 3.000000
v 4.000000 5.000000 6.000000
v 7.000000 8.000000 9.000000
vt 1.000000 0.500000
vt 0.500000 1.000000
vt 0.000000 0.000000
vt 1.000000 0.500000
vt 0.500000 1.000000
vt 0.000000 0.000000
vn 0.000000 1.000000 0.000000
vn 0.000000 0.000000 1.000000
vn 1.000000 0.000000 0.000000
vn 0.000000 1.000000 0.000000
vn 0.000000 0.000000 1.000000
vn 1.000000 0.000000 0.000000
usemtl red
f 1/1/1 2/2/2 3/3/3
usemtl blue
f 4/4/4 5/5/5 6/6/6
`, buf.String())
}

func TestWriteMaterials(t *testing.T) {
	// ARRANGE ================================================================
	buf := bytes.Buffer{}
	m := modeling.NewMesh(
		[]int{},
		nil, nil, nil,
		[]modeling.MeshMaterial{
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name:         "red",
					DiffuseColor: color.RGBA{1, 255, 3, 255},
				},
			},
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name:          "blue",
					AmbientColor:  color.RGBA{4, 5, 6, 255},
					SpecularColor: color.RGBA{7, 8, 9, 255},
				},
			},
		},
	)

	// ACT ====================================================================
	err := obj.WriteMaterials(m, &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t,
		`newmtl red
Kd 0.003922 1.000000 0.011765
Ns 0.000000
Ni 0.000000
d 1.000000

newmtl blue
Ka 0.015686 0.019608 0.023529
Ks 0.027451 0.031373 0.035294
Ns 0.000000
Ni 0.000000
d 1.000000

`, buf.String())
}
