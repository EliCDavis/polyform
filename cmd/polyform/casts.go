package main

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/cast"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func casts() *cast.Table {
	table := &cast.Table{}
	registerNumberCasts(table)
	registerNormCasts(table)
	registerTextureCasts(table)
	cast.Register(table, normals.RasterizeNormalmap)
	cast.Register(table, func(f sdf.ColoredField) sample.Vec3ToFloat { return f.Distance })
	cast.Register(table, func(m modeling.Mesh) *gltf.PolyformModel {
		return &gltf.PolyformModel{Name: "Mesh", Mesh: &m}
	})
	return table
}

func registerNumberCasts(table *cast.Table) {
	cast.Register(table, func(v int) float64 { return float64(v) })
	cast.Register(table, func(v float64) int { return int(math.Round(v)) })

	cast.Register(table, func(v bool) float64 { return boolNumber[float64](v) })
	cast.Register(table, func(v bool) int { return boolNumber[int](v) })
	cast.Register(table, func(v float64) bool { return v != 0 })
	cast.Register(table, func(v int) bool { return v != 0 })

	cast.Register(table, func(v vector2.Int) vector2.Float64 { return v.ToFloat64() })
	cast.Register(table, func(v vector2.Float64) vector2.Int { return v.RoundToInt() })
	cast.Register(table, func(v vector3.Int) vector3.Float64 { return v.ToFloat64() })
	cast.Register(table, func(v vector3.Float64) vector3.Int { return v.RoundToInt() })

	cast.Register(table, vector2.Fill[float64])
	cast.Register(table, vector2.Fill[int])
	cast.Register(table, vector3.Fill[float64])
	cast.Register(table, vector3.Fill[int])
	cast.Register(table, vector4.Fill[float64])

	cast.Register(table, func(v vector2.Float64) vector3.Float64 { return vector3.New(v.X(), v.Y(), 0) })
	cast.Register(table, func(v vector2.Int) vector3.Int { return vector3.New(v.X(), v.Y(), 0) })
}

func boolNumber[T int | float64](v bool) T {
	if v {
		return 1
	}
	return 0
}

func registerNormCasts(table *cast.Table) {
	cast.Register(table, norm.NewUFloat64)
	cast.Register(table, norm.NewSFloat64)
	cast.Register(table, func(v norm.UFloat64) float64 { return v.Value() })
	cast.Register(table, func(v norm.SFloat64) float64 { return v.Value() })
	cast.Register(table, func(v int) norm.UFloat64 { return norm.NewUFloat64(float64(v)) })
	cast.Register(table, func(v int) norm.SFloat64 { return norm.NewSFloat64(float64(v)) })
	cast.Register(table, func(v norm.UFloat64) norm.SFloat64 { return norm.NewSFloat64(v.Value()) })
}

func registerTextureCasts(table *cast.Table) {
	cast.Register(table, func(t texturing.Texture[float64]) texturing.Texture[coloring.Color] {
		return texturing.Convert(t, func(_, _ int, v float64) coloring.Color {
			return coloring.Color{R: v, G: v, B: v, A: 1}
		})
	})

	cast.Register(table, func(t texturing.Texture[bool]) texturing.Texture[float64] {
		return texturing.Convert(t, func(_, _ int, v bool) float64 {
			if v {
				return 1
			}
			return 0
		})
	})

	cast.Register(table, func(t texturing.Texture[float64]) texturing.Texture[bool] {
		return texturing.Convert(t, func(_, _ int, v float64) bool { return v != 0 })
	})

	cast.Register(table, func(t texturing.Texture[coloring.Color]) image.Image {
		return t.ToImage(func(c coloring.Color) color.Color {
			return coloring.Color{
				R: math.Max(0, math.Min(1, c.R)),
				G: math.Max(0, math.Min(1, c.G)),
				B: math.Max(0, math.Min(1, c.B)),
				A: math.Max(0, math.Min(1, c.A)),
			}
		})
	})

	cast.RegisterFallible(table, func(img image.Image) (texturing.Texture[coloring.Color], error) {
		if img == nil {
			return texturing.Empty[coloring.Color](0, 0), errors.New("no image to read into a texture")
		}
		return texturing.FromImage(img), nil
	})

	cast.Register(table, func(t texturing.Texture[bool]) texturing.Texture[coloring.Color] {
		return texturing.Convert(t, func(_, _ int, v bool) coloring.Color {
			if v {
				return coloring.Color{R: 1, G: 1, B: 1, A: 1}
			}
			return coloring.Color{A: 1}
		})
	})
}
