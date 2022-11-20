package triangulation

import (
	"log"
	"math"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

type Edge [2]int

func (e Edge) Length(points []vector.Vector2) float64 {
	return points[e[0]].Sub(points[e[1]]).Length()
}

func (e Edge) Connected(other Edge) bool {
	return e[1] == other[0] || e[0] == other[1]
}

type Triangle [3]int

func (t Triangle) Edges() []Edge {
	return []Edge{{t[0], t[1]}, {t[1], t[2]}, {t[2], t[0]}}
}

func (t Triangle) CounterClockwise(points []vector.Vector2) bool {
	a := points[t[0]]
	b := points[t[1]]
	c := points[t[2]]
	return (b.X()-a.X())*(c.Y()-a.Y())-(c.X()-a.X())*(b.Y()-a.Y()) > 0
}

func (t Triangle) InsideCircumcircle(p vector.Vector2, points []vector.Vector2) bool {
	// edges := t.Edges()
	// a := edges[0].Length(points)
	// b := edges[1].Length(points)
	// c := edges[2].Length(points)
	// radius := (a * b * c) / math.Sqrt((a+b+c)*(b+c-a)*(c+a-b)*(a+b-c))

	// https://stackoverflow.com/questions/39984709/how-can-i-check-wether-a-point-is-inside-the-circumcircle-of-3-points
	a := points[t[0]]
	b := points[t[1]]
	c := points[t[2]]

	ax_ := a.X() - p.X()
	ay_ := a.Y() - p.Y()
	bx_ := b.X() - p.X()
	by_ := b.Y() - p.Y()
	cx_ := c.X() - p.X()
	cy_ := c.Y() - p.Y()

	det := ((ax_*ax_+ay_*ay_)*(bx_*cy_-cx_*by_) -
		(bx_*bx_+by_*by_)*(ax_*cy_-cx_*ay_) +
		(cx_*cx_+cy_*cy_)*(ax_*by_-bx_*ay_))

	// log.Print(a, b, c, p, det, ccw(a, b, c))

	return det < 0
}

func ccw(a, b, c vector.Vector2) bool {
	return (b.X()-a.X())*(c.Y()-a.Y())-(c.X()-a.X())*(b.Y()-a.Y()) > 0
}

// SuperTriangle large enough to completely contain all the points in pointList
//
// TODO: This is just a guess. I've never really taken time to confirm this is
// a valid construction
func SuperTriangle(points []vector.Vector2) []vector.Vector2 {
	min := vector.NewVector2(math.Inf(1), math.Inf(1))
	max := vector.NewVector2(math.Inf(-1), math.Inf(-1))

	for _, v := range points {
		min = vector.NewVector2(
			math.Min(v.X(), min.X()),
			math.Min(v.Y(), min.Y()),
		)
		max = vector.NewVector2(
			math.Max(v.X(), max.X()),
			math.Max(v.Y(), max.Y()),
		)
	}

	height := max.Y() - min.Y()
	min = vector.NewVector2(min.X(), min.Y()-2)

	xMiddle := (min.X() + max.X()) / 2.
	width := max.X() - min.X()

	top := vector.NewVector2(
		xMiddle,
		min.Y()+(height*20),
	)

	left := vector.NewVector2(
		xMiddle-(width*20),
		min.Y(),
	)

	right := vector.NewVector2(
		xMiddle+(width*20),
		min.Y(),
	)
	return []vector.Vector2{left, top, right}
}

func containsSuperTriangleVertex(t Triangle, points []vector.Vector2) bool {
	superStart := len(points) - 3

	if t[0] >= superStart {
		return true
	}
	if t[1] >= superStart {
		return true
	}

	return t[2] >= superStart
}

var exists = struct{}{}

type SortByXComponent []vector.Vector2

func (a SortByXComponent) Len() int           { return len(a) }
func (a SortByXComponent) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByXComponent) Less(i, j int) bool { return a[i].X() < a[j].X() }

func fillHole(polygon []Edge, point int, triangulation map[Triangle]struct{}, points []vector.Vector2) {
	for _, edge := range polygon {

		if edge[0] == point || edge[1] == point {
			continue
		}

		triToAdd := Triangle{edge[0], edge[1], point}
		if triToAdd.CounterClockwise(points) {
			triToAdd = Triangle{edge[0], point, edge[1]}
			log.Print(triToAdd.CounterClockwise(points))
		}

		triangulation[triToAdd] = exists

	}
}

func BowyerWatson(pointsDirty []vector.Vector2) mesh.Mesh {
	if len(pointsDirty) < 3 {
		panic("can not tesselate without at least 3 points")
	}

	// TODO: Eli do a copy of the array instead so passed in points are not
	// effected by the sort
	points := pointsDirty
	// sort.Sort(SortByXComponent(points))

	points = append(pointsDirty, SuperTriangle(pointsDirty)...)
	triangulation := make(map[Triangle]struct{})
	triangulation[Triangle{len(points) - 3, len(points) - 2, len(points) - 1}] = exists

	for pi, point := range points {

		// We're at super triangle now
		if pi >= len(points)-3 {
			break
		}

		badTriangles := make([]Triangle, 0)

		// first find all the triangles that are no longer valid due to the insertion
		for triangle := range triangulation {
			if triangle.InsideCircumcircle(point, points) {
				badTriangles = append(badTriangles, triangle)
			}
		}

		polygon := make([]Edge, 0)

		// find the boundary of the polygonal hole
		for ti, triangle := range badTriangles {
			// edge is not shared by any other triangles in badTriangles
			for _, edge := range triangle.Edges() {

				notShared := true

				for oti, otherTriangle := range badTriangles {
					if ti != oti && notShared {
						for _, otherEdge := range otherTriangle.Edges() {
							if edge[0] == otherEdge[0] {
								if edge[1] == otherEdge[1] {
									notShared = false
								}
							}

							if edge[0] == otherEdge[1] {
								if edge[1] == otherEdge[0] {
									notShared = false
								}
							}
						}

						// if !notShared {
						// 	break
						// }
					}
				}

				if notShared {
					polygon = append(polygon, edge)
				}
			}
		}

		// remove them from the data structure
		for _, triangle := range badTriangles {
			delete(triangulation, triangle)
		}

		// re-triangulate the polygonal hole
		fillHole(polygon, pi, triangulation, points)
	}

	// done inserting points, now clean up
	for triangle := range triangulation {
		if containsSuperTriangleVertex(triangle, points) {
			delete(triangulation, triangle)
		}
	}

	tris := make([]int, 0, len(triangulation)*3)
	for triangle := range triangulation {
		tris = append(tris, triangle[0], triangle[1], triangle[2])
	}

	verts := make([]vector.Vector3, len(pointsDirty))
	uvs := make([]vector.Vector2, len(pointsDirty))
	for i, p := range pointsDirty {
		verts[i] = vector.NewVector3(p.X(), 0, p.Y())
		uvs[i] = vector.Vector2Zero()
	}

	return mesh.NewMesh(
		tris,
		verts,
		nil,
		[][]vector.Vector2{uvs},
	)
}
