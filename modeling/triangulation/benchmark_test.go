package triangulation_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
)

func overlappingCircles(rows int) []geometry.Shape {
	var circles []geometry.Shape
	for x := range rows {
		for y := range rows {
			circle := make(geometry.Shape, 32)
			for i := range circle {
				angle := 2 * math.Pi * float64(i) / float64(len(circle))
				circle[i] = vector2.New(float64(x)+0.8*math.Cos(angle), float64(y)+0.8*math.Sin(angle))
			}
			circles = append(circles, circle)
		}
	}
	return circles
}

func BenchmarkFillWindingOverlappingCircles(b *testing.B) {
	circles := overlappingCircles(10)
	for i := 0; i < b.N; i++ {
		if _, err := triangulation.FillWinding([][]geometry.Shape{circles}, func(w []int) bool { return w[0] != 0 }); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFillDifferenceBoardWithHoles(b *testing.B) {
	board := []geometry.Shape{{vector2.New(-1., -1.), vector2.New(10., -1.), vector2.New(10., 10.), vector2.New(-1., 10.)}}
	holes := overlappingCircles(10)
	for i := 0; i < b.N; i++ {
		if _, err := triangulation.FillDifference(board, holes); err != nil {
			b.Fatal(err)
		}
	}
}
