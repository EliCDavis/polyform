package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
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
	PolyformTexture: &gltf.PolyformTexture{
		URI: "pipe-normal.png",
	},
}

var pipeMrTexture = &gltf.PolyformTexture{
	URI: "pipe-mr.png",
}

func PipeMaterial(seed *rand.Rand) *gltf.PolyformMaterial {
	painted := seed.Float64() > 0.5

	if painted {
		metalic := chance.NewRange1D(.5, 1, seed).Value()
		rough := chance.NewRange1D(0, 1, seed).Value()
		return &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor:          pipeColor(),
				MetallicRoughnessTexture: pipeMrTexture,
				MetallicFactor:           &metalic,
				RoughnessFactor:          &rough,
			},
			NormalTexture: pipeNormalTexture,
		}
	} else {
		grey := byte(127 + (128 * rand.Float64()))
		metalic := chance.NewRange1D(.9, 1, seed).Value()
		rough := chance.NewRange1D(0, .25, seed).Value()
		return &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor:          color.RGBA{grey, grey, grey, 255},
				MetallicRoughnessTexture: pipeMrTexture,
				MetallicFactor:           &metalic,
				RoughnessFactor:          &rough,
			},
			NormalTexture: pipeNormalTexture,
		}
	}
}

type PipeNode = nodes.StructNode[generator.Artifact, PipeNodeData]

type PipeNodeData struct {
	Positions nodes.NodeOutput[[]vector3.Float64]

	NumberOfShelfs nodes.NodeOutput[int]
	ShelfSpacing   nodes.NodeOutput[float64]
	ShelfWidth     nodes.NodeOutput[float64]
	LegSpacing     nodes.NodeOutput[float64]

	PipeResolution    nodes.NodeOutput[int]
	PipeMinimumRadius nodes.NodeOutput[float64]
	PipeMaximumRadius nodes.NodeOutput[float64]

	LegWidth         nodes.NodeOutput[float64]
	LegHeight        nodes.NodeOutput[float64]
	FoundationWidth  nodes.NodeOutput[float64]
	FoundationHeight nodes.NodeOutput[float64]
}

