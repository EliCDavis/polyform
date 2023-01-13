package marching

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

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

func MultiSegmentLine(line []vector.Vector3, radius, strength float64) Field {
	if len(line) < 1 {
		panic("can not create a line segment field with less than 2 points")
	}

	bounds := modeling.NewAABB(line[0], vector.Vector3Zero())
	sdfs := make([]sample.Vec3ToFloat, 0, len(line)-1)
	for i := 1; i < len(line); i++ {
		start := line[i-1]
		end := line[i]

		boundsSize := vector.Vector3One().MultByConstant(radius + strength)
		bounds.EncapsulateBounds(modeling.NewAABB(start, boundsSize))
		bounds.EncapsulateBounds(modeling.NewAABB(end, boundsSize))

		sdfs = append(sdfs, sdf.Line(start, end, radius).Scale(strength))
	}

	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: func(v vector.Vector3) float64 {
				min := math.MaxFloat64
				for _, sdf := range sdfs {
					min = math.Min(min, sdf(v))
				}
				return min
			},
		},
	}
}
