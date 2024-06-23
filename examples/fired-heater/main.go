package main

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Segment struct {
	mesh   []gltf.PolyformModel
	height float64
}

type ChimneyNode = nodes.StructNode[Segment, ChimneyNodeData]

type ChimneyNodeData struct {
	FunnelWidth, FunnelHeight, TaperHeight, ShootWidth, ShootHeight nodes.NodeOutput[float64]
	Rows                                                            nodes.NodeOutput[int]
	Color                                                           nodes.NodeOutput[coloring.WebColor]
}

func (cn ChimneyNodeData) Process() (Segment, error) {
	taperHeight := cn.TaperHeight.Value()
	shootHeight := cn.ShootHeight.Value()
	funnelHeight := cn.FunnelHeight.Value()
	shootWidth := cn.ShootWidth.Value()
	funnelWidth := cn.FunnelWidth.Value()
	rows := cn.Rows.Value()
	color := cn.Color.Value()

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

	return Segment{
		mesh: []gltf.PolyformModel{
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
		},
		height: funnelHeight + taperHeight + shootHeight,
	}, nil
}

type ChasisNode = nodes.StructNode[Segment, ChasisNodeData]

type ChasisNodeData struct {
	Height, Width nodes.NodeOutput[float64]
	Rows, Columns nodes.NodeOutput[int]
	Color         nodes.NodeOutput[coloring.WebColor]
}

func (cn ChasisNodeData) Process() (Segment, error) {
	height := cn.Height.Value()
	width := cn.Width.Value()
	rows := cn.Rows.Value()
	columns := cn.Columns.Value()
	color := cn.Color.Value()

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

	return Segment{
		mesh: []gltf.PolyformModel{
			{
				Mesh: chasis,
				Material: &gltf.PolyformMaterial{
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color,
					},
				},
			},
		},
		height: height,
	}, nil
}

type LegsNode = nodes.StructNode[Segment, LegsNodeData]

type LegsNodeData struct {
	Height, Width nodes.NodeOutput[float64]
	NumLegs       nodes.NodeOutput[int]
	LegColor      nodes.NodeOutput[coloring.WebColor]
}

func (ln LegsNodeData) Process() (Segment, error) {
	height := ln.Height.Value()
	width := ln.Width.Value()
	numLegs := ln.NumLegs.Value()
	legColor := ln.LegColor.Value()

	columnHeight := 1.
	legHeight := height - columnHeight

	leg := primitives.Cube{Width: 1, Height: legHeight, Depth: 1}.
		UnweldedQuads().
		Translate(vector3.New(0, -(columnHeight / 2.), 0))

	return Segment{
		mesh: []gltf.PolyformModel{
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
		},
		height: height,
	}, nil
}

type FloorNode = nodes.StructNode[Segment, FloorNodeData]

type FloorNodeData struct {
	FloorHeight, Radius, WalkWidth nodes.NodeOutput[float64]
	FloorColor, RailingColor       nodes.NodeOutput[coloring.WebColor]
}

