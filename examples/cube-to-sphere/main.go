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

func main() {
	type CubeToSphereParams struct {
		Time       nodes.Node[float64]
		Resolution nodes.Node[float64]
	}

	params := CubeToSphereParams{
		Time: &generator.ParameterNode[float64]{
			Name:         "Time",
			DefaultValue: 0.,
		},
		Resolution: &generator.ParameterNode[float64]{
			Name:         "Mesh Resolution",
			DefaultValue: 30,
		},
	}

	cubeToSphereNode := nodes.Transformer(
		"Cube To Sphere Animation",
		params,
		func(in CubeToSphereParams) (modeling.Mesh, error) {
			time := math.Max(math.Min(in.Time.Data(), 1), 0)

			box := marching.Box(vector3.Float64{}, vector3.New(0.7, 0.5, 0.5), 1)
			sphere := marching.Sphere(vector3.Float64{}, 0.5*time, 1)

			return marching.
				CombineFields(box, sphere).
				March(modeling.PositionAttribute, in.Resolution.Data(), 0), nil
		},
	)

	smoothedMeshNode := SmoothNormalsNode(
		LaplacianSmoothingNode(
			cubeToSphereNode,
			&generator.ParameterNode[int]{
				Name:         "Smoothing Iterations",
				DefaultValue: 20,
			},
			&generator.ParameterNode[float64]{
				Name:         "Smoothing Factor",
				DefaultValue: .1,
			},
		),
	)

	app := generator.App{
		Name:        "Cube to Sphere",
		Description: "Smoothly blend a cube into a sphere",
		Version:     "1.0.0",
		Producers: map[string]nodes.Node[generator.Artifact]{
			"mesh.glb": GltfArtifactNode(smoothedMeshNode),
		},
	}

	err := app.Run()

	if err != nil {
		panic(err)
	}
}

type LaplacianSmoothTransformerParams struct {
	Attribute       nodes.Node[string]
	Iterations      nodes.Node[int]
	SmoothingFactor nodes.Node[float64]
	Mesh            nodes.Node[modeling.Mesh]
}

func LaplacianSmoothingNode(
	mesh nodes.Node[modeling.Mesh],
	iterations nodes.Node[int],
	factor nodes.Node[float64],
) *nodes.TransformerNode[LaplacianSmoothTransformerParams, modeling.Mesh] {
	return nodes.Transformer(
		"Laplacian Smoothing",
		LaplacianSmoothTransformerParams{
			Attribute:       nodes.Value[string](modeling.PositionAttribute),
			Iterations:      iterations,
			SmoothingFactor: factor,
			Mesh:            mesh,
		},
		func(in LaplacianSmoothTransformerParams) (modeling.Mesh, error) {
			return meshops.LaplacianSmooth(
				in.Mesh.Data(),
				in.Attribute.Data(),
				in.Iterations.Data(),
				in.SmoothingFactor.Data(),
			), nil
		},
	)
}

type MeshNodeParams struct {
	Mesh nodes.Node[modeling.Mesh]
}

func SmoothNormalsNode(
	mesh nodes.Node[modeling.Mesh],
) *nodes.TransformerNode[MeshNodeParams, modeling.Mesh] {
	return nodes.Transformer(
		"Smooth Normals",
		MeshNodeParams{
			Mesh: mesh,
		},
		func(in MeshNodeParams) (modeling.Mesh, error) {
			return meshops.SmoothNormals(
				in.Mesh.Data(),
			), nil
		},
	)
}

func GltfArtifactNode(mesh nodes.Node[modeling.Mesh]) *nodes.TransformerNode[MeshNodeParams, generator.Artifact] {
	return nodes.Transformer[MeshNodeParams, generator.Artifact](
		"To GLTF Artifact",
		MeshNodeParams{Mesh: mesh},
		func(in MeshNodeParams) (generator.Artifact, error) {
			return &generator.GltfArtifact{
				Scene: gltf.PolyformScene{
					Models: []gltf.PolyformModel{
						{
							Name: "Mesh",
							Mesh: in.Mesh.Data(),
						},
					},
				},
			}, nil
		},
	)
}
