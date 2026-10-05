package triangulation

import (
	"fmt"
	"slices"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
)

type Region struct {
	Points    []vector2.Float64
	Triangles []Triangle
}

// Fill triangulates the area the contours enclose by nonzero winding:
// overlapping contours union, and a contour wound against the one around it
// is a hole.
func Fill(contours ...geometry.Shape) (Region, error) {
	weld, segments := contourSegments(contours)
	if len(segments) == 0 {
		return Region{}, nil
	}
	tess, err := triangulate(weld, segments)
	if err != nil {
		return Region{}, err
	}
	tess.discardOutside()
	return tess.region(nil), nil
}

// FillDifference triangulates the area inside any of keep and outside every
// one of cut, however they overlap. Winding plays no part.
func FillDifference(keep, cut []geometry.Shape) (Region, error) {
	if len(keep) == 0 {
		return Region{}, nil
	}
	if !slices.ContainsFunc(keep, func(s geometry.Shape) bool { return s.SignedArea() != 0 }) {
		return Region{}, fmt.Errorf("none of the %d outlines to fill encloses any area", len(keep))
	}

	tess, err := triangulate(contourSegments(append(slices.Clone(keep), cut...)))
	if err != nil {
		return Region{}, err
	}

	// Every contour edge survives whole, so no triangle straddles one and a
	// single point inside it decides for all of it.
	return tess.region(func(tri Triangle) bool {
		centroid := tess.pts[tri[0]].Add(tess.pts[tri[1]]).Add(tess.pts[tri[2]]).Scale(1. / 3)
		return insideAny(keep, centroid) && !insideAny(cut, centroid)
	}), nil
}

func insideAny(shapes []geometry.Shape, p vector2.Float64) bool {
	return slices.ContainsFunc(shapes, func(s geometry.Shape) bool { return s.IsInside(p) })
}

func contourSegments(contours []geometry.Shape) (*geometry.PointWelder2D, [][2]int) {
	sets := make([][]vector2.Float64, len(contours))
	for i, contour := range contours {
		sets[i] = contour
	}
	weld := geometry.NewPointWelder2D(mergeTolerance(sets...))

	var segments [][2]int
	for _, contour := range contours {
		for i := range contour {
			a, b := weld.Index(contour[i]), weld.Index(contour[(i+1)%len(contour)])
			if a != b {
				segments = append(segments, [2]int{a, b})
			}
		}
	}
	return weld, segments
}

func (t *tessellation) region(kept func(Triangle) bool) Region {
	var triangles []Triangle
	used := make([]bool, len(t.pts))
	for _, tri := range t.ordered(func(tri Triangle) Triangle {
		clockwise := tri.clockwise(t.pts)
		return Triangle{clockwise[0], clockwise[2], clockwise[1]}
	}) {
		if kept != nil && !kept(tri) {
			continue
		}
		triangles = append(triangles, tri)
		for _, v := range tri {
			used[v] = true
		}
	}

	remap := make([]int, len(t.pts))
	var points []vector2.Float64
	for i, p := range t.pts {
		if used[i] {
			remap[i] = len(points)
			points = append(points, p)
		}
	}
	for i, tri := range triangles {
		triangles[i] = Triangle{remap[tri[0]], remap[tri[1]], remap[tri[2]]}
	}
	return Region{Points: points, Triangles: triangles}
}
