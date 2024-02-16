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
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type PointcloudNode struct {
	nodes.StructData[modeling.Mesh]

	Path nodes.NodeOutput[string]
}

func (pn *PointcloudNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: pn}
}

func (pn PointcloudNode) Process() (modeling.Mesh, error) {
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

	return plyMesh.Transform(
		meshops.RotateAttribute3DTransformer{
			Amount: quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		},
		meshops.CustomTransformer{
			Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
				q := quaternion.FromTheta(math.Pi, vector3.Forward[float64]())
				oldData := m.Float4Attribute(modeling.RotationAttribute)
				rotatedData := make([]vector4.Float64, oldData.Len())
				for i := 0; i < oldData.Len(); i++ {
					old := oldData.At(i)
					rot := q.Multiply(quaternion.New(vector3.New(old.X(), old.Y(), old.Z()), old.W()))
					rotatedData[i] = vector4.New(rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z(), rot.W())
				}

				return m.SetFloat4Attribute(modeling.RotationAttribute, rotatedData), nil
			},
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
				count := posData.Len()

				newPos := make([]vector3.Float64, count)

				baloonPos := bn.Position.Data()
				baloonRadius := bn.Radius.Data()
				baloonStrength := bn.Strength.Data()

				for i := 0; i < count; i++ {
					curPos := posData.At(i)
					dir := curPos.Sub(baloonPos)
					len := dir.Length()

					if len <= baloonRadius {
						newPos[i] = baloonPos.Add(dir.Scale(baloonStrength))
					} else {
						newPos[i] = curPos
					}
				}

				return m.SetFloat3Attribute(modeling.PositionAttribute, newPos), nil
			},
		}), nil
}

func main() {
	croppedCloud := meshops.CropAttribute3DNode{
		Mesh: (&PointcloudNode{
			Path: (&generator.ParameterNode[string]{
				Name:         "Pointcloud Path",
				DefaultValue: "C:/dev/projects/sfm/gaussian-splatting/output/84e0f0cd-f/point_cloud/iteration_30000/point_cloud.ply",
			}),
		}).Out(),
		AABB: (&BoundingBoxNode{
			LeftCutoff:    &generator.ParameterNode[float64]{Name: "Left Cutoff", DefaultValue: -10},
			DownCutoff:    &generator.ParameterNode[float64]{Name: "Bottom Cutoff", DefaultValue: -10},
			BackCutoff:    &generator.ParameterNode[float64]{Name: "Back Cutoff", DefaultValue: -10},
			RightCutoff:   &generator.ParameterNode[float64]{Name: "Right Cutoff", DefaultValue: 10},
			UpCutoff:      &generator.ParameterNode[float64]{Name: "Top Cutoff", DefaultValue: 10},
			ForwardCutoff: &generator.ParameterNode[float64]{Name: "Forward Cutoff", DefaultValue: 10},
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

	app := generator.App{
		Name:        "Edit Gaussian Splats",
		Version:     "1.0.0",
		Description: "Crop and Scale portions of Gaussian Splat data",
		Authors:     []generator.Author{{Name: "Eli C Davis"}},
		WebScene: &room.WebScene{
			Background: coloring.WebColor{0, 0, 0, 255},
			Fog: room.WebSceneFog{
				Color: coloring.WebColor{0, 0, 0, 255},
				Near:  5,
				Far:   25,
			},
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.splat": generator.SplatArtifactNode(baloonNode.Out()),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
