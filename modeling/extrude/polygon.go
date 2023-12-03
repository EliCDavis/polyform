package extrude

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
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

	// Vertices and normals ===================================================
	for i, p := range points {
		dir := pointDirections[i]

		rot := quaternion.RotationTo(lastDir, dir)
		// rot := quaternion.New(vector3.Zero[float64](), 1)
		// if dir.Dot(vector3.Down[float64]()) < 0.9999999 {
		// 	rot = quaternion.RotationTo(vector3.Up[float64](), dir)
		// }

		for sideIndex := 0; sideIndex < vertCount; sideIndex++ {
			point := circlePoints[sideIndex]
			point = lastRot.Rotate(point)
			point = rot.Rotate(point)

			vertices = append(vertices, point.Scale(p.Thickness).Add(p.Point))
			normals = append(normals, point)
		}

		lastRot = lastRot.Multiply(rot)
		lastDir = dir
	}

	// UVs ====================================================================
	validUVs := true
	for _, p := range points {
		if p.UV == nil {
			validUVs = false
			break
		}
	}

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

func ClosedCircleWithConstantThickness(sides int, thickness float64, path []vector3.Float64) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return polygon(sides, points, true)
}

func CircleWithConstantThickness(sides int, thickness float64, path []vector3.Float64) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return polygon(sides, points, false)
}

func CircleWithThickness(sides int, thickness []float64, path []vector3.Float64) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return polygon(sides, points, false)
}

func ClosedCircleWithThickness(sides int, thickness []float64, path []vector3.Float64) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return polygon(sides, points, true)
}

func Polygon(sides int, points []ExtrusionPoint) modeling.Mesh {
	return polygon(sides, points, false)
}
