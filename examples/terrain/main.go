package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/triangulation"
	"github.com/EliCDavis/vector"
)

func Texture(mapSize, textureSize int, height float64, name string, sampler noise.Sampler2D, colors coloring.ColorStack) {
	tex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	scaleFactor := float64(mapSize) / float64(textureSize)
	for x := 0; x < textureSize; x++ {
		for y := 0; y < textureSize; y++ {
			sample := sampler(vector.NewVector2(float64(x), float64(y)).MultByConstant(scaleFactor)) + (height / 2)
			tex.Set(x, y, colors.LinearSample(sample/height))
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
	n := 10000
	mapSize := 1000
	points := make([]vector.Vector2, n)
	for i := 0; i < n; i++ {
		points[i] = vector.Vector2Rnd().MultByConstant(float64(mapSize))
	}

	totalHeight := 200.
	perlinStack := noise.PerlinStack([]noise.Stack2DEntry{
		{Scalar: 1 / 300., Amplitude: totalHeight / 2},
		{Scalar: 1 / 150., Amplitude: totalHeight / 4},
		{Scalar: 1 / 75., Amplitude: totalHeight / 8},
		{Scalar: 1 / 37.5, Amplitude: totalHeight / 16},
	})

	textureName := "terrain.jpg"
	mat := mesh.Material{
		Name:            "Terrain",
		ColorTextureURI: &textureName,
	}

	Texture(
		mapSize,
		2048,
		totalHeight,
		textureName,
		noise.Sampler2D(perlinStack.Value),
		coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(3, 0.5, 0.5, color.RGBA{59, 120, 65, 255}),   // Grass
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{145, 145, 145, 255}), // Stone
			coloring.NewColorStackEntry(2, 0.5, 0.5, color.RGBA{224, 224, 224, 255}), // Mountain Top Snow
		}))

	terrain := triangulation.
		BowyerWatson(points).
		ModifyVertices(func(v vector.Vector3) vector.Vector3 {
			return v.SetY(perlinStack.Value(v.XZ()))
		}).
		ModifyUVs(func(v vector.Vector3, uv vector.Vector2) vector.Vector2 {
			return vector.NewVector2(v.X(), -v.Z()).DivByConstant(float64(mapSize))
		}).
		CalculateSmoothNormals().
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

	obj.WriteMesh(&terrain, "terrain.mtl", objFile)
	obj.WriteMaterials(&terrain, mtlFile)
}
