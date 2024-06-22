package main

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/quatn"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
	"github.com/EliCDavis/vector/vector3"
)

type BaloonNode = nodes.StructNode[modeling.Mesh, BaloonNodeData]

type BaloonNodeData struct {
	Mesh     nodes.NodeOutput[modeling.Mesh]
	Strength nodes.NodeOutput[float64]
	Radius   nodes.NodeOutput[float64]
	Position nodes.NodeOutput[vector3.Float64]
}

func (bn BaloonNodeData) Process() (modeling.Mesh, error) {
	return bn.Mesh.
		Value().
		Transform(meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				posData := m.Float3Attribute(modeling.PositionAttribute)
				scaleData := m.Float3Attribute(modeling.ScaleAttribute)
				count := posData.Len()

				newPos := make([]vector3.Float64, count)
				newScale := make([]vector3.Float64, count)

				baloonPos := bn.Position.Value()
				baloonRadius := bn.Radius.Value()
				baloonStrength := bn.Strength.Value()

				for i := 0; i < count; i++ {
					curPos := posData.At(i)
					curScale := scaleData.At(i)
					dir := curPos.Sub(baloonPos)
					len := dir.Length()

					if len <= baloonRadius {
						newPos[i] = baloonPos.Add(dir.Scale(baloonStrength))
						newScale[i] = curScale.Exp().Scale(baloonStrength).Log()
					} else {
						newPos[i] = curPos
						newScale[i] = curScale
					}
				}

				return m.
					SetFloat3Attribute(modeling.PositionAttribute, newPos).
					SetFloat3Attribute(modeling.ScaleAttribute, newScale), nil
			},
		}), nil
}

func main() {
	scale := &generator.ParameterNode[float64]{
		Name:         "Scale",
		DefaultValue: 1.,
	}

	pointcloud := &gausops.LoaderNode{
		Data: gausops.LoaderNodeData{
			Data: &generator.FileParameterNode{
				Name: "Splat File",
				CLI: &generator.CliParameterNodeConfig[string]{
					FlagName: "splat",
					Usage:    "Path to the guassian splat to load (PLY file)",
					// Default:  "./point_cloud/iteration_30000/point_cloud.ply",
				},
			},
		},
	}

	filteredCloud := &gausops.FilterNode{
		Data: gausops.FilterNodeData{
			Splat: pointcloud.Out(),
			MinOpacity: &generator.ParameterNode[float64]{
				Name:         "Minimum Opacity",
				DefaultValue: 0.,
			},
			MinVolume: &generator.ParameterNode[float64]{
				Name:         "Minimum Scale",
				DefaultValue: 0.,
			},
			MaxVolume: &generator.ParameterNode[float64]{
				Name:         "Maximum Scale",
				DefaultValue: 1000.,
			},
		},
	}

	rotateAmount := &quatn.FromTheta{
		Data: quatn.FromThetaData{
			Theta: &generator.ParameterNode[float64]{
				Name:         "Rotation",
				Description:  "How much to rotate the pointcloud by",
				DefaultValue: math.Pi,
			},
			Direction: &vecn3.New{
				Data: vecn3.NewData[float64]{
					X: &generator.ParameterNode[float64]{
						Name:         "Rotation Direction X",
						DefaultValue: 0,
					},
					Y: &generator.ParameterNode[float64]{
						Name:         "Rotation Direction Y",
						DefaultValue: 0,
					},
					Z: &generator.ParameterNode[float64]{
						Name:         "Rotation Direction Z",
						DefaultValue: 1,
					},
				},
			},
		},
	}

	rotatedCloud := &gausops.RotateAttributeNode{
		Data: gausops.RotateAttributeNodeData{
			Mesh: &meshops.RotateAttribute3DNode{
				Data: meshops.RotateAttribute3DNodeData{
					Mesh:   filteredCloud.Out(),
					Amount: rotateAmount,
				},
			},
			Amount: rotateAmount,
		},
	}

	croppedCloud := meshops.CropAttribute3DNode{
		Data: meshops.CropAttribute3DNodeData{
			Mesh: rotatedCloud,
			AABB: &generator.ParameterNode[geometry.AABB]{
				Name: "Keep Bounds",
				DefaultValue: geometry.NewAABBFromPoints(
					vector3.New(-10., -10., -10.),
					vector3.New(10., 10., 10.),
				),
			},
		},
	}

	baloonNode := BaloonNode{
		Data: BaloonNodeData{
			Mesh:     croppedCloud.Out(),
			Strength: &generator.ParameterNode[float64]{Name: "Baloon Strength", DefaultValue: .7},
			Radius:   &generator.ParameterNode[float64]{Name: "Baloon Radius", DefaultValue: .7},
			Position: &generator.ParameterNode[vector3.Float64]{
				Name:         "Baloon Position",
				DefaultValue: vector3.New(-0.344, 0.402, 5.363),
			},
		},
	}

	x := &vecn3.New{
		Data: vecn3.NewData[float64]{
			X: scale,
			Y: scale,
			Z: scale,
		},
	}

	scaleNode := meshops.ScaleAttribute3DNode{
		Data: meshops.ScaleAttribute3DNodeData{
			Mesh: &gausops.ScaleNode{
				Data: gausops.ScaleNodeData{
					Mesh:   baloonNode.Out(),
					Amount: x.Out(),
				},
			},
			Amount: x.Out(),
		},
	}

	colorGraded := gausops.ColorGradingLutNode{
		Data: gausops.ColorGradingLutNodeData{
			Mesh: scaleNode.Out(),
			LUT: &generator.ImageParameterNode{
				Name: "LUT",
				CLI: &generator.CliParameterNodeConfig[string]{
					FlagName: "lut",
					Usage:    "Path to the color grading LUT",
				},
			},
		},
	}

	app := generator.App{
		Name:        "Edit Gaussian Splats",
		Version:     "1.0.0",
		Description: "Crop and Scale portions of Gaussian Splat data",
		Authors:     []generator.Author{{Name: "Eli C Davis"}},
		WebScene: &room.WebScene{
			Background: coloring.Black(),
			Fog: room.WebSceneFog{
				Color: coloring.Black(),
				Near:  5,
				Far:   25,
			},
			Lighting:  coloring.White(),
			Ground:    coloring.White(),
			AntiAlias: false,
			XrEnabled: true,
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.splat": generator.NewSplatArtifactNode(colorGraded.Out()),
			"info.txt": generator.NewTextArtifactNode(&InfoNode{
				Data: InfoNodeData{
					Original: pointcloud.Out(),
					Final:    colorGraded.Out(),
				},
			}),
		},
		// AvailableNodes: generator.Nodes(),
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
