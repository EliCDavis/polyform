package main

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/basics"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func AbductionRing() nodes.NodeOutput[modeling.Mesh] {
	radius := &parameter.Float64{
		Name:         "Abduction Ring Radius",
		DefaultValue: 4.,
	}

	resolution := &parameter.Int{
		Name:         "Abduction Ring Path Resolution",
		DefaultValue: 120,
	}

	sample := &basics.SampleNode{
		Data: basics.SampleNodeData{
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
				Input: &basics.SampleNode{
					Data: basics.SampleNodeData{
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

	return &extrude.CircleNode{
		Data: extrude.CircleNodeData{
			Resolution: &parameter.Int{
				Name:         "Abduction Ring Resolution",
				DefaultValue: 20,
			},
			Radii: &ShiftNode{
				Data: ShiftNodeData{
					In: &CosNode{
						Data: CosNodeData{
							Input: &basics.SampleNode{
								Data: basics.SampleNodeData{
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

func main() {
	ringCount := &parameter.Int{
		Name:         "Ring Count",
		DefaultValue: 3,
	}

	scaleSample := &basics.SampleNode{
		Data: basics.SampleNodeData{
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

	contour := &repeat.MeshNode{
		Data: repeat.MeshNodeData{
			Mesh: &extrude.CircleNode{
				Data: extrude.CircleNodeData{
					Path: ufoOutline,
					Resolution: &parameter.Value[int]{
						Name:         "Contour Sides",
						DefaultValue: 20,
					},
					Radius: &parameter.Float64{
						Name:         "Contour Thickness",
						DefaultValue: .2,
					},
				},
			},
			Transforms: &repeat.CircleNode{
				Data: repeat.CircleNodeData{
					Times: &parameter.Int{
						Name:         "Countour Repeat Times",
						DefaultValue: 10,
					},
				},
			},
		},
	}

	allRings := &repeat.MeshNode{
		Data: repeat.MeshNodeData{
			Mesh: AbductionRing(),
			Transforms: &TRSNode{
				Data: TRSNodeData{
					Position: &VectorArrayNode{
						Data: VectoryArrayNodeData{
							Y: &basics.SampleNode{
								Data: basics.SampleNodeData{
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
			UVs: &primitives.StripUVsNode{
				Data: primitives.StripUVsNodeData{
					Start: &parameter.Vector2{
						Name:         "UV Start",
						DefaultValue: vector2.New(0., 0.5),
					},
					End: &parameter.Vector2{
						Name:         "UV End",
						DefaultValue: vector2.New(20, 0.5),
					},
				},
			},
		},
	}

	smoothedBody := &meshops.SmoothNormalsImplicitWeldNode{
		Data: meshops.SmoothNormalsImplicitWeldNodeData{
			Mesh: body,
			Distance: &parameter.Float64{
				Name:         "Weld Dist",
				DefaultValue: 0.0001,
			},
		},
	}

	ufoBodyMaterial := &GltfMaterialNode{
		Data: GltfMaterialNodeData{
			Color: &parameter.Color{
				Name:         "UFO Color",
				DefaultValue: coloring.White(),
			},
			MetallicFactor: &parameter.Float64{
				Name:         "UFO Metallic",
				DefaultValue: 1,
			},
			RoughnessFactor: &parameter.Float64{
				Name:         "UFO Roughness",
				DefaultValue: .3,
			},
			ColorTexture: &parameter.String{
				Name:         "Color Tex URI",
				DefaultValue: "brushed.png",
			},
			MetallicRoughnessTexture: &parameter.String{
				Name:         "Metalic Tex URI",
				DefaultValue: "rough.png",
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
			Ground:   coloring.WebColor{R: 0x51, G: 0x6e, B: 0x51, A: 255},
			Lighting: coloring.White(),
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"ufo.glb": &GltfArtifact{
				Data: GltfArtifactData{
					Models: []nodes.NodeOutput[gltf.PolyformModel]{
						&GltfModel{
							Data: GltfModelNodeData{
								Mesh:     smoothedBody,
								Material: ufoBodyMaterial,
							},
						},
						&GltfModel{
							Data: GltfModelNodeData{
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
							Data: GltfModelNodeData{
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
							Data: GltfModelNodeData{
								Mesh:     contour,
								Material: ufoBodyMaterial,
							},
						},
					},
				},
			},
			// "brushed.png": artifact.NewImageNode(&BrushedMetalNode{
			// 	Data: BrushedMetalNodeNodeData{
			// 		Dimensions: &parameter.Int{
			// 			Name:         "Tex Dimensions",
			// 			DefaultValue: 512,
			// 		},
			// 		Count: &parameter.Int{
			// 			Name:         "Brush Count",
			// 			DefaultValue: 50,
			// 		},
			// 		BrushSize: &parameter.Float64{
			// 			Name:         "Brush Size",
			// 			DefaultValue: 10,
			// 		},
			// 	},
			// }),
			"brushed.png": artifact.NewImageNode(&texturing.SeamlessPerlinNode{
				Data: texturing.SeamlessPerlinNodeData{
					Positive: &parameter.Color{
						Name:         "Positive",
						DefaultValue: coloring.Grey(222),
					},
					Negative: &parameter.Color{
						Name:         "Negative",
						DefaultValue: coloring.Grey(202),
					},
				},
			}),
			"rough.png": artifact.NewImageNode(&texturing.SeamlessPerlinNode{
				Data: texturing.SeamlessPerlinNodeData{
					Positive: &parameter.Color{
						Name: "Positive",
						DefaultValue: coloring.WebColor{
							R: 0,
							G: 70,  // Perfectly reflective
							B: 255, // Always metalic
							A: 255,
						},
					},
					Negative: &parameter.Color{
						Name: "Negative",
						DefaultValue: coloring.WebColor{
							R: 0,
							G: 200, // Somewhat reflective
							B: 255, // Always metalic
							A: 255,
						},
					},
				},
			}),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
