package main

import (
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CubeToSphereAnimation struct {
	nodes.StructData[modeling.Mesh]

	Time       nodes.NodeOutput[float64]
	Resolution nodes.NodeOutput[float64]
}

func (csa CubeToSphereAnimation) Process() (modeling.Mesh, error) {
	time := math.Max(math.Min(csa.Time.Data(), 1), 0)

	box := marching.Box(vector3.Float64{}, vector3.New(0.7, 0.5, 0.5), 1)
	sphere := marching.Sphere(vector3.Float64{}, 0.5*time, 1)

	return marching.
		CombineFields(box, sphere).
		March(modeling.PositionAttribute, csa.Resolution.Data(), 0), nil
}

func (csa *CubeToSphereAnimation) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: csa}
}

func main() {

	animation := CubeToSphereAnimation{
		Time: &generator.ParameterNode[float64]{
			Name:         "Time",
			DefaultValue: 0.,
		},
		Resolution: &generator.ParameterNode[float64]{
			Name:         "Mesh Resolution",
			DefaultValue: 30,
		},
	}

	smoothedMeshNode := meshops.SmoothNormalsNode{
		Mesh: (&meshops.LaplacianSmoothNode{
			Mesh: animation.Out(),
			Iterations: &generator.ParameterNode[int]{
				Name:         "Smoothing Iterations",
				DefaultValue: 20,
			},
			SmoothingFactor: &generator.ParameterNode[float64]{
				Name:         "Smoothing Factor",
				DefaultValue: .1,
			},
		}).SmoothedMesh(),
	}

	app := generator.App{
		Name:        "Cube to Sphere",
		Description: "Smoothly blend a cube into a sphere",
		Version:     "1.0.0",
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.glb": (&GltfArtifact{
				Mesh: smoothedMeshNode.SmoothedMesh(),
			}).Out(),
		},
	}

	err := app.Run()

	if err != nil {
		panic(err)
	}
}

type GltfArtifact struct {
	nodes.StructData[generator.Artifact]

	Mesh nodes.NodeOutput[modeling.Mesh]
}

func (csa GltfArtifact) Process() (generator.Artifact, error) {
	return &generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{
					Name: "Mesh",
					Mesh: csa.Mesh.Data(),
				},
			},
		},
	}, nil
}

func (csa *GltfArtifact) Out() nodes.NodeOutput[generator.Artifact] {
	return &nodes.StructNodeOutput[generator.Artifact]{Definition: csa}
}
