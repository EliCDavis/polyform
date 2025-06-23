package main

import (
	"image/color"
	"log"
	"time"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func main() {
	start := time.Now()

	// Colors
	gopherBlue := color.RGBA{R: 90, G: 218, B: 255, A: 255}
	gopherYellow := color.RGBA{R: 243, G: 245, B: 211, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	buttPosition := vector3.New(0., 0., 0.)
	hipPosition := vector3.New(0., 0.3, 0.1)
	headPosition := vector3.New(0., 1., 0.)

	armStart := vector3.New(0.7, 0.4, 0.45)
	hand := vector3.New(.9, 0.6, 0.6)

	legStart := vector3.New(.525, -0.5, 0.)
	footPos := vector3.New(.7, -0.6, 0.4)

	bodyRadius := .8
	gopherBody := marching.MultiSegmentLine(
		[]vector3.Float64{
			buttPosition,
			hipPosition,
			headPosition,
		},
		bodyRadius,
		1,
	).WithColor(gopherBlue)

	eyeWidth := 0.4
	gopherEye := marching.Sphere(vector3.New(eyeWidth, 1, 0.6), 0.4, 1).WithColor(white)
	gopherPupil := marching.Sphere(vector3.New(eyeWidth+.1, 1, .85), 0.2, 2).WithColor(black)

	armRadius := 0.1
	gopherArm := marching.Line(armStart, hand, armRadius, 1).WithColor(gopherYellow)
	gopherLeg := marching.Line(legStart, footPos, armRadius, 1).WithColor(gopherYellow)

	tailStart := vector3.New(0., -0.25, -.4)
	tailEnd := vector3.New(0., -0.4, -.8)
	tailOffset := tailEnd.Sub(tailStart)
	tailRadius := 0.15
	gopherTail := marching.Line(tailStart, tailEnd, tailRadius, 1).WithColor(gopherBlue)

	gopherEar := marching.Line(
		vector3.New(.3, 0.5, .2),
		vector3.New(.6, 1.7, 0.),
		0.2,
		1,
	).WithColor(gopherBlue)

	gopherNose := marching.Sphere(
		vector3.New(0., 0.6, .9),
		0.15,
		2,
	).WithColor(color.RGBA{0, 0, 0, 255})

	gopherJowel := marching.Sphere(
		vector3.New(.10, 0.5, 1.),
		0.15,
		2,
	).WithColor(gopherYellow)

	gopherTooth := marching.Box(
		vector3.New(.0, 0.35, .9),
		vector3.Fill(0.2),
		1,
	).WithColor(color.RGBA{255, 255, 255, 255})

	gopher := marching.MirrorAxis(marching.CombineFields(
		gopherBody,
		gopherEye,
		gopherPupil,
		gopherArm,
		gopherLeg,
		gopherTail,
		gopherEar,
		gopherNose,
		gopherJowel,
		gopherTooth,
	), marching.XAxis)

	mesh := gopher.March(modeling.PositionAttribute, 40, 0.).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      20,
				SmoothingFactor: .1,
			},
			meshops.SmoothNormalsTransformer{},
		)

	log.Printf("time to mesh: %s", time.Since(start))

	handOffset := hand.Sub(armStart)
	armDir := handOffset.Normalized()
	legDir := footPos.Sub(legStart).Normalized()

	rightArmJoint := animation.NewJoint(
		"Right Arm",
		armRadius,
		armStart,
		armDir,
		vector3.Forward[float64](),
		animation.NewJoint("Hand", armRadius, hand, armDir, vector3.Forward[float64]()),
	)

	rightLegJoint := animation.NewJoint(
		"Right Leg",
		armRadius,
		legStart,
		legDir,
		vector3.Forward[float64](),
		animation.NewJoint("Toes", armRadius, footPos, legDir, vector3.Forward[float64]()),
	)

	skeleton := animation.NewSkeleton(animation.NewJoint(
		"Hip",
		bodyRadius,
		hipPosition,
		vector3.Up[float64](),
		vector3.Forward[float64](),
		animation.NewJoint(
			"Head",
			bodyRadius,
			headPosition,
			vector3.Up[float64](),
			vector3.Forward[float64](),
		),
		animation.NewJoint(
			"Butt",
			bodyRadius,
			buttPosition,
			vector3.Up[float64](),
			vector3.Forward[float64](),
			rightLegJoint,
			animation.MirrorJoint(rightLegJoint, "Left Leg", animation.XAxis),
			animation.NewJoint(
				"Tail",
				tailRadius,
				tailStart,
				vector3.Up[float64](),
				vector3.Forward[float64](),
				animation.NewJoint(
					"Tip",
					tailRadius,
					tailEnd,
					vector3.Up[float64](),
					vector3.Forward[float64](),
				),
			),
		),
		rightArmJoint,
		animation.MirrorJoint(rightArmJoint, "Left Arm", animation.XAxis),
	))

	voxilization := gopher.Voxelize(modeling.PositionAttribute, 20, 0.05)
	mesh = animation.WeightMeshWithHeatDiffusion(mesh, skeleton, voxilization, 0.05, 100)

	tailWagAnimation := animation.NewSequence("Hip/Butt/Tail/Tip", []animation.Frame[vector3.Float64]{
		animation.NewFrame(0.1, tailOffset.Add(vector3.Right[float64]().Scale(0.2))),
		animation.NewFrame(0.2, tailOffset),
		animation.NewFrame(0.3, tailOffset.Add(vector3.Left[float64]().Scale(0.2))),
		animation.NewFrame(0.4, tailOffset),
		animation.NewFrame(0.5, tailOffset.Add(vector3.Right[float64]().Scale(0.2))),
	})

	animations := []animation.Sequence{
		tailWagAnimation,
	}

	err := gltf.SaveBinary("tmp/gopher/gopher.glb", gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name:       "Gopher",
				Mesh:       &mesh,
				Skeleton:   &skeleton,
				Animations: animations,
			},
		},
	}, nil)
	if err != nil {
		panic(err)
	}
}
