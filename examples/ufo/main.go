package main

import (
	"image/color"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func AbductionRing(radius, baseThickness, magnitude float64) modeling.Mesh {
	pathSize := 120
	path := make([]vector3.Float64, pathSize)
	thickness := make([]float64, pathSize)

	angleIncrement := (1.0 / float64(pathSize)) * 2.0 * math.Pi
	for i := 0; i < pathSize; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector3.New(math.Cos(angle)*radius, math.Sin(angle*5)*magnitude, math.Sin(angle)*radius)
		thickness[i] = (math.Sin(angle*8) * magnitude * 0.25) + baseThickness
	}

	mat := modeling.Material{
		Name:              "Abduction Ring",
		DiffuseColor:      color.RGBA{0, 255, 0, 255},
		AmbientColor:      color.RGBA{0, 255, 0, 255},
		SpecularColor:     color.Black,
		SpecularHighlight: 0,
		OpticalDensity:    5,
	}
	return extrude.
		ClosedCircleWithThickness(20, thickness, path).
		SetMaterial(mat)
}

func AbductionRing2() nodes.NodeOutput[modeling.Mesh] {
	radius := &parameter.Float64{
		Name:         "Abduction Ring Radius",
		DefaultValue: 4.,
	}

	resolution := &parameter.Int{
		Name:         "Abduction Ring Path Resolution",
		DefaultValue: 120,
	}

	sample := &SampleNode{
		Data: SampleNodeData{
			Samples: resolution,
			End: &parameter.Float64{
				Name:         "2Pi",
				DefaultValue: math.Pi * 2,
			},
		},
	}

	path := &VectorArrayNode{
		Data: VectoryArrayNodeData{
			X: &CosNode{Data: CosNodeData{Input: sample, Scale: radius}},
			Z: &SinNode{Data: SinNodeData{Input: sample, Scale: radius}},
			Y: &SinNode{Data: SinNodeData{
				Input: &SampleNode{
					Data: SampleNodeData{
						Samples: resolution,
						End: &parameter.Float64{
							Name:         "Abduction Ring Frequency",
							DefaultValue: math.Pi * 4 * 2,
						},
					},
				},
				Scale: &parameter.Float64{
					Name:         "Abduction Ring Amplitude",
					DefaultValue: .5,
				},
			}},
		},
	}

	return &extrude.PolygonNode{
		Data: extrude.PolygonNodeData{
			Sides: &parameter.Int{
				Name:         "Abduction Ring Resolution",
				DefaultValue: 20,
			},
			Thickness: &ShiftNode{
				Data: ShiftNodeData{
					In: &CosNode{
						Data: CosNodeData{
							Input: &SampleNode{
								Data: SampleNodeData{
									Samples: resolution,
									End: &parameter.Float64{
										Name:         "Abduction Ring Thickness Frequency",
										DefaultValue: math.Pi * 2 * 2 * 3,
									},
								},
							},
							Scale: &parameter.Float64{
								Name:         "Thickness Scale",
								DefaultValue: 0.25,
							},
						},
					},
					Shift: &parameter.Float64{
						Name:         "Thickness Shift",
						DefaultValue: .5,
					},
				},
			},
			Path: path,
			Closed: &parameter.Bool{
				Name:         "Closed",
				DefaultValue: true,
			},
		},
	}
}

func contour(positions []vector3.Float64, times int) modeling.Mesh {
	return repeat.Circle(extrude.CircleWithConstantThickness(7, .3, positions), times, 0)
}

func sideLights(numberOfLights int, radius float64) modeling.Mesh {
	sides := 8
	light := primitives.Cylinder{Sides: sides, Height: 0.5, Radius: 0.5}.ToMesh().
		Append(primitives.Cylinder{Sides: sides, Height: 0.25, Radius: 0.25}.ToMesh().Transform(
			meshops.TranslateAttribute3DTransformer{
				Amount: vector3.New(0., .35, 0.),
			},
		)).
		Rotate(quaternion.FromTheta(-math.Pi/2, vector3.Forward[float64]()))

	return repeat.Circle(light, numberOfLights, radius)
}

