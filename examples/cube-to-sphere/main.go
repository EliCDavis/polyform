package main

import (
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func main() {

	app := generator.App{
		Name:        "Cube to Sphere",
		Description: "Smoothly blend a cube into a sphere",
		Version:     "1.0.0",
		Generator: &generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					&generator.FloatParameter{
						Name: "Time",
						CLI: &generator.FloatCliParameterConfig{
							FlagName: "time",
							Usage:    "Percentage through the transition from cube to sphere, clamped between 0 and 1",
						},
					},
					&generator.IntParameter{
						Name:         "Resolution",
						DefaultValue: 30,
						CLI: &generator.IntCliParameterConfig{
							FlagName: "resolution",
							Usage:    "The resolution of the marching cubes algorithm, roughly translating to number of voxels per unit",
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"mesh.glb": func(c *generator.Context) (generator.Artifact, error) {
					time := math.Max(math.Min(c.Parameters.Float64("Time"), 1), 0)

					box := marching.Box(
						vector3.Zero[float64](),
						vector3.New(0.7, 0.5, 0.5), //.Scale(time), // Box size
						1,
					)

					sphere := marching.Sphere(
						vector3.Zero[float64](),
						0.5*time, // Sphere radius
						1,
					)

					field := marching.CombineFields(box, sphere).March(modeling.PositionAttribute, 30, 0)

					smoothedMesh := field.Transform(
						meshops.LaplacianSmoothTransformer{Iterations: 10, SmoothingFactor: 0.1},
						meshops.SmoothNormalsTransformer{},
					)

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "Mesh",
									Mesh: smoothedMesh,
								},
							},
						},
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
