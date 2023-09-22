package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func texture() image.Image {
	ctx := gg.NewContext(2, 2)
	ctx.SetColor(color.RGBA{
		R: 0,
		G: 0, // Roughness
		B: 0, // Metal - 0 = metal
		A: 255,
	})
	ctx.SetPixel(0, 0)
	ctx.SetPixel(1, 0)
	ctx.SetPixel(0, 1)
	ctx.SetPixel(1, 1)

	return ctx.Image()
}

func DiscoScene(c *generator.Context) (generator.Artifact, error) {
	ballParams := c.Parameters.Group("Disco Ball")
	tableParams := c.Parameters.Group("Table")
	discoball := primitives.
		UVSphereUnwelded(
			ballParams.Float64("radius"),
			ballParams.Int("rows"),
			ballParams.Int("column"),
		).Transform(meshops.FlatNormalsTransformer{})

	discoBallHeight := vector3.Up[float64]().Scale(ballParams.Float64("height"))
	discoNormals := discoball.Float3Attribute(modeling.NormalAttribute)

	discoball = discoball.ModifyFloat3Attribute(
		modeling.PositionAttribute,
		func(i int, v vector3.Float64) vector3.Float64 {
			return v.Add(discoNormals.At(i).Scale(ballParams.Float64("panel offset")))
		}).
		// Base connecting ball to the rod
		Append(primitives.
			Cylinder(15, 0.1, 0.2).
			Translate(vector3.New(0, ballParams.Float64("radius")+ballParams.Float64("panel offset"), 0)),
		).
		// Rod the ball is hanging from
		Append(primitives.
			Cylinder(4, 3., 0.025).
			Translate(vector3.New(0, ballParams.Float64("radius")+ballParams.Float64("panel offset")+1.5, 0)),
		).
		Translate(discoBallHeight)

	return generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{
					Name: "Disco Ball",
					Mesh: discoball,
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							MetallicFactor:  1,
							RoughnessFactor: 0,
							BaseColorFactor: ballParams.Color("Color"),
							MetallicRoughnessTexture: &gltf.PolyformTexture{
								URI: "mr.png",
							},
						},
					},
				},
				{
					Name: "Table",
					Mesh: primitives.Cylinder(
						tableParams.Int("Resolution"),
						tableParams.Float64("Thickness"),
						tableParams.Float64("Radius"),
					),
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							BaseColorFactor: tableParams.Color("Color"),
						},
					},
				},
			},
		},
	}, nil
}

func main() {
	app := generator.App{
		Name:        "Woodland Disco Romance",
		Version:     "1.0.0",
		Description: "Applying color pallettes to a sample room",
		Generator: generator.Generator{
			Parameters: &generator.GroupParameter{
				Name: "Disco",
				Parameters: []generator.Parameter{
					&generator.GroupParameter{
						Name: "Disco Ball",
						Parameters: []generator.Parameter{
							&generator.IntParameter{
								Name:         "rows",
								DefaultValue: 30,
							},

							&generator.IntParameter{
								Name:         "column",
								DefaultValue: 45,
							},

							&generator.FloatParameter{
								Name:         "radius",
								DefaultValue: 1,
							},

							&generator.FloatParameter{
								Name:         "panel offset",
								DefaultValue: 0.1,
							},

							&generator.FloatParameter{
								Name:         "height",
								DefaultValue: 5.5,
							},

							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0x85,
									G: 0x87,
									B: 0x82,
									A: 255,
								},
							},
						},
					},
					&generator.GroupParameter{
						Name: "Table",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Radius",
								DefaultValue: 4,
							},
							&generator.FloatParameter{
								Name:         "Thickness",
								DefaultValue: 0.1,
							},
							&generator.IntParameter{
								Name:         "Resolution",
								DefaultValue: 30,
							},
							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0x96,
									G: 0x77,
									B: 0x22,
									A: 255,
								},
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"disco.glb": DiscoScene,
				"mr.png": func(c *generator.Context) (generator.Artifact, error) {
					return generator.ImageArtifact{
						Image: texture(),
					}, nil
				},
			},
		},
	}
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
