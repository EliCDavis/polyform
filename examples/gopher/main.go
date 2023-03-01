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
)

func main() {
	start := time.Now()

	// Colors
	gopherBlue := color.RGBA{R: 90, G: 218, B: 255, A: 255}
	gopherYellow := color.RGBA{R: 243, G: 245, B: 211, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	headPosition := vector3.New(0., 1., 0.)

	gopherBody := marching.MultiSegmentLine(
		[]vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(0., 0.3, 0.1),
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

	gopherBottomLeg := marching.Line(
		vector3.New(0., -0.1, 0.),
		vector3.New(.7, -0.6, 0.4),
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

	jointPositions := []vector3.Float64{
		headPosition,
		// topLegStart,
		hand,
	}
	armDir := hand.Sub(topLegStart).Normalized()

	jointTree := modeling.NewPointCloud(
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: jointPositions,
		},
		nil,
		nil,
		nil,
	).OctTree()

	joinData := make([]vector3.Float64, mesh.AttributeLength())
	weightData := make([]vector3.Float64, mesh.AttributeLength())
	mesh.ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
		closestJointIndex, _ := jointTree.ClosestPoint(v)
		joinData[i] = vector3.New(float64(closestJointIndex), 0., 0.)
		weightData[i] = vector3.Right[float64]()
	})

	mesh = mesh.
		SetFloat3Attribute(modeling.JointAttribute, joinData).
		SetFloat3Attribute(modeling.WeightAttribute, weightData)

	skeleton := animation.NewJoint(
		headPosition,
		headPosition,
		vector3.Up[float64](),
		vector3.Forward[float64](),
		// animation.NewJoint(
		// 	topLegStart,
		// 	topLegStart.Sub(headPosition),
		// 	armDir,
		// 	vector3.Forward[float64](),
		// 	animation.NewJoint(hand, hand.Sub(topLegStart), armDir, vector3.Forward[float64]()),
		// ),
		animation.NewJoint(hand, hand.Sub(headPosition), armDir, vector3.Forward[float64]()),
	)

	animationWave := animation.NewSequence(1, []animation.Frame{
		// animation.NewFrame(0, hand),
		// animation.NewFrame(1, hand.Add(vector3.Up[float64]())),
		// animation.NewFrame(2, hand),
		animation.NewFrame(0, vector3.Zero[float64]()),
		animation.NewFrame(1, vector3.Up[float64]().Scale(0.5)),
		animation.NewFrame(2, vector3.Zero[float64]()),
	})

	animations := []animation.Sequence{
		animationWave,
	}

	err := gltf.SaveTextWithAnimations("tmp/gopher/gopher.gltf", mesh, &skeleton, animations)
	if err != nil {
		panic(err)
	}
}
