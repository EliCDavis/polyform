package main

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

type BerryConfig struct {
	relativeBerrySize float64
	colorPalette      coloring.ColorStack
	chanceOfBerry     float64
}

func Bristle(
	colorContext *gg.Context,
	specularContext *gg.Context,
	start, end vector.Vector2,
	branchWidth, chanceOfSnow float64,
	colors coloring.ColorStack,
	berry *BerryConfig,
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

	halfBranchWidth := branchWidth / 2.

	if depth > 1 {

		subBristles := 4
		currentStart := .1

		spacing := (1. - currentStart) / float64(subBristles)
		halfSpacing := spacing * 0.5

		for i := 0; i < subBristles; i++ {
			startPercentage := currentStart + (rand.Float64() * halfSpacing * 0.25)
			endPercentage := startPercentage + .2 + (rand.Float64() * .2)
			point := start.Add(dir.MultByConstant(startPercentage))

			// Draw Berries
			if berry != nil && rand.Float64() > berry.chanceOfBerry {
				colorContext.SetColor(berry.colorPalette.LinearSample(rand.Float64()))
				colorContext.DrawCircle(
					point.X()-halfBranchWidth+(halfBranchWidth*rand.Float64()*2),
					point.Y()-halfBranchWidth+(halfBranchWidth*rand.Float64()*2),
					branchWidth*berry.relativeBerrySize,
				)
				colorContext.Fill()
			}

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
				berry,
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
				berry,
				depth-1,
			)

			currentStart = startPercentage + (halfSpacing * 2)
		}
	}
}

func BranchTexture(
	colors coloring.ColorStack,
	textures *PBRTextures,
	imageSize float64,
	minSnow float64,
	maxSnow float64,
	numBranches int,
	numOfBerryBranches int,
	atlasSize int,
) *Atlas {

	if numOfBerryBranches > numBranches {
		panic(fmt.Errorf("berry branch count can't be greater than branch count (%d > %d)", numOfBerryBranches, numBranches))
	}

	colorContext := gg.NewContext(int(imageSize), int(imageSize))
	colorContext.SetRGBA(0, 0, 0, 0)
	colorContext.Clear()

	specularContext := gg.NewContext(int(imageSize), int(imageSize))
	specularContext.SetRGBA(0, 0, 0, 1)
	specularContext.Clear()

	branchImageSize := imageSize / float64(numBranches)
	halfBranchImageSize := branchImageSize / 2

	snowInc := (maxSnow - minSnow) / float64(numBranches*numBranches)

	workingAtlas := &Atlas{
		Name:       "Branches 0",
		BottomLeft: vector.Vector2Zero(),
		TopRight:   vector.Vector2One(),
		Entries:    make([]AtlasEntry, 0),
	}

	subAtlases := make([]*Atlas, 0)
	entryCount := 0

	berryConfig := &BerryConfig{
		relativeBerrySize: .5,
		colorPalette: coloring.NewColorStack(
			[]coloring.ColorStackEntry{
				coloring.NewColorStackEntry(1, 1, 1, color.RGBA{235, 64, 52, 255}),
				coloring.NewColorStackEntry(1, 1, 1, color.RGBA{235, 52, 98, 255}),
				coloring.NewColorStackEntry(1, 1, 1, color.RGBA{255, 102, 140, 255}),
			},
		),
		chanceOfBerry: .75,
	}

	for y := 0; y < numBranches; y++ {
		for x := 0; x < numBranches; x++ {

			var berryConfigForBranch *BerryConfig = nil
			if x < numOfBerryBranches {
				berryConfigForBranch = berryConfig
			}

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
				berryConfigForBranch,
				4,
			)

			workingAtlas.Entries = append(workingAtlas.Entries, AtlasEntry{
				BottomLeft: vector.NewVector2(float64(x)/float64(numBranches), float64(y)/float64(numBranches)),
				TopRight:   vector.NewVector2(float64(x+1)/float64(numBranches), float64(y+1)/float64(numBranches)),
			})

			entryCount++
			if entryCount == atlasSize {
				entryCount = 0
				subAtlases = append(subAtlases, workingAtlas)
				workingAtlas = &Atlas{
					Name:       fmt.Sprintf("Branches %d", len(subAtlases)),
					BottomLeft: vector.Vector2Zero(),
					TopRight:   vector.Vector2One(),
					Entries:    make([]AtlasEntry, 0),
				}
			}
		}
	}

	if len(workingAtlas.Entries) > 0 {
		subAtlases = append(subAtlases, workingAtlas)
	}

	textures.color = colorContext.Image()
	textures.normal = texturing.ToNormal(colorContext.Image())
	textures.specular = specularContext.Image()

	if len(subAtlases) > 1 {
		return &Atlas{
			Name:       "Branches",
			SubAtlas:   subAtlases,
			BottomLeft: vector.Vector2Zero(),
			TopRight:   vector.Vector2One(),
		}
	}
	return subAtlases[0]
}