func (fn FloorNodeData) Process() (Segment, error) {

	floorColor := fn.FloorColor.Value()
	railingColor := fn.RailingColor.Value()
	floorHeight := fn.FloorHeight.Value()
	radius := fn.Radius.Value()
	walkWidth := fn.WalkWidth.Value()

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

	return Segment{
		mesh: []gltf.PolyformModel{
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
		},
		height: floorHeight,
	}, nil
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

type CombineSegmentsNode = nodes.StructNode[generator.Artifact, CombineSegmentsNodeData]

type CombineSegmentsNodeData struct {
	Segments []nodes.NodeOutput[Segment]
}

func (csn CombineSegmentsNodeData) Process() (generator.Artifact, error) {
	offset := 0.
	final := make([]gltf.PolyformModel, 0)
	for _, segmentNode := range csn.Segments {
		segment := segmentNode.Value()
		offset += segment.height / 2
		for i, m := range segment.mesh {
			final = append(final, gltf.PolyformModel{
				Name:     segment.mesh[i].Name,
				Material: segment.mesh[i].Material,
				Mesh:     m.Mesh.Translate(vector3.New(0, offset, 0)),
			})
		}
		offset += segment.height / 2
	}
	return artifact.Gltf{
		Scene: gltf.PolyformScene{
			Models: final,
		},
	}, nil
}

func main() {
	baseColor := &parameter.Color{
		Name: "Base Color",
		DefaultValue: coloring.WebColor{
			R: 128,
			G: 128,
			B: 128,
			A: 255,
		},
	}

	chasisWidth := &parameter.Float64{
		Name:         "Chasis Width",
		DefaultValue: 7,
	}

	railingColor := &parameter.Color{
		Name: "Floor/Railing Color",
		DefaultValue: coloring.WebColor{
			R: 0xff,
			G: 0xf7,
			B: 0x00,
			A: 255,
		},
	}

	floorColor := &parameter.Color{
		Name: "Floor/Color",
		DefaultValue: coloring.WebColor{
			R: 0x30,
			G: 0x3b,
			B: 0x45,
			A: 255,
		},
	}

	floorHeight := &parameter.Float64{
		Name:         "Floor/Height",
		DefaultValue: .5,
	}

	firedheaterNode := &CombineSegmentsNode{
		Data: CombineSegmentsNodeData{
			Segments: []nodes.NodeOutput[Segment]{
				&LegsNode{
					Data: LegsNodeData{
						Height: &parameter.Float64{
							Name:         "Leg/Length",
							DefaultValue: 5.,
						},
						NumLegs: &parameter.Int{
							Name:         "Leg/Count",
							DefaultValue: 5,
						},
						LegColor: &parameter.Color{
							Name: "Leg/Color",
							DefaultValue: coloring.WebColor{
								R: 0x5f,
								G: 0x59,
								B: 0x54,
								A: 255,
							},
						},
						Width: &parameter.Float64{
							Name:         "Leg/Width",
							DefaultValue: 8.,
						},
					},
				},
				&FloorNode{
					Data: FloorNodeData{
						FloorHeight: floorHeight,
						Radius:      chasisWidth,
						WalkWidth: &parameter.Float64{
							Name:         "Floor/Lower Walkway Width",
							DefaultValue: 4.,
						},
						FloorColor:   floorColor,
						RailingColor: railingColor,
					},
				},
				&ChasisNode{
					Data: ChasisNodeData{
						Height: &parameter.Float64{
							Name:         "Chasis/Height",
							DefaultValue: 20.,
						},
						Width: chasisWidth,
						Rows: &parameter.Int{
							Name:         "Chasis/Rows",
							DefaultValue: 4,
						},
						Columns: &parameter.Int{
							Name:         "Chasis/Columns",
							DefaultValue: 7,
						},
						Color: baseColor,
					},
				},
				&FloorNode{
					Data: FloorNodeData{
						FloorHeight: floorHeight,
						Radius:      chasisWidth,
						WalkWidth: &parameter.Float64{
							Name:         "Floor/Upper Walkway Width",
							DefaultValue: 3.,
						},
						FloorColor:   floorColor,
						RailingColor: railingColor,
					},
				},
				&ChimneyNode{
					Data: ChimneyNodeData{
						FunnelWidth: chasisWidth,
						ShootWidth: &parameter.Float64{
							Name:         "Shoot Width",
							DefaultValue: 1,
						},
						FunnelHeight: &parameter.Float64{
							Name:         "Chimney/Base Height",
							DefaultValue: 4,
						},
						TaperHeight: &parameter.Float64{
							Name:         "Chimney/Taper Height",
							DefaultValue: 5,
						},
						ShootHeight: &parameter.Float64{
							Name:         "Chimney/Shoot Height",
							DefaultValue: 10,
						},
						Rows: &parameter.Int{
							Name:         "Chimney/Rows",
							DefaultValue: 4,
						},
						Color: baseColor,
					},
				},
			},
		},
	}

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
				Color: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0, A: 255},
			},
			Background: coloring.WebColor{R: 0x95, G: 0xcb, B: 0xf3, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xfd, B: 0xd1, A: 255},
			Ground:     coloring.WebColor{R: 0x87, G: 0x82, B: 0x78, A: 255},
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"firedheater.glb": firedheaterNode,
		},
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
