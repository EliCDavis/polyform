package main

import (
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Segment struct {
	mesh   []gltf.PolyformModel
	height float64
}

func Chimney(funnelWidth, funnelHeight, taperHeight, shootWidth, shootHeight float64, rows int, color color.RGBA) []gltf.PolyformModel {
	halfTotalHeight := (taperHeight + shootHeight + funnelHeight) / 2.
	path := []vector3.Float64{
		vector3.New(0, -halfTotalHeight, 0),
		vector3.New(0, -halfTotalHeight+funnelHeight, 0),
		vector3.New(.0, -halfTotalHeight+funnelHeight+taperHeight, 0),
		vector3.New(.0, halfTotalHeight, 0),
	}

	allRows := repeat.Line(
		primitives.Cylinder{Sides: 20, Height: 0.3, Radius: shootWidth + .3}.ToMesh(),
		vector3.New(0, -halfTotalHeight+funnelHeight+taperHeight, 0),
		vector3.New(0, (shootHeight*(float64(rows)/float64(rows+1)))-halfTotalHeight+funnelHeight+taperHeight, 0),
		rows-2,
	)

	widths := []float64{
		funnelWidth,
		funnelWidth,
		shootWidth,
		shootWidth,
	}

	return []gltf.PolyformModel{
		{
			Mesh: extrude.CircleWithThickness(20, widths, path).
				Append(allRows).
				Append(primitives.Cylinder{Sides: 20, Height: 0.3, Radius: funnelWidth + .3}.ToMesh().
					Translate(vector3.New(0, -halfTotalHeight+funnelHeight, 0))),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: color,
				},
			},
		},
	}

}

func Chasis(height, width float64, rows, columns int, color color.RGBA) []gltf.PolyformModel {
	chasis := primitives.Cylinder{Sides: 20, Height: height, Radius: width}.ToMesh()

	rowSpacing := height / float64(rows+1)
	for i := 1; i <= rows; i++ {
		pos := vector3.New(0, rowSpacing*float64(i)-(height/2.), 0)
		chasis = chasis.
			Append(primitives.Cylinder{Sides: 20, Height: 0.5, Radius: width + .3}.ToMesh().Translate(pos))
	}

	column := primitives.UnitCube().Scale(vector3.New(.2, height, .2))
	columnsMesh := repeat.Circle(column, columns, width)
	chasis = chasis.Append(columnsMesh)

	return []gltf.PolyformModel{
		{
			Mesh: chasis,
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: color,
				},
			},
		},
	}
}

func Legs(height, width float64, numLegs int, legColor color.RGBA) []gltf.PolyformModel {
	columnHeight := 1.
	legHeight := height - columnHeight

	leg := primitives.Cube{Width: 1, Height: legHeight, Depth: 1}.
		UnweldedQuads().
		Translate(vector3.New(0, -(columnHeight / 2.), 0))

	return []gltf.PolyformModel{
		{
			Name: "Legs",

			Mesh: primitives.
				Cylinder{Sides: 20, Height: columnHeight, Radius: width}.ToMesh().
				Translate(vector3.New(0, (height/2.)-(columnHeight/2.), 0)).
				Append(repeat.Circle(leg, numLegs, width-2.)),

			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: legColor,
				},
			},
		},
	}
}

func Floor(floorHeight, radius, walkWidth float64, floorColor, railingColor color.RGBA) []gltf.PolyformModel {
	numLegs := int(math.Round(2*math.Pi*radius) / 4)
	legHeight := 2.
	post := primitives.UnitCube().
		Scale(vector3.New(.1, legHeight, .1)).
		Translate(vector3.New(0, legHeight/2., 0))

	pathPointCount := numLegs * 2
	angleIncrement := (1.0 / float64(pathPointCount)) * 2.0 * math.Pi
	path := make([]vector3.Float64, pathPointCount)
	postRadius := radius + walkWidth - .1
	for i := 0; i < pathPointCount; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector3.New(math.Cos(angle)*postRadius, 0, math.Sin(angle)*postRadius)
	}
	railing := extrude.ClosedCircleWithConstantThickness(12, .05, flip(path))

	sides := 20
	angleIncrement = (1.0 / float64(sides)) * 2.0 * math.Pi
	shapePath := make([]vector3.Float64, sides)
	offset := radius + (walkWidth / 2)
	for i := 0; i < sides; i++ {
		angle := angleIncrement * float64(i)
		shapePath[i] = vector3.New(math.Cos(angle)*offset, 0, math.Sin(angle)*offset)
	}

	return []gltf.PolyformModel{
		{
			Name: "Railing",
			Mesh: repeat.Circle(post, numLegs, postRadius-.2).
				Append(railing.Translate(vector3.Up[float64]().Scale(legHeight))).
				Append(railing.Translate(vector3.Up[float64]().Scale(legHeight / 2))),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: railingColor,
				},
			},
		},
		{
			Name: "Floor",
			Mesh: extrude.ClosedShape(flip(PiShape(floorHeight, walkWidth)), shapePath),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorFactor: floorColor,
				},
			},
		},
	}
}

func flip[T any](arr []T) []T {
	final := make([]T, len(arr))
	for i, p := range arr {
		final[len(arr)-i-1] = p
	}
	return final
}

