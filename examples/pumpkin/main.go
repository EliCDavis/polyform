package main

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
)

func pumpkinSegmentField(maxWidth, topDip float64) marching.Field {
	return marching.VarryingThicknessLine([]marching.LinePoint{
		{
			Point:  vector3.New(0., .3, 0.),
			Radius: 0.1,
		},
		{
			Point:  vector3.New(0., .2, maxWidth/2),
			Radius: 0.1,
		},

		{
			Point:  vector3.New(0, 0.5, maxWidth),
			Radius: 0.3,
		},

		{
			Point:  vector3.New(0., .8, maxWidth*0.75),
			Radius: 0.2,
		},
		{
			Point:  vector3.New(0., 1-topDip, 0.),
			Radius: 0.1,
		},
	}, 1)
}

func pumpkinStem(maxWidth, minWidth, length, tipOffset float64) marching.Field {
	return marching.VarryingThicknessLine([]marching.LinePoint{
		{
			Point:  vector3.New(0., 0, 0.),
			Radius: maxWidth,
		},

		{
			Point:  vector3.New(0., length*.5, 0.),
			Radius: minWidth,
		},

		{
			Point:  vector3.New(tipOffset, length, 0.),
			Radius: minWidth,
		},
	}, 1)
}

func main() {

	app := generator.App{
		Name:        "Pumpkin",
		Version:     "0.0.1",
		Description: "Making a pumpkin for Haloween",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Near:  2,
				Far:   10,
				Color: coloring.WebColor{R: 0, G: 0, B: 0, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x4f, G: 0x6d, B: 0x55, A: 255},
			Background: coloring.WebColor{R: 0, G: 0, B: 0, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Generator: generator.Generator{
			Parameters: &generator.GroupParameter{
				Name: "Pumpkin",
				Parameters: []generator.Parameter{
					&generator.FloatParameter{
						Name:         "Cubes Per Unit",
						DefaultValue: 20,
					},

					&generator.IntParameter{
						Name:         "Wedges",
						DefaultValue: 10,
					},

					&generator.FloatParameter{
						Name:         "Wedge Spacing",
						DefaultValue: .1,
					},

					&generator.FloatParameter{
						Name:         "Max Width",
						DefaultValue: .3,
					},

					&generator.FloatParameter{
						Name:         "Top Dip",
						DefaultValue: .2,
					},

					&generator.ColorParameter{
						Name:         "Color",
						DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
					},
					&generator.GroupParameter{
						Name: "Stem",
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Color",
								DefaultValue: coloring.WebColor{R: 0x6d, G: 0x52, B: 0x40, A: 255},
							},
							&generator.FloatParameter{
								Name:         "Base Width",
								DefaultValue: 0.1,
							},
							&generator.FloatParameter{
								Name:         "Tip Width",
								DefaultValue: 0.06,
							},
							&generator.FloatParameter{
								Name:         "Length",
								DefaultValue: 0.3,
							},
							&generator.FloatParameter{
								Name:         "Tip Offset",
								DefaultValue: 0.1,
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"pumpkin.glb": func(c *generator.Context) (generator.Artifact, error) {

					pumpkinWedgeCanvas := marching.NewMarchingCanvas(c.Parameters.Float64("Cubes Per Unit"))
					pumpkinWedgeCanvas.AddFieldParallel(pumpkinSegmentField(
						c.Parameters.Float64("Max Width"),
						c.Parameters.Float64("Top Dip"),
					))
					pumpkinWedge := pumpkinWedgeCanvas.
						MarchParallel(0).
						Transform(meshops.LaplacianSmoothTransformer{
							Iterations:      20,
							SmoothingFactor: 0.1,
						})

					stemParams := c.Parameters.Group("Stem")
					stemCanvas := marching.NewMarchingCanvas(c.Parameters.Float64("Cubes Per Unit"))
					stemCanvas.AddFieldParallel(pumpkinStem(
						stemParams.Float64("Base Width"),
						stemParams.Float64("Tip Width"),
						stemParams.Float64("Length"),
						stemParams.Float64("Tip Offset"),
					))
					stem := stemCanvas.
						MarchParallel(0).
						Transform(
							meshops.LaplacianSmoothTransformer{
								Iterations:      20,
								SmoothingFactor: 0.1,
							},
							meshops.TranslateAttribute3DTransformer{
								Amount: vector3.New(0., 1-c.Parameters.Float64("Top Dip"), 0.),
							},
						)

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "Pumpkin",
									Mesh: repeat.
										Circle(
											pumpkinWedge,
											c.Parameters.Int("Wedges"),
											c.Parameters.Float64("Wedge Spacing"),
										).
										Transform(
											meshops.SmoothNormalsTransformer{},
										),
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorFactor: c.Parameters.Color("Color"),
										},
									},
								},
								{
									Name: "Stem",
									Mesh: stem,
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorFactor: stemParams.Color("Color"),
										},
									},
								},
							},
						},
					}, nil
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
