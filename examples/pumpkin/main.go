package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/colors"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/experimental"
	"github.com/EliCDavis/vector/vector3"
)

func jitterPositions(pos []vector3.Float64, amplitude, frequency float64) []vector3.Float64 {
	return vector3.Array[float64](pos).
		Modify(func(v vector3.Float64) vector3.Float64 {
			return vector3.New(
				noise.Perlin1D((v.X()*frequency)+0),
				noise.Perlin1D((v.Y()*frequency)+100),
				noise.Perlin1D((v.Z()*frequency)+200),
			).Scale(amplitude).Add(v)
		})
}

type MarchingCubes = nodes.Struct[modeling.Mesh, MarchingCubesData]

type MarchingCubesData struct {
	Field         nodes.NodeOutput[marching.Field]
	CubersPerUnit nodes.NodeOutput[float64]
}

func (mc MarchingCubesData) Process() (modeling.Mesh, error) {
	addFieldStart := time.Now()
	canvas := marching.NewMarchingCanvas(mc.CubersPerUnit.Value())
	canvas.AddField(mc.Field.Value())
	log.Printf("time to add field: %s", time.Since(addFieldStart))

	marchStart := time.Now()
	log.Println("starting march...")
	mesh := canvas.March(0)
	log.Printf("time to march: %s", time.Since(marchStart))
	return mesh, nil
}

type EdgeDetection = nodes.Struct[[][]float64, EdgeDetectionData]

type EdgeDetectionData struct {
	SrcImage  nodes.NodeOutput[image.Image]
	FillValue nodes.NodeOutput[float64]
}

func (ed EdgeDetectionData) Process() ([][]float64, error) {
	src := ed.SrcImage.Value()
	imageData := make([][]float64, src.Bounds().Dx())
	fillValue := ed.FillValue.Value()
	for i := 0; i < len(imageData); i++ {
		imageData[i] = make([]float64, src.Bounds().Dy())
	}

	texturing.Convolve(src, func(x, y int, kernel []color.Color) {
		if texturing.SimpleEdgeTest(kernel) {
			imageData[x][y] = 0
			return
		}

		if colors.RedEqual(kernel[4], 255) {
			imageData[x][y] = -fillValue
		} else {
			imageData[x][y] = fillValue
		}
	})

	return imageData, nil
}

func loadImage(imageData []byte) (image.Image, error) {
	imgBuf := bytes.NewBuffer(imageData)
	img, _, err := image.Decode(imgBuf)
	return img, err
}

type PropogateHeat = nodes.Struct[[][]float64, PropogateHeatData]

type PropogateHeatData struct {
	Data       nodes.NodeOutput[[][]float64]
	Iterations nodes.NodeOutput[int]
	Decay      nodes.NodeOutput[float64]
}

func (ph PropogateHeatData) Process() ([][]float64, error) {
	originalData := ph.Data.Value()
	iterations := ph.Iterations.Value()
	decay := ph.Decay.Value()

	data := make([][]float64, len(originalData))
	tempData := make([][]float64, len(data))
	for r := 0; r < len(tempData); r++ {
		data[r] = make([]float64, len(originalData[r]))
		copy(data[r], originalData[r])

		tempData[r] = make([]float64, len(data[r]))
	}

	for i := 0; i < iterations; i++ {
		toConvole := data
		toStore := tempData
		if i%2 == 1 {
			toConvole = tempData
			toStore = data
		}
		texturing.ConvolveArray[float64](toConvole, func(x, y int, kernel []float64) {
			if toConvole[x][y] == 0 {
				return
			}
			total := kernel[0] + kernel[1] + kernel[2] + kernel[3] + kernel[5] + kernel[6] + kernel[7] + kernel[8]
			toStore[x][y] = (total / 8) * decay
		})
	}

	if iterations%2 == 1 {
		return tempData, nil
	}
	return data, nil
}

func debugPropegation(data [][]float64, filename string) error {
	dst := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: len(data), Y: len(data[0])}})

	max := -math.MaxFloat64
	min := math.MaxFloat64
	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			max = math.Max(max, row[y])
			min = math.Min(min, row[y])
		}
	}

	delta := max - min

	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			val := row[y] / delta
			if val > 0 {
				dst.SetRGBA(x, y, color.RGBA{R: byte(val * 255), G: 0, B: 0, A: 255})
			} else {
				dst.SetRGBA(x, y, color.RGBA{R: 0, G: byte(val * -255), B: 0, A: 255})
			}
		}
	}

	imgFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer imgFile.Close()
	return png.Encode(imgFile, dst)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

//go:embed face.png
var facePNG []byte

