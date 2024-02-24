package main

import (
	"bufio"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type PointcloudLoaderNode struct {
	nodes.StructData[modeling.Mesh]

	Path nodes.NodeOutput[string]
}

func (pn *PointcloudLoaderNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: pn}
}

func (pn PointcloudLoaderNode) Process() (modeling.Mesh, error) {
	f, err := os.Open(pn.Path.Data())
	if err != nil {
		return modeling.EmptyMesh(modeling.PointTopology), err
	}
	defer f.Close()

	bufReader := bufio.NewReader(f)

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

type SplatEditNode struct {
	nodes.StructData[modeling.Mesh]

	SplatData  nodes.NodeOutput[modeling.Mesh]
	MinOpacity nodes.NodeOutput[float64]
	MinScale   nodes.NodeOutput[float64]
	Scale      nodes.NodeOutput[float64]
}

func (pn *SplatEditNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: pn}
}

func (pn SplatEditNode) Process() (modeling.Mesh, error) {
	return pn.SplatData.Data().Transform(

		// Filter out points that don't meet the opacity or scale criteria
		meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				minOpacity := pn.MinOpacity.Data()
				minScale := pn.MinScale.Data()

				opacity := m.Float1Attribute(modeling.OpacityAttribute)
				scale := m.Float3Attribute(modeling.ScaleAttribute)

				indicesKept := make([]int, 0)
				for i := 0; i < opacity.Len(); i++ {
					if opacity.At(i) >= minOpacity && scale.At(i).LengthSquared() >= minScale {
						indicesKept = append(indicesKept, i)
					}
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
					rot := q.Multiply(quaternion.New(vector3.New(old.Y(), old.Z(), old.W()), old.X()))
					rotatedData[i] = vector4.New(rot.W(), rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z())
				}

				return m.SetFloat4Attribute(modeling.RotationAttribute, rotatedData), nil
			},
		},

		// Scale the vertex data to meet their new positioning
		meshops.ModifyGaussianSplatScaleTransformer{
			Scale: vector3.Fill(pn.Scale.Data()),
		},
	), nil
}

type BoundingBoxNode struct {
	nodes.StructData[geometry.AABB]

	LeftCutoff    nodes.NodeOutput[float64]
	DownCutoff    nodes.NodeOutput[float64]
	BackCutoff    nodes.NodeOutput[float64]
	RightCutoff   nodes.NodeOutput[float64]
	UpCutoff      nodes.NodeOutput[float64]
	ForwardCutoff nodes.NodeOutput[float64]
}

func (bbn *BoundingBoxNode) Out() nodes.NodeOutput[geometry.AABB] {
	return &nodes.StructNodeOutput[geometry.AABB]{Definition: bbn}
}

func (bbn BoundingBoxNode) Process() (geometry.AABB, error) {
	return geometry.NewAABBFromPoints(
		vector3.New[float64](
			bbn.LeftCutoff.Data(),
			bbn.DownCutoff.Data(),
			bbn.BackCutoff.Data(),
		),
		vector3.New[float64](
			bbn.RightCutoff.Data(),
			bbn.UpCutoff.Data(),
			bbn.ForwardCutoff.Data(),
		),
	), nil
}

type BaloonNode struct {
	nodes.StructData[modeling.Mesh]

	Mesh     nodes.NodeOutput[modeling.Mesh]
	Strength nodes.NodeOutput[float64]
	Radius   nodes.NodeOutput[float64]
	Position nodes.NodeOutput[vector3.Float64]
}

func (bn *BaloonNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: bn}
}

func (bn BaloonNode) Process() (modeling.Mesh, error) {
	return bn.Mesh.
		Data().
		Transform(meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				posData := m.Float3Attribute(modeling.PositionAttribute)
				scaleData := m.Float3Attribute(modeling.ScaleAttribute)
				count := posData.Len()

				newPos := make([]vector3.Float64, count)
				newScale := make([]vector3.Float64, count)

				baloonPos := bn.Position.Data()
				baloonRadius := bn.Radius.Data()
				baloonStrength := bn.Strength.Data()

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
		Path: (&generator.ParameterNode[string]{
			Name:         "Pointcloud Path",
			DefaultValue: "./point_cloud/iteration_30000/point_cloud.ply",
			CLI: &generator.CliParameterNodeConfig[string]{
				FlagName: "splat",
				Usage:    "Path to the guassian splat to load (PLY file)",
			},
		}),
	}

	croppedCloud := meshops.CropAttribute3DNode{
		Mesh: (&SplatEditNode{
			SplatData: pointcloud.Out(),
			MinOpacity: &generator.ParameterNode[float64]{
				Name:         "Minimum Opacity",
				DefaultValue: 0.,
			},
			MinScale: &generator.ParameterNode[float64]{
				Name:         "Minimum Scale",
				DefaultValue: 0.,
			},
			Scale: scale,
		}).Out(),
		AABB: (&BoundingBoxNode{
			LeftCutoff:    &generator.ParameterNode[float64]{Name: "Left Cutoff", DefaultValue: -6}, // -6
			DownCutoff:    &generator.ParameterNode[float64]{Name: "Bottom Cutoff", DefaultValue: -5.6},
			BackCutoff:    &generator.ParameterNode[float64]{Name: "Back Cutoff", DefaultValue: 4},
			RightCutoff:   &generator.ParameterNode[float64]{Name: "Right Cutoff", DefaultValue: 3.5},
			UpCutoff:      &generator.ParameterNode[float64]{Name: "Top Cutoff", DefaultValue: 3},
			ForwardCutoff: &generator.ParameterNode[float64]{Name: "Forward Cutoff", DefaultValue: 14}, // 14
		}).Out(),
	}

	baloonNode := BaloonNode{
		Mesh:     croppedCloud.Out(),
		Strength: &generator.ParameterNode[float64]{Name: "Baloon Strength", DefaultValue: .7},
		Radius:   &generator.ParameterNode[float64]{Name: "Baloon Radius", DefaultValue: .7},
		Position: &generator.ParameterNode[vector3.Float64]{
			Name:         "Baloon Position",
			DefaultValue: vector3.New(-0.344, 0.402, 5.363),
		},
	}

	scaleNode := meshops.ScaleAttribute3DNode{
		Mesh: baloonNode.Out(),
		Amount: (&vecn3.New[float64]{
			X: scale,
			Y: scale,
			Z: scale,
		}).Out(),
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
			AntiAlias: false,
			XrEnabled: true,
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.splat": generator.SplatArtifactNode(scaleNode.Out()),
			"info.txt": generator.TextArtifactNode((&InfoNode{
				Original: pointcloud.Out(),
				Final:    scaleNode.Out(),
			}).Out()),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
