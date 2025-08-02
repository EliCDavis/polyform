package extrude

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func directionOfPoints(points []vector3.Float64) []vector3.Float64 {

	if len(points) == 0 {
		return nil
	}

	if len(points) == 1 {
		return []vector3.Vector[float64]{
			vector3.Up[float64](),
		}
	}

	directions := make([]vector3.Float64, len(points))

	for i, point := range points {
		if i == 0 {
			directions[i] = points[1].Sub(point).Normalized()
			continue
		}

		if i == len(points)-1 {
			directions[i] = point.Sub(points[i-1]).Normalized()
			continue
		}

		dirA := point.Sub(points[i-1]).Normalized()
		dirB := points[i+1].Sub(point).Normalized()
		directions[i] = dirA.Add(dirB).Normalized()
	}

	return directions
}

// TODO: Pretty sure this breaks for paths that have multiple points in the
// same direction.
func polygon(sides int, points []ExtrusionPoint, closed bool) modeling.Mesh {
	if len(points) < 2 {
		panic(fmt.Errorf("can not extrude polygon with %d points", len(points)))
	}

	if sides < 3 {
		panic(fmt.Errorf("can not extrude polygon with %d sides", sides))
	}

	vertCount := sides + 1
	vertices := make([]vector3.Float64, 0, len(points)*vertCount)
	normals := make([]vector3.Float64, 0, len(points)*vertCount)

	circlePoints := make([]vector3.Float64, vertCount)
	circlePoints[0] = vector3.Right[float64]()

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides+1; i++ {
		rot := quaternion.FromTheta(angleIncrement*float64(i), vector3.Up[float64]())
		circlePoints[i] = rot.Rotate(vector3.Right[float64]())
	}

	pointDirections := directionsOfExtrusionPoints(points)

	float2Data := map[string][]vector2.Float64{}

	lastDir := vector3.Up[float64]()
	lastRot := quaternion.New(vector3.Zero[float64](), 1)

	validUVs := true

	// Vertices and normals ===================================================
	for i, p := range points {
		if p.UV == nil {
			validUVs = false
		}

		dir := pointDirections[i]

		rot := quaternion.RotationTo(lastDir, dir)

		for sideIndex := 0; sideIndex < vertCount; sideIndex++ {
			point := circlePoints[sideIndex]
			point = lastRot.Rotate(point)
			point = rot.Rotate(point)

			vertices = append(vertices, point.Scale(p.Thickness).Add(p.Point))
			normals = append(normals, point)
		}

		lastRot = rot.Multiply(lastRot)
		lastDir = dir
	}

	// UVs ====================================================================
	if validUVs {
		uvs := make([]vector2.Float64, 0, len(points)*vertCount)
		for i, p := range points {

			var dirA vector2.Float64
			var dirB vector2.Float64

			if i == 0 {
				dirA = points[0].UV.Point
				dirB = points[1].UV.Point
			} else {
				dirA = points[i-1].UV.Point
				dirB = p.UV.Point
			}

			dir := dirB.Sub(dirA).Normalized()
			perp := vector2.New(dir.Y(), -dir.X()).
				Scale(p.UV.Thickness / 2.)

			// log.Print(perp)
			for sideIndex := 0; sideIndex < vertCount; sideIndex++ {
				percentUsed := ((float64(sideIndex) / float64(sides)) * 2) - 1.
				uvPoint := p.UV.Point.Add(perp.Scale(percentUsed))
				// log.Print(percentUsed, uvPoint)
				uvs = append(uvs, uvPoint)
			}
		}
		float2Data[modeling.TexCoordAttribute] = uvs
	}

	// Triangles ==============================================================
	tris := make([]int, 0, sides*2*3)

	for pathIndex, pathPoint := range points {
		bottom := pathIndex * vertCount
		top := (pathIndex + 1) * vertCount
		if pathIndex == len(points)-1 {
			if closed {
				top = 0
			} else {
				continue
			}
		}
		for sideIndex := 0; sideIndex < sides; sideIndex++ {
			topRight := top + sideIndex
			bottomRight := bottom + sideIndex

			topLeft := topRight + 1
			bottomLeft := bottomRight + 1

			// Figure out the normal of the triangle we're about to make, and
			// whether or not we need to flip it to point away from the center
			// of the extrusion
			dir := vertices[bottomLeft].Sub(vertices[topLeft]).
				Cross(vertices[topLeft].Sub(vertices[topRight]))

			a1 := topLeft
			a2 := topRight

			b1 := topRight
			b2 := bottomRight

			// we need to flip the windings...
			if dir.Dot(vertices[bottomLeft].Sub(pathPoint.Point)) < 0 {
				a1, a2 = a2, a1
				b1, b2 = b2, b1
			}

			tris = append(
				tris,

				bottomLeft,
				a1,
				a2,

				bottomLeft,
				b1,
				b2,
			)
		}
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Data(float2Data)
}

func Polygon(sides int, points []ExtrusionPoint) modeling.Mesh {
	return polygon(sides, points, false)
}

type Circle struct {
	Resolution int
	Radius     float64
	Radii      []float64
	ClosePath  bool
	Path       []vector3.Float64
}

func (c Circle) Extrude() modeling.Mesh {
	points := make([]ExtrusionPoint, len(c.Path))
	varrying := len(c.Radii) == len(c.Path)
	r := c.Radius
	for i, p := range c.Path {
		if varrying {
			r = c.Radii[i]
		}
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: r,
		}
	}
	return polygon(c.Resolution, points, false)
}

