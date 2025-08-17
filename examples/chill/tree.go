package main

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func Cone(base float64, points ...vector3.Float64) modeling.Mesh {
	length := vector3.Array[float64](points).Distance()
	extrusionPoints := make([]extrude.ExtrusionPoint, len(points))

	dist := 0.0
	for i := 0; i < len(points); i++ {
		if i > 0 {
			dist += points[i].Distance(points[i-1])
		}
		size := (1. - (dist / length))
		extrusionPoints[i] = extrude.ExtrusionPoint{
			Point:     points[i],
			Thickness: (base * size),
			UV: &extrude.ExtrusionPointUV{
				Point:     vector2.New(0., size),
				Thickness: size,
			},
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

	branches := modeling.EmptyMesh(modeling.TriangleTopology)
	for i := 0; i < branchCount; i++ {
		branchHeight := (heightCovered * rand.Float64()) + heightBare

		availableHeightUsed := (branchHeight - heightBare) / heightCovered

		trailOffGivenHeight := ((1 - availableHeightUsed) + .2)

		branchMaxWidth := (base) * 2 * trailOffGivenHeight * (1 + (.4 * rand.Float64()))

		dir := vector3.New(-1+(2*rand.Float64()), 0, -1+(2*rand.Float64())).
			Normalized().
			Scale(branchLength * trailOffGivenHeight)

		branchAtlasEntry := atlas.RandomEntry()

		// branchUV := vector2.New(0.25, 1).
		// 	Add(vector2.New(float64(xCordOfBranch)*.5, -yCordOfBranch*.5))

		branchUV := vector2.New(
			branchAtlasEntry.MinX()+(branchAtlasEntry.Width()/2),
			branchAtlasEntry.MaxY(),
		)
		branchUVLength := branchAtlasEntry.Height()

		branches = branches.Append(extrude.Line([]extrude.LinePoint{
			{
				Point:   vector3.New(0, branchHeight, 0),
				Up:      vector3.Up[float64](),
				Height:  -(height / 30),
				Width:   branchMaxWidth * 0.45,
				Uv:      branchUV,
				UvWidth: .25,
			},
			{
				Point:   dir.Scale(.25).SetY(branchHeight - 1),
				Up:      vector3.Up[float64](),
				Height:  -(height / 30),
				Width:   branchMaxWidth,
				Uv:      branchUV.Add(vector2.Down[float64]().Scale(branchUVLength * .25)),
				UvWidth: .5,
			},
			{
				Point:   dir.Scale(.5).SetY(branchHeight - 1.5),
				Up:      vector3.Up[float64](),
				Height:  -(height / 30),
				Width:   branchMaxWidth * 0.75,
				Uv:      branchUV.Add(vector2.Down[float64]().Scale(branchUVLength * .5)),
				UvWidth: .5,
			},
			{
				Point:   dir.SetY(branchHeight - 2),
				Up:      vector3.Up[float64](),
				Height:  0,
				Width:   branchMaxWidth * 0.35,
				Uv:      branchUV.Add(vector2.Down[float64]().Scale(branchUVLength)),
				UvWidth: .25,
			},
		}))
	}

	return Cone(
		base,
		vector3.New(0., 0., 0.),
		vector3.New(0, height, 0),
	).
		Transform(
			meshops.SmoothNormalsTransformer{},
		), branches
}

func TrunkTexture(imageSize int, colors coloring.Gradient, barkNoise sample.Vec2ToFloat, barkPBR *PBRTextures) {
	dc := gg.NewContext(imageSize, imageSize)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	df := noise.NewDistanceField(10, 10, vector2.Fill(float64(imageSize)))

	for x := 0; x < imageSize; x++ {
		for y := 0; y < imageSize; y++ {
			// sample := barkNoise(vector2.New(float64(x), float64(y)))

			sample := math.Min(df.Sample(vector2.New(float64(x), float64(y)))/(float64(imageSize)/10.), 1)

			dc.SetColor(colors.Sample(sample))
			dc.SetPixel(x, y)
		}
	}

	barkPBR.color = dc.Image()
	barkPBR.normal = texturing.ToNormal(dc.Image())
}
