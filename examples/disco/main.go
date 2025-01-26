package main

import (
	"image"
	"image/color"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/math/vector"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
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

type DiscoSceneNode = nodes.Struct[artifact.Artifact, DiscoSceneNodeData]

type DiscoSceneNodeData struct {
	People nodes.NodeOutput[int]

	DiscoBall nodes.NodeOutput[[]gltf.PolyformModel]

	Table          nodes.NodeOutput[gltf.PolyformModel]
	TableHeight    nodes.NodeOutput[float64]
	TableRadius    nodes.NodeOutput[float64]
	TableThickness nodes.NodeOutput[float64]

	Chair          nodes.NodeOutput[modeling.Mesh]
	ChairPositions nodes.NodeOutput[[]trs.TRS]
	ChairColor     nodes.NodeOutput[coloring.WebColor]

	Cushion          nodes.NodeOutput[modeling.Mesh]
	CushionPositions nodes.NodeOutput[[]trs.TRS]
	CushionColor     nodes.NodeOutput[coloring.WebColor]

	Plate          nodes.NodeOutput[modeling.Mesh]
	PlatePositions nodes.NodeOutput[[]trs.TRS]
	PlateColor     nodes.NodeOutput[coloring.WebColor]
	PlateThickness nodes.NodeOutput[float64]
}

func (dsn DiscoSceneNodeData) Process() (artifact.Artifact, error) {
	chairs := dsn.Chair.Value()
	cushions := dsn.Cushion.Value()
	plates := dsn.Plate.Value().Translate(vector3.New(
		0.,
		dsn.TableHeight.Value()+
			(dsn.TableThickness.Value()/2)+
			(dsn.PlateThickness.Value()/2),
		0.,
	))

	noMetal := 0.
	models := []gltf.PolyformModel{
		dsn.Table.Value(),
		{
			Name:         "Chair Frames",
			Mesh:         &chairs,
			GpuInstances: dsn.ChairPositions.Value(),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					MetallicFactor:  &noMetal,
					BaseColorFactor: dsn.ChairColor.Value(),
				},
			},
		},
		{
			Name:         "Chair Cushions",
			Mesh:         &cushions,
			GpuInstances: dsn.CushionPositions.Value(),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					MetallicFactor:  &noMetal,
					BaseColorFactor: dsn.CushionColor.Value(),
				},
			},
		},
		{
			Name:         "Plates",
			Mesh:         &plates,
			GpuInstances: dsn.PlatePositions.Value(),
			Material: &gltf.PolyformMaterial{
				Name: "Plate Mat",
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: dsn.PlateColor.Value(),
				},
				// PbrSpecularGlossiness: &gltf.PolyformPbrSpecularGlossiness{
				// 	GlossinessFactor: 1,
				// 	DiffuseFactor:    color.RGBA{R: 0, G: 0, B: 0, A: 125},
				// 	SpecularFactor:   color.RGBA{R: 0, G: 0, B: 0, A: 125},
				// },
				Extensions: []gltf.MaterialExtension{
					&gltf.PolyformTransmission{
						Factor: .8,
					},
				},
			},
		},
	}

	models = append(models, dsn.DiscoBall.Value()...)

	return gltf.Artifact{
		Scene: gltf.PolyformScene{
			Models: models,
		},
	}, nil
}

