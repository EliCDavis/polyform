package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func NewExtrusionPath(path []vector3.Float64, radius, uvScaling, offset float64) []extrude.ExtrusionPoint {
	allPoints := make([]extrude.ExtrusionPoint, len(path))

	distFromStart := 0.
	for i, p := range path {

		if i > 0 {
			distFromStart += p.Distance(path[i-1])
		}

		allPoints[i] = extrude.ExtrusionPoint{
			Point:     p,
			Thickness: radius,
			UV: &extrude.ExtrusionPointUV{
				Point:     vector2.New(0.5, (distFromStart*uvScaling)+offset),
				Thickness: 1,
			},
		}

	}
	return allPoints
}

type IBeam struct {
	Thickness float64
}

func (ib IBeam) Mesh() modeling.Mesh {
	d := primitives.Cube{
		Height: 1,
		Width:  ib.Thickness,
		Depth:  1,
		UVs:    primitives.DefaultCubeUVs(),
	}.UnweldedQuads()

	w := primitives.Cube{
		Height: 1,
		Depth:  ib.Thickness,
		Width:  1 - ib.Thickness,
		UVs:    primitives.DefaultCubeUVs(),
	}.UnweldedQuads()

	return d.Translate(vector3.New(0.5-(ib.Thickness/2), 0., 0.)).
		Append(d.Translate(vector3.New(-0.5+(ib.Thickness/2), 0., 0.))).
		Append(w)
}

type RackLeg struct {
	FoundationHeight float64
	FoundationWidth  float64

	Height float64
	Width  float64
}

func (rl RackLeg) Mesh(extraHeight float64) modeling.Mesh {
	foundation := primitives.
		Cube{
		Height: rl.FoundationHeight,
		Width:  rl.FoundationWidth,
		Depth:  rl.FoundationWidth,
		UVs:    primitives.DefaultCubeUVs(),
	}.UnweldedQuads().Translate(vector3.New(0, rl.FoundationHeight/2, 0))

	// leg := primitives.
	// 	Cube{
	// 	Height: rl.Height,
	// 	Width:  rl.Width,
	// 	Depth:  rl.Width,
	// }.UnweldedQuads().Translate(vector3.New(0, (rl.Height/2)+rl.FoundationHeight, 0))

	finalHeight := extraHeight + rl.Height

	leg := IBeam{
		Thickness: 0.1,
	}.
		Mesh().
		Transform(
			meshops.ScaleAttribute3DTransformer{
				Amount: vector3.New(rl.Width, finalHeight, rl.Width),
			},
			meshops.ScaleAttribute2DTransformer{
				Amount: vector2.New(finalHeight, rl.Width),
			},
			meshops.TranslateAttribute3DTransformer{
				Amount: vector3.New(0, (finalHeight/2)+rl.FoundationHeight, 0),
			},
		)

	return foundation.Append(leg)
}

type Rack struct {
	Leg          RackLeg
	LegPositions []vector3.Float64
	LegSpacing   float64
	Shelfs       []float64
	ShelfWidth   float64
}

func (r Rack) Mesh() modeling.Mesh {
	rack := modeling.EmptyMesh(modeling.TriangleTopology)

	pointDirections := extrude.DirectionsOfPoints(r.LegPositions)

	for i, pos := range r.LegPositions {
		legMesh := r.Leg.Mesh(pos.Y())

		offset := vector3.New[float64](0, 0, r.LegSpacing/2)

		legs := legMesh.Translate(offset).
			Append(legMesh.Translate(offset.Flip()))

		for _, height := range r.Shelfs {
			shelf := IBeam{Thickness: 0.1}.
				Mesh().
				Scale(vector3.New(r.ShelfWidth, r.LegSpacing, r.ShelfWidth)).
				Rotate(quaternion.FromTheta(math.Pi/2, vector3.Right[float64]())).
				Rotate(quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]())).
				Translate(vector3.New(0., height+pos.Y(), 0.))

			legs = legs.Append(shelf)
		}

		var dir = pointDirections[i]
		rot := quaternion.RotationTo(vector3.Right[float64](), dir.SetY(0))
		rack = rack.Append(legs.Rotate(rot).Translate(pos.SetY(0)))
	}

	for i := 0; i < len(r.LegPositions)-1; i++ {

		// dir := r.LegPositions[i+1].Sub(r.LegPositions[i])
		// len := dir.Length()

		// shelfing := modeling.EmptyMesh(modeling.TriangleTopology)
		// for _, height := range r.Shelfs {
		// 	shelf := IBeam{Thickness: 0.1}.
		// 		Mesh().
		// 		Scale(vector3.New(r.ShelfWidth, len, r.ShelfWidth)).
		// 		Rotate(quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]()))

		// 	shelfing = shelfing.
		// 		Append(shelf.Translate(vector3.New(len/2, height, r.LegSpacing/2))).
		// 		Append(shelf.Translate(vector3.New(len/2, height, -r.LegSpacing/2)))
		// }

		// rot := quaternion.RotationTo(vector3.Right[float64](), dir.Normalized())
		// rack = rack.Append(shelfing.Rotate(rot).Translate(r.LegPositions[i]))
	}

	return rack
}

