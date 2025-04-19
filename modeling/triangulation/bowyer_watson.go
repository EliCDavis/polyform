package triangulation

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Edge [2]int

func (e Edge) Length(points []vector2.Float64) float64 {
	return points[e[0]].Sub(points[e[1]]).Length()
}

func (e Edge) Connected(other Edge) bool {
	return e[1] == other[0] || e[0] == other[1]
}

type Triangle [3]int

func (t Triangle) Edges() []Edge {
	return []Edge{{t[0], t[1]}, {t[1], t[2]}, {t[2], t[0]}}
}

func (t Triangle) CounterClockwise(points []vector2.Float64) bool {
	a := points[t[0]]
	b := points[t[1]]
	c := points[t[2]]
	return (b.X()-a.X())*(c.Y()-a.Y())-(c.X()-a.X())*(b.Y()-a.Y()) > 0
}

func (t Triangle) Intersects(points []vector2.Float64, start, end vector2.Float64) []vector2.Float64 {
	intersections := make([]vector2.Float64, 0)

	intersects, point := intersection(points[t[0]], points[t[1]], start, end)
	if intersects {
		intersections = append(intersections, point)
	}

	intersects, point = intersection(points[t[1]], points[t[2]], start, end)
	if intersects {
		intersections = append(intersections, point)
	}

	intersects, point = intersection(points[t[2]], points[t[0]], start, end)
	if intersects {
		intersections = append(intersections, point)
	}

	return intersections
}

