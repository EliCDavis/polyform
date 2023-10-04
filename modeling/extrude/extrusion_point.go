package extrude

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ExtrusionPoint struct {
	Point     vector3.Float64
	Thickness float64
	UV        *ExtrusionPointUV
	Direction *ExtrusionPointDirection
}

type ExtrusionPointUV struct {
	Point     vector2.Float64
	Thickness float64
}

type ExtrusionPointDirection struct {
	Direction vector3.Float64
}

func directionsOfExtrusionPoints(points []ExtrusionPoint) []vector3.Float64 {
	if len(points) == 0 {
		return nil
	}

	if len(points) == 1 {
		dir := vector3.Up[float64]()

		if points[0].Direction != nil {
			dir = points[0].Direction.Direction
		}

		return []vector3.Vector[float64]{
			dir,
		}
	}

	directions := make([]vector3.Float64, len(points))

	// leftPoints := make([]vector3.Float64, len(points))
	// rightPoints := make([]vector3.Float64, len(points))
	// for i, point := range points {

	// }

	// for i, point := range points {
	// 	if i == 0 {
	// 		directions[i] = points[1].Point.Sub(point.Point).Normalized()
	// 		continue
	// 	}

	// 	if i == len(points)-1 {
	// 		directions[i] = point.Point.Sub(points[i-1].Point).Normalized()
	// 		continue
	// 	}

	// 	dirA := point.Point.Sub(points[i-1].Point).Normalized()
	// 	dirB := points[i+1].Point.Sub(point.Point).Normalized()
	// 	directions[i] = dirA.Add(dirB).Normalized()
	// }

	return directions
}
