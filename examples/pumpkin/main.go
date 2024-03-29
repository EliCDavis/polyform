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
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/colors"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type Albedo struct {
	nodes.StructData[image.Image]

	Positive nodes.NodeOutput[coloring.WebColor]
	Negative nodes.NodeOutput[coloring.WebColor]
}

func (an *Albedo) Process() (image.Image, error) {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	nR, nG, nB, _ := an.Negative.Data().RGBA()
	pR, pG, pB, _ := an.Positive.Data().RGBA()

	rRange := float64(pR>>8) - float64(nR>>8)
	gRange := float64(pG>>8) - float64(nG>>8)
	bRange := float64(pB>>8) - float64(nB>>8)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 0.5) + 0.5

			r := uint32(float64(nR) + (rRange * p))
			g := uint32(float64(nG) + (gRange * p))
			b := uint32(float64(nB) + (bRange * p))

			img.Set(x, y, color.RGBA{
				R: byte(r), // byte(len * 255),
				G: byte(g),
				B: byte(b),
				A: 255,
			})
		}
	}
	return img, nil
}

func (a *Albedo) Image() nodes.NodeOutput[image.Image] {
	return &nodes.StructNodeOutput[image.Image]{
		Definition: a,
	}
}

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

type MarchingCubes struct {
	nodes.StructData[modeling.Mesh]

	Field         nodes.NodeOutput[marching.Field]
	CubersPerUnit nodes.NodeOutput[float64]
}

func (mc *MarchingCubes) Mesh() nodes.NodeOutput[modeling.Mesh] {
	return nodes.StructNodeOutput[modeling.Mesh]{
		Definition: mc,
	}
}

func (mc MarchingCubes) Process() (modeling.Mesh, error) {
	addFieldStart := time.Now()
	canvas := marching.NewMarchingCanvas(mc.CubersPerUnit.Data())
	canvas.AddField(mc.Field.Data())
	log.Printf("time to add field: %s", time.Since(addFieldStart))

	marchStart := time.Now()
	log.Println("starting march...")
	mesh := canvas.March(0)
	log.Printf("time to march: %s", time.Since(marchStart))
	return mesh, nil
}

type EdgeDetection struct {
	nodes.StructData[[][]float64]

	SrcImage  nodes.NodeOutput[image.Image]
	FillValue nodes.NodeOutput[float64]
}

