package main

import (
	"math"

	"github.com/EliCDavis/polyform/formats/colmap"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

type Cache struct {
	Filepath *string
	Mesh     modeling.Mesh
}

var cache Cache = Cache{Mesh: modeling.EmptyMesh(modeling.PointTopology)}

func load(filepath string) (modeling.Mesh, error) {
	if cache.Filepath != nil && *cache.Filepath == filepath {
		return cache.Mesh, nil
	}

	colmapData, err := colmap.LoadSparsePointData(filepath)
	if err != nil {
		return cache.Mesh, err
	}

	cache.Filepath = &filepath
	cache.Mesh = colmapData
	return cache.Mesh, nil
}

func main() {
	app := generator.App{
		Name:        "Crop COLMAP Sparse Pointcloud",
		Version:     "1.0.0",
		Description: "Crop COLMAP Sparse Featurematch data from the reconstruction step",
		Authors:     []generator.Author{{Name: "Eli C Davis"}},
		Generator: &generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					// "C:/dev/sfm/chris2023/sparse/0/points3D.bin"
					&generator.StringParameter{
						Name: "COLMAP data",
						CLI: &generator.StringCliParameterConfig{
							FlagName: "colmap-data",
						},
					},
					&generator.FloatParameter{Name: "Bottom Cutoff", DefaultValue: -10},
					&generator.FloatParameter{Name: "Top Cutoff", DefaultValue: 10},
					&generator.FloatParameter{Name: "Left Cutoff", DefaultValue: -10},
					&generator.FloatParameter{Name: "Right Cutoff", DefaultValue: 10},
					&generator.FloatParameter{Name: "Forward Cutoff", DefaultValue: 10},
					&generator.FloatParameter{Name: "Back Cutoff", DefaultValue: -10},
				},
			},
			Producers: map[string]generator.Producer{
				"pointcloud.glb": func(c *generator.Context) (generator.Artifact, error) {
					params := c.Parameters

					colmapData, err := load(params.String("COLMAP data"))
					if err != nil {
						return nil, err
					}

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Mesh: colmapData.Transform(

										// COLMAP Pointdata is upside down
										meshops.RotateAttribute3DTransformer{
											Amount: quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
										},

										// Put it in the correct colorspace for GLTF
										meshops.VertexColorSpaceTransformer{
											Transformation: meshops.VertexColorSpaceSRGBToLinear,
										},

										// Crop points outside bounding box
										meshops.CropAttribute3DTransformer{
											BoundingBox: geometry.NewAABBFromPoints(
												vector3.New(
													params.Float64("Left Cutoff"),
													params.Float64("Bottom Cutoff"),
													params.Float64("Back Cutoff"),
												),
												vector3.New(
													params.Float64("Right Cutoff"),
													params.Float64("Top Cutoff"),
													params.Float64("Forward Cutoff"),
												),
											),
										},
									),
								},
							},
						},
					}, nil
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
