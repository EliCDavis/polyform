package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func rotate(m modeling.Mesh, q quaternion.Quaternion) modeling.Mesh {
	return m.Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    q,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    q,
		},
	)
}

func DefaultCubeUVs() *CubeUVs {
	oneThird := 1. / 3.
	return &CubeUVs{
		Top: &StripUVs{
			Start: vector2.New(0., 0.5),
			End:   vector2.New(0.25, 0.5),
			Width: oneThird,
		},
		Front: &StripUVs{
			Start: vector2.New(0.25, 0.5),
			End:   vector2.New(0.5, 0.5),
			Width: oneThird,
		},
		Bottom: &StripUVs{
			Start: vector2.New(0.5, 0.5),
			End:   vector2.New(0.75, 0.5),
			Width: oneThird,
		},
		Back: &StripUVs{
			Start: vector2.New(0.75, 0.5),
			End:   vector2.New(0.1, 0.5),
			Width: oneThird,
		},
		Left: &StripUVs{
			Start: vector2.New(0.375, oneThird*2),
			End:   vector2.New(0.375, 1.),
			Width: 0.25,
		},
		Right: &StripUVs{
			Start: vector2.New(0.375, oneThird),
			End:   vector2.New(0.375, 0.),
			Width: 0.25,
		},
	}
}

var cubeVertIndices = []int{
	// Back
	0, 2, 6,
	0, 6, 4,

	// Left
	1, 3, 2,
	1, 2, 0,

	// Right
	4, 6, 7,
	4, 7, 5,

	// Top
	2, 3, 7,
	2, 7, 6,

	// Bottom
	1, 0, 4,
	1, 4, 5,

	// Front
	5, 7, 3,
	5, 3, 1,
}

type CubeUVs struct {
	Top    EuclideanUVSpace
	Bottom EuclideanUVSpace
	Left   EuclideanUVSpace
	Right  EuclideanUVSpace
	Front  EuclideanUVSpace
	Back   EuclideanUVSpace
}

type Cube struct {
	Height     float64
	Width      float64
	Depth      float64
	Dimensions int
	UVs        *CubeUVs
}

func UnitCube() modeling.Mesh {
	return Cube{
		Height: 1,
		Width:  1,
		Depth:  1,
	}.Welded()
}

/*
Builds a cube

The Cube Indices:

	  3 ------- 7
	/ |      /  |

2  |     6   |
|  1     |   5
| /      | /
0 ------ 4
*/
func (c Cube) UnweldedQuads() modeling.Mesh {
	halfW := c.Width / 2
	halfH := c.Height / 2
	halfD := c.Depth / 2
	var topUV, bottomUV, leftUV, rightUV, frontUV, backUV EuclideanUVSpace = nil, nil, nil, nil, nil, nil
	if c.UVs != nil {
		topUV = c.UVs.Top
		bottomUV = c.UVs.Bottom
		leftUV = c.UVs.Left
		rightUV = c.UVs.Right
		frontUV = c.UVs.Front
		backUV = c.UVs.Back
	}

	top := Quad{
		UVs:     topUV,
		Width:   c.Width,
		Depth:   c.Depth,
		Columns: c.Dimensions,
		Rows:    c.Dimensions,
	}.ToMesh().Translate(vector3.New(0., halfH, 0.))

	bottom := rotate(
		Quad{UVs: bottomUV, Width: c.Width, Depth: c.Depth, Columns: c.Dimensions, Rows: c.Dimensions}.ToMesh(),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
	).Translate(vector3.New(0., -halfH, 0.))

	left := rotate(
		Quad{UVs: leftUV, Width: c.Height, Depth: c.Depth, Columns: c.Dimensions, Rows: c.Dimensions}.ToMesh(),
		quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]()),
	).Translate(vector3.New(-halfW, 0., 0.))

	right := rotate(
		Quad{UVs: rightUV, Width: c.Height, Depth: c.Depth, Columns: c.Dimensions, Rows: c.Dimensions}.ToMesh(),
		quaternion.FromTheta(math.Pi*(3./2.), vector3.Forward[float64]()),
	).Translate(vector3.New(halfW, 0, 0.))

	front := rotate(
		Quad{UVs: frontUV, Width: c.Width, Depth: c.Height, Columns: c.Dimensions, Rows: c.Dimensions}.ToMesh(),
		quaternion.FromTheta(math.Pi*(3./2.), vector3.Left[float64]()),
	).Translate(vector3.New(0., 0., halfD))

	back := rotate(
		Quad{UVs: backUV, Width: c.Width, Depth: c.Height, Columns: c.Dimensions, Rows: c.Dimensions}.ToMesh(),
		quaternion.FromTheta(math.Pi*(1./2.), vector3.Left[float64]()),
	).Translate(vector3.New(0., 0., -halfD))

	return top.Append(bottom).Append(left).Append(right).Append(front).Append(back)
}

func (c Cube) Welded() modeling.Mesh {
	halfW := c.Width / 2
	halfH := c.Height / 2
	halfD := c.Depth / 2

	potentialVerts := []vector3.Float64{
		// bottom, back, left
		vector3.New(-halfW, -halfH, -halfD),
		// bottom, front, left
		vector3.New(-halfW, -halfH, halfD),
		// top, back, left
		vector3.New(-halfW, halfH, -halfD),
		// top, front, left
		vector3.New(-halfW, halfH, halfD),

		// bottom, back, right
		vector3.New(halfW, -halfH, -halfD),
		// bottom, front, right
		vector3.New(halfW, -halfH, halfD),
		// top, back, right
		vector3.New(halfW, halfH, -halfD),
		// top, front, right
		vector3.New(halfW, halfH, halfD),
	}

	v2Data := make(map[string][]vector2.Float64)
	if c.UVs != nil {
		v2Data[modeling.TexCoordAttribute] = c.calcUVs()
	}

	return modeling.NewTriangleMesh(cubeVertIndices).
		SetFloat3Data(map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: potentialVerts,
			modeling.NormalAttribute:   vector3.Array[float64](potentialVerts).Normalized(),
		}).
		SetFloat2Data(v2Data)

}

