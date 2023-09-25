package ply_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestWriteASCII(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment Created with github.com/EliCDavis/polyform
element vertex 8
property float nx
property float ny
property float nz
property float x
property float y
property float z
element face 12
property list uchar int vertex_indices
end_header
-0.577350 -0.577350 -0.577350 -0.500000 -0.500000 -0.500000
-0.577350 -0.577350 0.577350 -0.500000 -0.500000 0.500000
-0.577350 0.577350 -0.577350 -0.500000 0.500000 -0.500000
-0.577350 0.577350 0.577350 -0.500000 0.500000 0.500000
0.577350 -0.577350 -0.577350 0.500000 -0.500000 -0.500000
0.577350 -0.577350 0.577350 0.500000 -0.500000 0.500000
0.577350 0.577350 -0.577350 0.500000 0.500000 -0.500000
0.577350 0.577350 0.577350 0.500000 0.500000 0.500000
3 0 2 6
3 0 6 4
3 1 3 2
3 1 2 0
3 4 6 7
3 4 7 5
3 2 3 7
3 2 7 6
3 1 0 4
3 1 4 5
3 5 7 3
3 5 3 1
`
	cube := primitives.Cube{
		Height: 1,
		Width:  1,
		Depth:  1,
	}.Welded()

	buf := bytes.Buffer{}
	err := ply.WriteASCII(&buf, cube)

	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}

func TestWriteASCIIWithTextureData(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment TextureFile tri.png
comment Created with github.com/EliCDavis/polyform
element vertex 3
property float x
property float y
property float z
element face 1
property list uchar int vertex_indices
property list uchar float texcoord
end_header
0.000000 0.000000 0.000000
0.000000 1.000000 0.000000
1.000000 1.000000 0.000000
3 0 1 2 6 0.000000 0.000000 0.000000 1.000000 1.000000 1.000000
`

	imgName := "tri.png"
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 1., 0.),
			},
		}).
		SetFloat2Data(map[string][]vector2.Vector[float64]{
			modeling.TexCoordAttribute: []vector2.Float64{
				vector2.New(0., 0.),
				vector2.New(0., 1.),
				vector2.New(1., 1.),
			},
		}).
		SetMaterials([]modeling.MeshMaterial{
			{
				PrimitiveCount: 1,
				Material: &modeling.Material{
					Name:            "example",
					ColorTextureURI: &imgName,
				},
			},
		})

	buf := bytes.Buffer{}
	err := ply.WriteASCII(&buf, tri)

	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}