func main() {

	personCount := &parameter.Int{
		Name:         "People",
		DefaultValue: 6,
	}

	tableRadius := &parameter.Float64{
		Name:         "Table/Radius",
		DefaultValue: 3,
	}

	chairHeight := &parameter.Float64{
		Name:         "Chair/Height",
		DefaultValue: 1,
	}

	chairTableSpacing := &parameter.Float64{
		Name:         "Table/Spacing From Table",
		DefaultValue: .1,
	}

	chairWidth := &parameter.Float64{
		Name:         "Chair/Width",
		DefaultValue: 1.3,
	}

	chairLength := &parameter.Float64{
		Name:         "Chair/Length",
		DefaultValue: 1,
	}

	cushionInset := &parameter.Float64{
		Name:         "Cushion/Inset",
		DefaultValue: .05,
	}

	tableHeight := &parameter.Float64{
		Name:         "Table/Height",
		DefaultValue: 1.75,
	}

	chairPosition := &math.SumNode{
		Data: math.SumData[float64]{
			Values: []nodes.NodeOutput[float64]{
				tableRadius,
				chairTableSpacing,
			},
		},
	}

	cushionThickness := &parameter.Float64{
		Name:         "Cushion/Thickness",
		DefaultValue: .1,
	}

	cusionPositions := &repeat.CircleNode{
		Data: repeat.CircleNodeData{
			Times:  personCount,
			Radius: chairPosition,
		},
	}

	cushion := meshops.TranslateAttribute3DNode{
		Data: meshops.TranslateAttribute3DNodeData{
			Mesh: &CushionNode{
				Data: CushionNodeData{
					Thickness: cushionThickness,
					Width: &math.DifferenceNode{
						Data: math.DifferenceData[float64]{
							A: chairWidth,
							B: cushionInset,
						},
					},
					Length: &math.DifferenceNode{
						Data: math.DifferenceData[float64]{
							A: chairLength,
							B: cushionInset,
						},
					},
				},
			},
			Amount: &vector.New{
				Data: vector.NewData[float64]{
					Y: &math.SumNode{
						Data: math.SumData[float64]{
							Values: []nodes.NodeOutput[float64]{
								chairHeight,
								&math.DivideNode{
									Data: math.DivideData[float64]{
										Dividend: cushionThickness,
										Divisor:  nodes.Value[float64](2),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	chairs := &ChairNode{
		Data: ChairNodeData{
			Height: chairHeight,
			Width:  chairWidth,
			Length: chairLength,
			Thickness: &parameter.Float64{
				Name:         "Chair/Thickness",
				DefaultValue: .1,
			},
			BackHeight: &parameter.Float64{
				Name:         "Chair/BackHeight",
				DefaultValue: 2,
			},
			BackingPieceHeight: &parameter.Float64{
				Name:         "Chair/BackingPiece Hieght",
				DefaultValue: .4,
			},
			BackingPieceHeightPegs: &parameter.Int{
				Name:         "Chair/Backing Piece Height Pegs",
				DefaultValue: 4,
			},
			LegRadius: &parameter.Float64{
				Name:         "Chair/Leg Radius",
				DefaultValue: .05,
			},
			LegInset: &parameter.Float64{
				Name:         "Chair/Leg Inset",
				DefaultValue: .1,
			},
		},
	}

	plateRadius := &parameter.Float64{
		Name:         "Plate/Radius",
		DefaultValue: .3,
	}

	plateThickenss := &parameter.Float64{
		Name:         "Plate/Thickness",
		DefaultValue: .01,
	}

	plate := &PlateNode{
		Data: PlateNodeData{
			Thickness: plateThickenss,
			Radius:    plateRadius,
			Resolution: &parameter.Int{
				Name:         "Plate/Resolution",
				DefaultValue: 20,
			},
		},
	}

	platePositions := &repeat.CircleNode{
		Data: repeat.CircleNodeData{
			Times: personCount,
			Radius: &math.DifferenceNode{
				Data: math.DifferenceData[float64]{
					A: tableRadius,
					B: &math.SumNode{
						Data: math.SumData[float64]{
							Values: []nodes.NodeOutput[float64]{
								&parameter.Float64{
									Name:         "Plate/Table Inset",
									DefaultValue: 0.1,
								},
								plateRadius,
							},
						},
					},
				},
			},
		},
	}

	discoScene := &DiscoSceneNode{
		Data: DiscoSceneNodeData{
			Plate:          plate,
			PlatePositions: platePositions,
			PlateThickness: plateThickenss,
			PlateColor: &parameter.Color{
				Name:         "Plate Color",
				DefaultValue: coloring.WebColor{R: 225, G: 225, B: 225, A: 255},
			},

			TableRadius: tableRadius,
			TableHeight: tableHeight,
			TableThickness: &parameter.Float64{
				Name:         "Table/Thickness",
				DefaultValue: .1,
			},

			Chair: chairs.Out(),
			ChairPositions: &repeat.CircleNode{
				Data: repeat.CircleNodeData{
					Radius: chairPosition,
					Times:  personCount,
				},
			},
			ChairColor: &parameter.Color{
				Name:         "Chair Color",
				DefaultValue: coloring.WebColor{R: 0x21, G: 0x21, B: 0x21, A: 255},
			},

			Cushion:          cushion.Out(),
			CushionPositions: cusionPositions,
			CushionColor: &parameter.Color{
				Name:         "Cushion Color",
				DefaultValue: coloring.WebColor{R: 225, G: 225, B: 225, A: 255},
			},
			People: personCount,
			DiscoBall: &DiscoBallNode{
				Data: DiscoBallNodeData{
					Radius: &parameter.Float64{
						Name:         "Ball/Radius",
						DefaultValue: 1,
					},
					PanelOffset: &parameter.Float64{
						Name:         "Ball/Offset",
						DefaultValue: .1,
					},
					Height: &parameter.Float64{
						Name:         "Ball/Height",
						DefaultValue: 6,
					},
					Rows: &parameter.Int{
						Name:         "Ball/Rows",
						DefaultValue: 20,
					},
					Columns: &parameter.Int{
						Name:         "Ball/Columns",
						DefaultValue: 24,
					},
					Color: &parameter.Color{
						Name:         "Ball/Color",
						DefaultValue: coloring.WebColor{R: 127, G: 127, B: 127, A: 255},
					},
				},
			},
			Table: &TableNode{
				Data: TableNodeData{
					Radius: tableRadius,
					Thickness: &parameter.Float64{
						Name:         "Table/Thickness",
						DefaultValue: .1,
					},
					Height: tableHeight,
					Resolution: &parameter.Int{
						Name:         "Table/Resolution",
						DefaultValue: 20,
					},
					Color: &parameter.Color{
						Name:         "Table/Color",
						DefaultValue: coloring.WebColor{R: 0xea, G: 0xba, B: 0x76, A: 255},
					},
				},
			},
		},
	}

	app := generator.App{
		Name:        "Woodland Disco Romance",
		Version:     "1.0.0",
		Description: "Applying color pallettes to a sample room",
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			"disco.glb": discoScene.Out(),
			"metal.png": basics.NewImageNode(nodes.Value(texture(1, 0))),
			"rough.png": basics.NewImageNode(nodes.Value(texture(0, 1))),
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
