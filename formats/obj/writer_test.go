package obj_test

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteObj_EmptyMesh(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.EmptyMesh(modeling.TriangleTopology)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t, `# Created with github.com/EliCDavis/polyform
`, buf.String())
}

func TestWriteObj_NoNormalsOrUVs(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New[float64](1., 2., 3.),
				vector3.New[float64](4., 5., 6.),
				vector3.New[float64](7., 8., 9.),
			},
		})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
f 1 2 3
`, buf.String())
}

func TestWriteObj_NoUVs(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New[float64](1, 2, 3),
				vector3.New[float64](4, 5, 6),
				vector3.New[float64](7, 8, 9),
			},
			modeling.NormalAttribute: {
				vector3.New[float64](0, 1, 0),
				vector3.New[float64](0, 0, 1),
				vector3.New[float64](1, 0, 0),
			},
		})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
vn 0 1 0
vn 0 0 1
vn 1 0 0
f 1//1 2//2 3//3
`, buf.String())
}

func TestWriteObj_NoNormals(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New[float64](1, 2, 3),
			vector3.New[float64](4, 5, 6),
			vector3.New[float64](7, 8, 9),
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, []vector2.Float64{
			vector2.New[float64](1., 0.5),
			vector2.New[float64](0.5, 1.),
			vector2.New[float64](0., 0.),
		})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
vt 1 0.5
vt 0.5 1
vt 0 0
f 1/1 2/2 3/3
`, buf.String())
}

func TestWriteObj(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New[float64](1, 2, 3),
				vector3.New[float64](4, 5, 6),
				vector3.New[float64](7, 8, 9),
			},
			modeling.NormalAttribute: {
				vector3.New[float64](0, 1, 0),
				vector3.New[float64](0, 0, 1),
				vector3.New[float64](1, 0, 0),
			},
		}).
		SetFloat2Data(
			map[string][]vector2.Float64{
				modeling.TexCoordAttribute: {
					vector2.New[float64](1, 0.5),
					vector2.New[float64](0.5, 1),
					vector2.New[float64](0, 0),
				},
			},
		)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.WriteMesh(m, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
vt 1 0.5
vt 0.5 1
vt 0 0
vn 0 1 0
vn 0 0 1
vn 1 0 0
f 1/1/1 2/2/2 3/3/3
`, buf.String())
}

func TestWriteObjWithSingleMaterial(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New[float64](1, 2, 3),
				vector3.New[float64](4, 5, 6),
				vector3.New[float64](7, 8, 9),
			},
			modeling.NormalAttribute: {
				vector3.New[float64](0, 1, 0),
				vector3.New[float64](0, 0, 1),
				vector3.New[float64](1, 0, 0),
			},
		}).
		SetFloat2Data(map[string][]vector2.Float64{
			modeling.TexCoordAttribute: {
				vector2.New[float64](1, 0.5),
				vector2.New[float64](0.5, 1),
				vector2.New[float64](0, 0),
			},
		})
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.Write(obj.Scene{
		Objects: []obj.Object{
			{
				Entries: []obj.Entry{
					{
						Mesh: m,
						Material: &obj.Material{
							Name: "red",
						},
					},
				},
			},
		},
	}, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
vt 1 0.5
vt 0.5 1
vt 0 0
vn 0 1 0
vn 0 0 1
vn 1 0 0
usemtl red
f 1/1/1 2/2/2 3/3/3
`, buf.String())
}

func TestWriteObjWithMultipleMaterials(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(1., 2, 3),
				vector3.New(4., 5, 6),
				vector3.New(7., 8, 9),
			},
			modeling.NormalAttribute: {
				vector3.New(0., 1, 0),
				vector3.New(0., 0, 1),
				vector3.New(1., 0, 0),
			},
		}).
		SetFloat2Data(map[string][]vector2.Float64{
			modeling.TexCoordAttribute: {
				vector2.New(1, 0.5),
				vector2.New(0.5, 1),
				vector2.New(0., 0),
			},
		})

	scene := obj.Scene{
		Objects: []obj.Object{
			{
				Entries: []obj.Entry{
					{
						Mesh: m,
						Material: &obj.Material{
							Name: "red",
						},
					},
					{
						Mesh: m,
						Material: &obj.Material{
							Name: "blue",
						},
					},
				},
			},
		},
	}

	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := obj.Write(scene, "", &buf)

	// ASSERT =================================================================
	require.NoError(t, err)

	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
v 1 2 3
v 4 5 6
v 7 8 9
vt 1 0.5
vt 0.5 1
vt 0 0
vn 0 1 0
vn 0 0 1
vn 1 0 0
v 1 2 3
v 4 5 6
v 7 8 9
vt 1 0.5
vt 0.5 1
vt 0 0
vn 0 1 0
vn 0 0 1
vn 1 0 0
usemtl red
f 1/1/1 2/2/2 3/3/3
usemtl blue
f 4/4/4 5/5/5 6/6/6
`, buf.String())
}

func TestWriteMaterials(t *testing.T) {
	// ARRANGE ================================================================
	buf := bytes.Buffer{}
	scene := obj.Scene{
		Objects: []obj.Object{
			{
				Entries: []obj.Entry{
					{
						Material: &obj.Material{
							Name:         "red",
							DiffuseColor: color.RGBA{1, 255, 3, 255},
						},
					},
					{
						Material: &obj.Material{
							Name:          "blue",
							AmbientColor:  color.RGBA{4, 5, 6, 255},
							SpecularColor: color.RGBA{7, 8, 9, 255},
						},
					},
				},
			},
		},
	}

	// ACT ====================================================================
	err := obj.WriteMaterials(scene, &buf)

	// ASSERT =================================================================
	require.NoError(t, err)
	assert.Equal(t,
		`# Created with github.com/EliCDavis/polyform
newmtl red
Kd 0.004 1 0.012
Ns 0
Ni 0
d 1

newmtl blue
Ka 0.016 0.02 0.024
Ks 0.027 0.031 0.035
Ns 0
Ni 0
d 1

`, buf.String())
}
