package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
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

type IBeam struct {
	Thickness float64
}

func (ib IBeam) Mesh() modeling.Mesh {
	d := primitives.Cube{
		Height: 1,
		Width:  ib.Thickness,
		Depth:  1,
	}.UnweldedQuads()

	w := primitives.Cube{
		Height: 1,
		Depth:  ib.Thickness,
		Width:  1 - ib.Thickness,
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
		Scale(vector3.New(rl.Width, finalHeight, rl.Width)).
		Translate(vector3.New(0, (finalHeight/2)+rl.FoundationHeight, 0))

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

func Pipe(params generator.GroupParameter) []gltf.PolyformModel {

	gltfModels := make([]gltf.PolyformModel, 0)

	pipeParams := params.Group("Pipes")
	rackParams := params.Group("Rack")
	legParams := rackParams.Group("Leg")

	pipeSides := pipeParams.Int("Sides")
	legWidth := legParams.Float64("Width")

	radius := chance.NewRange1D(
		pipeParams.Float64("Min Radius"),
		pipeParams.Float64("Max Radius"),
		rand.New(rand.NewSource(0)),
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
			pipes = pipes.Append(extrude.Polygon(pipeSides, NewExtrusionPath(p, pipeRadius)))
		}

		gltfModels = append(gltfModels, gltf.PolyformModel{
			Name: "Pipes",
			Mesh: pipes,
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: pipeColor(),
				},
			},
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
					// &generator.GroupParameter{
					// 	Name: "Rack",
					// 	Parameters: []generator.Parameter{
					// 		&generator.FloatParameter{
					// 			Name:         "Leg Height",
					// 			DefaultValue: 10,
					// 		},
					// 		&generator.FloatParameter{
					// 			Name:         "Height",
					// 			DefaultValue: 10,
					// 		},
					// 	},
					// },
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
			},
		},
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
