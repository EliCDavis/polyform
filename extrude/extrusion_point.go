package extrude

import "github.com/EliCDavis/vector"

type ExtrusionPoint struct {
	Point       vector.Vector3
	Thickness   float64
	UvPoint     vector.Vector2
	UvThickness float64
}
