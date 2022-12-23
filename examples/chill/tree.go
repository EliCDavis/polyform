package main

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

func Cone(base float64, points ...vector.Vector3) modeling.Mesh {
	length := vector.Vector3Array(points).Distance()
	extrusionPoints := make([]extrude.ExtrusionPoint, len(points))

	dist := 0.0
	for i := 0; i < len(points); i++ {
		if i > 0 {
			dist += points[i].Distance(points[i-1])
		}
		size := (1. - (dist / length))
		extrusionPoints[i] = extrude.ExtrusionPoint{
			Point:       points[i],
			Thickness:   (base * size),
			UvThickness: size,
			UvPoint:     vector.NewVector2(0, size),
		}
	}

	return extrude.Polygon(16, extrusionPoints)
}

func Tree(
	height, base, percentageCovered float64,
	branchCount int,
	atlas *Atlas,
) (modeling.Mesh, modeling.Mesh) {
	percentBare := 1 - percentageCovered

	heightCovered := height * percentageCovered
	heightBare := height * percentBare

	// branchCount := 3
	branchLength := height * 0.25 * (.8 + (.4 * rand.Float64()))

	branches := modeling.EmptyMesh()
	for i := 0; i < branchCount; i++ {
		branchHeight := (heightCovered * rand.Float64()) + heightBare

		availableHeightUsed := (branchHeight - heightBare) / heightCovered

		trailOffGivenHeight := ((1 - availableHeightUsed) + .2)

		branchMaxWidth := (base) * 2 * trailOffGivenHeight * (1 + (.4 * rand.Float64()))

		dir := vector.NewVector3(-1+(2*rand.Float64()), 0, -1+(2*rand.Float64())).
			Normalized().
			MultByConstant(branchLength * trailOffGivenHeight)

		branchAtlasEntry := atlas.RandomEntry()

		// branchUV := vector.NewVector2(0.25, 1).
		// 	Add(vector.NewVector2(float64(xCordOfBranch)*.5, -yCordOfBranch*.5))

		branchUV := vector.NewVector2(
			branchAtlasEntry.MinX()+(branchAtlasEntry.Width()/2),
			branchAtlasEntry.MaxY(),
		)
		branchUVLength := branchAtlasEntry.Height()

		branches = branches.Append(extrude.Line([]extrude.LinePoint{
			{
				Point:   vector.NewVector3(0, branchHeight, 0),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   branchMaxWidth * 0.45,
				Uv:      branchUV,
				UvWidth: .25,
			},
			{
				Point:   dir.MultByConstant(.25).SetY(branchHeight - 1),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   branchMaxWidth,
				Uv:      branchUV.Add(vector.Vector2Down().MultByConstant(branchUVLength * .25)),
				UvWidth: .5,
			},
			{
				Point:   dir.MultByConstant(.5).SetY(branchHeight - 1.5),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   branchMaxWidth * 0.75,
				Uv:      branchUV.Add(vector.Vector2Down().MultByConstant(branchUVLength * .5)),
				UvWidth: .5,
			},
			{
				Point:   dir.SetY(branchHeight - 2),
				Up:      vector.Vector3Up(),
				Height:  0,
				Width:   branchMaxWidth * 0.35,
				Uv:      branchUV.Add(vector.Vector2Down().MultByConstant(branchUVLength)),
				UvWidth: .25,
			},
		}))
	}

	return Cone(
		base,
		vector.NewVector3(0, 0, 0),
		vector.NewVector3(0, height, 0),
	).
		CalculateSmoothNormals(), branches
}

func TrunkTexture(imageSize int, colors coloring.ColorStack, barkNoise noise.Sampler2D, barkPBR *PBRTextures) {
	dc := gg.NewContext(imageSize, imageSize)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	df := noise.NewDistanceField(10, 10, vector.Vector2One().MultByConstant(float64(imageSize)))

	for x := 0; x < imageSize; x++ {
		for y := 0; y < imageSize; y++ {
			// sample := barkNoise(vector.NewVector2(float64(x), float64(y)))

			sample := math.Min(df.Sample(vector.NewVector2(float64(x), float64(y)))/(float64(imageSize)/10.), 1)

			dc.SetColor(colors.LinearSample(sample))
			dc.SetPixel(x, y)
		}
	}

	barkPBR.color = dc.Image()
	barkPBR.normal = texturing.ToNormal(dc.Image())
}