type CircleAlongSpline struct {
	CircleResolution int
	Radius           float64
	Radii            []float64
	ClosePath        bool

	Spline           curves.Spline
	SplineResolution int
	UVs              *primitives.StripUVs
}

func (c CircleAlongSpline) Extrude() modeling.Mesh {
	points := make([]ExtrusionPoint, c.SplineResolution)
	varrying := len(c.Radii) == c.SplineResolution
	r := c.Radius
	inc := c.Spline.Length() / float64(c.SplineResolution-1)
	for i := 0; i < c.SplineResolution; i++ {
		if varrying {
			r = c.Radii[i]
		}
		points[i] = ExtrusionPoint{
			Point:     c.Spline.At(inc * float64(i)),
			Thickness: r,
		}
	}

	if c.UVs != nil {
		uvInc := 1. / float64(c.SplineResolution-1)

		for i := range c.SplineResolution {
			points[i].UV = &ExtrusionPointUV{
				Point:     c.UVs.At(uvInc * float64(i)),
				Thickness: c.UVs.Width,
			}
		}
	}

	return polygon(c.CircleResolution, points, c.ClosePath)
}

type CircleNode struct {
	Closed     nodes.Output[bool]
	Resolution nodes.Output[int]
	Radius     nodes.Output[float64]
	Radii      nodes.Output[[]float64]
	Path       nodes.Output[[]vector3.Float64]
}

func (pnd CircleNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if pnd.Path == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	circle := Circle{
		Radius:     nodes.TryGetOutputValue(out, pnd.Radius, 1.0),
		Resolution: max(3, nodes.TryGetOutputValue(out, pnd.Resolution, 3)),
		ClosePath:  nodes.TryGetOutputValue(out, pnd.Closed, false),
		Path:       nodes.GetOutputValue(out, pnd.Path),
		Radii:      nodes.TryGetOutputValue(out, pnd.Radii, nil),
	}
	out.Set(circle.Extrude())
}

type CircleAlongSplineNode struct {
	Closed           nodes.Output[bool]
	CircleResolution nodes.Output[int]
	Radius           nodes.Output[float64]
	Radii            nodes.Output[[]float64]
	Spline           nodes.Output[curves.Spline]
	SplineResolution nodes.Output[int]
	UVs              nodes.Output[primitives.StripUVs]
}

func (pnd CircleAlongSplineNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if pnd.Spline == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	spline := nodes.GetOutputValue(out, pnd.Spline)
	if spline == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	circle := CircleAlongSpline{
		Radius:           nodes.TryGetOutputValue(out, pnd.Radius, 1.0),
		CircleResolution: max(3, nodes.TryGetOutputValue(out, pnd.CircleResolution, 3)),
		ClosePath:        nodes.TryGetOutputValue(out, pnd.Closed, false),
		Spline:           spline,
		SplineResolution: max(3, nodes.TryGetOutputValue(out, pnd.SplineResolution, 3)),
		Radii:            nodes.TryGetOutputValue(out, pnd.Radii, nil),
		UVs:              nodes.TryGetOutputReference(out, pnd.UVs, nil),
	}

	out.Set(circle.Extrude())
}
