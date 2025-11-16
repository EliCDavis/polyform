package sdf

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

// Field meant to represent "nothing".
func nullField(f vector3.Float64) float64 {
	return math.Inf(1)
}
