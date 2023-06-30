package noise

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/EliCDavis/vector/vector2"
)

type DistanceField struct {
	points         []vector2.Float64
	xCells, yCells int
	size           vector2.Float64
	spacing        vector2.Float64
}

func NewDistanceField(xCells, yCells int, size vector2.Float64) *DistanceField {
	if xCells <= 0 {
		panic(fmt.Errorf("invalid distance field x cell count: %d", xCells))
	}

	if yCells <= 0 {
		panic(fmt.Errorf("invalid distance field y cell count: %d", yCells))
	}

	if size.X() <= 0 {
		panic(fmt.Errorf("invalid distance field width: %f", size.X()))
	}

	if size.Y() <= 0 {
		panic(fmt.Errorf("invalid distance field height: %f", size.Y()))
	}

	spacing := vector2.New(
		size.X()/float64(xCells),
		size.Y()/float64(yCells),
	)

	points := make([]vector2.Float64, xCells*yCells)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for y := 0; y < yCells; y++ {
		for x := 0; x < xCells; x++ {
			points[(xCells*y)+x] = vector2.
				Rand(r).
				MultByVector(spacing).
				Add(spacing.MultByVector(vector2.New(float64(x), float64(y))))
		}
	}

	return &DistanceField{
		points:  points,
		xCells:  xCells,
		yCells:  yCells,
		size:    size,
		spacing: spacing,
	}
}

func (df DistanceField) point(x, y int) vector2.Float64 {
	offset := vector2.Zero[float64]()

	cleanX := x
	if cleanX < 0 {
		cleanX = df.xCells + x
		offset = offset.SetX(-df.size.X())
	} else if cleanX >= df.xCells {
		cleanX = cleanX - df.xCells
		offset = offset.SetX(df.size.X())
	}

	cleanY := y
	if cleanY < 0 {
		cleanY = df.yCells + y
		offset = offset.SetY(-df.size.Y())
	} else if cleanY >= df.yCells {
		cleanY = cleanY - df.yCells
		offset = offset.SetY(df.size.Y())
	}

	return df.points[(df.xCells*cleanY)+cleanX].Add(offset)
}

func (df DistanceField) Sample(in vector2.Float64) float64 {
	cellX := int(math.Floor(in.X() / (df.size.X() / float64(df.xCells))))
	cellY := int(math.Floor(in.Y() / (df.size.Y() / float64(df.yCells))))
	minDist := math.MaxFloat64
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			dist := df.point(cellX+x, cellY+y).Distance(in)
			if dist < minDist {
				minDist = dist
			}
		}
	}
	return minDist
}