func pipeColor() color.RGBA {
	candidates := []color.RGBA{
		{255, 0, 0, 255},
		{255, 255, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{255, 255, 255, 255},
		{200, 200, 200, 255},
	}

	return candidates[rand.Intn(len(candidates))]
}

var pipeNormalTexture = &gltf.PolyformNormal{
	PolyformTexture: gltf.PolyformTexture{
		URI: "pipe-normal.png",
	},
}

var pipeMrTexture = &gltf.PolyformTexture{
	URI: "pipe-mr.png",
}

func PipeMaterial(seed *rand.Rand) *gltf.PolyformMaterial {

	sd := seed.Float64()
	log.Println(sd)
	painted := sd > 0.5

	if painted {
		return &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor:          pipeColor(),
				MetallicRoughnessTexture: pipeMrTexture,
				MetallicFactor:           chance.NewRange1D(.5, 1, seed).Value(),
				RoughnessFactor:          chance.NewRange1D(0, 1, seed).Value(),
			},
			NormalTexture: pipeNormalTexture,
		}
	} else {
		grey := byte(127 + (128 * rand.Float64()))
		return &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor:          color.RGBA{grey, grey, grey, 255},
				MetallicRoughnessTexture: pipeMrTexture,
				MetallicFactor:           chance.NewRange1D(.9, 1, seed).Value(),
				RoughnessFactor:          chance.NewRange1D(0, .25, seed).Value(),
			},
			NormalTexture: pipeNormalTexture,
		}
	}
}

