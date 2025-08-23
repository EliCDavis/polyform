package main

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func reSample(val float64, originalRange, newRange vector2.Float64) float64 {
	percent := (val - originalRange.X()) / (originalRange.Y() - originalRange.X())
	return ((newRange.Y() - newRange.X()) * percent) + newRange.X()
}

func TerrainTexture(
	textureSize int,
	mapSize float64,
	textures *PBRTextures,
	colors coloring.Gradient[color.Color],
	startPos vector3.Float64,
	landNoise sample.Vec2ToFloat,
) {
	tex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))
	specTex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))
	normalSourceTex := image.NewRGBA(image.Rect(0, 0, textureSize, textureSize))

	imageDimensions := vector2.Fill(float64(textureSize))
	df := noise.NewDistanceField(30, 30, imageDimensions)
	df2 := noise.NewDistanceField(60, 60, imageDimensions)
	df3 := noise.NewDistanceField(80, 80, imageDimensions)
	df4 := noise.NewDistanceField(160, 160, imageDimensions)
	df5 := noise.NewDistanceField(240, 240, imageDimensions)
	df6 := noise.NewDistanceField(480, 480, imageDimensions)

	colorSampleFunc := func(samplePos vector2.Float64) float64 {
		return df.Sample(samplePos) -
			(df2.Sample(samplePos) / 2) +
			(df3.Sample(samplePos) / 4) +
			(df4.Sample(samplePos) / 4) +
			(df5.Sample(samplePos) / 8)
	}
	// colorSampleFunc = func(samplePos vector2.Float64) float64 {
	// 	return 0.5
	// }

	// scaleFactor := mapSize / float64(textureSize)

	for x := 0; x < textureSize; x++ {
		for y := 0; y < textureSize; y++ {
			pixel := vector2.New(float64(x), float64(y))

			colorSample := colorSampleFunc(pixel)
			clampedSample := modeling.Clamp(colorSample/(float64(textureSize)/40.), 0, 1)
			tex.Set(x, y, colors.Sample(clampedSample))

			// worldSpacePos := pixel.Scale(scaleFactor)
			// height := landNoise(worldSpacePos)

			spec := uint8((reSample(1.-clampedSample, vector2.New(.0, 1.), vector2.New(0.5, 0.75)) * .65) * 255)

			clampedSample = modeling.Clamp((colorSample+(df6.Sample(pixel)/2))/(float64(textureSize)/40.), 0, 1)
			nrml := uint8((reSample(1.-clampedSample, vector2.New(0., 1.), vector2.New(0.2, 0.75)) * .85) * 255)

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
	terrain modeling.Mesh,
	textures *PBRTextures,
	trail Trail,
	forestWidth float64,
	terrainImageSize int,
	snowColors coloring.Gradient[color.Color],
) modeling.Mesh {
	if len(trail.Segments) == 0 {
		return terrain
	}

	pixelsPerMeter := float64(terrainImageSize) / forestWidth
	dc := gg.NewContextForImage(textures.color)

	for _, seg := range trail.Segments {
		dc.SetColor(color.RGBA{70, 75, 80, 80})
		dc.SetLineWidth(pixelsPerMeter * seg.Width)
		dc.DrawLine(
			seg.StartX*pixelsPerMeter,
			seg.StartY*pixelsPerMeter,
			seg.EndX*pixelsPerMeter,
			seg.EndY*pixelsPerMeter,
		)
		dc.Stroke()
	}

	textures.color = dc.Image()

	return terrain.
		ModifyFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
			heightAdj := 0.

			for _, seg := range trail.Segments {
				line := geometry.NewLine2D(
					vector2.New(
						seg.StartX,
						seg.StartY,
					),
					vector2.New(
						seg.EndX,
						seg.EndY,
					),
				)
				p := v.XZ()
				dist := line.ClosestPointOnLine(p).Distance(p)
				if dist > 30 {
					continue
				}
				heightAdj += SnakeOut(dist, -seg.Depth, 2, seg.Width)
			}

			return v.SetY(v.Y() + heightAdj)
		}).
		Transform(
			meshops.SmoothNormalsTransformer{},
		)
}

func Terrain(forestWidth float64, height sample.Vec2ToFloat, textures *PBRTextures) (obj.Object, vector3.Float64) {
	n := 10000
	mapRadius := forestWidth / 2
	mapOffset := vector2.New(mapRadius, mapRadius)

	points := make([]vector2.Float64, n)
	for i := 0; i < n; i++ {
		points[i] = randomVec2Radial().
			Scale(mapRadius).
			Add(mapOffset)
	}

	heightFunc := sample.Vec2ToFloat(func(v vector2.Float64) float64 {
		return height(v)
	})

	maxHeight := vector3.New(0., -math.MaxFloat64, 0.)

	uvs := make([]vector2.Float64, len(points))

	terrain := triangulation.
		BowyerWatson(points).
		ModifyFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
			height := heightFunc(v.XZ())
			val := v.SetY(height)
			if height > maxHeight.Y() {
				maxHeight = val
			}

			uvs[i] = vector2.New(v.X(), -v.Z()).
				DivByConstant(forestWidth)

			return val
		}).
		Transform(
			meshops.SmoothNormalsTransformer{},
		).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

	return obj.Object{
		Name: "Terrain",
		Entries: []obj.Entry{
			obj.Entry{
				Mesh:     terrain,
				Material: textures.Material(),
			},
		},
	}, maxHeight
}
