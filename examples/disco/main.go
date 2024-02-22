package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
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

type DiscoSceneNode struct {
	nodes.StructData[generator.Artifact]

	People nodes.NodeOutput[int]

	DiscoBall nodes.NodeOutput[[]gltf.PolyformModel]

	Table          nodes.NodeOutput[gltf.PolyformModel]
	TableHeight    nodes.NodeOutput[float64]
	TableRadius    nodes.NodeOutput[float64]
	TableThickness nodes.NodeOutput[float64]

	Chairs     nodes.NodeOutput[modeling.Mesh]
	ChairColor nodes.NodeOutput[coloring.WebColor]

	Cushion      nodes.NodeOutput[modeling.Mesh]
	CushionColor nodes.NodeOutput[coloring.WebColor]

	Plates         nodes.NodeOutput[modeling.Mesh]
	PlateColor     nodes.NodeOutput[coloring.WebColor]
	PlateThickness nodes.NodeOutput[float64]
}

func (cn *DiscoSceneNode) Out() nodes.NodeOutput[generator.Artifact] {
	return &nodes.StructNodeOutput[generator.Artifact]{Definition: cn}
}

func (dsn DiscoSceneNode) Process() (generator.Artifact, error) {
	chairs := dsn.Chairs.Data()

	models := []gltf.PolyformModel{
		dsn.Table.Data(),
		{
			Name: "Chair Frames",
			Mesh: chairs,
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					MetallicFactor:  0,
					RoughnessFactor: 1,
					BaseColorFactor: dsn.ChairColor.Data(),
				},
			},
		},
		{
			Name: "Chair Cushions",
			Mesh: dsn.Cushion.Data(),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					MetallicFactor:  0,
					RoughnessFactor: 1,
					BaseColorFactor: dsn.CushionColor.Data(),
				},
			},
		},
		{
			Name: "Plates",
			Mesh: dsn.Plates.Data().Translate(vector3.New(
				0.,
				dsn.TableHeight.Data()+
					(dsn.TableThickness.Data()/2)+
					(dsn.PlateThickness.Data()/2),
				0.,
			)),
			Material: &gltf.PolyformMaterial{
				Name: "Plate Mat",
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: dsn.PlateColor.Data(),
				},
				// PbrSpecularGlossiness: &gltf.PolyformPbrSpecularGlossiness{
				// 	GlossinessFactor: 1,
				// 	DiffuseFactor:    color.RGBA{R: 0, G: 0, B: 0, A: 125},
				// 	SpecularFactor:   color.RGBA{R: 0, G: 0, B: 0, A: 125},
				// },
				Extensions: []gltf.MaterialExtension{
					&gltf.PolyformTransmission{
						TransmissionFactor: .8,
					},
				},
			},
		},
	}

	models = append(models, dsn.DiscoBall.Data()...)

	return generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: models,
		},
	}, nil
}