func UfoBody(outerRadius float64, portalRadius float64, frameSections int) modeling.Mesh {
	path := []vector3.Float64{
		vector3.Up[float64]().Scale(-1),
		vector3.Up[float64]().Scale(2),

		vector3.Up[float64]().Scale(0.5),
		vector3.Up[float64]().Scale(3),
		vector3.Up[float64]().Scale(4),
		vector3.Up[float64]().Scale(5),

		vector3.Up[float64]().Scale(5.5),
		vector3.Up[float64]().Scale(5.5),
	}
	thickness := []float64{
		0,
		portalRadius - 1,

		portalRadius,
		outerRadius,
		outerRadius,
		portalRadius,

		portalRadius,
		portalRadius - 1,
	}

	domeResolution := 10
	domeHight := 2
	domeStartHeight := 5.5
	domeStartWidth := portalRadius - 1
	halfPi := math.Pi / 2.
	domePath := make([]vector3.Float64, 0)
	domePath = append(domePath, path[len(path)-1])
	domeThickness := make([]float64, 0)
	domeThickness = append(domeThickness, thickness[len(thickness)-1])
	for i := 0; i < domeResolution; i++ {
		percent := float64(i+1) / float64(domeResolution)

		height := math.Sin(percent*halfPi) * float64(domeHight)
		domePath = append(domePath, vector3.Up[float64]().Scale(height+domeStartHeight))

		cosResult := math.Cos(percent * halfPi)
		domeThickness = append(domeThickness, (cosResult * domeStartWidth))
	}

	mat := modeling.Material{
		Name:              "UFO Body",
		DiffuseColor:      color.RGBA{128, 128, 128, 255},
		AmbientColor:      color.RGBA{128, 128, 128, 255},
		SpecularColor:     color.RGBA{128, 128, 128, 255},
		SpecularHighlight: 100,
		OpticalDensity:    1,
	}

	domeMat := modeling.Material{
		Name:              "UFO Dome",
		DiffuseColor:      color.RGBA{0, 0, 255, 255},
		AmbientColor:      color.RGBA{0, 0, 255, 255},
		SpecularColor:     color.RGBA{0, 0, 255, 255},
		SpecularHighlight: 100,
		Transparency:      0.2,
		OpticalDensity:    2,
	}

	return extrude.CircleWithThickness(20, thickness, path).
		Append(contour([]vector3.Float64{
			vector3.New(thickness[2], path[2].Y(), 0),
			vector3.New(thickness[3], path[3].Y(), 0),
			vector3.New(thickness[4], path[4].Y(), 0),
			vector3.New(thickness[5], path[5].Y(), 0),
		}, frameSections)).
		Append(primitives.
			Cylinder{
			Sides:  20,
			Height: 1,
			Radius: outerRadius + 1,
		}.ToMesh().
			Translate(vector3.New(0., 3.5, 0.))).
		Append(extrude.ClosedCircleWithConstantThickness(8, .25, repeat.CirclePoints(frameSections, portalRadius)).
			Translate(vector3.Up[float64]().Scale(0.5))).
		Append(sideLights(frameSections, outerRadius+1).Translate(vector3.New(0., 3.5, 0.))).
		SetMaterial(mat).
		Append(extrude.CircleWithThickness(20, domeThickness, domePath).SetMaterial(domeMat))
}

func main2() {
	ufoOuterRadius := 10.
	ufoportalRadius := 4.
	ring := AbductionRing(ufoportalRadius, 0.5, 0.5)
	ringSpacing := vector3.New(0., 3., 0.)
	final := ring.
		Append(ring.
			Scale(vector3.Fill(.75)).
			Translate(ringSpacing.Scale(1)).
			Rotate(quaternion.FromTheta(0.3, vector3.Down[float64]()))).
		Append(ring.
			Scale(vector3.Fill(.5)).
			Translate(ringSpacing.Scale(2)).
			Rotate(quaternion.FromTheta(0.5, vector3.Down[float64]()))).
		Append(UfoBody(ufoOuterRadius, ufoportalRadius, 8).Translate(ringSpacing.Scale(2.5)))

	mtlFile, err := os.Create("ufo.mtl")
	if err != nil {
		panic(err)
	}
	defer mtlFile.Close()

	objFile, err := os.Create("ufo.obj")
	if err != nil {
		panic(err)
	}
	defer objFile.Close()

	obj.WriteMesh(final, "ufo.mtl", objFile)
	obj.WriteMaterialsFromMesh(final, mtlFile)
}

