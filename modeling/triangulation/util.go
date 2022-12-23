package triangulation

import "github.com/EliCDavis/vector"

// https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
func intersection(p0, p1, p2, p3 vector.Vector2) (bool, vector.Vector2) {
	s1 := p1.Sub(p0)
	s2 := p3.Sub(p2)

	div := (-s2.X()*s1.Y() + s1.X()*s2.Y())

	if div == 0 {
		return false, vector.Vector2Zero()
	}

	s := (-s1.Y()*(p0.X()-p2.X()) + s1.X()*(p0.Y()-p2.Y())) / div
	t := (s2.X()*(p0.Y()-p2.Y()) - s2.Y()*(p0.X()-p2.X())) / div

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		return true, p0.Add(s1.MultByConstant(t))
	}

	return false, vector.Vector2Zero()
}