func main() {

	personCount := &generator.ParameterNode[int]{
		Name:         "People",
		DefaultValue: 6,
	}

	tableRadius := &generator.ParameterNode[float64]{
		Name:         "Table/Radius",
		DefaultValue: 3,
	}

	chairHeight := &generator.ParameterNode[float64]{
		Name:         "Chair/Height",
		DefaultValue: 1,
	}

	chairTableSpacing := &generator.ParameterNode[float64]{
		Name:         "Table/Spacing From Table",
		DefaultValue: .1,
	}

	chairWidth := &generator.ParameterNode[float64]{
		Name:         "Chair/Width",
		DefaultValue: 1.3,
	}

	chairLength := &generator.ParameterNode[float64]{
		Name:         "Chair/Length",
		DefaultValue: 1,
	}

	cushionInset := &generator.ParameterNode[float64]{
		Name:         "Cushion/Inset",
		DefaultValue: .05,
	}

	tableHeight := &generator.ParameterNode[float64]{
		Name:         "Table/Height",
		DefaultValue: 1.75,
	}

	chairPosition := (&nodes.Sum[float64]{
		Values: []nodes.NodeOutput[float64]{
			tableRadius,
			chairTableSpacing,
		},
	}).Out()

	cushionThickness := &generator.ParameterNode[float64]{
		Name:         "Cushion/Thickness",
		DefaultValue: .1,
	}

	cushion := meshops.TranslateAttribute3DNode{
		Mesh: (&repeat.CircleNode{
			Mesh: (&CushionNode{
				Thickness: cushionThickness,
				Width: (&nodes.Difference[float64]{
					A: chairWidth,
					B: cushionInset,
				}).Out(),
				Length: (&nodes.Difference[float64]{
					A: chairLength,
					B: cushionInset,
				}).Out(),
			}).Out(),
			Times:  personCount,
			Radius: chairPosition,
		}).Out(),
		Amount: (&vecn3.New[float64]{
			Y: (&nodes.Sum[float64]{
				Values: []nodes.NodeOutput[float64]{
					chairHeight,
					(&nodes.Divide[float64]{
						Dividend: cushionThickness,
						Divisor:  nodes.Value[float64](2),
					}).Out(),
				},
			}).Out(),
		}).Out(),
	}

	chairs := repeat.CircleNode{
		Mesh: (&ChairNode{
			Height: chairHeight,
			Width:  chairWidth,
			Length: chairLength,
			Thickness: &generator.ParameterNode[float64]{
				Name:         "Chair/Thickness",
				DefaultValue: .1,
			},
			BackHeight: &generator.ParameterNode[float64]{
				Name:         "Chair/BackHeight",
				DefaultValue: 2,
			},
			BackingPieceHeight: &generator.ParameterNode[float64]{
				Name:         "Chair/BackingPiece Hieght",
				DefaultValue: .4,
			},
			BackingPieceHeightPegs: &generator.ParameterNode[int]{
				Name:         "Chair/Backing Piece Height Pegs",
				DefaultValue: 4,
			},
			LegRadius: &generator.ParameterNode[float64]{
				Name:         "Chair/Leg Radius",
				DefaultValue: .05,
			},
			LegInset: &generator.ParameterNode[float64]{
				Name:         "Chair/Leg Inset",
				DefaultValue: .1,
			},
		}).Out(),
		Radius: chairPosition,
		Times:  personCount,
	}

	plateRadius := &generator.ParameterNode[float64]{
		Name:         "Plate/Radius",
		DefaultValue: .3,
	}

	plateThickenss := &generator.ParameterNode[float64]{
		Name:         "Plate/Thickness",
		DefaultValue: .01,
	}

	plates := repeat.CircleNode{
		Times: personCount,
		Mesh: (&PlateNode{
			Thickness: plateThickenss,
			Radius:    plateRadius,
			Resolution: &generator.ParameterNode[int]{
				Name:         "Plate/Resolution",
				DefaultValue: 20,
			},
		}).Out(),
		Radius: (&nodes.Difference[float64]{
			A: tableRadius,
			B: (&nodes.Sum[float64]{
				Values: []nodes.NodeOutput[float64]{
					&generator.ParameterNode[float64]{
						Name:         "Plate/Table Inset",
						DefaultValue: 0.1,
					},
					plateRadius,
				},
			}).Out(),
		}).Out(),
	}

	discoScene := DiscoSceneNode{
		Plates:         plates.Out(),
		PlateThickness: plateThickenss,
		PlateColor: &generator.ParameterNode[coloring.WebColor]{
			Name:         "Plate Color",
			DefaultValue: coloring.WebColor{R: 225, G: 225, B: 225, A: 255},
		},

		TableRadius: tableRadius,
		TableHeight: tableHeight,
		TableThickness: &generator.ParameterNode[float64]{
			Name:         "Table/Thickness",
			DefaultValue: .1,
		},

		Chairs: chairs.Out(),
		ChairColor: &generator.ParameterNode[coloring.WebColor]{
			Name:         "Chair Color",
			DefaultValue: coloring.WebColor{R: 0x21, G: 0x21, B: 0x21, A: 255},
		},

		Cushion: cushion.Out(),
		CushionColor: &generator.ParameterNode[coloring.WebColor]{
			Name:         "Cushion Color",
			DefaultValue: coloring.WebColor{R: 225, G: 225, B: 225, A: 255},
		},
		People: personCount,
		DiscoBall: (&DiscoBallNode{
			Radius: &generator.ParameterNode[float64]{
				Name:         "Ball/Radius",
				DefaultValue: 1,
			},
			PanelOffset: &generator.ParameterNode[float64]{
				Name:         "Ball/Offset",
				DefaultValue: .1,
			},
			Height: &generator.ParameterNode[float64]{
				Name:         "Ball/Height",
				DefaultValue: 6,
			},
			Rows: &generator.ParameterNode[int]{
				Name:         "Ball/Rows",
				DefaultValue: 20,
			},
			Columns: &generator.ParameterNode[int]{
				Name:         "Ball/Columns",
				DefaultValue: 24,
			},
			Color: &generator.ParameterNode[coloring.WebColor]{
				Name:         "Ball/Color",
				DefaultValue: coloring.WebColor{R: 127, G: 127, B: 127, A: 255},
			},
		}).Out(),
		Table: (&TableNode{
			Radius: tableRadius,
			Thickness: &generator.ParameterNode[float64]{
				Name:         "Table/Thickness",
				DefaultValue: .1,
			},
			Height: tableHeight,
			Resolution: &generator.ParameterNode[int]{
				Name:         "Table/Resolution",
				DefaultValue: 20,
			},
			Color: &generator.ParameterNode[coloring.WebColor]{
				Name:         "Table/Color",
				DefaultValue: coloring.WebColor{R: 0xea, G: 0xba, B: 0x76, A: 255},
			},
		}).Out(),
	}

	app := generator.App{
		Name:        "Woodland Disco Romance",
		Version:     "1.0.0",
		Description: "Applying color pallettes to a sample room",
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"disco.glb": discoScene.Out(),
			"metal.png": generator.ImageArtifactNode(nodes.Value[image.Image](texture(1, 0))),
			"rough.png": generator.ImageArtifactNode(nodes.Value[image.Image](texture(0, 1))),
		},
	}
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
