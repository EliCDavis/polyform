package geometry

import (
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Each point is filed once, in its home cell; a query reads the cells its
// tolerance ball touches, at most two per axis.
type pointWelder[T comparable] struct {
	tolerance float64
	cell      float64
	dims      int
	coords    func(T) [3]float64
	distance  func(a, b T) float64
	buckets   map[[3]int64][]int
	exact     map[T]int
	points    []T
}

const cellsPerTolerance = 8

func newPointWelder[T comparable](
	tolerance float64,
	dims int,
	coords func(T) [3]float64,
	distance func(a, b T) float64,
) pointWelder[T] {
	w := pointWelder[T]{
		tolerance: tolerance,
		cell:      tolerance * cellsPerTolerance,
		dims:      dims,
		coords:    coords,
		distance:  distance,
	}
	if tolerance > 0 {
		w.buckets = make(map[[3]int64][]int)
	} else {
		w.exact = make(map[T]int)
	}
	return w
}

func (w *pointWelder[T]) cellOf(c [3]float64) [3]int64 {
	var key [3]int64
	for i := 0; i < w.dims; i++ {
		key[i] = int64(math.Floor(c[i] / w.cell))
	}
	return key
}

func (w *pointWelder[T]) find(p T) (int, bool) {
	if w.exact != nil {
		i, ok := w.exact[p]
		return i, ok
	}

	c := w.coords(p)
	var lo, hi [3]float64
	for i := 0; i < w.dims; i++ {
		lo[i] = c[i] - w.tolerance
		hi[i] = c[i] + w.tolerance
	}
	from, to := w.cellOf(lo), w.cellOf(hi)

	var key [3]int64
	for key[0] = from[0]; key[0] <= to[0]; key[0]++ {
		for key[1] = from[1]; key[1] <= to[1]; key[1]++ {
			for key[2] = from[2]; key[2] <= to[2]; key[2]++ {
				for _, i := range w.buckets[key] {
					if w.distance(w.points[i], p) <= w.tolerance {
						return i, true
					}
				}
			}
		}
	}
	return -1, false
}

func (w *pointWelder[T]) add(p T) int {
	id := len(w.points)
	w.points = append(w.points, p)

	if w.exact != nil {
		if _, taken := w.exact[p]; !taken {
			w.exact[p] = id
		}
		return id
	}

	key := w.cellOf(w.coords(p))
	w.buckets[key] = append(w.buckets[key], id)
	return id
}

func (w *pointWelder[T]) index(p T) int {
	if i, ok := w.find(p); ok {
		return i
	}
	return w.add(p)
}

// PointWelder3D hands out one index per distinct point, where points within
// tolerance of one already added count as it. Tolerance zero matches exactly.
type PointWelder3D struct {
	pointWelder[vector3.Float64]
}

func NewPointWelder3D(tolerance float64) *PointWelder3D {
	return &PointWelder3D{newPointWelder(tolerance, 3,
		func(p vector3.Float64) [3]float64 { return [3]float64{p.X(), p.Y(), p.Z()} },
		func(a, b vector3.Float64) float64 { return a.Distance(b) },
	)}
}

// Add records the point under a new index even when one within tolerance exists.
func (w *PointWelder3D) Add(p vector3.Float64) int { return w.add(p) }

func (w *PointWelder3D) Find(p vector3.Float64) (int, bool) { return w.find(p) }

// Index returns the index of the first point added within tolerance, adding p if none is.
func (w *PointWelder3D) Index(p vector3.Float64) int { return w.index(p) }

func (w *PointWelder3D) Points() []vector3.Float64 { return w.points }

type PointWelder2D struct {
	pointWelder[vector2.Float64]
}

func NewPointWelder2D(tolerance float64) *PointWelder2D {
	return &PointWelder2D{newPointWelder(tolerance, 2,
		func(p vector2.Float64) [3]float64 { return [3]float64{p.X(), p.Y(), 0} },
		func(a, b vector2.Float64) float64 { return a.Distance(b) },
	)}
}

func (w *PointWelder2D) Add(p vector2.Float64) int { return w.add(p) }

func (w *PointWelder2D) Find(p vector2.Float64) (int, bool) { return w.find(p) }

func (w *PointWelder2D) Index(p vector2.Float64) int { return w.index(p) }

func (w *PointWelder2D) Points() []vector2.Float64 { return w.points }
