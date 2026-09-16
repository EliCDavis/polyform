package predicate

import (
	"math"
	"math/big"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const filter = 1e-10

// big.Float panics on NaN/infinity
func representable(values ...float64) bool {
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

// Derive required precision based on inputs
//
// TODO: replace with Shewchuk's adaptive predicates.
type arithmetic struct {
	precision uint
}

func exactArithmeticFor(values ...float64) arithmetic {
	lowest, highest := math.MaxInt32, math.MinInt32
	for _, v := range values {
		if v == 0 {
			continue
		}
		_, exponent := math.Frexp(v)
		lowest = min(lowest, exponent)
		highest = max(highest, exponent)
	}
	spread := 0
	if highest > lowest {
		spread = highest - lowest
	}
	return arithmetic{precision: uint(4*(54+spread) + 8)}
}

func (a arithmetic) float(f float64) *big.Float {
	return new(big.Float).SetPrec(a.precision).SetFloat64(f)
}

func (a arithmetic) sub(x, y float64) *big.Float {
	return new(big.Float).SetPrec(a.precision).Sub(a.float(x), a.float(y))
}

func (a arithmetic) mul(x, y *big.Float) *big.Float {
	return new(big.Float).SetPrec(a.precision).Mul(x, y)
}

func (a arithmetic) add(x, y *big.Float) *big.Float {
	return new(big.Float).SetPrec(a.precision).Add(x, y)
}

func (a arithmetic) minus(x, y *big.Float) *big.Float {
	return new(big.Float).SetPrec(a.precision).Sub(x, y)
}

// Orient2D is positive when a, b, c wind counter clockwise, negative when
// clockwise, and exactly zero when collinear. The magnitude is twice the
// signed area of abc.
func Orient2D(a, b, c vector2.Float64) float64 {
	left := (b.X() - a.X()) * (c.Y() - a.Y())
	right := (c.X() - a.X()) * (b.Y() - a.Y())
	det := left - right

	if magnitude := math.Abs(left) + math.Abs(right); math.Abs(det) > filter*magnitude {
		return det
	}
	if !representable(a.X(), a.Y(), b.X(), b.Y(), c.X(), c.Y()) {
		return math.NaN()
	}
	return orient2DExact(a, b, c)
}

func orient2DExact(a, b, c vector2.Float64) float64 {
	x := exactArithmeticFor(a.X(), a.Y(), b.X(), b.Y(), c.X(), c.Y())
	det := x.minus(
		x.mul(x.sub(b.X(), a.X()), x.sub(c.Y(), a.Y())),
		x.mul(x.sub(c.X(), a.X()), x.sub(b.Y(), a.Y())),
	)
	answer, _ := det.Float64()
	return answer
}

// Orient3D reports which side of the plane through the corners the point
// falls on: positive, negative, or exactly zero when it lies in the plane.
func Orient3D(corner1, corner2, corner3, point vector3.Float64) float64 {
	toPointX, toPointY, toPointZ := point.X()-corner1.X(), point.Y()-corner1.Y(), point.Z()-corner1.Z()
	toCorner2X, toCorner2Y, toCorner2Z := corner2.X()-corner1.X(), corner2.Y()-corner1.Y(), corner2.Z()-corner1.Z()
	toCorner3X, toCorner3Y, toCorner3Z := corner3.X()-corner1.X(), corner3.Y()-corner1.Y(), corner3.Z()-corner1.Z()

	normalX := toCorner2Y*toCorner3Z - toCorner2Z*toCorner3Y
	normalY := toCorner2Z*toCorner3X - toCorner2X*toCorner3Z
	normalZ := toCorner2X*toCorner3Y - toCorner2Y*toCorner3X

	det := toPointX*normalX + toPointY*normalY + toPointZ*normalZ
	magnitude := math.Abs(toPointX*normalX) + math.Abs(toPointY*normalY) + math.Abs(toPointZ*normalZ)

	if math.Abs(det) > filter*magnitude {
		return det
	}
	if !representable(
		corner1.X(), corner1.Y(), corner1.Z(), corner2.X(), corner2.Y(), corner2.Z(),
		corner3.X(), corner3.Y(), corner3.Z(), point.X(), point.Y(), point.Z()) {
		return math.NaN()
	}
	return orient3DExact(corner1, corner2, corner3, point)
}

func orient3DExact(corner1, corner2, corner3, point vector3.Float64) float64 {
	x := exactArithmeticFor(
		corner1.X(), corner1.Y(), corner1.Z(), corner2.X(), corner2.Y(), corner2.Z(),
		corner3.X(), corner3.Y(), corner3.Z(), point.X(), point.Y(), point.Z())
	toPointX, toPointY, toPointZ := x.sub(point.X(), corner1.X()), x.sub(point.Y(), corner1.Y()), x.sub(point.Z(), corner1.Z())
	toCorner2X, toCorner2Y, toCorner2Z := x.sub(corner2.X(), corner1.X()), x.sub(corner2.Y(), corner1.Y()), x.sub(corner2.Z(), corner1.Z())
	toCorner3X, toCorner3Y, toCorner3Z := x.sub(corner3.X(), corner1.X()), x.sub(corner3.Y(), corner1.Y()), x.sub(corner3.Z(), corner1.Z())

	det := x.add(x.add(
		x.mul(toPointX, x.minus(x.mul(toCorner2Y, toCorner3Z), x.mul(toCorner2Z, toCorner3Y))),
		x.mul(toPointY, x.minus(x.mul(toCorner2Z, toCorner3X), x.mul(toCorner2X, toCorner3Z)))),
		x.mul(toPointZ, x.minus(x.mul(toCorner2X, toCorner3Y), x.mul(toCorner2Y, toCorner3X))))

	answer, _ := det.Float64()
	return answer
}

// InCircle is positive when point is inside the circle through the corners,
// negative when outside, and exactly zero when all four are cocircular. The
// corners must wind counter clockwise or the sign inverts.
func InCircle(corner1, corner2, corner3, point vector2.Float64) float64 {
	toCorner1X, toCorner1Y := corner1.X()-point.X(), corner1.Y()-point.Y()
	toCorner2X, toCorner2Y := corner2.X()-point.X(), corner2.Y()-point.Y()
	toCorner3X, toCorner3Y := corner3.X()-point.X(), corner3.Y()-point.Y()

	minor1 := toCorner2X*toCorner3Y - toCorner3X*toCorner2Y
	minor2 := toCorner3X*toCorner1Y - toCorner1X*toCorner3Y
	minor3 := toCorner1X*toCorner2Y - toCorner2X*toCorner1Y

	lift1 := toCorner1X*toCorner1X + toCorner1Y*toCorner1Y
	lift2 := toCorner2X*toCorner2X + toCorner2Y*toCorner2Y
	lift3 := toCorner3X*toCorner3X + toCorner3Y*toCorner3Y

	det := lift1*minor1 + lift2*minor2 + lift3*minor3

	magnitude := lift1*(math.Abs(toCorner2X*toCorner3Y)+math.Abs(toCorner3X*toCorner2Y)) +
		lift2*(math.Abs(toCorner3X*toCorner1Y)+math.Abs(toCorner1X*toCorner3Y)) +
		lift3*(math.Abs(toCorner1X*toCorner2Y)+math.Abs(toCorner2X*toCorner1Y))

	if math.Abs(det) > filter*magnitude {
		return det
	}
	if !representable(
		corner1.X(), corner1.Y(), corner2.X(), corner2.Y(),
		corner3.X(), corner3.Y(), point.X(), point.Y()) {
		return math.NaN()
	}
	return inCircleExact(corner1, corner2, corner3, point)
}

func inCircleExact(corner1, corner2, corner3, point vector2.Float64) float64 {
	x := exactArithmeticFor(
		corner1.X(), corner1.Y(), corner2.X(), corner2.Y(),
		corner3.X(), corner3.Y(), point.X(), point.Y())
	toCorner1X, toCorner1Y := x.sub(corner1.X(), point.X()), x.sub(corner1.Y(), point.Y())
	toCorner2X, toCorner2Y := x.sub(corner2.X(), point.X()), x.sub(corner2.Y(), point.Y())
	toCorner3X, toCorner3Y := x.sub(corner3.X(), point.X()), x.sub(corner3.Y(), point.Y())

	minor := func(firstX, firstY, secondX, secondY *big.Float) *big.Float {
		return x.minus(x.mul(firstX, secondY), x.mul(secondX, firstY))
	}

	lift1 := x.add(x.mul(toCorner1X, toCorner1X), x.mul(toCorner1Y, toCorner1Y))
	lift2 := x.add(x.mul(toCorner2X, toCorner2X), x.mul(toCorner2Y, toCorner2Y))
	lift3 := x.add(x.mul(toCorner3X, toCorner3X), x.mul(toCorner3Y, toCorner3Y))

	det := x.add(x.add(
		x.mul(lift1, minor(toCorner2X, toCorner2Y, toCorner3X, toCorner3Y)),
		x.mul(lift2, minor(toCorner3X, toCorner3Y, toCorner1X, toCorner1Y))),
		x.mul(lift3, minor(toCorner1X, toCorner1Y, toCorner2X, toCorner2Y)))

	answer, _ := det.Float64()
	return answer
}

// SegmentsCross reports whether two segments pass through each other.
// Touching, at an endpoint or along the other's interior, does not count.
func SegmentsCross(firstStart, firstEnd, secondStart, secondEnd vector2.Float64) bool {
	secondStartSide := Orient2D(firstStart, firstEnd, secondStart)
	secondEndSide := Orient2D(firstStart, firstEnd, secondEnd)
	firstStartSide := Orient2D(secondStart, secondEnd, firstStart)
	firstEndSide := Orient2D(secondStart, secondEnd, firstEnd)

	return ((secondStartSide > 0) != (secondEndSide > 0)) &&
		((firstStartSide > 0) != (firstEndSide > 0)) &&
		secondStartSide != 0 && secondEndSide != 0 &&
		firstStartSide != 0 && firstEndSide != 0
}
