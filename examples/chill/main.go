package main

import (
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/extrude"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/texturing"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

func Bristle(dc *gg.Context, start, end vector.Vector2, branchWidth, chanceOfSnow float64, colors coloring.ColorStack, depth int) {
	dc.SetColor(color.RGBA{99, 62, 10, 255})
	dc.SetLineWidth(branchWidth)
	dc.DrawLine(start.X(), start.Y(), end.X(), end.Y())
	dc.Stroke()

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

		w := branchWidth / 7.

		if rand.Float64() <= chanceOfSnow {
			dc.SetColor(color.RGBA{255, 255, 255, 255})
		} else {
			dc.SetColor(colors.LinearSample(rand.Float64()))
		}

		dc.SetLineWidth(w)
		dc.DrawLine(point.X(), point.Y(), endPoint.X(), endPoint.Y())
		dc.Stroke()
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
				dc,
				point,
				rightBristleEnd,
				branchWidth/2,
				chanceOfSnow,
				colors,
				depth-1,
			)

			Bristle(
				dc,
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

func BranchTexture(colors coloring.ColorStack, imageSize float64) {
	dc := gg.NewContext(int(imageSize), int(imageSize))
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

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
				dc,
				start,
				start.Add(vector.NewVector2(0, branchImageSize*.8)),
				20.,
				minSnow+(snowInc*float64(x+(y*numBranches))),
				colors,
				4,
			)
		}
	}

	dc.SavePNG("branch.png")

	normal := texturing.ToNormal(dc.Image())
	f, _ := os.Create("branch_normal.png")
	png.Encode(f, normal)
}

func TrunkTexture(imageSize int, colors coloring.ColorStack, barkNoise noise.Sampler2D) error {
	dc := gg.NewContext(imageSize, imageSize)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	for x := 0; x < imageSize; x++ {
		for y := 0; y < imageSize; y++ {
			sample := barkNoise(vector.NewVector2(float64(x), (float64(y))))
			dc.SetColor(colors.LinearSample(sample))
			dc.SetPixel(x, y)
		}
	}

	return dc.SavePNG("bark.png")
}

func Cone(base float64, points ...vector.Vector3) mesh.Mesh {
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
			UvPoint:     vector.NewVector2(0, size*3),
		}
	}

	return extrude.Polygon(16, extrusionPoints)
}

func Tree(height, base, percentageCovered float64, branchSnowNoise noise.Sampler2D, pos vector.Vector3) mesh.Mesh {
	percentBare := 1 - percentageCovered

	heightCovered := height * percentageCovered
	heightBare := height * percentBare

	branchCount := 200 + int(rand.Float64()*300)
	branchLength := height * 0.25 * (.8 + (.4 * rand.Float64()))

	branches := mesh.EmptyMesh()
	for i := 0; i < branchCount; i++ {
		branchHeight := (heightCovered * rand.Float64()) + heightBare

		availableHeightUsed := (branchHeight - heightBare) / heightCovered

		trailOffGivenHeight := ((1 - availableHeightUsed) + .2)

		branchMaxWidth := (base) * 2 * trailOffGivenHeight * (1 + (.4 * rand.Float64()))

		dir := vector.NewVector3(-1+(2*rand.Float64()), 0, -1+(2*rand.Float64())).
			Normalized().
			MultByConstant(branchLength * trailOffGivenHeight)

		branchIndex := int(math.Floor(4 * branchSnowNoise(pos.XZ().Add(dir.XZ()))))
		xCordOfBranch := branchIndex % 2
		yCordOfBranch := math.Floor(float64(branchIndex) / 2.)

		branchUV := vector.NewVector2(0.25, 1).
			Add(vector.NewVector2(float64(xCordOfBranch)*.5, -yCordOfBranch*.5))
		branchUVLength := 0.5

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

	branchImage := "branch.png"
	branchNormalImage := "branch_normal.png"
	barkImage := "bark.png"

	branches = branches.SetMaterial(mesh.Material{
		Name:             "Branches",
		DiffuseColor:     color.RGBA{0, 143, 45, 255},
		ColorTextureURI:  &branchImage,
		NormalTextureURI: &branchNormalImage,
	})

	return Cone(
		base,
		vector.NewVector3(0, 0, 0),
		vector.NewVector3(0, height, 0),
	).
		CalculateSmoothNormals().
		SetMaterial(mesh.Material{
			Name:            "Trunk",
			DiffuseColor:    color.RGBA{99, 62, 10, 255},
			ColorTextureURI: &barkImage,
		}).
		Append(branches).
		Translate(pos)
}

func main() {
	numTree := 1
	forestWidth := 100.
	forest := mesh.EmptyMesh()

	for i := 0; i < numTree; i++ {

		treePos := vector.NewVector3(
			rand.Float64()*forestWidth,
			0,
			rand.Float64()*forestWidth,
		)

		forest = forest.Append(
			Tree(
				20+(25*rand.Float64()),
				0.5+(rand.Float64()*2),
				.7+(.2*rand.Float64()),
				noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
					{Scalar: 1 / 150., Amplitude: 1. / 2},
					{Scalar: 1 / 75., Amplitude: 1. / 4},
					{Scalar: 1 / 37.5, Amplitude: 1. / 8},
					{Scalar: 1 / 18., Amplitude: 1. / 16},
				}).Value),
				treePos,
			),
		)
	}

	// BranchTexture(coloring.NewColorStack([]coloring.ColorStackEntry{
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{12, 89, 36, 255}),
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{3, 191, 0, 255}),
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{2, 69, 23, 255}),
	// }), 2048)
	TrunkTexture(
		1024,
		coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{115, 87, 71, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{97, 61, 41, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{102, 78, 44, 255}),
		}),
		noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
			{Scalar: 1 / 50., Amplitude: 1. / 2},
			{Scalar: 1 / 25., Amplitude: 1. / 4},
			{Scalar: 1 / 12.5, Amplitude: 1. / 8},
			{Scalar: 1 / 7.25, Amplitude: 1. / 16},
		}).Value),
	)

	obj.Save("chill.obj", forest)
}
