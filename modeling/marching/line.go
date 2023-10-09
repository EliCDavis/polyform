package marching

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type thickLinePrimitive struct {
	start, end vector3.Float64
	radius     float64
}

func (l thickLinePrimitive) BoundingBox() geometry.AABB {
	aabb := geometry.NewAABBFromPoints(l.start, l.end)
	aabb.Expand(l.radius * math.Sqrt2 * 2)
	return aabb
}

func (l thickLinePrimitive) ClosestPoint(point vector3.Float64) vector3.Float64 {
	line3d := geometry.NewLine3D(l.start, l.end)
	return line3d.ClosestPointOnLine(point)
}

func Line(start, end vector3.Float64, radius, strength float64) Field {
	bounds := geometry.NewAABBFromPoints(start, end)
	bounds.Expand(radius * 2)
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Line(start, end, radius).Scale(strength),
		},
	}
}

func MultiSegmentLine(linePoints []vector3.Float64, radius, strength float64) Field {
	if len(linePoints) < 2 {
		panic("can not create a line segment field with less than 2 points")
	}

	thickLines := make([]trees.Element, len(linePoints)-1)
	for i := 1; i < len(linePoints); i++ {
		thickLines[i-1] = &thickLinePrimitive{
			start:  linePoints[i-1],
			end:    linePoints[i],
			radius: radius,
		}
	}

	octree := trees.NewOctree(thickLines)
	bounds := geometry.NewAABBFromPoints(linePoints...)
	bounds.Expand(radius * math.Sqrt2 * 2)
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: func(v vector3.Float64) float64 {
				lineIndexes := octree.ElementsContainingPoint(v)

				min := math.MaxFloat64
				for _, l := range lineIndexes {
					val := sdf.Line(linePoints[l], linePoints[l+1], radius).Scale(strength)(v)
					min = math.Min(min, val)
				}
				return min
			},
		},
	}
}

func VarryingThicknessLine(linePoints []sdf.LinePoint, strength float64) Field {
	if len(linePoints) < 2 {
		panic("can not create a line segment field with less than 2 points")
	}

	bounds := geometry.NewAABB(linePoints[0].Point, vector3.Zero[float64]())
	for i := 1; i < len(linePoints); i++ {
		start := linePoints[i-1]
		end := linePoints[i]

		boundsSize := vector3.Fill(math.Max(start.Radius, end.Radius) + strength)
		bounds.EncapsulateBounds(geometry.NewAABB(start.Point, boundsSize))
		bounds.EncapsulateBounds(geometry.NewAABB(end.Point, boundsSize))

	}

	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.VarryingThicknessLine(linePoints).Scale(strength),
		},
	}
}
