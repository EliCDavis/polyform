package main

import (
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/trs"
	vec3Node "github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type GlbArtifactNode = nodes.Struct[artifact.Artifact, GlbArtifactNodeData]

type GlbArtifactNodeData struct {
	Plank          nodes.NodeOutput[modeling.Mesh]
	PlankPositions nodes.NodeOutput[[]trs.TRS]
	PlankColor     nodes.NodeOutput[coloring.WebColor]
	Rail           nodes.NodeOutput[modeling.Mesh]
	Rail2          nodes.NodeOutput[modeling.Mesh]
}

func (gan GlbArtifactNodeData) Process() (artifact.Artifact, error) {
	railMetal := 1.
	railRough := 0.4
	plankMetal := 0.
	planks := gan.Plank.Value()
	rails := gan.Rail.Value().Append(gan.Rail2.Value())
	scene := gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name:         "Planks",
				Mesh:         &planks,
				GpuInstances: gan.PlankPositions.Value(),
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
				Mesh: &rails,
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

	return gltf.Artifact{
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

	widthShift := &math.Multiply{
		Data: math.MultiplyData[float64]{
			A: width,
			B: &parameter.Float64{Name: "Spacing", DefaultValue: .3},
		},
	}

	inverseWidthShift := &math.Multiply{
		Data: math.MultiplyData[float64]{
			A: widthShift,
			B: &parameter.Float64{Name: "Flip", DefaultValue: -1},
		},
	}

	shift := vec3Node.New{
		Data: vec3Node.NewNodeData[float64]{
			X: widthShift,
			Y: height,
		},
	}

	path := &parameter.Vector3Array{
		Name: "Path",
		DefaultValue: []vector3.Vector[float64]{
			vector3.New(0., 0., 0.),
			vector3.New(0., 0., 3.),
			vector3.New(0., -1, 6.),
			vector3.New(0., 0., 9.),
			vector3.New(0., 0., 12.),
			// vector3.New(0., 0., 12.),
			// vector3.New(0., 0., 15.),
		},
	}

	splineAlpha := &parameter.Float64{Name: "Alpha", DefaultValue: .5}

	railSpline := &curves.CatmullRomSplineNode{
		Data: curves.CatmullRomSplineNodeData{
			Points: &vec3Node.ShiftArrayNode{
				Data: vec3Node.ShiftArrayNodeData[float64]{
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

	numPlanks := &math.DivideNode{
		Data: math.DivideData[float64]{
			Dividend: splineLength,
			Divisor:  &parameter.Float64{Name: "Planks Per Meter", DefaultValue: 1},
		},
	}

	gltfNode := &GlbArtifactNode{
		Data: GlbArtifactNodeData{
			Plank: plank,
			PlankColor: &parameter.Color{
				Name:         "Plank Color",
				DefaultValue: coloring.WebColor{R: 70, G: 46, B: 37, A: 255},
			},
			PlankPositions: &repeat.SplineNode{
				Data: repeat.SplineNodeData{
					Curve: pathSpline,
					Times: &math.Round{
						Data: math.RoundData[float64]{
							A: numPlanks,
						},
					},
				},
			},
			Rail: rail,
			Rail2: &extrude.CircleAlongSplineNode{
				Data: extrude.CircleAlongSplineNodeData{
					Spline: &curves.CatmullRomSplineNode{
						Data: curves.CatmullRomSplineNodeData{
							Points: &vec3Node.ShiftArrayNode{
								Data: vec3Node.ShiftArrayNodeData[float64]{
									Array: path,
									Amount: &vec3Node.New{
										Data: vec3Node.NewNodeData[float64]{
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
		Authors: []schema.Author{
			{
				Name: "Eli Davis",
				ContactInfo: []schema.AuthorContact{
					{
						Medium: "twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		WebScene: &schema.WebScene{
			Background: coloring.WebColor{R: 0x91, G: 0xd2, B: 0xed},
			Ground:     coloring.WebColor{R: 0x80, G: 0xac, B: 0x8a},
			Lighting:   coloring.WebColor{R: 0xFF, G: 0xFF, B: 0xFF},
			Fog: schema.WebSceneFog{
				Color: coloring.WebColor{R: 0x91, G: 0xd2, B: 0xed},
				Near:  10,
				Far:   50,
			},
		},
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			"rails.glb": gltfNode.Out(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
