package main

import (
	"image/color"
	"math/rand"

	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/texturing"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

func Bristle(
	colorContext *gg.Context,
	specularContext *gg.Context,
	start, end vector.Vector2,
	branchWidth, chanceOfSnow float64,
	colors coloring.ColorStack,
	depth int,
) {
	colorContext.SetColor(color.RGBA{99, 62, 10, 255})
	colorContext.SetLineWidth(branchWidth)
	colorContext.DrawLine(start.X(), start.Y(), end.X(), end.Y())
	colorContext.Stroke()

	w := branchWidth / 7.
	colorContext.SetLineWidth(w)
	specularContext.SetLineWidth(w * .75)

	dir := end.Sub(start)
	right := dir.Perpendicular()
	for i := 0; i < 300; i++ {
		startPercentage := rand.Float64()
		endPercentage := startPercentage + (.1 * (1 - startPercentage))
		point := start.Add(dir.MultByConstant(startPercentage))

		side := 1
		if rand.Float64() <= .5 {
			side = -1
		}

		needleLength := .05 + (.1 * (1. - endPercentage))

		endPoint := start.Add(dir.MultByConstant((endPercentage) + (rand.Float64() * .05))).
			Add(right.MultByConstant(needleLength * float64(side)))

		if rand.Float64() <= chanceOfSnow {
			colorContext.SetColor(color.RGBA{255, 255, 255, 255})
			specularContext.SetColor(color.RGBA{65, 65, 65, 127})
		} else {
			colorContext.SetColor(colors.LinearSample(rand.Float64()))
			specularContext.SetColor(color.RGBA{0, 0, 0, 127})
		}

		colorContext.DrawLine(point.X(), point.Y(), endPoint.X(), endPoint.Y())
		colorContext.Stroke()

		specularContext.DrawLine(point.X(), point.Y(), endPoint.X(), endPoint.Y())
		specularContext.Stroke()
	}

	if depth > 1 {

		subBristles := 4
		currentStart := .1

		spacing := (1. - currentStart) / float64(subBristles)
		halfSpacing := spacing * 0.5

		for i := 0; i < subBristles; i++ {
			startPercentage := currentStart + (rand.Float64() * halfSpacing * 0.25)
			endPercentage := startPercentage + .2 + (rand.Float64() * .2)
			point := start.Add(dir.MultByConstant(startPercentage))

			generalSize := (1. - startPercentage) * .5

			rightBristleEnd := start.Add(dir.MultByConstant(endPercentage)).
				Add(right.MultByConstant(generalSize * (.8 + (rand.Float64() * .4))))

			leftBristleEnd := start.Add(dir.MultByConstant(endPercentage)).
				Sub(right.MultByConstant(generalSize * (.8 + (rand.Float64() * .4))))

			Bristle(
				colorContext,
				specularContext,
				point,
				rightBristleEnd,
				branchWidth/2,
				chanceOfSnow,
				colors,
				depth-1,
			)

			Bristle(
				colorContext,
				specularContext,
				point,
				leftBristleEnd,
				branchWidth/2,
				chanceOfSnow,
				colors,
				depth-1,
			)

			currentStart = startPercentage + (halfSpacing * 2)
		}
	}
}

func BranchTexture(colors coloring.ColorStack, textures *PBRTextures, imageSize float64) {
	colorContext := gg.NewContext(int(imageSize), int(imageSize))
	colorContext.SetRGBA(0, 0, 0, 0)
	colorContext.Clear()

	specularContext := gg.NewContext(int(imageSize), int(imageSize))
	specularContext.SetRGBA(0, 0, 0, 1)
	specularContext.Clear()

	numBranches := 2
	branchImageSize := imageSize / float64(numBranches)
	halfBranchImageSize := branchImageSize / 2

	minSnow := .2
	maxSnow := .9
	snowInc := (maxSnow - minSnow) / float64(numBranches*numBranches)

	for x := 0; x < numBranches; x++ {
		for y := 0; y < numBranches; y++ {
			start := vector.NewVector2(halfBranchImageSize, 0).
				Add(vector.NewVector2(float64(x)*branchImageSize, float64(y)*branchImageSize))

			Bristle(
				colorContext,
				specularContext,
				start,
				start.Add(vector.NewVector2(0, branchImageSize*.8)),
				20.,
				minSnow+(snowInc*float64(x+(y*numBranches))),
				colors,
				4,
			)
		}
	}

	textures.color = colorContext.Image()
	textures.normal = texturing.ToNormal(colorContext.Image())
	textures.specular = specularContext.Image()
}
