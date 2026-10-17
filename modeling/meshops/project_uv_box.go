package meshops

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

var errProjectionNeedsTriangles = errors.New("projection needs a triangle mesh")

// Returns each new vertex's source vertex and group alongside the mesh.
func splitByGroup(m modeling.Mesh, group func(corner int) int) (modeling.Mesh, []int, []int) {
	type key struct{ vertex, group int }
	indices := m.Indices()
	remapped := make([]int, indices.Len())
	seen := map[key]int{}
	var sources, groups []int
	for corner := 0; corner < indices.Len(); corner++ {
		k := key{indices.At(corner), group(corner)}
		index, ok := seen[k]
		if !ok {
			index = len(sources)
			seen[k] = index
			sources = append(sources, k.vertex)
			groups = append(groups, k.group)
		}
		remapped[corner] = index
	}

	v4 := map[string][]vector4.Float64{}
	for _, attr := range m.Float4Attributes() {
		data := m.Float4Attribute(attr)
		values := make([]vector4.Float64, len(sources))
		for i, source := range sources {
			values[i] = data.At(source)
		}
		v4[attr] = values
	}
	v3 := map[string][]vector3.Float64{}
	for _, attr := range m.Float3Attributes() {
		data := m.Float3Attribute(attr)
		values := make([]vector3.Float64, len(sources))
		for i, source := range sources {
			values[i] = data.At(source)
		}
		v3[attr] = values
	}
	v2 := map[string][]vector2.Float64{}
	for _, attr := range m.Float2Attributes() {
		data := m.Float2Attribute(attr)
		values := make([]vector2.Float64, len(sources))
		for i, source := range sources {
			values[i] = data.At(source)
		}
		v2[attr] = values
	}
	v1 := map[string][]float64{}
	for _, attr := range m.Float1Attributes() {
		data := m.Float1Attribute(attr)
		values := make([]float64, len(sources))
		for i, source := range sources {
			values[i] = data.At(source)
		}
		v1[attr] = values
	}

	split := modeling.NewMesh(m.Topology(), remapped).
		SetFloat4Data(v4).
		SetFloat3Data(v3).
		SetFloat2Data(v2).
		SetFloat1Data(v1)
	return split, sources, groups
}

func triangleNormal(m modeling.Mesh, triangle int) vector3.Float64 {
	tri := m.Tri(triangle)
	a := tri.P1Vec3Attr(modeling.PositionAttribute)
	b := tri.P2Vec3Attr(modeling.PositionAttribute)
	c := tri.P3Vec3Attr(modeling.PositionAttribute)
	return b.Sub(a).Cross(c.Sub(a))
}

const (
	facePosX = iota
	faceNegX
	facePosY
	faceNegY
	facePosZ
	faceNegZ
)

func dominantFace(n vector3.Float64) int {
	x, y, z := math.Abs(n.X()), math.Abs(n.Y()), math.Abs(n.Z())
	switch {
	case x >= y && x >= z && x > 0:
		if n.X() > 0 {
			return facePosX
		}
		return faceNegX
	case y >= z && y > 0:
		if n.Y() > 0 {
			return facePosY
		}
		return faceNegY
	case n.Z() < 0:
		return faceNegZ
	}
	return facePosZ
}

// Each face reads its texture upright and unmirrored when seen from outside.
func boxFaceUV(p vector3.Float64, face int) vector2.Float64 {
	switch face {
	case facePosX:
		return vector2.New(-p.Z(), p.Y())
	case faceNegX:
		return vector2.New(p.Z(), p.Y())
	case facePosY:
		return vector2.New(p.X(), -p.Z())
	case faceNegY:
		return vector2.New(p.X(), p.Z())
	case faceNegZ:
		return vector2.New(-p.X(), p.Y())
	}
	return vector2.New(p.X(), p.Y())
}

func BoxProjectUV(m modeling.Mesh, scale, offset vector2.Float64) (modeling.Mesh, error) {
	if !m.HasFloat3Attribute(modeling.PositionAttribute) || m.PrimitiveCount() == 0 {
		return m, nil
	}
	if m.Topology() != modeling.TriangleTopology {
		return m, errProjectionNeedsTriangles
	}

	positions := m.Float3Attribute(modeling.PositionAttribute)
	faces := make([]int, m.PrimitiveCount())
	for triangle := range faces {
		faces[triangle] = dominantFace(triangleNormal(m, triangle))
	}

	split, sources, groups := splitByGroup(m, func(corner int) int { return faces[corner/3] })
	su, sv := nonZero(scale.X()), nonZero(scale.Y())
	uvs := make([]vector2.Float64, len(sources))
	for i, source := range sources {
		uv := boxFaceUV(positions.At(source), groups[i])
		uvs[i] = vector2.New(uv.X()/su+offset.X(), uv.Y()/sv+offset.Y())
	}
	return split.SetFloat2Attribute(modeling.TexCoordAttribute, uvs), nil
}

func nonZero(v float64) float64 {
	if v == 0 {
		return 1
	}
	return v
}

type BoxProjectUVNode struct {
	Mesh   nodes.Output[modeling.Mesh]
	Scale  nodes.Output[vector2.Float64] `description:"World size of one texture repeat along u and v, in meters. Defaults to 1 x 1."`
	Offset nodes.Output[vector2.Float64] `description:"Added to every uv after scaling. Defaults to zero."`
}

func (BoxProjectUVNode) Description() string {
	return "Gives a mesh UVs by projecting each face from the axis it faces most, like a box around it, so no side is left edge-on."
}

func (BoxProjectUVNode) Keywords() []string {
	return []string{"unwrap", "triplanar", "cube"}
}

func (n BoxProjectUVNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}
	mesh, err := BoxProjectUV(
		nodes.GetOutputValue(out, n.Mesh),
		nodes.TryGetOutputValue(out, n.Scale, vector2.One[float64]()),
		nodes.TryGetOutputValue(out, n.Offset, vector2.Zero[float64]()),
	)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
