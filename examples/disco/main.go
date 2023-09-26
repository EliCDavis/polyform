package main

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func texture(metal float64, roughness float64) image.Image {
	ctx := gg.NewContext(2, 2)
	ctx.SetColor(color.RGBA{
		R: 0,
		G: uint8(roughness * 255),   // Roughness
		B: byte((1. - metal) * 255), // Metal - 0 = metal
		A: 255,
	})
	ctx.SetPixel(0, 0)
	ctx.SetPixel(1, 0)
	ctx.SetPixel(0, 1)
	ctx.SetPixel(1, 1)

	return ctx.Image()
}

func Plate(params generator.GroupParameter) modeling.Mesh {
	return primitives.Cylinder{
		Sides:  params.Int("Resolution"),
		Height: params.Float64("Thickness"),
		Radius: params.Float64("Radius"),
	}.ToMesh()
}

func Chair(params generator.GroupParameter) modeling.Mesh {

	chairHeight := params.Float64("Height")
	chairWidth := params.Float64("Width")
	chairLength := params.Float64("Length")

	halfHeight := chairHeight / 2
	halfWidth := chairWidth / 2
	halfLength := params.Float64("Length") / 2

	// LEGS ===================================================================

	legParams := params.Group("Leg")
	legRadius := legParams.Float64("Radius")
	legInset := legParams.Float64("Inset")

	leg := primitives.Cylinder{
		Sides:  8,
		Height: params.Float64("Height"),
		Radius: legRadius,
	}.ToMesh()

	legRadiusAndInset := legRadius + legInset

	legSupportFrontBackRotation := modeling.UnitQuaternionFromTheta(math.Pi/2, vector3.Forward[float64]())
	legFrontBackSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairWidth - (legRadiusAndInset * 2),
		Radius: legRadius / 2,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    legSupportFrontBackRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    legSupportFrontBackRotation,
		},
	)

	legSupportLeftRightRotation := modeling.UnitQuaternionFromTheta(math.Pi/2, vector3.Right[float64]())
	legLeftRightSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairLength - (legRadiusAndInset * 2),
		Radius: legRadius / 2,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    legSupportLeftRightRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    legSupportLeftRightRotation,
		},
	)

	// BACK ===================================================================

	backParams := params.Group("Back")
	backHeight := backParams.Float64("Height")
	halfBackHeight := backHeight / 2

	backPeg := primitives.Cylinder{
		Sides:  8,
		Height: backHeight,
		Radius: legRadius,
	}.ToMesh()

	backSupportRotation := modeling.UnitQuaternionFromTheta(math.Pi/2, vector3.Forward[float64]())
	backSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairWidth - (legRadiusAndInset * 2),
		Radius: legRadius / 1.1,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    backSupportRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    backSupportRotation,
		},
	)

	backSupportPegHeight := backHeight * backParams.Float64("Backing Piece Height")
	backSupportPeg := primitives.Cylinder{
		Sides:  8,
		Height: backSupportPegHeight,
		Radius: legRadius / 1.4,
	}.ToMesh()

	backSupportPegs := repeat.LineExlusive(
		backSupportPeg,
		vector3.New(halfWidth-legRadiusAndInset, 0, halfLength-legRadiusAndInset),
		vector3.New(-halfWidth+legRadiusAndInset, 0, halfLength-legRadiusAndInset),
		backParams.Int("Backing Piece Pegs"),
	)

	return primitives.Cube{
		Height: params.Float64("Thickness"),
		Width:  chairWidth,
		Depth:  params.Float64("Length"),
		UVs:    primitives.DefaultCubeUVs(),
	}.
		UnweldedQuads().
		Translate(vector3.New(0, chairHeight, 0)).
		// LEGS ===============================================================
		Append(leg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, halfHeight, -halfLength+legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, halfHeight, halfLength-legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, halfHeight, -halfLength+legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, halfHeight, halfLength-legRadiusAndInset),
		)).

		// LEG SUPPORT ========================================================
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.85, halfLength-legRadiusAndInset),
		)).
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.6, -halfLength+legRadiusAndInset),
		)).
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.3, -halfLength+legRadiusAndInset),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight*0.45, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight*0.7, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight*0.45, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight*0.7, 0),
		)).

		// BACK ===============================================================
		Append(backPeg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backPeg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0, chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0, chairHeight+halfBackHeight+backSupportPegHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupportPegs.Translate(
			vector3.New(0, chairHeight+halfBackHeight+(backSupportPegHeight/2), 0),
		)).
		Append(backSupport.Translate(
			vector3.New(0, chairHeight+(halfBackHeight*0.8), halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0, chairHeight+(halfBackHeight*0.55), halfLength-legRadiusAndInset),
		))

}

