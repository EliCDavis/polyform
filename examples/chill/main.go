package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/coloring"
	"github.com/EliCDavis/mesh/formats/obj"
	"github.com/EliCDavis/mesh/noise"
	"github.com/EliCDavis/vector"
	"github.com/urfave/cli/v2"
)

func branchColorPallettes() map[string]coloring.ColorStack {
	return map[string]coloring.ColorStack{
		"green": coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{12, 89, 36, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{3, 191, 0, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{2, 69, 23, 255}),
		}),
		"dead": coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{234, 226, 126, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{228, 179, 95, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{171, 120, 52, 255}),
		}),
		"red": coloring.NewColorStack([]coloring.ColorStackEntry{
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{168, 50, 62, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{224, 94, 107, 255}),
			coloring.NewColorStackEntry(1, 1, 1, color.RGBA{204, 22, 86, 255}),
		}),
	}
}

func randomVec2Radial() vector.Vector2 {
	theta := rand.Float64() * 2 * math.Pi
	return vector.
		NewVector2(math.Cos(theta), math.Sin(theta)).
		MultByConstant(math.Sqrt(rand.Float64()))
}

func calcTreePositions(count int, forestWidth float64, terrainHeight noise.Sampler2D, path Trail) []vector.Vector3 {
	positions := make([]vector.Vector3, 0)
	for i := 0; i < count; i++ {
		xz := randomVec2Radial().
			MultByConstant((forestWidth / 2) * .8).
			Add(vector.NewVector2(forestWidth/2, forestWidth/2))
		y := terrainHeight(xz) - 1

		invalid := false
		for _, seg := range path.Segments {
			line := mesh.NewLine2D(
				vector.NewVector2(
					seg.StartX,
					seg.StartY,
				),
				vector.NewVector2(
					seg.EndX,
					seg.EndY,
				),
			)
			dist := line.ClosestPointOnLine(xz).Distance(xz)
			if dist < seg.Width {
				invalid = true
			}

		}

		if !invalid {
			positions = append(positions, vector.NewVector3(xz.X(), y, xz.Y()))

		}

	}
	return positions
}

func initSeed(ctx *cli.Context) {
	if ctx.IsSet("seed") {
		rand.Seed(ctx.Int64("seed"))
	} else {
		seed := time.Now().UnixNano()
		rand.Seed(seed)
		log.Printf("Generating with Seed: %d\n", seed)
	}
}

