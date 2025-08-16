package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func sigmoid(xScale, x, xShift, yShift float64) float64 {
	denominator := 1 + math.Pow(math.E, -xScale*(x-xShift))
	return (-1 / denominator) + yShift
}

func Texture(textureSize int, mapSize, height, waterLevel float64, name string, landNoise, waterNoise sample.Vec2ToFloat, landColors, waterColors coloring.ColorStack) {
	tex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	scaleFactor := mapSize / float64(textureSize)
	for x := range textureSize {
		for y := range textureSize {
			samplePos := vector2.New(float64(x), float64(y)).Scale(scaleFactor)

			sample := landNoise(samplePos)
			if sample <= waterLevel {
				tex.Set(x, y, waterColors.LinearSample(waterNoise(samplePos)))
			} else {
				tex.Set(x, y, landColors.LinearSample((sample-waterLevel)/(height-waterLevel)))
			}
		}
	}

	texOut, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer texOut.Close()

	err = jpeg.Encode(texOut, tex, &jpeg.Options{Quality: 100})
	if err != nil {
		panic(err)
	}
}

func main() {
	n := 5000
	mapSize := 3000.
	mapRadius := mapSize / 2
	mapOffset := vector2.New(mapRadius, mapRadius)
	totalHeight := 200.
	waterLevel := 15.
	points := make([]vector2.Float64, n)
	for i := range n {
		theta := rand.Float64() * 2 * math.Pi
		points[i] = vector2.
			New(math.Cos(theta), math.Sin(theta)).
			Scale(mapRadius * math.Sqrt(rand.Float64())).
			Add(mapOffset)
	}

	perlinStack := noise.PerlinStack(
		noise.Stack2DEntry{Scalar: 1 / 300., Amplitude: totalHeight / 2},
		noise.Stack2DEntry{Scalar: 1 / 150., Amplitude: totalHeight / 4},
		noise.Stack2DEntry{Scalar: 1 / 75., Amplitude: totalHeight / 8},
		noise.Stack2DEntry{Scalar: 1 / 37.5, Amplitude: totalHeight / 16},
	)

	heightFunc := sample.Vec2ToFloat(func(v vector2.Float64) float64 {
		rollOff := sigmoid(20, v.Sub(mapOffset).Length()/mapRadius, .5, 1)
		return math.Max(perlinStack.Value(v)*rollOff, waterLevel)
	})

	textureName := "terrain.jpg"
	mat := obj.Material{
		Name:            "Terrain",
		ColorTextureURI: &textureName,
	}

	Texture(
		2048,
		mapSize,
		totalHeight,
		waterLevel,
		textureName,
		heightFunc,
		sample.Vec2ToFloat(noise.PerlinStack(
			noise.Stack2DEntry{Scalar: 1 / 300., Amplitude: 1. / 2},
			noise.Stack2DEntry{Scalar: 1 / 150., Amplitude: 1. / 4},
			noise.Stack2DEntry{Scalar: 1 / 75., Amplitude: 1. / 8},
			noise.Stack2DEntry{Scalar: 1 / 37.5, Amplitude: 1. / 16},
		).Value),
		coloring.NewColorStack(
			coloring.NewColorStackEntry(0.1, 0.5, 0.7, color.RGBA{199, 237, 255, 255}), // Water Foam
			coloring.NewColorStackEntry(0.5, 0.5, 0.1, color.RGBA{209, 191, 138, 255}), // Sand
			coloring.NewColorStackEntry(3, 0.1, 0.5, color.RGBA{59, 120, 65, 255}),     // Grass
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{145, 145, 145, 255}),   // Stone
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{224, 224, 224, 255}),   // Mountain Top Snow
		),
		coloring.NewColorStack(
			coloring.NewColorStackEntry(1, 0.8, 0.8, color.RGBA{0, 174, 255, 255}),
			coloring.NewColorStackEntry(0.5, 0.8, 0.8, color.RGBA{84, 201, 255, 255}),
		),
	)

	uvs := make([]vector2.Float64, len(points))

	terrain := triangulation.
		BowyerWatson(points).
		ModifyFloat3AttributeParallel(modeling.PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
			uvs[i] = vector2.New(v.X(), -v.Z()).
				DivByConstant(mapSize)

			return v.SetY(heightFunc(v.XZ()))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

	err := obj.Save("terrain.obj", obj.Scene{
		Objects: []obj.Object{
			{
				Entries: []obj.Entry{
					{
						Mesh:     terrain,
						Material: &mat,
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
