package csg_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Attributes that are linear in position: a cut anywhere on a face has to
// reproduce them exactly, or the weights are wrong.
func withLinearFields(m modeling.Mesh) modeling.Mesh {
	positions := m.Float3Attribute(modeling.PositionAttribute)
	uvs := make([]vector2.Float64, positions.Len())
	heights := make([]float64, positions.Len())
	tints := make([]vector4.Float64, positions.Len())
	for i := range uvs {
		p := positions.At(i)
		uvs[i] = vector2.New(p.X(), p.Y())
		heights[i] = p.Z()
		tints[i] = vector4.New(p.X(), p.Y(), p.Z(), 1)
	}
	return m.
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs).
		SetFloat1Attribute("height", heights).
		SetFloat4Attribute(modeling.ColorAttribute, tints)
}

func TestAttributesFollowTheCut(t *testing.T) {
	box := withLinearFields(cube(vector3.Zero[float64](), 2))
	ball := withLinearFields(primitives.UVSphere(1.28, 12, 18))

	for name, op := range map[string]func(modeling.Mesh, modeling.Mesh) (modeling.Mesh, error){
		"union": csg.Union, "intersect": csg.Intersect, "subtract": csg.Subtract,
	} {
		t.Run(name, func(t *testing.T) {
			out, err := op(box, ball)
			require.NoError(t, err)

			positions := out.Float3Attribute(modeling.PositionAttribute)
			uvs := out.Float2Attribute(modeling.TexCoordAttribute)
			heights := out.Float1Attribute("height")
			tints := out.Float4Attribute(modeling.ColorAttribute)
			require.Greater(t, positions.Len(), 0)

			for i := 0; i < positions.Len(); i++ {
				p := positions.At(i)
				assert.InDeltaf(t, p.X(), uvs.At(i).X(), 1e-9, "vertex %d u", i)
				assert.InDeltaf(t, p.Y(), uvs.At(i).Y(), 1e-9, "vertex %d v", i)
				assert.InDeltaf(t, p.Z(), heights.At(i), 1e-9, "vertex %d height", i)
				assert.InDeltaf(t, p.X(), tints.At(i).X(), 1e-9, "vertex %d tint", i)
				assert.InDeltaf(t, 1, tints.At(i).W(), 1e-9, "vertex %d tint alpha", i)
			}
		})
	}
}

func TestCarriedNormalsFaceTheWayTheSurfaceDoes(t *testing.T) {
	box := cube(vector3.Zero[float64](), 2)
	ball := primitives.UVSphere(1.28, 12, 18)
	require.True(t, box.HasFloat3Attribute(modeling.NormalAttribute), "sanity: the cube carries normals")

	out, err := csg.Subtract(box, ball)
	require.NoError(t, err)

	indices := out.Indices()
	positions := out.Float3Attribute(modeling.PositionAttribute)
	normals := out.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < indices.Len(); i += 3 {
		a, b, c := positions.At(indices.At(i)), positions.At(indices.At(i+1)), positions.At(indices.At(i+2))
		facing := b.Sub(a).Cross(c.Sub(a))
		for k := 0; k < 3; k++ {
			n := normals.At(indices.At(i + k))
			assert.InDeltaf(t, 1, n.Length(), 1e-9, "triangle %d corner %d is not unit", i/3, k)
			assert.Positivef(t, n.Dot(facing), "triangle %d corner %d points into the solid", i/3, k)
		}
	}
}

func TestOutputCarriesEveryAttributeOfEitherInput(t *testing.T) {
	box := cube(vector3.Zero[float64](), 2).
		SetFloat1Attribute("only on a", make([]float64, 24))
	ball := primitives.UVSphere(1.28, 12, 18).
		SetFloat2Attribute("only on b", make([]vector2.Float64, primitives.UVSphere(1.28, 12, 18).AttributeLength()))

	out, err := csg.Union(box, ball)
	require.NoError(t, err)

	assert.True(t, out.HasFloat1Attribute("only on a"))
	assert.True(t, out.HasFloat2Attribute("only on b"))
	assert.True(t, out.HasFloat3Attribute(modeling.NormalAttribute))
	for _, name := range out.Float1Attributes() {
		assert.Equal(t, out.AttributeLength(), out.Float1Attribute(name).Len(), name)
	}
	for _, name := range out.Float2Attributes() {
		assert.Equal(t, out.AttributeLength(), out.Float2Attribute(name).Len(), name)
	}
	for _, name := range out.Float3Attributes() {
		assert.Equal(t, out.AttributeLength(), out.Float3Attribute(name).Len(), name)
	}
}

func TestOutputFallsBackToFlatNormals(t *testing.T) {
	strip := func(m modeling.Mesh) modeling.Mesh {
		return modeling.NewTriangleMesh(readIndices(m)).
			SetFloat3Attribute(modeling.PositionAttribute, readPositions(m))
	}
	out, err := csg.Union(strip(cube(vector3.Zero[float64](), 2)), strip(primitives.UVSphere(1.28, 12, 18)))
	require.NoError(t, err)

	indices := out.Indices()
	positions := out.Float3Attribute(modeling.PositionAttribute)
	normals := out.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < indices.Len(); i += 3 {
		a, b, c := positions.At(indices.At(i)), positions.At(indices.At(i+1)), positions.At(indices.At(i+2))
		flat := b.Sub(a).Cross(c.Sub(a)).Normalized()
		for k := 0; k < 3; k++ {
			assert.InDeltaf(t, 1, normals.At(indices.At(i+k)).Dot(flat), 1e-9, "triangle %d corner %d", i/3, k)
		}
	}
}

func readIndices(m modeling.Mesh) []int {
	indices := m.Indices()
	out := make([]int, indices.Len())
	for i := range out {
		out[i] = indices.At(i)
	}
	return out
}

func readPositions(m modeling.Mesh) []vector3.Float64 {
	positions := m.Float3Attribute(modeling.PositionAttribute)
	out := make([]vector3.Float64, positions.Len())
	for i := range out {
		out[i] = positions.At(i)
	}
	return out
}
