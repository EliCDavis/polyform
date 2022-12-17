package main

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/mesh/texturing"
	"github.com/EliCDavis/mesh/triangulation"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

func reSample(val float64, originalRange, newRange vector.Vector2) float64 {
	percent := (val - originalRange.X()) / (originalRange.Y() - originalRange.X())
	return ((newRange.Y() - newRange.X()) * percent) + newRange.X()
}

func TerrainTexture(
	textureSize int,
	mapSize float64,
	textures *PBRTextures,
	colors coloring.ColorStack,
	startPos vector.Vector3,
	landNoise noise.Sampler2D,
) {
	tex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))
	specTex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))
	normalSourceTex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	imageDimensions := vector.Vector2One().MultByConstant(float64(textureSize))
	df := noise.NewDistanceField(30, 30, imageDimensions)
	df2 := noise.NewDistanceField(60, 60, imageDimensions)
	df3 := noise.NewDistanceField(80, 80, imageDimensions)
	df4 := noise.NewDistanceField(160, 160, imageDimensions)
	df5 := noise.NewDistanceField(240, 240, imageDimensions)
	df6 := noise.NewDistanceField(480, 480, imageDimensions)

	colorSampleFunc := func(samplePos vector.Vector2) float64 {
		return df.Sample(samplePos) -
			(df2.Sample(samplePos) / 2) +
			(df3.Sample(samplePos) / 4) +
			(df4.Sample(samplePos) / 4) +
			(df5.Sample(samplePos) / 8)
	}
	// colorSampleFunc = func(samplePos vector.Vector2) float64 {
	// 	return 0.5
	// }

	// scaleFactor := mapSize / float64(textureSize)

	for x := 0; x < textureSize; x++ {
		for y := 0; y < textureSize; y++ {
			pixel := vector.NewVector2(float64(x), float64(y))

			colorSample := colorSampleFunc(pixel)
			clampedSample := mesh.Clamp(colorSample/(float64(textureSize)/40.), 0, 1)
			tex.Set(x, y, colors.LinearSample(clampedSample))

			// worldSpacePos := pixel.MultByConstant(scaleFactor)
			// height := landNoise(worldSpacePos)

			spec := uint8((reSample(1.-clampedSample, vector.NewVector2(0, 1), vector.NewVector2(0.5, 0.75)) * .65) * 255)

			clampedSample = mesh.Clamp((colorSample+(df6.Sample(pixel)/2))/(float64(textureSize)/40.), 0, 1)
			nrml := uint8((reSample(1.-clampedSample, vector.NewVector2(0, 1), vector.NewVector2(0.2, 0.75)) * .85) * 255)

			specTex.Set(x, y, color.RGBA{
				R: spec,
				G: spec,
				B: spec,
				A: 255,
			})

			normalSourceTex.Set(x, y, color.RGBA{
				R: nrml,
				G: nrml,
				B: nrml,
				A: 255,
			})
		}
	}

	textures.color = tex
	textures.normal = texturing.ToNormal(texturing.BoxBlurNTimes(normalSourceTex, 5))
	textures.specular = specTex

}

func SnakeOut(x, amplitude, iterations, scale float64) float64 {
	x2pi := x * 2. * math.Pi
	return math.Sin(x2pi/scale) * ((scale * amplitude) / x2pi)
}

func DrawTrail(
	terrain mesh.Mesh,
	textures *PBRTextures,
	path []vector.Vector2,
	forestWidth float64,
	terrainImageSize int,
	snowColors coloring.ColorStack,
) mesh.Mesh {

	if len(path) < 2 {
		return terrain
	}

	pathWidth := 20.

	pixelsPerMeter := float64(terrainImageSize) / forestWidth

	dc := gg.NewContextForImage(textures.color)
	for i := 1; i < len(path); i++ {
		dc.SetColor(color.RGBA{70, 75, 80, 80})
		dc.SetLineWidth(pixelsPerMeter * pathWidth)
		dc.DrawLine(
			path[i-1].X()*pixelsPerMeter,
			path[i-1].Y()*pixelsPerMeter,
			path[i].X()*pixelsPerMeter,
			path[i].Y()*pixelsPerMeter,
		)
		dc.Stroke()
	}
	textures.color = dc.Image()

	return terrain.
		ModifyVertices(func(v vector.Vector3) vector.Vector3 {
			heightAdj := 0.

			for i := 1; i < len(path); i++ {
				line := mesh.NewLine2D(path[i], path[i-1])
				p := v.XZ()
				dist := line.ClosestPointOnLine(p).Distance(p)
				if dist > 30 {
					continue
				}
				heightAdj += SnakeOut(dist, -10, 2, pathWidth)
			}

			return v.SetY(v.Y() + heightAdj)
		}).
		CalculateSmoothNormals()
}

func Terrain(forestWidth float64, height noise.Sampler2D, textures *PBRTextures) (mesh.Mesh, vector.Vector3) {
	n := 10000
	mapRadius := forestWidth / 2
	mapOffset := vector.NewVector2(mapRadius, mapRadius)

	points := make([]vector.Vector2, n)
	for i := 0; i < n; i++ {
		points[i] = randomVec2Radial().
			MultByConstant(mapRadius).
			Add(mapOffset)
	}

	heightFunc := noise.Sampler2D(func(v vector.Vector2) float64 {
		return height(v)
	})

	maxHeight := vector.NewVector3(0, -math.MaxFloat64, 0)

	terrain := triangulation.
		BowyerWatson(points).
		ModifyVertices(func(v vector.Vector3) vector.Vector3 {
			height := heightFunc(v.XZ())
			val := v.SetY(height)
			if height > maxHeight.Y() {
				maxHeight = val
			}
			return val
		}).
		CalculateSmoothNormals().
		ModifyUVs(func(v vector.Vector3, uv vector.Vector2) vector.Vector2 {
			return vector.NewVector2(v.X(), -v.Z()).
				DivByConstant(forestWidth)
		}).
		SetMaterial(textures.Material())

	return terrain, maxHeight
}
