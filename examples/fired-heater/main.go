package main

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
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

type ChimneyNode struct {
	nodes.StructData[Segment]

	FunnelWidth, FunnelHeight, TaperHeight, ShootWidth, ShootHeight nodes.NodeOutput[float64]
	Rows                                                            nodes.NodeOutput[int]
	Color                                                           nodes.NodeOutput[coloring.WebColor]
}

func (cn *ChimneyNode) Out() nodes.NodeOutput[Segment] {
	return &nodes.StructNodeOutput[Segment]{Definition: cn}
}

func (cn ChimneyNode) Process() (Segment, error) {
	taperHeight := cn.TaperHeight.Data()
	shootHeight := cn.ShootHeight.Data()
	funnelHeight := cn.FunnelHeight.Data()
	shootWidth := cn.ShootWidth.Data()
	funnelWidth := cn.FunnelWidth.Data()
	rows := cn.Rows.Data()
	color := cn.Color.Data()

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

type ChasisNode struct {
	nodes.StructData[Segment]

	Height, Width nodes.NodeOutput[float64]
	Rows, Columns nodes.NodeOutput[int]
	Color         nodes.NodeOutput[coloring.WebColor]
}

func (cn *ChasisNode) Out() nodes.NodeOutput[Segment] {
	return &nodes.StructNodeOutput[Segment]{Definition: cn}
}

func (cn ChasisNode) Process() (Segment, error) {
	height := cn.Height.Data()
	width := cn.Width.Data()
	rows := cn.Rows.Data()
	columns := cn.Columns.Data()
	color := cn.Color.Data()

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

type LegsNode struct {
	nodes.StructData[Segment]

	Height, Width nodes.NodeOutput[float64]
	NumLegs       nodes.NodeOutput[int]
	LegColor      nodes.NodeOutput[coloring.WebColor]
}

func (ln *LegsNode) Out() nodes.NodeOutput[Segment] {
	return &nodes.StructNodeOutput[Segment]{Definition: ln}
}

func (ln LegsNode) Process() (Segment, error) {
	height := ln.Height.Data()
	width := ln.Width.Data()
	numLegs := ln.NumLegs.Data()
	legColor := ln.LegColor.Data()

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

type FloorNode struct {
	nodes.StructData[Segment]

	FloorHeight, Radius, WalkWidth nodes.NodeOutput[float64]
	FloorColor, RailingColor       nodes.NodeOutput[coloring.WebColor]
}

func (fn *FloorNode) Out() nodes.NodeOutput[Segment] {
	return &nodes.StructNodeOutput[Segment]{Definition: fn}
}

func (fn FloorNode) Process() (Segment, error) {

	floorColor := fn.FloorColor.Data()
	railingColor := fn.RailingColor.Data()
	floorHeight := fn.FloorHeight.Data()
	radius := fn.Radius.Data()
	walkWidth := fn.WalkWidth.Data()

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

type CombineSegmentsNode struct {
	nodes.StructData[generator.Artifact]

	Segments []nodes.NodeOutput[Segment]
}

func (bn *CombineSegmentsNode) Out() nodes.NodeOutput[generator.Artifact] {
	return &nodes.StructNodeOutput[generator.Artifact]{Definition: bn}
}

func (csn CombineSegmentsNode) Process() (generator.Artifact, error) {
	offset := 0.
	final := make([]gltf.PolyformModel, 0)
	for _, segmentNode := range csn.Segments {
		segment := segmentNode.Data()
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
	return generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: final,
		},
	}, nil
}

func main() {
	baseColor := &generator.ParameterNode[coloring.WebColor]{
		Name: "Base Color",
		DefaultValue: coloring.WebColor{
			R: 128,
			G: 128,
			B: 128,
			A: 255,
		},
	}

	chasisWidth := &generator.ParameterNode[float64]{
		Name:         "Chasis Width",
		DefaultValue: 7,
	}

	railingColor := &generator.ParameterNode[coloring.WebColor]{
		Name: "Floor/Railing Color",
		DefaultValue: coloring.WebColor{
			R: 0xff,
			G: 0xf7,
			B: 0x00,
			A: 255,
		},
	}

	floorColor := &generator.ParameterNode[coloring.WebColor]{
		Name: "Floor/Color",
		DefaultValue: coloring.WebColor{
			R: 0x30,
			G: 0x3b,
			B: 0x45,
			A: 255,
		},
	}

	floorHeight := &generator.ParameterNode[float64]{
		Name:         "Floor/Height",
		DefaultValue: .5,
	}

	firedheaterNode := CombineSegmentsNode{
		Segments: []nodes.NodeOutput[Segment]{
			(&LegsNode{
				Height: &generator.ParameterNode[float64]{
					Name:         "Leg/Length",
					DefaultValue: 5.,
				},
				NumLegs: &generator.ParameterNode[int]{
					Name:         "Leg/Count",
					DefaultValue: 5,
				},
				LegColor: &generator.ParameterNode[coloring.WebColor]{
					Name: "Leg/Color",
					DefaultValue: coloring.WebColor{
						R: 0x5f,
						G: 0x59,
						B: 0x54,
						A: 255,
					},
				},
				Width: &generator.ParameterNode[float64]{
					Name:         "Leg/Width",
					DefaultValue: 8.,
				},
			}).Out(),
			(&FloorNode{
				FloorHeight: floorHeight,
				Radius:      chasisWidth,
				WalkWidth: &generator.ParameterNode[float64]{
					Name:         "Floor/Lower Walkway Width",
					DefaultValue: 4.,
				},
				FloorColor:   floorColor,
				RailingColor: railingColor,
			}).Out(),
			(&ChasisNode{
				Height: &generator.ParameterNode[float64]{
					Name:         "Chasis/Height",
					DefaultValue: 20.,
				},
				Width: chasisWidth,
				Rows: &generator.ParameterNode[int]{
					Name:         "Chasis/Rows",
					DefaultValue: 4,
				},
				Columns: &generator.ParameterNode[int]{
					Name:         "Chasis/Columns",
					DefaultValue: 7,
				},
				Color: baseColor,
			}).Out(),
			(&FloorNode{
				FloorHeight: floorHeight,
				Radius:      chasisWidth,
				WalkWidth: &generator.ParameterNode[float64]{
					Name:         "Floor/Upper Walkway Width",
					DefaultValue: 3.,
				},
				FloorColor:   floorColor,
				RailingColor: railingColor,
			}).Out(),
			(&ChimneyNode{
				FunnelWidth: chasisWidth,
				ShootWidth: &generator.ParameterNode[float64]{
					Name:         "Shoot Width",
					DefaultValue: 1,
				},
				FunnelHeight: &generator.ParameterNode[float64]{
					Name:         "Chimney/Base Height",
					DefaultValue: 4,
				},
				TaperHeight: &generator.ParameterNode[float64]{
					Name:         "Chimney/Taper Height",
					DefaultValue: 5,
				},
				ShootHeight: &generator.ParameterNode[float64]{
					Name:         "Chimney/Shoot Height",
					DefaultValue: 10,
				},
				Rows: &generator.ParameterNode[int]{
					Name:         "Chimney/Rows",
					DefaultValue: 4,
				},
				Color: baseColor,
			}).Out(),
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
			"firedheater.glb": firedheaterNode.Out(),
		},
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