func (c Cube) calcUVs() []vector2.Float64 {
	if c.UVs == nil {
		return nil
	}

	/*
		Builds a cube with no normals

		The Cube Indices:

			  3 ------- 7
			/ |      /  |

		2  |     6   |
		|  1     |   5
		| /      | /
		0 ------ 4
	*/
	uvs := make([]vector2.Float64, 8)

	if c.UVs.Top != nil {
		uvs[2] = c.UVs.Top.AtXY(vector2.New(0., 0.))
		uvs[6] = c.UVs.Top.AtXY(vector2.New(1., 0.))
		uvs[3] = c.UVs.Top.AtXY(vector2.New(0., 1.))
		uvs[7] = c.UVs.Top.AtXY(vector2.New(1., 1.))
	}

	if c.UVs.Bottom != nil {
		uvs[0] = c.UVs.Bottom.AtXY(vector2.New(0., 0.))
		uvs[4] = c.UVs.Bottom.AtXY(vector2.New(1., 0.))
		uvs[1] = c.UVs.Bottom.AtXY(vector2.New(0., 1.))
		uvs[5] = c.UVs.Bottom.AtXY(vector2.New(1., 1.))
	}

	if c.UVs.Left != nil {
		uvs[1] = c.UVs.Left.AtXY(vector2.New(0., 0.))
		uvs[0] = c.UVs.Left.AtXY(vector2.New(1., 0.))
		uvs[3] = c.UVs.Left.AtXY(vector2.New(0., 1.))
		uvs[2] = c.UVs.Left.AtXY(vector2.New(1., 1.))
	}

	if c.UVs.Right != nil {
		uvs[4] = c.UVs.Right.AtXY(vector2.New(0., 0.))
		uvs[5] = c.UVs.Right.AtXY(vector2.New(1., 0.))
		uvs[6] = c.UVs.Right.AtXY(vector2.New(0., 1.))
		uvs[7] = c.UVs.Right.AtXY(vector2.New(1., 1.))
	}

	if c.UVs.Front != nil {
		uvs[0] = c.UVs.Front.AtXY(vector2.New(0., 0.))
		uvs[4] = c.UVs.Front.AtXY(vector2.New(1., 0.))
		uvs[2] = c.UVs.Front.AtXY(vector2.New(0., 1.))
		uvs[6] = c.UVs.Front.AtXY(vector2.New(1., 1.))
	}

	if c.UVs.Back != nil {
		uvs[5] = c.UVs.Back.AtXY(vector2.New(0., 0.))
		uvs[1] = c.UVs.Back.AtXY(vector2.New(1., 0.))
		uvs[7] = c.UVs.Back.AtXY(vector2.New(0., 1.))
		uvs[3] = c.UVs.Back.AtXY(vector2.New(1., 1.))
	}

	return uvs
}

type CubeNode struct {
	Width      nodes.Output[float64]
	Height     nodes.Output[float64]
	Depth      nodes.Output[float64]
	Dimensions nodes.Output[int]
	UVs        nodes.Output[CubeUVs]
}

func (c CubeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	strip := &StripUVs{
		Start: vector2.New(0, 0.5),
		End:   vector2.New(1, 0.5),
		Width: 1.,
	}
	cube := Cube{
		Height:     nodes.TryGetOutputValue(out, c.Height, 1.),
		Width:      nodes.TryGetOutputValue(out, c.Width, 1.),
		Depth:      nodes.TryGetOutputValue(out, c.Depth, 1.),
		Dimensions: max(1, nodes.TryGetOutputValue(out, c.Dimensions, 1)),
		UVs: nodes.TryGetOutputReference(out, c.UVs, &CubeUVs{
			Top:    strip,
			Bottom: strip,
			Left:   strip,
			Right:  strip,
			Front:  strip,
			Back:   strip,
		}),
	}
	out.Set(cube.UnweldedQuads())
}

// CubeUVs

type CubeUVsNode struct {
	Top    nodes.Output[StripUVs]
	Bottom nodes.Output[StripUVs]
	Left   nodes.Output[StripUVs]
	Right  nodes.Output[StripUVs]
	Front  nodes.Output[StripUVs]
	Back   nodes.Output[StripUVs]
}

func (cnd CubeUVsNode) Uv(out *nodes.StructOutput[CubeUVs]) {
	uvs := CubeUVs{}

	if cnd.Top != nil {
		uvs.Top = nodes.GetOutputValue(out, cnd.Top)
	}

	if cnd.Bottom != nil {
		uvs.Bottom = nodes.GetOutputValue(out, cnd.Bottom)
	}

	if cnd.Left != nil {
		uvs.Left = nodes.GetOutputValue(out, cnd.Left)
	}

	if cnd.Right != nil {
		uvs.Right = nodes.GetOutputValue(out, cnd.Right)
	}

	if cnd.Front != nil {
		uvs.Front = nodes.GetOutputValue(out, cnd.Front)
	}

	if cnd.Back != nil {
		uvs.Back = nodes.GetOutputValue(out, cnd.Back)
	}

	out.Set(uvs)
}