func Pipe(params generator.GroupParameter) []gltf.PolyformModel {

	gltfModels := make([]gltf.PolyformModel, 0)

	pipeParams := params.Group("Pipes")
	rackParams := params.Group("Rack")
	legParams := rackParams.Group("Leg")

	pipeSides := pipeParams.Int("Sides")
	legWidth := legParams.Float64("Width")

	randSeed := rand.New(rand.NewSource(time.Now().Unix()))
	radius := chance.NewRange1D(
		pipeParams.Float64("Min Radius"),
		pipeParams.Float64("Max Radius"),
		randSeed,
	)

	pipeUvOffset := chance.NewRange1D(
		0,
		10.,
		randSeed,
	)

	path := rackParams.Vector3Array("Positions")

	legHeight := legParams.Float64("Height")
	numShelfs := rackParams.Int("Number of Shelfs")
	shelfSpacing := rackParams.Float64("Shelf Spacing")

	shelfHeights := make([]float64, numShelfs)
	for i := 0; i < numShelfs; i++ {
		shelfHeights[i] = legHeight - 0.5 - (float64(i) * shelfSpacing)
	}

	legSpacing := rackParams.Float64("Leg Spacing")
	shelfWidth := rackParams.Float64("Shelf Width")

	innerRackWidth := legSpacing - legWidth

	base := modeling.EmptyMesh(modeling.TriangleTopology)

	for _, shelfHeight := range shelfHeights {
		pipeRadius := radius.Value()

		halfAvailableSpace := (innerRackWidth - (pipeRadius * 2)) / 2

		numPipes := int(math.Floor(halfAvailableSpace / pipeRadius))

		start := vector2.New(0.0, -halfAvailableSpace)
		end := vector2.New(0.0, halfAvailableSpace)
		dir := end.Sub(start)
		inc := 1. / float64(numPipes-1)

		stencil := make([]vector2.Float64, numPipes)
		for i := 0; i < numPipes; i++ {
			stencil[i] = start.Add(dir.Scale(inc * float64(i)))
		}

		subPaths := extrude.PathPoints(
			stencil,
			vector3.Float64Array(path).Add(vector3.New(0., shelfHeight+pipeRadius+(shelfWidth/2), 0.)),
		)

		pipes := modeling.EmptyMesh(modeling.TriangleTopology)
		for _, p := range subPaths {
			pipes = pipes.Append(extrude.Polygon(pipeSides, NewExtrusionPath(p, pipeRadius, 0.75, pipeUvOffset.Value())))
		}

		gltfModels = append(gltfModels, gltf.PolyformModel{
			Name:     "Pipes",
			Mesh:     pipes,
			Material: PipeMaterial(randSeed),
		})
	}

	rack := Rack{
		Leg: RackLeg{
			FoundationHeight: legParams.Float64("Foundation Height"),
			FoundationWidth:  legParams.Float64("Foundation Width"),

			Height: legHeight,
			Width:  legWidth,
		},
		LegPositions: path,
		LegSpacing:   legSpacing,
		Shelfs:       shelfHeights,
		ShelfWidth:   shelfWidth,
	}

	gltfModels = append(gltfModels, gltf.PolyformModel{
		Name: "Rack",
		Mesh: base.Append(rack.Mesh()),
		Material: &gltf.PolyformMaterial{
			Name: "Rack",
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor: color.RGBA{200, 200, 200, 255},
				MetallicRoughnessTexture: &gltf.PolyformTexture{
					URI: "ibeam-mr.png",
				},
				MetallicFactor:  1,
				RoughnessFactor: 0,
			},
			NormalTexture: &gltf.PolyformNormal{
				PolyformTexture: gltf.PolyformTexture{
					URI: "ibeam-normal.png",
				},
			},
		},
	})

	return gltfModels
}

