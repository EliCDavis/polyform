package main

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func TranslateExtrusion(points []extrude.ExtrusionPoint, amount vector3.Float64) []extrude.ExtrusionPoint {
	finalPoints := make([]extrude.ExtrusionPoint, len(points))

	for i, p := range points {
		finalPoints[i] = p
		finalPoints[i].Point = finalPoints[i].Point.Add(amount)
	}

	return finalPoints
}

func NewExtrusionPath(path []vector3.Float64, radius float64) []extrude.ExtrusionPoint {
	allPoints := make([]extrude.ExtrusionPoint, len(path))
	for i, p := range path {
		allPoints[i] = extrude.ExtrusionPoint{
			Point:     p,
			Thickness: radius,
		}
	}
	return allPoints
}

func Pipe(params generator.GroupParameter) gltf.PolyformModel {

	height := params.Float64("Height")
	radius := params.Float64("Radius")

	masterPath := []vector3.Float64{
		vector3.Zero[float64](),
		vector3.New(0, height, 0),
		vector3.New(height, height, 0),
		vector3.New(height, height*2, 0),
		vector3.New(0, height*2, 0),
		// vector3.New(0, height*2, height*2),
		// vector3.New(0, height, height*2),
		// vector3.New(0, height, height),
	}

	subPaths := extrude.PathPoints(
		[]vector2.Vector[float64]{
			vector2.New(0.0, 0.0),
			vector2.New(0.0, 0.5),
			vector2.New(0.0, 1.0),

			vector2.New(0.5, 0.0),
			vector2.New(0.5, 0.5),
			vector2.New(0.5, 1.0),

			vector2.New(1.0, 0.0),
			vector2.New(1.0, 0.5),
			vector2.New(1.0, 1.0),
		},
		masterPath,
	)

	base := modeling.EmptyMesh(modeling.TriangleTopology)

	for _, p := range subPaths {
		base = base.Append(extrude.Polygon(params.Int("Sides"), NewExtrusionPath(p, radius)))
	}

	return gltf.PolyformModel{
		Name: "Pipe",
		Mesh: base,
	}
}

func main() {

	app := generator.App{
		Name:        "Structure",
		Version:     "1.0.0",
		Description: "ProcJam 2023 Submission",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
				ContactInfo: []generator.AuthorContact{
					{
						Medium: "Twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Near:  2,
				Far:   30,
				Color: coloring.WebColor{R: 0x9f, G: 0xb0, B: 0xc1, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x7c, G: 0x83, B: 0x7d, A: 255},
			Background: coloring.WebColor{R: 0x9f, G: 0xb0, B: 0xc1, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Generator: &generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					&generator.IntParameter{
						Name:         "Sides",
						DefaultValue: 16,
					},
					&generator.FloatParameter{
						Name:         "Height",
						DefaultValue: 3,
					},
					&generator.FloatParameter{
						Name:         "Radius",
						DefaultValue: .5,
					},
				},
			},
			Producers: map[string]generator.Producer{
				"structure.glb": func(c *generator.Context) (generator.Artifact, error) {

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								Pipe(*c.Parameters),
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