func (ed EdgeDetection) Process() ([][]float64, error) {
	src := ed.SrcImage.Data()
	imageData := make([][]float64, src.Bounds().Dx())
	fillValue := ed.FillValue.Data()
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

func (ed *EdgeDetection) ImageData() nodes.NodeOutput[[][]float64] {
	return nodes.StructNodeOutput[[][]float64]{
		Definition: ed,
	}
}

func loadImage(imageData []byte) (image.Image, error) {
	imgBuf := bytes.NewBuffer(imageData)
	img, _, err := image.Decode(imgBuf)
	return img, err
}

type PropogateHeat struct {
	nodes.StructData[[][]float64]

	Data       nodes.NodeOutput[[][]float64]
	Iterations nodes.NodeOutput[int]
	Decay      nodes.NodeOutput[float64]
}

func (ph *PropogateHeat) ImageData() nodes.NodeOutput[[][]float64] {
	return nodes.StructNodeOutput[[][]float64]{
		Definition: ph,
	}
}

func (ph PropogateHeat) Process() ([][]float64, error) {
	originalData := ph.Data.Data()
	iterations := ph.Iterations.Data()
	decay := ph.Decay.Data()

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
	maxHeatNode := &generator.ParameterNode[float64]{
		Name:         "Max Heat",
		DefaultValue: 100.,
	}
	img, err := loadImage(facePNG)
	check(err)

	edgeDetection := &EdgeDetection{
		SrcImage:  nodes.Value(img),
		FillValue: maxHeatNode,
	}

	imgData := PropogateHeat{
		Data: edgeDetection.ImageData(),
		Iterations: &generator.ParameterNode[int]{
			Name:         "Iterations",
			DefaultValue: 250,
		},
		Decay: &generator.ParameterNode[float64]{
			Name:         "Decay",
			DefaultValue: 0.9999,
		},
	}
	check(debugPropegation(imgData.ImageData().Data(), "debug.png"))

	topDip := &generator.ParameterNode[float64]{
		Name:         "Top Dip",
		DefaultValue: .2,
	}

	pumpkinField := PumpkinField{
		MaxWidth: &generator.ParameterNode[float64]{
			Name:         "Max Width",
			DefaultValue: .3,
		},
		TopDip: topDip,
		DistanceFromCenter: &generator.ParameterNode[float64]{
			Name:         "Wedge Spacing",
			DefaultValue: .1,
		},
		WedgeLineRadius: &generator.ParameterNode[float64]{
			Name:         "Wedge Radius",
			DefaultValue: .3,
		},
		Sides: &generator.ParameterNode[int]{
			Name:         "Wedges",
			DefaultValue: 10,
		},
		UseImageField: &generator.ParameterNode[bool]{
			Name:         "Carve",
			DefaultValue: true,
		},
		ImageField: imgData.ImageData(),
	}

	pumpkinMesh := &SphericalUVMapping{
		Mesh: (&meshops.SmoothNormalsNode{
			Mesh: (&meshops.LaplacianSmoothNode{
				Mesh: (&MarchingCubes{
					Field: pumpkinField.Field(),
					CubersPerUnit: &generator.ParameterNode[float64]{
						Name:         "Pumpkin Resolution",
						DefaultValue: 20,
					},
				}).Mesh(),
				Iterations: &generator.ParameterNode[int]{
					Name:         "Smoothing Iterations",
					DefaultValue: 20,
				},
				SmoothingFactor: &generator.ParameterNode[float64]{
					Name:         "Smoothing Factor",
					DefaultValue: .1,
				},
			}).SmoothedMesh(),
		}).SmoothedMesh(),
	}

	textureDimensions := &generator.ParameterNode[int]{
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
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"pumpkin.glb": (&PumpkinGLBArtifact{
				PumpkinBody: pumpkinMesh.SphericalMesh(),
				LightColor: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Light Color",
					DefaultValue: coloring.WebColor{R: 0xf4, G: 0xf5, B: 0xad, A: 255},
				},
				PumpkinStem: (&StemMesh{
					StemResolution: &generator.ParameterNode[float64]{
						Name:         "Stem Resolution",
						DefaultValue: 100,
					},
					TopDip: topDip,
				}).Mesh(),
			}).Artifact(),
			"pumpkin.png": generator.NewImageArtifactNode((&Albedo{
				Positive: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Base Color",
					DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
				},
				Negative: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Negative Color",
					DefaultValue: coloring.WebColor{R: 0xf7, G: 0x71, B: 0x02, A: 255},
				},
			}).Image()),
			"stem.png": generator.NewImageArtifactNode((&Albedo{
				Positive: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Stem Base Color",
					DefaultValue: coloring.WebColor{R: 0xce, G: 0xa2, B: 0x7e, A: 255},
				},
				Negative: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Stem Negative Color",
					DefaultValue: coloring.WebColor{R: 0x7d, G: 0x53, B: 0x2c, A: 255},
				},
			}).Image()),
			"normal.png": (&NormalImage{
				NumberOfLines: &generator.ParameterNode[int]{
					Name:         "Number of Lines",
					DefaultValue: 20,
				},
				NumberOfWarts: &generator.ParameterNode[int]{
					Name:         "Number of Warts",
					DefaultValue: 50,
				},
			}).Image(),
			"stem-normal.png": (&StemNormalImage{
				NumberOfLines: &generator.ParameterNode[int]{
					Name:         "Stem Normal Line Count",
					DefaultValue: 30,
				},
			}).Image(),
			"roughness.png": (&MetalRoughness{
				Roughness: &generator.ParameterNode[float64]{
					Name:         "Pumpkin Roughness",
					DefaultValue: 0.75,
				},
			}).Image(),
			"stem-roughness.png": (&StemRoughness{
				Dimensions: textureDimensions,
				Roughness:  nodes.Value(0.78),
			}).Image(),
			// "uvMap.png": nodes.InputFromFunc(func() generator.Artifact {
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

	if err := app.Run(); err != nil {
		panic(err)
	}
}