func main() {
	maxHeatNode := &parameter.Float64{
		Name:         "Max Heat",
		DefaultValue: 100.,
	}
	img, err := loadImage(facePNG)
	check(err)

	edgeDetection := &EdgeDetection{
		Data: EdgeDetectionData{
			SrcImage:  nodes.Value(img),
			FillValue: maxHeatNode,
		},
	}

	imgData := &PropogateHeat{
		Data: PropogateHeatData{
			Data: edgeDetection,
			Iterations: &parameter.Int{
				Name:         "Iterations",
				DefaultValue: 250,
			},
			Decay: &parameter.Float64{
				Name:         "Decay",
				DefaultValue: 0.9999,
			},
		},
	}
	check(debugPropegation(imgData.Value(), "debug.png"))

	topDip := &parameter.Float64{
		Name:         "Top Dip",
		DefaultValue: .2,
	}

	pumpkinField := &PumpkinField{
		Data: PumpkinFieldData{
			MaxWidth: &parameter.Float64{
				Name:         "Max Width",
				DefaultValue: .3,
			},
			TopDip: topDip,
			DistanceFromCenter: &parameter.Float64{
				Name:         "Wedge Spacing",
				DefaultValue: .1,
			},
			WedgeLineRadius: &parameter.Float64{
				Name:         "Wedge Radius",
				DefaultValue: .3,
			},
			Sides: &parameter.Int{
				Name:         "Wedges",
				DefaultValue: 10,
			},
			UseImageField: &parameter.Bool{
				Name:         "Carve",
				DefaultValue: true,
			},
			ImageField: imgData,
		},
	}

	pumpkinMesh := &SphericalUVMapping{
		Data: SphericalUVMappingData{
			Mesh: &meshops.SmoothNormalsNode{
				Data: meshops.SmoothNormalsNodeData{
					Mesh: &meshops.LaplacianSmoothNode{
						Data: meshops.LaplacianSmoothNodeData{
							Mesh: &MarchingCubes{
								Data: MarchingCubesData{
									Field: pumpkinField,
									CubersPerUnit: &parameter.Float64{
										Name:         "Pumpkin Resolution",
										DefaultValue: 20,
									},
								},
							},
							Iterations: &parameter.Int{
								Name:         "Smoothing Iterations",
								DefaultValue: 20,
							},
							SmoothingFactor: &parameter.Float64{
								Name:         "Smoothing Factor",
								DefaultValue: .1,
							},
						},
					},
				},
			},
		},
	}

	textureDimensions := &parameter.Int{
		Name:         "Texture Dimension",
		DefaultValue: 1024,
	}

	app := generator.App{
		Name:        "Pumpkin",
		Version:     "1.0.0",
		Description: "Making a pumpkin for Haloween",
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
				Far:   10,
				Color: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x4f, G: 0x6d, B: 0x55, A: 255},
			Background: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Producers: map[string]nodes.NodeOutput[artifact.Artifact]{
			"pumpkin.glb": &PumpkinGLBArtifact{
				Data: PumpkinGLBArtifactData{
					PumpkinBody: pumpkinMesh,
					LightColor: &parameter.Color{
						Name:         "Light Color",
						DefaultValue: coloring.WebColor{R: 0xf4, G: 0xf5, B: 0xad, A: 255},
					},
					PumpkinStem: &StemMesh{
						Data: StemMeshData{
							StemResolution: &parameter.Float64{
								Name:         "Stem Resolution",
								DefaultValue: 100,
							},
							TopDip: topDip,
						},
					},
				},
			},
			"pumpkin.png": basics.NewImageNode(&experimental.SeamlessPerlinNode{
				Data: experimental.SeamlessPerlinNodeData{
					Positive: &parameter.Color{
						Name:         "Base Color",
						DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
					},
					Negative: &parameter.Color{
						Name:         "Negative Color",
						DefaultValue: coloring.WebColor{R: 0xf7, G: 0x71, B: 0x02, A: 255},
					},
				},
			}),
			"stem.png": basics.NewImageNode(&experimental.SeamlessPerlinNode{
				Data: experimental.SeamlessPerlinNodeData{
					Positive: &parameter.Color{
						Name:         "Stem Base Color",
						DefaultValue: coloring.WebColor{R: 0xce, G: 0xa2, B: 0x7e, A: 255},
					},
					Negative: &parameter.Color{
						Name:         "Stem Negative Color",
						DefaultValue: coloring.WebColor{R: 0x7d, G: 0x53, B: 0x2c, A: 255},
					},
				},
			}),
			"normal.png": &NormalImage{
				Data: NormalImageData{
					NumberOfLines: &parameter.Int{
						Name:         "Number of Lines",
						DefaultValue: 20,
					},
					NumberOfWarts: &parameter.Int{
						Name:         "Number of Warts",
						DefaultValue: 50,
					},
				},
			},
			"stem-normal.png": &StemNormalImage{
				Data: StemNormalImageData{
					NumberOfLines: &parameter.Int{
						Name:         "Stem Normal Line Count",
						DefaultValue: 30,
					},
				},
			},
			"roughness.png": &MetalRoughness{
				Data: MetalRoughnessData{
					Roughness: &parameter.Float64{
						Name:         "Pumpkin Roughness",
						DefaultValue: 0.75,
					},
				},
			},
			"stem-roughness.png": &StemRoughness{
				Data: StemRoughnessData{
					Dimensions: textureDimensions,
					Roughness:  nodes.Value(0.78),
				},
			},
			// "uvMap.png": nodes.InputFromFunc(func() artifact.Artifact {
			// 	img := texturing.DebugUVTexture{
			// 		ImageResolution:      1024,
			// 		BoardResolution:      10,
			// 		NegativeCheckerColor: color.RGBA{0, 0, 0, 255},

			// 		PositiveCheckerColor: color.RGBA{255, 0, 0, 255},
			// 		XColorScale:          color.RGBA{0, 255, 0, 255},
			// 		YColorScale:          color.RGBA{0, 0, 255, 255},
			// 	}.Image()
			// 	return &generator.ImageArtifact{Image: img}
			// }),
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
