package marching

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/trees"
	"github.com/EliCDavis/vector"
)

type thickLinePrimitive struct {
	start, end vector.Vector3
	radius     float64
}

func (l thickLinePrimitive) BoundingBox(atr string) modeling.AABB {
	aabb := modeling.NewAABBFromPoints(l.start, l.end)
	aabb.Expand(l.radius)
	return aabb
}

func (l thickLinePrimitive) ClosestPoint(atr string, point vector.Vector3) vector.Vector3 {
	line3d := geometry.NewLine3D(l.start, l.end)
	return line3d.ClosestPointOnLine(point)
}

func Line(start, end vector.Vector3, radius, strength float64) Field {
	boundsSize := vector.Vector3One().MultByConstant(radius + strength)
	bounds := modeling.NewAABB(start, boundsSize)
	bounds.EncapsulateBounds(modeling.NewAABB(end, boundsSize))
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.Line(start, end, radius).Scale(strength),
		},
	}
}

func MultiSegmentLine(linePoints []vector.Vector3, radius, strength float64) Field {
	if len(linePoints) < 2 {
		panic("can not create a line segment field with less than 2 points")
	}

	thickLines := make([]modeling.Primitive, len(linePoints)-1)
	for i := 1; i < len(linePoints); i++ {
		thickLines[i-1] = &thickLinePrimitive{
			start:  linePoints[i-1],
			end:    linePoints[i],
			radius: radius,
		}
	}

	octree := trees.FromPrimitives(thickLines, modeling.PositionAttribute)
	bounds := modeling.NewAABBFromPoints(linePoints...)
	bounds.Expand(radius)
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: func(v vector.Vector3) float64 {
				if !bounds.Contains(v) {
					return 0
				}

				closestIndex, _ := octree.ClosestPoint(v)

				return sdf.Line(linePoints[closestIndex], linePoints[closestIndex+1], radius).Scale(strength)(v)
			},
		},
	}

	// ==============================================================

	// bounds := modeling.NewAABB(line[0], vector.Vector3Zero())
	// sdfs := make([]sample.Vec3ToFloat, 0, len(line)-1)
	// for i := 1; i < len(line); i++ {
	// 	start := line[i-1]
	// 	end := line[i]

	// 	boundsSize := vector.Vector3One().MultByConstant(radius + strength)
	// 	bounds.EncapsulateBounds(modeling.NewAABB(start, boundsSize))
	// 	bounds.EncapsulateBounds(modeling.NewAABB(end, boundsSize))

	// 	sdfs = append(sdfs, sdf.Line(start, end, radius).Scale(strength))
	// }

	// return Field{
	// 	Domain: bounds,
	// 	Float1Functions: map[string]sample.Vec3ToFloat{
	// 		modeling.PositionAttribute: func(v vector.Vector3) float64 {
	// 			min := math.MaxFloat64
	// 			for _, sdf := range sdfs {
	// 				min = math.Min(min, sdf(v))
	// 			}
	// 			return min
	// 		},
	// 	},
	// }
}
