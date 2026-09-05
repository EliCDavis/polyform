package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Tube is a hollow cylinder: a ring with real wall thickness, centered on
// the origin and running along the Y axis.
type Tube struct {
	Sides       int
	Height      float64
	InnerRadius float64
	OuterRadius float64
}

func (t Tube) ToMesh() modeling.Mesh {
	sides := max(t.Sides, 3)
	inner := math.Min(t.InnerRadius, t.OuterRadius)
	outer := math.Max(t.InnerRadius, t.OuterRadius)
	halfHeight := t.Height / 2

	angleIncrement := (1.0 / float64(sides)) * 2.0 * math.Pi

	// Every surface gets its own copy of the ring so wall and cap normals
	// stay independent.
	ringCount := sides + 1
	vertices := make([]vector3.Float64, 0, ringCount*8)
	normals := make([]vector3.Float64, 0, ringCount*8)
	uvs := make([]vector2.Float64, 0, ringCount*8)

	ring := func(radius, y float64, normal func(cos, sin float64) vector3.Float64, uv func(i int, cos, sin float64) vector2.Float64) int {
		start := len(vertices)
		for i := 0; i <= sides; i++ {
			angle := angleIncrement * float64(i)
			cos, sin := math.Cos(angle), math.Sin(angle)
			vertices = append(vertices, vector3.New(cos*radius, y, sin*radius))
			normals = append(normals, normal(cos, sin))
			uvs = append(uvs, uv(i, cos, sin))
		}
		return start
	}

	outward := func(cos, sin float64) vector3.Float64 { return vector3.New(cos, 0., sin) }
	inward := func(cos, sin float64) vector3.Float64 { return vector3.New(-cos, 0., -sin) }
	up := func(cos, sin float64) vector3.Float64 { return vector3.Up[float64]() }
	down := func(cos, sin float64) vector3.Float64 { return vector3.Down[float64]() }

	wallUV := func(v float64) func(int, float64, float64) vector2.Float64 {
		return func(i int, cos, sin float64) vector2.Float64 {
			return vector2.New(float64(i)/float64(sides), v)
		}
	}
	capUV := func(radius float64) func(int, float64, float64) vector2.Float64 {
		scale := 0.
		if outer > 0 {
			scale = (radius / outer) * 0.5
		}
		return func(i int, cos, sin float64) vector2.Float64 {
			return vector2.New(0.5+cos*scale, 0.5+sin*scale)
		}
	}

	outerTop := ring(outer, halfHeight, outward, wallUV(1))
	outerBottom := ring(outer, -halfHeight, outward, wallUV(0))
	innerTop := ring(inner, halfHeight, inward, wallUV(1))
	innerBottom := ring(inner, -halfHeight, inward, wallUV(0))
	capOuterTop := ring(outer, halfHeight, up, capUV(outer))
	capInnerTop := ring(inner, halfHeight, up, capUV(inner))
	capOuterBottom := ring(outer, -halfHeight, down, capUV(outer))
	capInnerBottom := ring(inner, -halfHeight, down, capUV(inner))

	tris := make([]int, 0, sides*8*3)
	// quad stitches segment i between two rings, wound so its face normal
	// points the same way as the rings' vertex normals.
	quad := func(ringA, ringB, i int) {
		al, ar := ringA+i-1, ringA+i
		bl, br := ringB+i-1, ringB+i
		tris = append(tris, bl, al, ar, bl, ar, br)
	}

	for i := 1; i <= sides; i++ {
		quad(outerTop, outerBottom, i)
		quad(innerBottom, innerTop, i)
		quad(capInnerTop, capOuterTop, i)
		quad(capOuterBottom, capInnerBottom, i)
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, vertices).
		SetFloat3Attribute(modeling.NormalAttribute, normals).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

type TubeNode struct {
	Sides       nodes.Output[int]
	Height      nodes.Output[float64]
	InnerRadius nodes.Output[float64]
	OuterRadius nodes.Output[float64]
}

func (t TubeNode) Description() string {
	return "A hollow cylinder along the Y axis: pipe, washer, ring, rim. Unlike a capped-off Cylinder this has real wall thickness, so it stays visible edge-on. A short Height gives a flat annulus."
}

func (t TubeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	tube := Tube{
		Sides:       nodes.TryGetOutputValue(out, t.Sides, 20),
		Height:      nodes.TryGetOutputValue(out, t.Height, 1),
		InnerRadius: nodes.TryGetOutputValue(out, t.InnerRadius, 0.25),
		OuterRadius: nodes.TryGetOutputValue(out, t.OuterRadius, 0.5),
	}
	out.Set(tube.ToMesh())
}
