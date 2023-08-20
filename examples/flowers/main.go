package main

import (
	"image/color"
	"math"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

type FlowerTip func(percent float64) float64

func basicTip(percent float64) float64 {
	return math.Sin(percent * math.Pi)
}

func crocusTip(percent float64) float64 {
	finalPercent := math.Sin(percent * math.Pi)
	return finalPercent * finalPercent * finalPercent
}

func marigoldTip(percent float64) float64 {
	// I fucked around in desomos
	finalPercent := math.Sin(percent * math.Pi)
	return (finalPercent * 1.8898) - (finalPercent * finalPercent * finalPercent)
}

func pedal(baseWidth, midWidth, pedalLength, tipLength float64, arcVertCount int, tip FlowerTip) modeling.Mesh {
	halfBaseWidth := baseWidth / 2.
	halfMidWidth := midWidth / 2.

	verts := []vector3.Float64{
		vector3.New(-halfBaseWidth, 0., 0.),
		vector3.New(-halfMidWidth, 0., pedalLength),
	}

	xInx := midWidth / (float64(arcVertCount) + 2.)
	for i := 0; i < arcVertCount; i++ {
		x := -halfMidWidth + (xInx * float64(i+1))

		zPercentage := (float64(i+1) / float64(arcVertCount+2))
		z := pedalLength + (tip(zPercentage) * tipLength)
		verts = append(verts, vector3.New(x, 0., z))
	}

	verts = append(
		verts,
		vector3.New(halfMidWidth, 0., pedalLength),
		vector3.New(halfBaseWidth, 0., 0.),
	)

	faces := make([]int, 0)
	for i := 1; i < arcVertCount+3; i++ {
		faces = append(faces, 0, i, i+1)
	}

	normals := make([]vector3.Float64, len(verts))
	uvs := make([]vector2.Float64, len(verts))
	for i, v := range verts {
		zPercent := v.Z() / (pedalLength + tipLength)
		xPercent := math.Abs(v.X()) / midWidth
		uvs[i] = vector2.New(math.Max(xPercent, zPercent), (v.X()/midWidth)+0.5)
		normals[i] = vector3.Up[float64]()
	}

	return modeling.
		NewTriangleMesh(faces).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat3Attribute(modeling.NormalAttribute, normals).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func flower(numPedals int, radius, pitch float64) modeling.Mesh {
	q := modeling.UnitQuaternionFromTheta(pitch, vector3.Left[float64]())
	return repeat.Circle(pedal(0.2, 0.4, 0.4, 0.1, 10, marigoldTip).Rotate(q), numPedals, radius)
}

func texture(textureName string) error {
	yellowTip := color.RGBA{255, 175, 0, 255}
	flowerColor := coloring.NewColorStack(
		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{228, 0, 0, 255}),
		coloring.NewColorStackEntry(3, 1, 1, color.RGBA{143, 3, 1, 255}),
		coloring.NewColorStackEntry(1, 1, 1, yellowTip),
	)

	flowerImg := flowerColor.Image(300, 300)
	ctx := gg.NewContextForImage(flowerImg)
	ctx.SetColor(yellowTip)
	ctx.SetLineWidth(2)
	ctx.DrawLine(50, 150, 200, 150)
	ctx.Stroke()

	imgPath := path.Join("tmp/flowers/", textureName)
	err := os.MkdirAll(path.Dir(imgPath), os.ModeDir)
	if err != nil {
		return err
	}

	return ctx.SavePNG(imgPath)
}

func points(r *rand.Rand, width, height float64, count int) []vector3.Float64 {
	pts := make([]vector3.Float64, count)
	for i := 0; i < count; i++ {
		pts[i] = vector3.Rand(r).MultByVector(vector3.New(width, 0, height))
	}
	return pts
}

func main() {
	textureName := "flower.png"
	texture(textureName)
	singleFlower := flower(11, 0.25, 0.1).
		Append(flower(8, 0.15, 0.3).Scale(vector3.Fill(0.9))).
		Append(flower(5, 0.05, 0.6).Scale(vector3.Fill(0.8)))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	flowerPos := points(r, 3, 3, 3)

	allFlowers := modeling.EmptyMesh(modeling.TriangleTopology)
	for _, v := range flowerPos {
		allFlowers = allFlowers.Append(singleFlower.
			Translate(v).
			Scale(vector3.Fill(0.5 + (rand.Float64() * 1))),
		)
	}

	allFlowers = allFlowers.SetMaterial(modeling.Material{
		ColorTextureURI: &textureName,
	})

	gltf.SaveText("tmp/flowers/flowers.gltf", gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "Flowers",
				Mesh: allFlowers,
				Material: &gltf.PolyformMaterial{
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.White,
					},
				},
			},
		},
	})
}
