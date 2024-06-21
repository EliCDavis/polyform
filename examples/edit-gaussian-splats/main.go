package main

import (
	"bufio"
	"bytes"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type PointcloudLoaderNode = nodes.StructNode[modeling.Mesh, PointcloudLoaderNodeData]

type PointcloudLoaderNodeData struct {
	Data nodes.NodeOutput[[]byte]
}

func (pn PointcloudLoaderNodeData) Process() (modeling.Mesh, error) {
	bufReader := bufio.NewReader(bytes.NewReader(pn.Data.Value()))

	header, err := ply.ReadHeader(bufReader)
	if err != nil {
		return modeling.EmptyMesh(modeling.PointTopology), err
	}

	reader := header.BuildReader(bufReader)
	plyMesh, err := reader.ReadMesh(ply.GuassianSplatVertexAttributes)
	if err != nil {
		return modeling.EmptyMesh(modeling.PointTopology), err
	}
	return *plyMesh, err
}

type SplatEditNode = nodes.StructNode[modeling.Mesh, SplatEditNodeData]

type SplatEditNodeData struct {
	SplatData  nodes.NodeOutput[modeling.Mesh]
	MinOpacity nodes.NodeOutput[float64]
	MinScale   nodes.NodeOutput[float64]
	MaxScale   nodes.NodeOutput[float64]
	Scale      nodes.NodeOutput[float64]
}

func (pn SplatEditNodeData) Process() (modeling.Mesh, error) {
	return pn.SplatData.Value().Transform(

		// Filter out points that don't meet the opacity or scale criteria
		meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				minOpacity := pn.MinOpacity.Value()
				minScale := pn.MinScale.Value()
				maxScale := math.Inf(0)
				if pn.MaxScale != nil {
					maxScale = pn.MaxScale.Value()
				}

				opacity := m.Float1Attribute(modeling.OpacityAttribute)
				scale := m.Float3Attribute(modeling.ScaleAttribute)

				indicesKept := make([]int, 0)
				for i := 0; i < opacity.Len(); i++ {
					if opacity.At(i) < minOpacity {
						continue
					}

					length := scale.At(i).Exp().LengthSquared()
					if length < minScale || length > maxScale {
						continue
					}
					indicesKept = append(indicesKept, i)
				}

				return m.SetIndices(indicesKept), nil
			},
		},
		meshops.RemovedUnreferencedVerticesTransformer{},

		meshops.RotateAttribute3DTransformer{
			Amount: quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		},

		// Gaussian splat has rotational data on a per vertex basis that needs
		// to be individually rotated
		meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				q := quaternion.FromTheta(math.Pi, vector3.Forward[float64]())
				oldData := m.Float4Attribute(modeling.RotationAttribute)
				rotatedData := make([]vector4.Float64, oldData.Len())
				for i := 0; i < oldData.Len(); i++ {
					old := oldData.At(i)
					rot := q.Multiply(quaternion.New(vector3.New(old.Y(), old.Z(), old.W()), old.X())).Normalize()
					rotatedData[i] = vector4.New(rot.W(), rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z())
				}

				return m.SetFloat4Attribute(modeling.RotationAttribute, rotatedData), nil
			},
		},

		// Scale the vertex data to meet their new positioning
		gausops.ScaleTransformer{
			Scale: vector3.Fill(pn.Scale.Value()),
		},
	), nil
}

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

	pointcloud := &PointcloudLoaderNode{
		Data: PointcloudLoaderNodeData{
			// Data: &FileNode{
			// 	Data: FileNodeData{
			// 		Path: &generator.ParameterNode[string]{
			// 			Name:         "Pointcloud Path",
			// 			DefaultValue: "./point_cloud/iteration_30000/point_cloud.ply",
			// 			CLI: &generator.CliParameterNodeConfig[string]{
			// 				FlagName: "splat",
			// 				Usage:    "Path to the guassian splat to load (PLY file)",
			// 			},
			// 		},
			// 	},
			// },
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

	croppedCloud := meshops.CropAttribute3DNode{
		Data: meshops.CropAttribute3DNodeData{
			Mesh: &SplatEditNode{
				Data: SplatEditNodeData{
					SplatData: pointcloud.Out(),
					MinOpacity: &generator.ParameterNode[float64]{
						Name:         "Minimum Opacity",
						DefaultValue: 0.,
					},
					MinScale: &generator.ParameterNode[float64]{
						Name:         "Minimum Scale",
						DefaultValue: 0.,
					},
					MaxScale: &generator.ParameterNode[float64]{
						Name:         "Maximum Scale",
						DefaultValue: 1000.,
					},
					Scale: scale,
				},
			},
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
			Mesh:   baloonNode.Out(),
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