func main() {

	app := generator.App{
		Name:        "Structure",
		Version:     "1.0.0",
		Description: "ProcJam 2023 Submission",
		Authors: []generator.Author{
			{
				Name:        "Eli C Davis",
				ContactInfo: []generator.AuthorContact{{Medium: "Twitter", Value: "@EliCDavis"}},
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Near:  2,
				Far:   40,
				Color: coloring.WebColor{R: 0x9f, G: 0xb0, B: 0xc1, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x7c, G: 0x83, B: 0x7d, A: 255},
			Background: coloring.WebColor{R: 0x9f, G: 0xb0, B: 0xc1, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Generator: &generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					&generator.GroupParameter{
						Name: "Pipes",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Min Radius",
								DefaultValue: 0.05,
							},

							&generator.FloatParameter{
								Name:         "Max Radius",
								DefaultValue: 0.15,
							},

							&generator.IntParameter{
								Name:         "Sides",
								DefaultValue: 16,
							},
						},
					},
					&generator.GroupParameter{
						Name: "Rack",
						Parameters: []generator.Parameter{
							&generator.GroupParameter{
								Name: "Leg",
								Parameters: []generator.Parameter{
									&generator.FloatParameter{
										Name:         "Height",
										DefaultValue: 8,
									},
									&generator.FloatParameter{
										Name:         "Width",
										DefaultValue: 0.5,
									},

									&generator.FloatParameter{
										Name:         "Foundation Height",
										DefaultValue: 0.1,
									},
									&generator.FloatParameter{
										Name:         "Foundation Width",
										DefaultValue: 1.0,
									},
								},
							},

							&generator.FloatParameter{
								Name:         "Leg Spacing",
								DefaultValue: 2.,
							},

							&generator.IntParameter{
								Name:         "Number of Shelfs",
								DefaultValue: 3,
							},

							&generator.FloatParameter{
								Name:         "Shelf Width",
								DefaultValue: 0.2,
							},

							&generator.FloatParameter{
								Name:         "Shelf Spacing",
								DefaultValue: 0.5,
							},
							&generator.VectorArrayParameter{
								Name: "Positions",
								DefaultValue: []vector3.Vector[float64]{
									vector3.New(4*0, 0., 0.),
									vector3.New(4*1, 0., 0.),
									vector3.New(4*2, 0., 4),
									vector3.New(4*3, 0., 4),
								},
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"structure.glb": func(c *generator.Context) (generator.Artifact, error) {
					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: Pipe(*c.Parameters),
						},
					}, nil
				},
				"pipe-mr.png": func(c *generator.Context) (generator.Artifact, error) {
					dim := 256
					n := noise.NewTilingNoise(dim, 1/64., 3)
					img := image.NewRGBA(image.Rect(0, 0, dim, dim))

					for x := 0; x < dim; x++ {
						for y := 0; y < dim; y++ {
							val := n.Noise(x, y)
							p := (val * 128) + 128

							p = 255 - (p * 0.75)

							img.Set(x, y, color.RGBA{
								R: 0,
								G: 50 + byte(p/2.), //roughness (0-smooth, 1-rough)
								B: 255,             //metallness
								A: 255,
							})
						}
					}

					return generator.ImageArtifact{
						Image: img,
					}, nil
				},
				"ibeam-normal.png": func(c *generator.Context) (generator.Artifact, error) {
					return generator.ImageArtifact{
						Image: ibeamNormalImage(),
					}, nil
				},
				"pipe-normal.png": func(c *generator.Context) (generator.Artifact, error) {
					return generator.ImageArtifact{
						Image: pipeNormalImage(),
					}, nil
				},
				"ibeam-mr.png": func(c *generator.Context) (generator.Artifact, error) {
					dim := 256
					n := noise.NewTilingNoise(dim, 1/64., 3)
					img := image.NewRGBA(image.Rect(0, 0, dim, dim))

					for x := 0; x < dim; x++ {
						for y := 0; y < dim; y++ {
							val := n.Noise(x, y)
							p := (val * 128) + 128

							p = 255 - (p * 0.75)

							img.Set(x, y, color.RGBA{
								R: 0,
								G: 70 + byte(p/3.), //roughness (0-smooth, 1-rough)
								B: 255,             //metallness
								A: 255,
							})
						}
					}

					return generator.ImageArtifact{
						Image: img,
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

func ibeamNormalImage() image.Image {
	dim := 256
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(256, 1/64., 3)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 128) + 128

			img.Set(x, y, color.RGBA{
				R: byte(p), // byte(len * 255),
				G: byte(p),
				B: byte(p),
				A: 255,
			})
		}
	}

	return texturing.BoxBlurNTimes(texturing.ToNormal(img), 5)
}

func pipeNormalImage() image.Image {
	dim := 256
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(256, 1/64., 3)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 128) + 128

			img.Set(x, y, color.RGBA{
				R: byte(p), // byte(len * 255),
				G: byte(p),
				B: byte(p),
				A: 255,
			})
		}
	}

	img = texturing.ToNormal(img)

	bolts := 7
	boltRadius := 6.
	boltInc := float64(dim) / float64(bolts)
	halfBoltInc := boltInc / 2

	lines := 3
	lineWidth := 7.
	lineInc := float64(dim) / float64(lines)
	halfLineInc := lineInc / 2
	for lineIndex := 0; lineIndex < lines; lineIndex++ {
		l := (lineInc * float64(lineIndex)) + halfLineInc
		normals.Line{
			Start:           vector2.New(0., l),
			End:             vector2.New(float64(dim), l),
			Width:           lineWidth,
			NormalDirection: normals.Additive,
		}.Round(img)

		for boltIndex := 0; boltIndex < bolts; boltIndex++ {
			b := (boltInc * float64(boltIndex)) + halfBoltInc

			normals.Sphere{
				Center:    vector2.New(b, l-lineWidth-boltRadius-(boltRadius)),
				Radius:    boltRadius,
				Direction: normals.Additive,
			}.Draw(img)
		}
	}

	return texturing.BoxBlurNTimes(img, 5)
}
