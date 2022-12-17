package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/vector"
)

func randomVec2Radial() vector.Vector2 {
	theta := rand.Float64() * 2 * math.Pi
	return vector.
		NewVector2(math.Cos(theta), math.Sin(theta)).
		MultByConstant(math.Sqrt(rand.Float64()))
}

func calcTreePositions(count int, forestWidth float64, terrainHeight noise.Sampler2D, path []vector.Vector2) []vector.Vector3 {
	positions := make([]vector.Vector3, 0)
	for i := 0; i < count; i++ {
		xz := randomVec2Radial().
			MultByConstant((forestWidth / 2) * .8).
			Add(vector.NewVector2(forestWidth/2, forestWidth/2))
		y := terrainHeight(xz) - 1

		invalid := false
		for i := 1; i < len(path) && !invalid; i++ {
			line := mesh.NewLine2D(path[i], path[i-1])
			dist := line.ClosestPointOnLine(xz).Distance(xz)
			if dist < 20 {
				invalid = true
			}

		}

		if !invalid {
			positions = append(positions, vector.NewVector3(xz.X(), y, xz.Y()))

		}

	}
	return positions
}

func main() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	log.Printf("Generating with Seed: %d\n", seed)

	terrainPBR := PBRTextures{name: "terrain"}
	branchPBR := PBRTextures{name: "branch"}

	totalHeight := 200.
	terrainHeight := noise.PerlinStack([]noise.Stack2DEntry{
		{Scalar: 1 / 300., Amplitude: totalHeight / 2},
		{Scalar: 1 / 150., Amplitude: totalHeight / 8},
		{Scalar: 1 / 75., Amplitude: totalHeight / 16},
		{Scalar: 1 / 37.5, Amplitude: totalHeight / 32},
	})

	numTree := 200
	forestWidth := 400.
	terrain, maxTerrainValue := Terrain(forestWidth, terrainHeight.Value, &terrainPBR)

	snowColors := coloring.NewColorStack([]coloring.ColorStackEntry{
		coloring.NewColorStackEntry(10, 1, 1, color.RGBA{255, 255, 255, 255}),
		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{245, 247, 255, 255}),
	})

	terrainImageSize := 1024

	TerrainTexture(
		terrainImageSize,
		forestWidth,
		&terrainPBR,
		snowColors,
		maxTerrainValue,
		terrainHeight.Value,
	)

	snowPath := []vector.Vector2{
		vector.NewVector2(forestWidth/2, forestWidth/2),
		vector.NewVector2(forestWidth, forestWidth),
	}

	terrain = DrawTrail(
		terrain,
		&terrainPBR,
		snowPath,
		forestWidth,
		terrainImageSize,
		snowColors,
	)

	treePositions := calcTreePositions(numTree, forestWidth, terrainHeight.Value, snowPath)
	for _, pos := range treePositions {
		terrain = terrain.Append(
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
				pos,
				branchPBR,
			),
		)
	}

	// BranchTexture(coloring.NewColorStack([]coloring.ColorStackEntry{
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{12, 89, 36, 255}),
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{3, 191, 0, 255}),
	// 	coloring.NewColorStackEntry(1, 1, 1, color.RGBA{2, 69, 23, 255}),
	// }), branchPBR, 2048)

	// TrunkTexture(
	// 	1024,
	// 	coloring.NewColorStack([]coloring.ColorStackEntry{
	// 		// coloring.NewColorStackEntry(1, 1, 1, color.RGBA{115, 87, 71, 255}),
	// 		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{97, 61, 41, 255}),
	// 		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{102, 78, 44, 255}),
	// 	}),
	// 	noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
	// 		{Scalar: 1 / 50., Amplitude: 1. / 2},
	// 		{Scalar: 1 / 25., Amplitude: 1. / 4},
	// 		{Scalar: 1 / 12.5, Amplitude: 1. / 8},
	// 		{Scalar: 1 / 7.25, Amplitude: 1. / 16},
	// 	}).Value),
	// )

	terrainPBR.Save()
	branchPBR.Save()
	obj.Save("chill.obj", terrain)
}