func DiscoScene(c *generator.Context) (generator.Artifact, error) {
	ballParams := c.Parameters.Group("Disco Ball")
	discoballRadius := ballParams.Float64("radius")

	tableParams := c.Parameters.Group("Table")

	discoball := primitives.
		UVSphereUnwelded(
			discoballRadius,
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
		Append(primitives.UVSphere(
			discoballRadius+(ballParams.Float64("panel offset")/2),
			ballParams.Int("rows"),
			ballParams.Int("column"),
		)).
		Translate(discoBallHeight)

	discoballAttachment := // Base connecting ball to the rod
		primitives.
			Cylinder{
			Sides:  15,
			Height: 0.1,
			Radius: 0.2,
			UVs: &primitives.CylinderUVs{
				Top: &primitives.CircleUVs{
					Center: vector2.New(0.5, 0.5),
					Radius: 0.5,
				},
				Bottom: &primitives.CircleUVs{
					Center: vector2.New(0.5, 0.5),
					Radius: 0.5,
				},
				Side: &primitives.StripUVs{
					Start: vector2.New(0.5, 0.),
					End:   vector2.New(0.5, 1.),
					Width: 0.5,
				},
			},
		}.ToMesh().
			Translate(vector3.New(0, discoballRadius+ballParams.Float64("panel offset"), 0)).

			// Rod that the ball is hanging from
			Append(primitives.
				Cylinder{
				Sides:  4,
				Height: 3,
				Radius: 0.025,
				UVs: &primitives.CylinderUVs{
					Top: &primitives.CircleUVs{
						Center: vector2.New(0.5, 0.5),
						Radius: 0.5,
					},
					Bottom: &primitives.CircleUVs{
						Center: vector2.New(0.5, 0.5),
						Radius: 0.5,
					},
					Side: &primitives.StripUVs{
						Start: vector2.New(0.5, 0.),
						End:   vector2.New(0.5, 1.),
						Width: 0.5,
					},
				},
			}.ToMesh().
				Translate(vector3.New(0., discoballRadius+ballParams.Float64("panel offset")+1.5, 0.)),
			).
			Translate(discoBallHeight)

	chairParams := c.Parameters.Group("Chair")
	chairs := repeat.Circle(
		Chair(*chairParams),
		c.Parameters.Int("People"),
		tableParams.Float64("Radius")+chairParams.Float64("Table Spacing"),
	)

	cushionParams := chairParams.Group("Cushion")
	cushionThickness := cushionParams.Float64("Thickness")
	cushionColor := cushionParams.Color("Color")
	cushionInset := cushionParams.Float64("Inset")

	cushion := primitives.Cube{
		Height: cushionThickness,
		Width:  chairParams.Float64("Width") - cushionInset,
		Depth:  chairParams.Float64("Length") - cushionInset,
		UVs:    primitives.DefaultCubeUVs(),
	}.UnweldedQuads().Translate(vector3.New(0., chairParams.Float64("Height")+(cushionThickness/2), 0.))

	chairCushions := repeat.Circle(
		cushion,
		c.Parameters.Int("People"),
		tableParams.Float64("Radius")+chairParams.Float64("Table Spacing"),
	)

	plateParams := c.Parameters.Group("Plate")
	plates := repeat.Circle(
		Plate(*plateParams),
		c.Parameters.Int("People"),
		tableParams.Float64("Radius")-plateParams.Float64("Table Inset")-plateParams.Float64("Radius"),
	)

	return generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{
					Name: "Disco Ball",
					Mesh: discoball,
					Material: &gltf.PolyformMaterial{
						Name: "Disco Ball",
						Extras: map[string]any{
							"threejs-material": "phong",
						},
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							MetallicFactor:  1,
							RoughnessFactor: 0,
							BaseColorFactor: ballParams.Color("Color"),
							MetallicRoughnessTexture: &gltf.PolyformTexture{
								URI: "metal.png",
							},
						},
					},
				},
				{
					Name: "Disco Ball Attachment",
					Mesh: discoballAttachment,
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							MetallicFactor:  1,
							RoughnessFactor: 0,
							BaseColorFactor: color.Black,
							MetallicRoughnessTexture: &gltf.PolyformTexture{
								URI: "metal.png",
							},
						},
					},
				},
				{
					Name: "Chair Frames",
					Mesh: chairs,
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							MetallicFactor:  0,
							RoughnessFactor: 1,
							BaseColorFactor: chairParams.Color("Color"),
						},
					},
				},
				{
					Name: "Chair Cushions",
					Mesh: chairCushions,
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							MetallicFactor:  0,
							RoughnessFactor: 1,
							BaseColorFactor: cushionColor,
						},
					},
				},
				{
					Name: "Plates",
					Mesh: plates.Translate(vector3.New(
						0.,
						tableParams.Float64("Height")+
							(tableParams.Float64("Thickness")/2)+
							(plateParams.Float64("Thickness")/2),
						0.,
					)),
					Material: &gltf.PolyformMaterial{
						Name: "Plate Mat",
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							BaseColorFactor: plateParams.Color("Color"),
						},
					},
				},
				{
					Name: "Table",
					Mesh: primitives.Cylinder{
						Sides:  tableParams.Int("Resolution"),
						Height: tableParams.Float64("Thickness"),
						Radius: tableParams.Float64("Radius"),
						UVs: &primitives.CylinderUVs{
							Top: &primitives.CircleUVs{
								Center: vector2.New(0.5, 0.5),
								Radius: 0.5,
							},
							Bottom: &primitives.CircleUVs{
								Center: vector2.New(0.5, 0.5),
								Radius: 0.5,
							},
							Side: &primitives.StripUVs{
								Start: vector2.New(0.5, 0.),
								End:   vector2.New(0.5, 1.),
								Width: 0.5,
							},
						},
					}.
						ToMesh().
						Translate(vector3.New(0., tableParams.Float64("Height"), 0.)).
						Append(primitives.Cylinder{
							Sides:  tableParams.Int("Resolution"),
							Height: tableParams.Float64("Height"),
							Radius: tableParams.Float64("Radius") / 8,
							UVs: &primitives.CylinderUVs{
								Top: &primitives.CircleUVs{
									Center: vector2.New(0.5, 0.5),
									Radius: 0.5,
								},
								Bottom: &primitives.CircleUVs{
									Center: vector2.New(0.5, 0.5),
									Radius: 0.5,
								},
								Side: &primitives.StripUVs{
									Start: vector2.New(0.5, 0.),
									End:   vector2.New(0.5, 1.),
									Width: 0.5,
								},
							},
						}.ToMesh().Translate(vector3.New(0., tableParams.Float64("Height")/2, 0.))),

					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							BaseColorFactor: tableParams.Color("Color"),
							RoughnessFactor: 1,
							MetallicRoughnessTexture: &gltf.PolyformTexture{
								URI: "rough.png",
							},
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
					&generator.IntParameter{
						Name:         "People",
						DefaultValue: 5,
					},
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
								DefaultValue: 7.5,
							},

							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0xea,
									G: 0xff,
									B: 0xe0,
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
							&generator.FloatParameter{
								Name:         "Height",
								DefaultValue: 3,
							},
							&generator.IntParameter{
								Name:         "Resolution",
								DefaultValue: 30,
							},
							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0xd1,
									G: 0xd1,
									B: 0xd1,
									A: 255,
								},
							},
						},
					},
					&generator.GroupParameter{
						Name: "Plate",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Radius",
								DefaultValue: .5,
							},
							&generator.FloatParameter{
								Name:         "Thickness",
								DefaultValue: 0.03,
							},
							&generator.FloatParameter{
								Name:         "Table Inset",
								DefaultValue: .45,
							},
							&generator.IntParameter{
								Name:         "Resolution",
								DefaultValue: 30,
							},
							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0x8a,
									G: 0x8a,
									B: 0x8a,
									A: 255,
								},
							},
						},
					},
					&generator.GroupParameter{
						Name: "Chair",
						Parameters: []generator.Parameter{

							&generator.FloatParameter{
								Name:         "Thickness",
								DefaultValue: 0.1,
							},

							&generator.FloatParameter{
								Name:         "Height",
								DefaultValue: 1.6,
							},

							&generator.FloatParameter{
								Name:         "Table Spacing",
								DefaultValue: 0.5,
							},

							&generator.FloatParameter{
								Name:         "Width",
								DefaultValue: 1.5,
							},

							&generator.FloatParameter{
								Name:         "Length",
								DefaultValue: 1.2,
							},

							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0x21,
									G: 0x21,
									B: 0x21,
									A: 255,
								},
							},

							&generator.GroupParameter{
								Name: "Leg",
								Parameters: []generator.Parameter{
									&generator.FloatParameter{
										Name:         "Radius",
										DefaultValue: 0.05,
									},
									&generator.FloatParameter{
										Name:         "Inset",
										DefaultValue: 0.05,
									},
								},
							},

							&generator.GroupParameter{
								Name: "Back",
								Parameters: []generator.Parameter{
									&generator.FloatParameter{
										Name:         "Height",
										DefaultValue: 2,
									},
									&generator.FloatParameter{
										Name:         "Backing Piece Height",
										DefaultValue: 0.4,
									},

									&generator.IntParameter{
										Name:         "Backing Piece Pegs",
										DefaultValue: 5,
									},
								},
							},

							&generator.GroupParameter{
								Name: "Cushion",
								Parameters: []generator.Parameter{
									&generator.FloatParameter{
										Name:         "Thickness",
										DefaultValue: .2,
									},
									&generator.FloatParameter{
										Name:         "Inset",
										DefaultValue: .05,
									},
									&generator.ColorParameter{
										Name:         "Color",
										DefaultValue: coloring.WebColor{R: 255, G: 255, B: 255, A: 255},
									},
								},
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"disco.glb": DiscoScene,
				"metal.png": func(c *generator.Context) (generator.Artifact, error) {
					return generator.ImageArtifact{
						Image: texture(1, 0),
					}, nil
				},
				"rough.png": func(c *generator.Context) (generator.Artifact, error) {
					return generator.ImageArtifact{
						Image: texture(0, 1),
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