func (t Triangle) InsideCircumcircle(p vector2.Float64, points []vector2.Float64) bool {
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

func ccw(a, b, c vector2.Float64) bool {
	return (b.X()-a.X())*(c.Y()-a.Y())-(c.X()-a.X())*(b.Y()-a.Y()) > 0
}

// SuperTriangle large enough to completely contain all the points in pointList
//
// TODO: This is just a guess. I've never really taken time to confirm this is
// a valid construction
func SuperTriangle(points []vector2.Float64) []vector2.Float64 {
	min := vector2.New(math.Inf(1), math.Inf(1))
	max := vector2.New(math.Inf(-1), math.Inf(-1))

	for _, v := range points {
		min = vector2.New(
			math.Min(v.X(), min.X()),
			math.Min(v.Y(), min.Y()),
		)
		max = vector2.New(
			math.Max(v.X(), max.X()),
			math.Max(v.Y(), max.Y()),
		)
	}

	height := max.Y() - min.Y()
	min = vector2.New(min.X(), min.Y()-2)

	xMiddle := (min.X() + max.X()) / 2.
	width := max.X() - min.X()

	top := vector2.New(
		xMiddle,
		min.Y()+(height*20),
	)

	left := vector2.New(
		xMiddle-(width*20),
		min.Y(),
	)

	right := vector2.New(
		xMiddle+(width*20),
		min.Y(),
	)
	return []vector2.Float64{left, top, right}
}

func containsSuperTriangleVertex(t Triangle, points []vector2.Float64) bool {
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

type SortByXComponent []vector2.Float64

func (a SortByXComponent) Len() int           { return len(a) }
func (a SortByXComponent) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByXComponent) Less(i, j int) bool { return a[i].X() < a[j].X() }

func fillHole(polygon []Edge, point int, triangulation map[Triangle]struct{}, points []vector2.Float64) {
	for _, edge := range polygon {

		if edge[0] == point || edge[1] == point {
			continue
		}

		triToAdd := Triangle{edge[0], edge[1], point}
		if triToAdd.CounterClockwise(points) {
			triToAdd = Triangle{edge[0], point, edge[1]}
			// log.Print(triToAdd.CounterClockwise(points))
		}

		triangulation[triToAdd] = exists

	}
}

func bowyerWatson(pointsDirty []vector2.Float64) map[Triangle]struct{} {
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

	return triangulation
}

func ConstrainedBowyerWatson(pointsDirty []vector2.Float64, constraints []Constraint) modeling.Mesh {
	finalPoints := pointsDirty
	// finalPoints = append(finalPoints, constraints[0].shape...)
	triangulation := bowyerWatson(finalPoints)
	trisToAdd := make(map[Triangle]struct{})

	for _, constraint := range constraints {
		for triangle := range triangulation {

			containsP1 := constraint.contains(finalPoints[triangle[0]])
			containsP2 := constraint.contains(finalPoints[triangle[1]])
			containsP3 := constraint.contains(finalPoints[triangle[2]])

			totalPointsContained := containsP1 + containsP2 + containsP3

			if totalPointsContained == 0 {
				delete(triangulation, triangle)
				continue
			}

			if totalPointsContained == 3 {
				continue
			}

			for i := range constraint.shape {
				edgeEnd := constraint.shape[i]

				left := i - 1
				if i == 0 {
					left = len(constraint.shape) - 1
				}

				edgeStart := constraint.shape[left]

				intersections := triangle.Intersects(
					finalPoints,
					edgeStart,
					edgeEnd,
				)

				if len(intersections) != 2 {
					continue
				}

				delete(triangulation, triangle)

				// log.Println(totalPointsContained, len(intersections))

				if totalPointsContained == 1 {
					finalPoints = append(finalPoints, intersections...)
					pointContained := triangle[0]
					if containsP2 == 1 {
						pointContained = triangle[1]
					}
					if containsP3 == 1 {
						pointContained = triangle[2]
					}
					tri := Triangle{pointContained, len(finalPoints) - 1, len(finalPoints) - 2}
					if tri.CounterClockwise(finalPoints) {
						tri = Triangle{pointContained, len(finalPoints) - 2, len(finalPoints) - 1}
					}
					trisToAdd[tri] = exists
				} else {

					km1 := triangle[2]
					k := triangle[0]
					kp1 := triangle[1]

					if containsP2 == 0 {
						km1 = triangle[0]
						k = triangle[1]
						kp1 = triangle[2]
					}
					if containsP3 == 0 {
						km1 = triangle[1]
						k = triangle[2]
						kp1 = triangle[0]
					}

					_, cPoint := intersection(finalPoints[km1], finalPoints[k], edgeStart, edgeEnd)
					_, dPoint := intersection(finalPoints[k], finalPoints[kp1], edgeStart, edgeEnd)
					finalPoints = append(finalPoints, cPoint, dPoint)

					trisToAdd[Triangle{km1, len(finalPoints) - 2, kp1}] = exists
					trisToAdd[Triangle{len(finalPoints) - 2, len(finalPoints) - 1, kp1}] = exists
				}
			}
		}
	}

	tris := make([]int, 0, len(triangulation)*3)
	for triangle := range triangulation {
		tris = append(tris, triangle[0], triangle[1], triangle[2])
	}
	for triangle := range trisToAdd {
		tris = append(tris, triangle[0], triangle[1], triangle[2])
	}

	verts := make([]vector3.Float64, len(finalPoints))
	uvs := make([]vector2.Float64, len(finalPoints))
	for i, p := range finalPoints {
		verts[i] = vector3.New(p.X(), 0, p.Y())
		uvs[i] = vector2.Zero[float64]()
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func BowyerWatson(pointsDirty []vector2.Float64) modeling.Mesh {
	triangulation := bowyerWatson(pointsDirty)
	tris := make([]int, 0, len(triangulation)*3)
	for triangle := range triangulation {
		tris = append(tris, triangle[0], triangle[1], triangle[2])
	}

	verts := make([]vector3.Float64, len(pointsDirty))
	uvs := make([]vector2.Float64, len(pointsDirty))
	for i, p := range pointsDirty {
		verts[i] = vector3.New(p.X(), 0, p.Y())
		uvs[i] = vector2.Zero[float64]()
	}

	return modeling.
		NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}
