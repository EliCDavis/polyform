package geometry

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Each point is filed once, in its home cell; a query reads the cells its
// tolerance ball touches, at most two per axis.
type pointWelder[T comparable] struct {
	tolerance float64
	cell      float64
	extent    vector3.Float64
	coords    func(T) vector3.Float64
	distance  func(a, b T) float64
	buckets   map[vector3.Int][]int
	exact     map[T]int
	points    []T
}

const cellsPerTolerance = 8

func newPointWelder[T comparable](
	tolerance float64,
	extent vector3.Float64,
	coords func(T) vector3.Float64,
	distance func(a, b T) float64,
) pointWelder[T] {
	w := pointWelder[T]{
		tolerance: tolerance,
		cell:      tolerance * cellsPerTolerance,
		extent:    extent.Scale(tolerance),
		coords:    coords,
		distance:  distance,
	}
	if tolerance > 0 {
		w.buckets = make(map[vector3.Int][]int)
	} else {
		w.exact = make(map[T]int)
	}
	return w
}

func (w *pointWelder[T]) Tolerance() float64 { return w.tolerance }

func (w *pointWelder[T]) cellOf(c vector3.Float64) vector3.Int {
	return c.DivByConstant(w.cell).FloorToInt()
}

func (w *pointWelder[T]) find(p T) (int, bool) {
	if w.exact != nil {
		i, ok := w.exact[p]
		return i, ok
	}

	c := w.coords(p)
	from, to := w.cellOf(c.Sub(w.extent)), w.cellOf(c.Add(w.extent))

	for x := from.X(); x <= to.X(); x++ {
		for y := from.Y(); y <= to.Y(); y++ {
			for z := from.Z(); z <= to.Z(); z++ {
				for _, i := range w.buckets[vector3.New(x, y, z)] {
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
	return &PointWelder3D{newPointWelder(tolerance, vector3.One[float64](),
		func(p vector3.Float64) vector3.Float64 { return p },
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
	return &PointWelder2D{newPointWelder(tolerance, vector3.New(1., 1, 0),
		func(p vector2.Float64) vector3.Float64 { return vector3.New(p.X(), p.Y(), 0) },
		func(a, b vector2.Float64) float64 { return a.Distance(b) },
	)}
}

func (w *PointWelder2D) Add(p vector2.Float64) int { return w.add(p) }

func (w *PointWelder2D) Find(p vector2.Float64) (int, bool) { return w.find(p) }

func (w *PointWelder2D) Index(p vector2.Float64) int { return w.index(p) }

func (w *PointWelder2D) Points() []vector2.Float64 { return w.points }