func main() {
	ringCount := &parameter.Int{
		Name:         "Ring Count",
		DefaultValue: 3,
	}

	scaleSample := &SampleNode{
		Data: SampleNodeData{
			Start: &parameter.Float64{
				Name:         "Start Ring Scale",
				DefaultValue: 1,
			},
			End: &parameter.Float64{
				Name:         "End Ring Scale",
				DefaultValue: .3,
			},
			Samples: ringCount,
		},
	}

	portalRadius := 4.
	ufoOuterRadius := 10.

	ufoOutline := &parameter.Vector3Array{
		Name: "UFO Outline",
		DefaultValue: []vector3.Vector[float64]{
			vector3.New(0., -1., 0.),
			vector3.New(portalRadius-1, 2., 0.),

			vector3.New(portalRadius, 0.5, 0.),
			vector3.New(ufoOuterRadius, 3., 0.),
			vector3.New(ufoOuterRadius, 4., 0.),
			vector3.New(portalRadius, 5., 0.),

			vector3.New(portalRadius, 5.5, 0.),
			vector3.New(0, 5.5, 0.),
		},
	}

	contour := &repeat.CircleNode{
		Data: repeat.CircleNodeData{
			Mesh: &extrude.PolygonNode{
				Data: extrude.PolygonNodeData{
					Path: ufoOutline,
					Sides: &parameter.Value[int]{
						Name:         "Contour Sides",
						DefaultValue: 20,
					},
					ThicknessScale: &parameter.Float64{
						Name:         "Contour Thickness",
						DefaultValue: .2,
					},
				},
			},
			Times: &parameter.Int{
				Name:         "Countour Repeat Times",
				DefaultValue: 10,
			},
		},
	}

	allRings := &repeat.Node{
		Data: repeat.NodeData{
			Mesh: AbductionRing2(),
			Position: &VectorArrayNode{
				Data: VectoryArrayNodeData{
					Y: &SampleNode{
						Data: SampleNodeData{
							Start: &parameter.Float64{
								Name:         "Ring Start Position",
								DefaultValue: -10,
							},
							End: &parameter.Float64{
								Name:         "Ring End Position",
								DefaultValue: -3,
							},
							Samples: ringCount,
						},
					},
				},
			},
			Scale: &VectorArrayNode{
				Data: VectoryArrayNodeData{
					X: scaleSample,
					Y: scaleSample,
					Z: scaleSample,
				},
			},
		},
	}

	body := &extrude.ScrewNode{
		Data: extrude.ScrewNodeData{
			Line: ufoOutline,
			Segments: &parameter.Int{
				Name:         "UFO Resolution",
				DefaultValue: 40,
			},
			Distance: &parameter.Value[float64]{
				Name: "Height",
			},
			Revolutions: &parameter.Value[float64]{
				Name:         "Revolutions",
				DefaultValue: 1,
			},
		},
	}

	smoothedBody := &meshops.SmoothNormalsNode{
		Data: meshops.SmoothNormalsNodeData{
			Mesh: body,
		},
	}

	ufoBodyMaterial := &GltfMaterialNode{
		Data: GltfMaterialNodeData{
			Color: &parameter.Color{
				Name:         "UFO Color",
				DefaultValue: coloring.WebColor{R: 225, G: 225, B: 225, A: 255},
			},
			MetallicFactor: &parameter.Float64{
				Name:         "UFO Metallic",
				DefaultValue: 1,
			},
			RoughnessFactor: &parameter.Float64{
				Name:         "UFO Roughness",
				DefaultValue: .4,
			},
			Clearcoat: &GltfMaterialClearcoatExtensionNode{
				Data: GltfMaterialClearcoatExtensionNodeData{
					ClearcoatFactor: &parameter.Float64{
						Name:         "UFO ClearcoatFactor",
						DefaultValue: .25,
					},
					ClearcoatRoughnessFactor: &parameter.Float64{
						Name:         "UFO ClearcoatRoughnessFactor",
						DefaultValue: .15,
					},
				},
			},
		},
	}

	abductionRingColor := &parameter.Color{
		Name:         "Abduction Ring Color",
		DefaultValue: coloring.Green(),
	}

	backgroundColor := coloring.WebColor{R: 0x26, G: 0x22, B: 0x69, A: 255}
	app := generator.App{
		Name:        "UFO",
		Version:     "1.0.0",
		Description: "Demo for different GLTF material extensions",
		WebScene: &room.WebScene{
			Background: backgroundColor,
			Fog: room.WebSceneFog{
				Color: backgroundColor,
				Near:  10,
				Far:   150,
			},
			Ground:   coloring.WebColor{R: 0x2e, G: 0x47, B: 0x2e, A: 255},
			Lighting: coloring.White(),
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"ufo.glb": &GltfArtifact{
				Data: GltfArtifactData{
					Models: []nodes.NodeOutput[gltf.PolyformModel]{
						&GltfModel{
							Data: GltfModelData{
								Mesh:     smoothedBody,
								Material: ufoBodyMaterial,
							},
						},
						&GltfModel{
							Data: GltfModelData{
								Mesh: allRings,
								Material: &GltfMaterialNode{
									Data: GltfMaterialNodeData{
										Color:          abductionRingColor,
										EmissiveFactor: abductionRingColor,
										EmissiveStrength: &parameter.Float64{
											Name:         "Abduction Ring Emissive Strength",
											DefaultValue: 3,
										},
									},
								},
							},
						},
						&GltfModel{
							Data: GltfModelData{
								Mesh: &meshops.TranslateAttribute3DNode{
									Data: meshops.TranslateAttribute3DNodeData{
										Amount: &parameter.Vector3{
											Name:         "Dome Position",
											DefaultValue: vector3.New(0., 5., 0.),
										},
										Mesh: &meshops.SmoothNormalsNode{
											Data: meshops.SmoothNormalsNodeData{Mesh: &primitives.HemisphereNode{
												Data: primitives.HemisphereNodeData{
													Radius: &parameter.Float64{
														Name:         "Hemisphere Radius",
														DefaultValue: 3.1,
													},
													Rows: &parameter.Int{
														Name:         "UFO Dome Rows",
														DefaultValue: 40,
													},
													Columns: &parameter.Int{
														Name:         "UFO Dome Columns",
														DefaultValue: 40,
													},
												},
											},
											},
										},
									},
								},
								Material: &GltfMaterialNode{
									Data: GltfMaterialNodeData{
										Color: &parameter.Color{
											Name:         "Dome Color",
											DefaultValue: coloring.Grey(200),
										},
										MetallicFactor: &parameter.Float64{
											Name:         "Dome Metalic",
											DefaultValue: 0,
										},
										Transmission: &GltfMaterialTransmissionExtensionNode{
											Data: GltfMaterialTransmissionExtensionNodeData{
												TransmissionFactor: &parameter.Float64{
													Name:         "Dome Transmission",
													DefaultValue: .9,
												},
											},
										},
										Volume: &GltfMaterialVolumeExtensionNode{
											Data: GltfMaterialVolumeExtensionNodeData{
												ThicknessFactor: &parameter.Float64{
													Name:         "Dome Thickness",
													DefaultValue: .5,
												},
											},
										},
										RoughnessFactor: &parameter.Float64{
											Name:         "Dome Roughness",
											DefaultValue: .2,
										},
										IndexOfRefraction: &parameter.Float64{
											Name:         "Dome IOR",
											DefaultValue: 1.52,
										},
									},
								},
							},
						},
						&GltfModel{
							Data: GltfModelData{
								Mesh:     contour,
								Material: ufoBodyMaterial,
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
