package main

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/quatn"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
	"github.com/EliCDavis/vector/vector3"
)

func main() {
	scale := &parameter.Float64{Name: "Scale", DefaultValue: 1}

	fileNode := &parameter.File{
		Name: "Splat File",
		CLI: &parameter.CliConfig[string]{
			FlagName: "splat",
			Usage:    "Path to the guassian splat to load (PLY file)",
			// Default:  "./point_cloud/iteration_30000/point_cloud.ply",
		},
	}

	pointcloud := &gausops.LoaderNode{
		Data: gausops.LoaderNodeData{
			Data: fileNode,
		},
	}

	// filteredCloud := &gausops.FilterNode{
	// 	Data: gausops.FilterNodeData{
	// 		Splat:      pointcloud.Out(),
	// 		MinOpacity: &parameter.Float64{Name: "Minimum Opacity"},
	// 		MinVolume:  &parameter.Float64{Name: "Minimum Scale"},
	// 		MaxVolume: &parameter.Float64{
	// 			Name:         "Maximum Scale",
	// 			DefaultValue: 1000.,
	// 		},
	// 	},
	// }

	rotateAmount := &quatn.FromTheta{
		Data: quatn.FromThetaData{
			Theta: &parameter.Float64{
				Name:         "Rotation",
				Description:  "How much to rotate the pointcloud by",
				DefaultValue: math.Pi,
			},
			Direction: &vecn3.New{
				Data: vecn3.NewData[float64]{
					X: &parameter.Float64{
						Name: "Rotation Direction X",
					},
					Y: &parameter.Float64{
						Name: "Rotation Direction Y",
					},
					Z: &parameter.Float64{
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
					Mesh:   pointcloud.Out(),
					Amount: rotateAmount,
				},
			},
			Amount: rotateAmount,
		},
	}

	croppedCloud := meshops.CropAttribute3DNode{
		Data: meshops.CropAttribute3DNodeData{
			Mesh: rotatedCloud,
			AABB: &parameter.AABB{
				Name: "Keep Bounds",
				DefaultValue: geometry.NewAABBFromPoints(
					vector3.New(-10., -10., -10.),
					vector3.New(10., 10., 10.),
				),
			},
		},
	}

	baloonNode := &gausops.ScaleWithinRegionNode{
		Data: gausops.ScaleWithinRegionNodeData{
			Mesh:   croppedCloud.Out(),
			Scale:  &parameter.Float64{Name: "Baloon Strength", DefaultValue: .7},
			Radius: &parameter.Float64{Name: "Baloon Radius", DefaultValue: .7},
			Position: &parameter.Vector3{
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
			LUT: &parameter.Image{
				Name: "LUT",
				CLI: &parameter.CliConfig[string]{
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
		Producers: map[string]nodes.NodeOutput[artifact.Artifact]{
			"mesh.ply": artifact.NewSplatPlyNode(pointcloud.Out()),
			"info.txt": artifact.NewTextNode(&InfoNode{
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
