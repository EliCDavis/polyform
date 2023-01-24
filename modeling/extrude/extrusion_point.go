package extrude

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ExtrusionPoint struct {
	Point       vector3.Float64
	Thickness   float64
	UvPoint     vector2.Float64
	UvThickness float64
}