func main() {

	app := &cli.App{
		Name: "chill",
		Authors: []*cli.Author{
			{
				Name:  "Eli Davis",
				Email: "eli@recolude.com",
			},
		},
		Version:     "1.0.0",
		Description: "ProcJam 2022 Submission",
		Commands: []*cli.Command{
			{
				Name:        "tree",
				Aliases:     []string{"t"},
				Description: "Creates a single tree",
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:        "height",
						Usage:       "The height of the tree",
						DefaultText: "random value between 20 and 45",
					},
					&cli.Float64Flag{
						Name:        "base",
						Usage:       "The width of the tree trunk",
						DefaultText: "random value between 0.5 and 2.5",
					},
					&cli.Float64Flag{
						Name:        "covered",
						Usage:       "Percent of the tree's trunk covered by it's branches",
						DefaultText: "random value between 0.7 and 0.9",
					},
					&cli.IntFlag{
						Name:        "branches",
						Usage:       "Number of branches the tree will have",
						DefaultText: "random value between 200 and 500",
					},
					&cli.StringFlag{
						Name:  "branch-pallette",
						Usage: "The color pallette to use for coloring the tree",
						Value: "green",
					},

					&cli.Float64Flag{
						Name:  "min-snow",
						Usage: "Minimum percentage of snow to use on the branch textures",
						Value: .4,
					},
					&cli.Float64Flag{
						Name:  "max-snow",
						Usage: "Maximum percentage of snow to use on the branch textures",
						Value: .7,
					},

					&cli.Int64Flag{
						Name:        "seed",
						Usage:       "The seed fpr the random number generator to use",
						DefaultText: "clock time",
					},
					&cli.StringFlag{
						Name:        "out",
						Usage:       "Path to write the tree obj/mtl/pngs to",
						DefaultText: ".",
						Value:       ".",
					},
					&cli.StringFlag{
						Name:        "name",
						Usage:       "Name of the files that will be generated",
						DefaultText: "tree",
						Value:       "tree",
					},
				},
				Action: func(ctx *cli.Context) error {
					filePrefixes := path.Join(ctx.String("out"), ctx.String("name"))

					initSeed(ctx)

					var treeHeight float64
					if ctx.IsSet("height") {
						treeHeight = ctx.Float64("height")
					} else {
						treeHeight = 20 + (25 * rand.Float64())
					}

					var treeBase float64
					if ctx.IsSet("base") {
						treeBase = ctx.Float64("base")
					} else {
						treeBase = 0.5 + (rand.Float64() * 2)
					}

					var treeCovered float64
					if ctx.IsSet("covered") {
						treeCovered = ctx.Float64("covered")
					} else {
						treeCovered = .7 + (.2 * rand.Float64())
					}

					var branches int
					if ctx.IsSet("branches") {
						branches = ctx.Int("branches")
					} else {
						branches = 200 + int(300*rand.Float64())
					}

					branchPBR := PBRTextures{
						name: ctx.String("name") + "_branch",
						path: ctx.String("out"),
					}

					barkPBR := PBRTextures{
						name: ctx.String("name") + "_bark",
						path: ctx.String("out"),
					}

					atlas := BranchTexture(
						branchColorPallettes()[ctx.String("branch-pallette")],
						&branchPBR, 2048,
						ctx.Float64("min-snow"),
						ctx.Float64("max-snow"),
						2,
						1,
						4,
					)

					trunk, branch := Tree(
						treeHeight,
						treeBase,
						treeCovered,
						branches,
						atlas,
					)

					tree := trunk.SetMaterial(barkPBR.Material()).
						Append(branch.SetMaterial(branchPBR.Material()))

					TrunkTexture(
						1024,
						coloring.NewColorStack([]coloring.ColorStackEntry{
							// coloring.NewColorStackEntry(1, 1, 1, color.RGBA{115, 87, 71, 255}),
							coloring.NewColorStackEntry(1, 1, 1, color.RGBA{71, 43, 6, 255}),
							coloring.NewColorStackEntry(1, 1, 1, color.RGBA{94, 63, 21, 255}),
						}),
						noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
							{Scalar: 1 / 50., Amplitude: 1. / 2},
							{Scalar: 1 / 25., Amplitude: 1. / 4},
							{Scalar: 1 / 12.5, Amplitude: 1. / 8},
							{Scalar: 1 / 7.25, Amplitude: 1. / 16},
						}).Value),
						&barkPBR,
					)

					err := branchPBR.Save()
					if err != nil {
						return err
					}

					err = barkPBR.Save()
					if err != nil {
						return err
					}

					return obj.Save(fmt.Sprintf("%s.obj", filePrefixes), tree)
				},
			},
			{
				Name:        "forest",
				Aliases:     []string{"f"},
				Description: "Creates a forest of trees",
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  "max-height",
						Usage: "Max height of the terrain",
						Value: 200,
					},
					&cli.Float64Flag{
						Name:  "forest-width",
						Usage: "Diameter of forest",
						Value: 400,
					},
					&cli.IntFlag{
						Name:  "tree-count",
						Usage: "Number of trees the forest will contain",
						Value: 200,
					},
					&cli.StringFlag{
						Name:  "trail",
						Usage: "Path to a JSON file containing trail data",
					},

					&cli.Float64Flag{
						Name:  "min-tree-height",
						Usage: "The minimum height of a tree in the forest",
						Value: 20,
					},
					&cli.Float64Flag{
						Name:  "min-tree-base",
						Usage: "The minimum width of a trunk of a tree in the forest",
						Value: 0.5,
					},
					&cli.Float64Flag{
						Name:  "min-tree-covered",
						Usage: "The minimum percent of a tree's trunk covered by it's branches in the forest",
						Value: 0.7,
					},
					&cli.IntFlag{
						Name:  "min-tree-branches",
						Usage: "The minimum number of branches a tree will have",
						Value: 200,
					},

					&cli.Float64Flag{
						Name:  "max-tree-height",
						Usage: "The maximum height of a tree in the forest",
						Value: 45,
					},
					&cli.Float64Flag{
						Name:  "max-tree-base",
						Usage: "The maximum width of a trunk of a tree in the forest",
						Value: 2.5,
					},
					&cli.Float64Flag{
						Name:  "max-tree-covered",
						Usage: "The maximum percent of a tree's trunk covered by it's branches in the forest",
						Value: 0.9,
					},
					&cli.IntFlag{
						Name:  "max-tree-branches",
						Usage: "The maximum number of branches a tree will have",
						Value: 500,
					},
					&cli.StringFlag{
						Name:  "branch-pallette",
						Usage: "The color pallette to use for coloring the tree",
						Value: "green",
					},

					&cli.Float64Flag{
						Name:  "min-snow",
						Usage: "Minimum percentage of snow to use on the branch textures",
						Value: .2,
					},
					&cli.Float64Flag{
						Name:  "max-snow",
						Usage: "Maximum percentage of snow to use on the branch textures",
						Value: .9,
					},

					&cli.Int64Flag{
						Name:        "seed",
						Usage:       "The seed fpr the random number generator to use",
						DefaultText: "clock time",
					},
					&cli.StringFlag{
						Name:        "out",
						Usage:       "Path to write the tree obj/mtl/pngs to",
						DefaultText: ".",
						Value:       ".",
					},
					&cli.StringFlag{
						Name:        "name",
						Usage:       "Name of the files that will be generated",
						DefaultText: "forest",
						Value:       "forest",
					},
				},
				Action: func(ctx *cli.Context) error {
					filePrefixes := path.Join(ctx.String("out"), ctx.String("name"))

					initSeed(ctx)

					terrainPBR := PBRTextures{
						name: ctx.String("name") + "terrain",
						path: ctx.String("out"),
					}
					branchPBR := PBRTextures{
						name: ctx.String("name") + "branch",
						path: ctx.String("out"),
					}
					barkPBR := PBRTextures{
						name: ctx.String("name") + "_bark",
						path: ctx.String("out"),
					}

					totalHeight := ctx.Float64("max-height")
					terrainHeight := noise.PerlinStack([]noise.Stack2DEntry{
						{Scalar: 1 / 300., Amplitude: totalHeight / 2},
						{Scalar: 1 / 150., Amplitude: totalHeight / 8},
						{Scalar: 1 / 75., Amplitude: totalHeight / 16},
						{Scalar: 1 / 37.5, Amplitude: totalHeight / 32},
					})

					numTree := ctx.Int("tree-count")
					forestWidth := ctx.Float64("forest-width")
					terrain, maxTerrainValue := Terrain(forestWidth, terrainHeight.Value, &terrainPBR)

					snowColors := coloring.NewColorStack([]coloring.ColorStackEntry{
						coloring.NewColorStackEntry(10, 1, 1, color.RGBA{255, 255, 255, 255}),
						coloring.NewColorStackEntry(1, 1, 1, color.RGBA{245, 247, 255, 255}),
					})

					terrainImageSize := 1024 * 4

					TerrainTexture(
						terrainImageSize,
						forestWidth,
						&terrainPBR,
						snowColors,
						maxTerrainValue,
						terrainHeight.Value,
					)

					var snowPath Trail

					if ctx.IsSet("trail") {

						trailFileData, err := ioutil.ReadFile(ctx.String("trail"))
						if err != nil {
							return err
						}

						if err := json.Unmarshal(trailFileData, &snowPath); err != nil {
							return err
						}

						terrain = DrawTrail(
							terrain,
							&terrainPBR,
							snowPath,
							forestWidth,
							terrainImageSize,
							snowColors,
						)
					}

					atlas := BranchTexture(
						branchColorPallettes()[ctx.String("branch-pallette")],
						&branchPBR,
						2048,
						ctx.Float64("min-snow"),
						ctx.Float64("max-snow"),
						4,
						1,
						4,
					)

					treePositions := calcTreePositions(numTree, forestWidth, terrainHeight.Value, snowPath)

					minTreeHeight := ctx.Float64("min-tree-height")
					maxTreeHeight := ctx.Float64("max-tree-height")

					minTreeBase := ctx.Float64("min-tree-base")
					maxTreeBase := ctx.Float64("max-tree-base")

					minTreeCovered := ctx.Float64("min-tree-covered")
					maxTreeCovered := ctx.Float64("max-tree-covered")

					minTreeBranches := ctx.Int("min-tree-branches")
					maxTreeBranches := ctx.Int("max-tree-branches")

					treeColorNoise := noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
						{Scalar: 1 / 150., Amplitude: 1. / 2},
						{Scalar: 1 / 75., Amplitude: 1. / 4},
						{Scalar: 1 / 37.5, Amplitude: 1. / 8},
						{Scalar: 1 / 18., Amplitude: 1. / 16},
					}).Value)

					trunks := mesh.EmptyMesh()
					branches := mesh.EmptyMesh()

					for _, pos := range treePositions {
						trunk, branch := Tree(
							minTreeHeight+((maxTreeHeight-minTreeHeight)*rand.Float64()),
							minTreeBase+((maxTreeBase-minTreeBase)*rand.Float64()),
							minTreeCovered+((maxTreeCovered-minTreeCovered)*rand.Float64()),
							minTreeBranches+int(float64(maxTreeBranches-minTreeBranches)*rand.Float64()),
							atlas.SubAtlas[int(math.Floor(treeColorNoise(pos.XZ())*float64(len(atlas.SubAtlas))))],
						)
						trunks = trunks.Append(trunk.Translate(pos))
						branches = branches.Append(branch.Translate(pos))
					}
					terrain = terrain.
						Append(trunks.SetMaterial(barkPBR.Material())).
						Append(branches.SetMaterial(branchPBR.Material()))

					TrunkTexture(
						1024,
						coloring.NewColorStack([]coloring.ColorStackEntry{
							coloring.NewColorStackEntry(1, 1, 1, color.RGBA{71, 43, 6, 255}),
							coloring.NewColorStackEntry(1, 1, 1, color.RGBA{94, 63, 21, 255}),
						}),
						noise.Sampler2D(noise.PerlinStack([]noise.Stack2DEntry{
							{Scalar: 1 / 50., Amplitude: 1. / 2},
							{Scalar: 1 / 25., Amplitude: 1. / 4},
							{Scalar: 1 / 12.5, Amplitude: 1. / 8},
							{Scalar: 1 / 7.25, Amplitude: 1. / 16},
						}).Value),
						&barkPBR,
					)

					err := terrainPBR.Save()
					if err != nil {
						return err
					}
					err = branchPBR.Save()
					if err != nil {
						return err
					}

					err = barkPBR.Save()
					if err != nil {
						return err
					}
					return obj.Save(fmt.Sprintf("%s.obj", filePrefixes), terrain)
				},
			},
			{
				Name:        "word",
				Description: "Writes out the word 'chill' in the trail segment format",
				Action: func(ctx *cli.Context) error {
					c := []vector.Vector2{
						vector.NewVector2(1, 1),
						vector.NewVector2(0, 0.5),
						vector.NewVector2(1, 0),
					}

					h := []vector.Vector2{
						vector.NewVector2(0, 0),
						vector.NewVector2(0, 1),
						vector.NewVector2(0, 0.5),
						vector.NewVector2(1, 0.5),
						vector.NewVector2(1, 1),
						vector.NewVector2(1, 0),
					}

					i := []vector.Vector2{
						vector.NewVector2(0.5, 0),
						vector.NewVector2(0.5, 1),
					}

					l := []vector.Vector2{
						vector.NewVector2(0, 0),
						vector.NewVector2(0, 1),
						vector.NewVector2(1, 1),
					}

					word := [][]vector.Vector2{c, h, i, l, l}

					characterWidth := 35.
					height := 90.
					characterSpacing := 40.
					terrainSize := 500.
					offset := vector.NewVector2(terrainSize/2., terrainSize/2.).
						Sub(vector.NewVector2((characterWidth+characterSpacing)*0.5*float64(len(word)), height/2))

					trail := Trail{
						Segments: make([]TrailSegment, 0),
					}

					for charIndex, character := range word {
						characterOffset := vector.NewVector2(float64(charIndex)*(characterSpacing+characterWidth), 0)
						for pIndex := 1; pIndex < len(character); pIndex++ {
							start := vector.NewVector2(
								character[pIndex-1].X()*characterWidth,
								character[pIndex-1].Y()*height,
							).
								Add(characterOffset).
								Add(offset)

							end := vector.NewVector2(
								character[pIndex].X()*characterWidth,
								character[pIndex].Y()*height,
							).
								Add(characterOffset).
								Add(offset)

							trail.Segments = append(trail.Segments, TrailSegment{
								Width:  30,
								Depth:  15,
								StartX: start.X(),
								StartY: start.Y(),
								EndX:   end.X(),
								EndY:   end.Y(),
							})
						}
					}

					strB, err := json.Marshal(trail)
					if err != nil {
						return err
					}
					fmt.Fprint(ctx.App.Writer, string(strB))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
