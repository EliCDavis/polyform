package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/triangulation"
	"github.com/EliCDavis/vector"
)

func sigmoid(xScale, x, xShift, yShift float64) float64 {
	denominator := 1 + math.Pow(math.E, -xScale*(x-xShift))
	return (-1 / denominator) + yShift
}

func Texture(textureSize int, mapSize, height, waterLevel float64, name string, landNoise, waterNoise noise.Sampler2D, landColors, waterColors coloring.ColorStack) {
	tex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	scaleFactor := mapSize / float64(textureSize)
	for x := 0; x < textureSize; x++ {
		for y := 0; y < textureSize; y++ {
			samplePos := vector.NewVector2(float64(x), float64(y)).MultByConstant(scaleFactor)

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
	mapOffset := vector.NewVector2(mapRadius, mapRadius)
	totalHeight := 200.
	waterLevel := 15.
	points := make([]vector.Vector2, n)
	for i := 0; i < n; i++ {
		theta := rand.Float64() * 2 * math.Pi
		points[i] = vector.
			NewVector2(math.Cos(theta), math.Sin(theta)).
			MultByConstant(mapRadius * math.Sqrt(rand.Float64())).
			Add(mapOffset)
	}

	perlinStack := noise.PerlinStack([]noise.Stack2DEntry{
		{Scalar: 1 / 300., Amplitude: totalHeight / 2},
		{Scalar: 1 / 150., Amplitude: totalHeight / 4},
		{Scalar: 1 / 75., Amplitude: totalHeight / 8},
		{Scalar: 1 / 37.5, Amplitude: totalHeight / 16},
	})

	heightFunc := noise.Sampler2D(func(v vector.Vector2) float64 {
		rollOff := sigmoid(20, v.Sub(mapOffset).Length()/mapRadius, .5, 1)
		return math.Max(perlinStack.Value(v)*rollOff, waterLevel)
	})

	textureName := "terrain.jpg"
	mat := mesh.Material{
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
		noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
			{Scalar: 1 / 300., Amplitude: 1. / 2},
			{Scalar: 1 / 150., Amplitude: 1. / 4},
			{Scalar: 1 / 75., Amplitude: 1. / 8},
			{Scalar: 1 / 37.5, Amplitude: 1. / 16},
		}).Value),
		coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(0.1, 0.5, 0.7, color.RGBA{199, 237, 255, 255}), // Water Foam
			coloring.NewColorStackEntry(0.5, 0.5, 0.1, color.RGBA{209, 191, 138, 255}), // Sand
			coloring.NewColorStackEntry(3, 0.1, 0.5, color.RGBA{59, 120, 65, 255}),     // Grass
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{145, 145, 145, 255}),   // Stone
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{224, 224, 224, 255}),   // Mountain Top Snow
		}),
		coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(1, 0.8, 0.8, color.RGBA{0, 174, 255, 255}),
			coloring.NewColorStackEntry(0.5, 0.8, 0.8, color.RGBA{84, 201, 255, 255}),
		}),
	)

	terrain := triangulation.
		BowyerWatson(points).
		ModifyVertices(func(v vector.Vector3) vector.Vector3 {
			return v.SetY(heightFunc(v.XZ()))
		}).
		ModifyUVs(func(v vector.Vector3, uv vector.Vector2) vector.Vector2 {
			return vector.NewVector2(v.X(), -v.Z()).
				DivByConstant(mapSize)
		}).
		SetMaterial(mat)

	objFile, err := os.Create("terrain.obj")
	if err != nil {
		panic(err)
	}
	defer objFile.Close()

	mtlFile, err := os.Create("terrain.mtl")
	if err != nil {
		panic(err)
	}
	defer mtlFile.Close()

	obj.WriteMesh(terrain, "terrain.mtl", objFile)
	obj.WriteMaterials(terrain, mtlFile)
}
