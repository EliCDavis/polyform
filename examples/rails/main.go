package main

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/vecn/vecn3"
	"github.com/EliCDavis/vector/vector3"
)

type GlbArtifactNode = nodes.StructNode[generator.Artifact, GlbArtifactNodeData]

type GlbArtifactNodeData struct {
	Planks     nodes.NodeOutput[modeling.Mesh]
	PlankColor nodes.NodeOutput[coloring.WebColor]
	Rail       nodes.NodeOutput[modeling.Mesh]
	Rail2      nodes.NodeOutput[modeling.Mesh]
}

func (gan GlbArtifactNodeData) Process() (generator.Artifact, error) {
	railMetal := 1.
	railRough := 0.4
	plankMetal := 0.
	scene := gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "Planks",
				Mesh: gan.Planks.Value(),
				Material: &gltf.PolyformMaterial{
					Name: "Planks",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: gan.PlankColor.Value(),
						MetallicFactor:  &plankMetal,
					},
				},
			},
			{
				Name: "Rails",
				Mesh: gan.Rail.Value().Append(gan.Rail2.Value()),
				Material: &gltf.PolyformMaterial{
					Name: "Rails",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						MetallicFactor:  &railMetal,
						RoughnessFactor: &railRough,
					},
				},
			},
		},
	}

	return artifact.Gltf{
		Scene: scene,
	}, nil
}

func main() {

	width := &parameter.Float64{Name: "Width", DefaultValue: 1.}
	height := &parameter.Float64{Name: "Height", DefaultValue: .1}
	depth := &parameter.Float64{Name: "Depth", DefaultValue: .5}

	plank := &primitives.CubeNode{
		Data: primitives.CubeNodeData{
			Width:  width,
			Height: height,
			Depth:  depth,
		},
	}

	widthShift := &nodes.Multiply{
		Data: nodes.MultiplyData[float64]{
			A: width,
			B: &parameter.Float64{Name: "Spacing", DefaultValue: .3},
		},
	}

	inverseWidthShift := &nodes.Multiply{
		Data: nodes.MultiplyData[float64]{
			A: widthShift,
			B: &parameter.Float64{Name: "Flip", DefaultValue: -1},
		},
	}

	shift := vecn3.New{
		Data: vecn3.NewData[float64]{
			X: widthShift,
			Y: height,
		},
	}

	path := &parameter.Vector3Array{
		Name: "Path",
		DefaultValue: []vector3.Vector[float64]{
			vector3.New(0., 0., 0.),
			vector3.New(0., 0., 3.),
			vector3.New(0., 1., 6.),
			vector3.New(0., 0., 9.),
			vector3.New(0., 0., 12.),
			// vector3.New(0., 0., 15.),
		},
	}

	splineAlpha := &parameter.Float64{Name: "Alpha", DefaultValue: .5}

	railSpline := &curves.CatmullRomSplineNode{
		Data: curves.CatmullRomSplineNodeData{
			Points: &vecn3.ShiftArrayNode{
				Data: vecn3.ShiftArrayNodeData[float64]{
					Array:  path,
					Amount: &shift,
				},
			},
			Alpha: splineAlpha,
		},
	}

	railSplineResolution := &parameter.Int{Name: "Spline Resolution", DefaultValue: 50}
	railRadius := &parameter.Float64{Name: "Radius", DefaultValue: .05}
	railCircleResolution := &parameter.Int{Name: "Circle Resolution", DefaultValue: 10}

	rail := &extrude.CircleAlongSplineNode{
		Data: extrude.CircleAlongSplineNodeData{
			Spline:           railSpline,
			SplineResolution: railSplineResolution,
			Radius:           railRadius,
			CircleResolution: railCircleResolution,
		},
	}

	pathSpline := &curves.CatmullRomSplineNode{
		Data: curves.CatmullRomSplineNodeData{
			Points: path,
			Alpha:  splineAlpha,
		},
	}

	splineLength := &curves.LengthNode{
		Data: curves.LengthNodeData{
			Spline: pathSpline,
		},
	}

	numPlanks := &nodes.DivideNode{
		Data: nodes.DivideData[float64]{
			Dividend: splineLength,
			Divisor:  &parameter.Float64{Name: "Planks Per Meter", DefaultValue: 1},
		},
	}

	planks := &repeat.SplineNode{
		Data: repeat.SplineNodeData{
			Mesh:  plank,
			Curve: pathSpline,
			Times: &nodes.Round{
				Data: nodes.RoundData[float64]{
					A: numPlanks,
				},
			},
		},
	}

	gltfNode := &GlbArtifactNode{
		Data: GlbArtifactNodeData{
			Planks: planks,
			PlankColor: &parameter.Color{
				Name:         "Plank Color",
				DefaultValue: coloring.WebColor{R: 70, G: 46, B: 37, A: 255},
			},
			Rail: rail,
			Rail2: &extrude.CircleAlongSplineNode{
				Data: extrude.CircleAlongSplineNodeData{
					Spline: &curves.CatmullRomSplineNode{
						Data: curves.CatmullRomSplineNodeData{
							Points: &vecn3.ShiftArrayNode{
								Data: vecn3.ShiftArrayNodeData[float64]{
									Array: path,
									Amount: &vecn3.New{
										Data: vecn3.NewData[float64]{
											X: inverseWidthShift,
											Y: height,
										},
									},
								},
							},
							Alpha: splineAlpha,
						},
					},
					SplineResolution: railSplineResolution,
					Radius:           railRadius,
					CircleResolution: railCircleResolution,
				},
			},
		},
	}

	app := generator.App{
		Name:        "Rail Road Demo",
		Version:     "0.0.1",
		Description: "Small demo that let's you build a rail road track",
		Authors: []generator.Author{
			{
				Name: "Eli Davis",
				ContactInfo: []generator.AuthorContact{
					{
						Medium: "twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		WebScene: &room.WebScene{
			Background: coloring.WebColor{R: 0x91, G: 0xd2, B: 0xed},
			Ground:     coloring.WebColor{R: 0x80, G: 0xac, B: 0x8a},
			Lighting:   coloring.WebColor{R: 0xFF, G: 0xFF, B: 0xFF},
			Fog: room.WebSceneFog{
				Color: coloring.WebColor{R: 0x91, G: 0xd2, B: 0xed},
				Near:  10,
				Far:   50,
			},
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.glb": gltfNode.Out(),
			// mrTexturePath: artifact.NewImageNode(nodes.FuncValue(mrTexture)),
			// collarAlbedoPath: artifact.NewImageNode(&CollarAlbedoTextureNode{
			// 	Data: CollarAlbedoTextureNodeData{
			// 		BaseColor: &parameter.Color{
			// 			Name:         "Collar/Base Color",
			// 			DefaultValue: coloring.WebColor{46, 46, 46, 255},
			// 		},
			// 		StitchColor: &parameter.Color{
			// 			Name:         "Collar/Stitch Color",
			// 			DefaultValue: coloring.WebColor{10, 10, 10, 255},
			// 		},
			// 	},
			// }),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}

}
