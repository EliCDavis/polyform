package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type TubeUVs struct {
	Outer  *StripUVs
	Inner  *StripUVs
	Top    *CircleUVs
	Bottom *CircleUVs
}

var (
	defaultTubeWallUVs = StripUVs{Start: vector2.New(0., 0.5), End: vector2.New(1., 0.5), Width: 1}
	defaultTubeCapUVs  = CircleUVs{Center: vector2.Fill(0.5), Radius: 0.5}
)

type Tube struct {
	Sides       int
	Height      float64
	InnerRadius float64
	OuterRadius float64

	UVs *TubeUVs
}

func (t Tube) ToMesh() modeling.Mesh {
	sides := max(t.Sides, 3)
	inner := math.Min(t.InnerRadius, t.OuterRadius)
	outer := math.Max(t.InnerRadius, t.OuterRadius)
	halfHeight := t.Height / 2

	angleIncrement := (1.0 / float64(sides)) * 2.0 * math.Pi

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

	layout := t.UVs
	if layout == nil {
		layout = &TubeUVs{}
	}
	outerWall, innerWall := defaultTubeWallUVs, defaultTubeWallUVs
	if layout.Outer != nil {
		outerWall = *layout.Outer
	}
	if layout.Inner != nil {
		innerWall = *layout.Inner
	}
	topCap, bottomCap := defaultTubeCapUVs, defaultTubeCapUVs
	if layout.Top != nil {
		topCap = *layout.Top
	}
	if layout.Bottom != nil {
		bottomCap = *layout.Bottom
	}

	wallUV := func(strip StripUVs, top bool) func(int, float64, float64) vector2.Float64 {
		startLeft, dir := strip.StartLeft(), strip.Dir()
		acrossOffset := vector2.Zero[float64]()
		if !top {
			acrossOffset = strip.LeftToRight()
		}
		return func(i int, cos, sin float64) vector2.Float64 {
			along := float64(i) / float64(sides)
			return startLeft.Add(dir.Scale(along)).Add(acrossOffset)
		}
	}

	capUV := func(circle CircleUVs, radius float64) func(int, float64, float64) vector2.Float64 {
		scale := 0.
		if outer > 0 {
			scale = (radius / outer) * circle.Radius
		}
		return func(i int, cos, sin float64) vector2.Float64 {
			return vector2.New(cos*scale, sin*scale).Add(circle.Center)
		}
	}

	outerTop := ring(outer, halfHeight, outward, wallUV(outerWall, true))
	outerBottom := ring(outer, -halfHeight, outward, wallUV(outerWall, false))
	innerTop := ring(inner, halfHeight, inward, wallUV(innerWall, true))
	innerBottom := ring(inner, -halfHeight, inward, wallUV(innerWall, false))
	capOuterTop := ring(outer, halfHeight, up, capUV(topCap, outer))
	capInnerTop := ring(inner, halfHeight, up, capUV(topCap, inner))
	capOuterBottom := ring(outer, -halfHeight, down, capUV(bottomCap, outer))
	capInnerBottom := ring(inner, -halfHeight, down, capUV(bottomCap, inner))

	tris := make([]int, 0, sides*8*3)
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

	mesh := modeling.NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, vertices).
		SetFloat3Attribute(modeling.NormalAttribute, normals)

	if t.UVs != nil {
		mesh = mesh.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
	}
	return mesh
}

type TubeUVsNode struct {
	Outer  nodes.Output[StripUVs]
	Inner  nodes.Output[StripUVs]
	Top    nodes.Output[CircleUVs]
	Bottom nodes.Output[CircleUVs]
}

func (n TubeUVsNode) Description() string {
	return "UV layout for the Tube primitive. The walls unroll into strips; the caps are annuli and take a circular layout. An unwired port covers the whole 0..1 square."
}

func (n TubeUVsNode) Out(out *nodes.StructOutput[TubeUVs]) {
	out.Set(TubeUVs{
		Outer:  nodes.TryGetOutputReference(out, n.Outer, nil),
		Inner:  nodes.TryGetOutputReference(out, n.Inner, nil),
		Top:    nodes.TryGetOutputReference(out, n.Top, nil),
		Bottom: nodes.TryGetOutputReference(out, n.Bottom, nil),
	})
}

type TubeNode struct {
	Sides       nodes.Output[int]
	Height      nodes.Output[float64]
	InnerRadius nodes.Output[float64]
	OuterRadius nodes.Output[float64]
	UVs         nodes.Output[TubeUVs]
}

func (t TubeNode) Description() string {
	return "A hollow cylinder along the Y axis."
}

func (t TubeNode) Keywords() []string {
	return []string{"pipe", "washer", "ring"}
}

func (t TubeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	tube := Tube{
		Sides:       nodes.TryGetOutputValue(out, t.Sides, 20),
		Height:      nodes.TryGetOutputValue(out, t.Height, 1),
		InnerRadius: nodes.TryGetOutputValue(out, t.InnerRadius, 0.25),
		OuterRadius: nodes.TryGetOutputValue(out, t.OuterRadius, 0.5),
		UVs:         nodes.TryGetOutputReference(out, t.UVs, nil),
	}
	out.Set(tube.ToMesh())
}