func (p PipeNodeData) Process() (generator.Artifact, error) {
	gltfModels := make([]gltf.PolyformModel, 0)

	pipeSides := p.PipeResolution.Value()
	legWidth := p.LegWidth.Value()

	randSeed := rand.New(rand.NewSource(time.Now().Unix()))
	radius := chance.NewRange1D(
		p.PipeMinimumRadius.Value(),
		p.PipeMaximumRadius.Value(),
		randSeed,
	)

	pipeUvOffset := chance.NewRange1D(
		0,
		10.,
		randSeed,
	)

	path := p.Positions.Value()

	legHeight := p.LegHeight.Value()
	numShelfs := p.NumberOfShelfs.Value()
	shelfSpacing := p.ShelfSpacing.Value()

	shelfHeights := make([]float64, numShelfs)
	for i := 0; i < numShelfs; i++ {
		shelfHeights[i] = legHeight - 0.5 - (float64(i) * shelfSpacing)
	}

	legSpacing := p.LegSpacing.Value()
	shelfWidth := p.ShelfWidth.Value()

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
			Mesh:     &pipes,
			Material: PipeMaterial(randSeed),
		})
	}

	rack := Rack{
		Leg: RackLeg{
			FoundationHeight: p.FoundationHeight.Value(),
			FoundationWidth:  p.FoundationWidth.Value(),

			Height: legHeight,
			Width:  legWidth,
		},
		LegPositions: path,
		LegSpacing:   legSpacing,
		Shelfs:       shelfHeights,
		ShelfWidth:   shelfWidth,
	}

	roughness := 0.
	rackMesh := base.Append(rack.Mesh())
	gltfModels = append(gltfModels, gltf.PolyformModel{
		Name: "Rack",
		Mesh: &rackMesh,
		Material: &gltf.PolyformMaterial{
			Name: "Rack",
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor: color.RGBA{200, 200, 200, 255},
				MetallicRoughnessTexture: &gltf.PolyformTexture{
					URI: "ibeam-mr.png",
				},
				RoughnessFactor: &roughness,
			},
			NormalTexture: &gltf.PolyformNormal{
				PolyformTexture: &gltf.PolyformTexture{
					URI: "ibeam-normal.png",
				},
			},
		},
	})

	return artifact.Gltf{
		Scene: gltf.PolyformScene{
			Models: gltfModels,
		},
	}, nil
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
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"pipe-normal.png": artifact.NewImageNode((&PipeNormalsNode{
				Data: PipeNormalsNodeData{
					BlurIterations: &parameter.Int{
						Name:         "Pipe Normal/Blur Iterations",
						DefaultValue: 7,
					},

					LineCount: &parameter.Int{
						Name:         "Pipe Normal/Line Count",
						DefaultValue: 3,
					},
					LineWidth: &parameter.Float64{
						Name:         "Pipe Normal/Line Width",
						DefaultValue: 7,
					},
					BoltCount: &parameter.Int{
						Name:         "Pipe Normal/Bolt Count",
						DefaultValue: 7,
					},
					BoltRadius: &parameter.Float64{
						Name:         "Pipe Normal/Bolt Radius",
						DefaultValue: 6.,
					},
				},
			}).Out()),
			"pipe-mr.png": &MetallicRoughnessNode{
				Data: MetallicRoughnessNodeData{
					Octaves: &parameter.Int{
						Name:         "Pipe Metallic/Noise Octaves",
						DefaultValue: 3,
					},
					MinimumRoughness: &parameter.Float64{
						Name:         "Pipe Metallic/Minimum Roughness",
						DefaultValue: 0.2,
					},
					MaximumRoughness: &parameter.Float64{
						Name:         "Pipe Metallic/Maximum Roughness",
						DefaultValue: 0.5,
					},
				},
			},
			"ibeam-mr.png": &MetallicRoughnessNode{
				Data: MetallicRoughnessNodeData{
					Octaves: &parameter.Int{
						Name:         "Pipe Metallic/Noise Octaves",
						DefaultValue: 3,
					},
					MinimumRoughness: &parameter.Float64{
						Name:         "Pipe Metallic/Minimum Roughness",
						DefaultValue: 0.4,
					},
					MaximumRoughness: &parameter.Float64{
						Name:         "Pipe Metallic/Maximum Roughness",
						DefaultValue: 0.7,
					},
				},
			},
			"ibeam-normal.png": artifact.NewImageNode(&PerlinNoiseNormalsNode{
				Data: PerlinNoiseNormalsNodeData{
					Octaves: &parameter.Int{
						Name:         "IBeam Normal/Noise Octaves",
						DefaultValue: 3,
					},
					BlurIterations: &parameter.Int{
						Name:         "IBeam Normal/Blur Iterations",
						DefaultValue: 5,
					},
				},
			}),
			"structure.glb": &PipeNode{
				Data: PipeNodeData{
					PipeMinimumRadius: &parameter.Float64{
						Name:         "Pipe/Minimum Radius",
						DefaultValue: 0.05,
					},
					PipeMaximumRadius: &parameter.Float64{
						Name:         "Pipe/Maximum Radius",
						DefaultValue: 0.15,
					},
					PipeResolution: &parameter.Int{
						Name:         "Pipe/Sides",
						DefaultValue: 16,
					},
					LegHeight: &parameter.Float64{
						Name:         "Leg/Height",
						DefaultValue: 8,
					},
					LegWidth: &parameter.Float64{
						Name:         "Leg Width",
						DefaultValue: 0.5,
					},
					FoundationHeight: &parameter.Float64{
						Name:         "Leg/Foundation Height",
						DefaultValue: 0.1,
					},
					FoundationWidth: &parameter.Float64{
						Name:         "Leg/Foundation Width",
						DefaultValue: 1,
					},
					LegSpacing: &parameter.Float64{
						Name:         "Rack/Leg Spacing",
						DefaultValue: 2.,
					},
					NumberOfShelfs: &parameter.Int{
						Name:         "Rack/Number of Shelfs",
						DefaultValue: 3,
					},
					ShelfWidth: &parameter.Float64{
						Name:         "Rack/Shelf Width",
						DefaultValue: .2,
					},
					ShelfSpacing: &parameter.Float64{
						Name:         "Rack/Shelf Spacing",
						DefaultValue: 0.5,
					},
					Positions: &parameter.Value[[]vector3.Float64]{
						Name: "Positions",
						DefaultValue: []vector3.Vector[float64]{
							vector3.New(4*0., 0., 0.),
							vector3.New(4*1., 0., 0.),
							vector3.New(4*2., 0., 4.),
							vector3.New(4*3., 0., 4.),
						},
					},
				},
			},
		},
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