func PiShape(height, width float64) []vector2.Float64 {
	halfWidth := (width / 2.)
	topHeight := height / 2.
	bottomHeight := -topHeight
	nubHeight := bottomHeight - topHeight
	nubSize := halfWidth / 3.

	return []vector2.Float64{
		vector2.New(-halfWidth, topHeight),
		vector2.New(halfWidth, topHeight),
		vector2.New(halfWidth, bottomHeight),

		vector2.New(halfWidth-nubSize, bottomHeight),
		vector2.New(halfWidth-nubSize, nubHeight),
		vector2.New(halfWidth-nubSize-nubSize, nubHeight),
		vector2.New(halfWidth-nubSize-nubSize, bottomHeight),

		vector2.New(-halfWidth+nubSize+nubSize, bottomHeight),
		vector2.New(-halfWidth+nubSize+nubSize, nubHeight),
		vector2.New(-halfWidth+nubSize, nubHeight),
		vector2.New(-halfWidth+nubSize, bottomHeight),

		vector2.New(-halfWidth, bottomHeight),
	}
}

func PutTogetherSegments(segments ...Segment) []gltf.PolyformModel {
	offset := 0.
	final := make([]gltf.PolyformModel, 0)
	for _, segment := range segments {
		offset += segment.height / 2
		for i, m := range segment.mesh {
			segment.mesh[i].Mesh = m.Mesh.Translate(vector3.New(0, offset, 0))
			final = append(final, segment.mesh[i])
		}
		offset += segment.height / 2
	}
	return final
}

func main() {
	app := generator.App{
		Name:        "Fired Heater",
		Version:     "1.0.0",
		Description: "Idk making a fired heater",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Far:   150,
				Near:  10,
				Color: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0},
			},
			Background: coloring.WebColor{R: 0x95, G: 0xcb, B: 0xf3},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xfd, B: 0xd1},
			Ground:     coloring.WebColor{R: 0x87, G: 0x82, B: 0x78},
		},
		Generator: generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					&generator.ColorParameter{
						Name:         "Color",
						DefaultValue: coloring.WebColor{R: 128, G: 128, B: 128, A: 255},
					},

					&generator.GroupParameter{
						Name: "Leg",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Length",
								DefaultValue: 5.,
							},
							&generator.ColorParameter{
								Name: "Color",
								DefaultValue: coloring.WebColor{
									R: 0x5f,
									G: 0x59,
									B: 0x54,
									A: 255,
								},
							},
							&generator.IntParameter{
								Name:         "Count",
								DefaultValue: 8,
							},
						},
					},

					&generator.GroupParameter{
						Name: "Floor",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Height",
								DefaultValue: .5,
							},
							&generator.ColorParameter{
								Name: "Floor Color",
								DefaultValue: coloring.WebColor{
									R: 0x30,
									G: 0x3b,
									B: 0x45,
									A: 255,
								},
							},
							&generator.ColorParameter{
								Name: "Railing Color",
								DefaultValue: coloring.WebColor{
									R: 0xff,
									G: 0xf7,
									B: 0x00,
									A: 255,
								},
							},

							&generator.FloatParameter{
								Name:         "Lower Walkway Width",
								DefaultValue: 4.,
							},

							&generator.FloatParameter{
								Name:         "Upper Walkway Width",
								DefaultValue: 3.,
							},
						},
					},

					&generator.GroupParameter{
						Name: "Chasis",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Height",
								DefaultValue: 20,
							},
							&generator.FloatParameter{
								Name:         "Width",
								DefaultValue: 7,
							},
							&generator.IntParameter{
								Name:         "Rows",
								DefaultValue: 4,
							},
							&generator.IntParameter{
								Name:         "Columns",
								DefaultValue: 7,
							},
						},
					},

					&generator.GroupParameter{
						Name: "Chimney",
						Parameters: []generator.Parameter{
							&generator.FloatParameter{
								Name:         "Base Height",
								DefaultValue: 4.,
							},
							&generator.FloatParameter{
								Name:         "Taper Height",
								DefaultValue: 5.,
							},
							&generator.FloatParameter{
								Name:         "Shoot Height",
								DefaultValue: 10.,
							},
							&generator.IntParameter{
								Name:         "Rows",
								DefaultValue: 4.,
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"firedheater.glb": func(c *generator.Context) (generator.Artifact, error) {

					baseColor := c.Parameters.Color("Color")

					legParams := c.Parameters.Group("Leg")
					legsHeight := legParams.Float64("Length")

					floorParams := c.Parameters.Group("Floor")
					floorHeight := floorParams.Float64("Height")

					chasisParams := c.Parameters.Group("Chasis")
					chasisWidth := chasisParams.Float64("Width")
					chasisHeight := chasisParams.Float64("Height")

					chimneyParams := c.Parameters.Group("Chimney")

					firedheater := PutTogetherSegments(
						Segment{
							mesh: Legs(
								legsHeight,
								8.,
								legParams.Int("Count"),
								legParams.Color("Color"),
							),
							height: legsHeight,
						},
						Segment{
							mesh: Floor(
								floorHeight,
								chasisWidth,
								floorParams.Float64("Lower Walkway Width"),
								floorParams.Color("Floor Color"),
								floorParams.Color("Railing Color"),
							),
							height: floorHeight,
						},
						Segment{
							mesh: Chasis(
								chasisHeight,
								chasisWidth,
								chasisParams.Int("Rows"),
								chasisParams.Int("Columns"),
								baseColor,
							),
							height: chasisHeight,
						},
						Segment{
							mesh: Floor(
								floorHeight,
								chasisWidth,
								floorParams.Float64("Upper Walkway Width"),
								floorParams.Color("Floor Color"),
								floorParams.Color("Railing Color"),
							),
							height: floorHeight,
						},
						Segment{
							mesh: Chimney(
								chasisWidth,
								chimneyParams.Float64("Base Height"),
								chimneyParams.Float64("Taper Height"),
								chasisWidth/6,
								chimneyParams.Float64("Shoot Height"),
								chimneyParams.Int("Rows"),
								baseColor,
							),
							height: chimneyParams.Float64("Base Height") +
								chimneyParams.Float64("Taper Height") +
								chimneyParams.Float64("Shoot Height"),
						},
					)

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: firedheater,
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
