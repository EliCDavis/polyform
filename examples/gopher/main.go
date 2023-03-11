package main

import (
	"image/color"
	"log"
	"time"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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

	gopherBody := marching.MultiSegmentLine(
		[]vector3.Float64{
			buttPosition,
			hipPosition,
			headPosition,
		},
		.8,
		1,
	).WithColor(gopherBlue)

	eyeWidth := 0.4
	gopherEye := marching.Sphere(vector3.New(eyeWidth, 1, 0.6), 0.4, 1).WithColor(white)
	gopherPupil := marching.Sphere(vector3.New(eyeWidth+.1, 1, .85), 0.2, 2).WithColor(black)

	topLegStart := vector3.New(0.7, 0.4, 0.45)
	hand := vector3.New(.9, 0.6, 0.6)

	gopherTopLeg := marching.Line(
		topLegStart,
		hand,
		0.1,
		1,
	).WithColor(gopherYellow)

	bottomLegStart := vector3.New(0., -0.1, 0.)
	toesPos := vector3.New(.7, -0.6, 0.4)
	gopherBottomLeg := marching.Line(
		bottomLegStart,
		toesPos,
		0.1,
		1,
	).WithColor(gopherYellow)

	gopherTail := marching.Line(
		vector3.New(0., -0.1, 0.),
		vector3.New(0., -0.4, -.8),
		0.15,
		1,
	).WithColor(gopherBlue)

	gopherOuterEar := marching.Line(
		vector3.New(.3, 0.5, .2),
		vector3.New(.6, 1.7, 0.),
		0.2,
		1,
	).WithColor(gopherBlue)

	gopherInnerEar := marching.Line(
		vector3.New(.3, 0.5, .2),
		vector3.New(.6, 1.7, .1),
		0.07,
		0.2,
	).WithColor(gopherBlue)

	gopherEar := gopherOuterEar.Modify(
		modeling.PositionAttribute,
		gopherInnerEar,
		func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat {
			return func(v vector3.Float64) float64 {
				return a(v)
				// return a(v) * b(v)
			}
		}).WithColor(gopherBlue)

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
		vector3.One[float64]().Scale(0.2),
		1,
	).WithColor(color.RGBA{255, 255, 255, 255})

	gopher := marching.MirrorAxis(marching.CombineFields(
		gopherBody,
		gopherEye,
		gopherPupil,
		gopherTopLeg,
		gopherBottomLeg,
		gopherTail,
		gopherEar,
		gopherNose,
		gopherJowel,
		gopherTooth,
	), marching.XAxis)

	mesh := gopher.March(modeling.PositionAttribute, 40, 0.).
		SmoothLaplacian(10, .1).
		CalculateSmoothNormals().
		SetMaterial(modeling.Material{
			Name:         "Gopher",
			DiffuseColor: color.RGBA{R: 90, G: 218, B: 255, A: 255},
		})

	log.Printf("time to mesh: %s", time.Since(start))

	armDir := hand.Sub(topLegStart).Normalized()
	legDir := toesPos.Sub(bottomLegStart).Normalized()
	skeleton := animation.NewSkeleton(animation.NewJoint(
		"Hip",
		hipPosition,
		vector3.Up[float64](),
		vector3.Forward[float64](),
		animation.NewJoint(
			"Head",
			headPosition,
			vector3.Up[float64](),
			vector3.Forward[float64](),
		),
		animation.NewJoint(
			"Arm",
			topLegStart,
			armDir,
			vector3.Forward[float64](),
			animation.NewJoint("Hand", hand, armDir, vector3.Forward[float64]()),
		),
		animation.NewJoint(
			"Leg",
			bottomLegStart,
			legDir,
			vector3.Forward[float64](),
			animation.NewJoint("Toes", toesPos, legDir, vector3.Forward[float64]()),
		),
	))

	joinData := make([]vector4.Float64, mesh.AttributeLength())
	weightData := make([]vector4.Float64, mesh.AttributeLength())
	mesh.ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
		joints := skeleton.ClosestJoints(v)
		distances := make([]float64, len(joints))

		total := 0.
		for i, j := range joints {
			distances[i] = skeleton.WorldPosition(j).Distance(v)
			total += distances[i]
		}

		for i := range joints {
			distances[i] = 1. - (distances[i] / total)
		}

		joint := vector4.New(float64(joints[0]), float64(joints[1]), float64(joints[2]), float64(joints[3]))
		weight := vector4.New(distances[0], distances[1], distances[2], distances[3])

		// joinData[i] = vector3.New(float64(closestJointIndex), 0., 0.)
		// weightData[i] = vector3.Right[float64]()

		// d1 := jointPositions[0].Distance(v)
		// d2 := jointPositions[2].Distance(v)
		// total := d1 + d2
		joinData[i] = joint
		weightData[i] = weight
	})

	mesh = mesh.
		SetFloat4Attribute(modeling.JointAttribute, joinData).
		SetFloat4Attribute(modeling.WeightAttribute, weightData)

	animationWave := animation.NewSequence("Hip/Arm/Hand", []animation.Frame{
		// animation.NewFrame(0, hand),
		// animation.NewFrame(1, hand.Add(vector3.Up[float64]())),
		// animation.NewFrame(2, hand),
		animation.NewFrame(0, vector3.Zero[float64]().Add(hand)),
		animation.NewFrame(1, vector3.Up[float64]().Scale(0.5).Add(hand)),
		animation.NewFrame(2, vector3.Zero[float64]().Add(hand)),
	})

	animations := []animation.Sequence{
		animationWave,
	}

	err := gltf.SaveTextWithAnimations("tmp/gopher/gopher.gltf", mesh, &skeleton, animations)
	if err != nil {
		panic(err)
	}
}
