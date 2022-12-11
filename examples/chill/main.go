package main

import (
	"image/color"
	"math/rand"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/extrude"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/vector"
	"github.com/fogleman/gg"
)

func BranchTexture() {
	const W = 1024
	const H = 1024
	halfHeight := H / 2.
	dc := gg.NewContext(W, H)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	branchWidth := 20.

	dc.SetColor(color.RGBA{99, 62, 10, 255})
	dc.SetLineWidth(branchWidth)
	dc.DrawLine(W/2, 0, W/2, H)
	dc.Stroke()

	chanceOfSnow := 0.3

	for i := 0; i < 200; i++ {
		x1 := 0.5*W - (branchWidth / 2) + (branchWidth * rand.Float64())
		y1 := rand.Float64() * H
		x2 := rand.Float64() * W
		y2 := y1 - ((rand.Float64() * halfHeight) + halfHeight)

		w := rand.Float64()*4 + 1

		if rand.Float64() <= chanceOfSnow {
			dc.SetColor(color.RGBA{255, 255, 255, 255})
		} else {
			dc.SetColor(color.RGBA{0, 143, 45, 255})
		}

		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	dc.SetColor(color.RGBA{255, 255, 255, 255})
	for i := 0; i < 30; i++ {
		x1 := 0.5*W - (branchWidth / 2) + (branchWidth * rand.Float64())
		y1 := rand.Float64() * H
		x2 := rand.Float64() * W
		y2 := y1 - ((rand.Float64() * halfHeight) + halfHeight)

		w := rand.Float64()*4 + 1

		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	dc.SavePNG("branch.png")
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
			UvPoint:     vector.NewVector2(0, size),
		}
	}

	return extrude.Polygon(16, extrusionPoints)
}

func Tree(height, base, percentageCovered float64) mesh.Mesh {
	percentBare := 1 - percentageCovered

	heightCovered := height * percentageCovered
	heightBare := height * percentBare

	branchCount := 300
	branchLength := height * 0.25 * (.8 + (.4 * rand.Float64()))

	branches := mesh.EmptyMesh()
	for i := 0; i < branchCount; i++ {
		branchHeight := (heightCovered * rand.Float64()) + heightBare

		availableHeightUsed := (branchHeight - heightBare) / heightCovered

		trailOffGivenHeight := ((1 - availableHeightUsed) + .2)

		dir := vector.NewVector3(-1+(2*rand.Float64()), 0, -1+(2*rand.Float64())).
			Normalized().
			MultByConstant(branchLength * trailOffGivenHeight)

		branches = branches.Append(extrude.Line([]extrude.LinePoint{
			{
				Point:   vector.NewVector3(0, branchHeight, 0),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   (base) * trailOffGivenHeight * 0.35,
				Uv:      vector.NewVector2(0.5, 0),
				UvWidth: 1,
			},
			{
				Point:   dir.MultByConstant(.25).SetY(branchHeight - 1),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   (base) * 2 * trailOffGivenHeight,
				Uv:      vector.NewVector2(0.5, .25),
				UvWidth: 1,
			},
			{
				Point:   dir.MultByConstant(.5).SetY(branchHeight - 1.5),
				Up:      vector.Vector3Up(),
				Height:  -(height / 30),
				Width:   (base) * 2 * trailOffGivenHeight * 0.75,
				Uv:      vector.NewVector2(0.5, .75),
				UvWidth: 1,
			},
			{
				Point:   dir.SetY(branchHeight - 2),
				Up:      vector.Vector3Up(),
				Height:  0,
				Width:   (base) * trailOffGivenHeight * 0.35,
				Uv:      vector.NewVector2(0.5, 1),
				UvWidth: 1,
			},
		}))
	}

	branchImage := "branch.png"

	branches = branches.SetMaterial(mesh.Material{
		Name:            "Branches",
		DiffuseColor:    color.RGBA{0, 143, 45, 255},
		ColorTextureURI: &branchImage,
	})

	return Cone(
		base,
		vector.NewVector3(0, 0, 0),
		vector.NewVector3(0, height, 0),
	).
		SetMaterial(mesh.Material{
			Name:         "Trunk",
			DiffuseColor: color.RGBA{99, 62, 10, 255},
		}).
		Append(branches)
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
			).
				Translate(treePos),
		)
	}

	BranchTexture()

	obj.Save("chill.obj", forest)
}
