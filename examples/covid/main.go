package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"path"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func tendrilField(start, direction vector3.Float64, radius, length float64, plumbs int) marching.Field {
	endPoint := start.Add(direction.Scale(length))
	fields := []marching.Field{}

	angleIncrement := (math.Pi * 2) / float64(plumbs)
	perpendicular := direction.Perpendicular().Normalized()
	for i := 0; i < plumbs; i++ {
		rot := modeling.UnitQuaternionFromTheta(float64(i)*angleIncrement, direction)

		plumbRadius := radius * (.7 + (rand.Float64() * .2))

		plumbSite := rot.
			Rotate(perpendicular).
			Scale(0.5).
			Add(start.Add(direction.Scale(length * 1.2)))
		fields = append(fields, marching.Sphere(plumbSite, plumbRadius, 1))
	}

	return marching.Line(start, endPoint, radius, 1).Combine(fields...)
}

func virusField(center vector3.Float64, virusWidth, time float64) marching.Field {
	fields := []marching.Field{}

	tendrilCount := 20 + int(math.Round(20*rand.Float64()))
	tendrilRadius := .5 //.2 + (.4 * rand.Float64())
	tendrilLength := .7 + (.2 * rand.Float64())
	tendrilSites := repeat.FibonacciSpherePoints(tendrilCount, virusWidth*time)
	for _, site := range tendrilSites {
		direction := site.Normalized()
		start := center.Add(site)
		numberOfPlumbs := 3 + int(math.Round(2*rand.Float64()))
		fields = append(fields, tendrilField(start, direction, tendrilRadius*time, tendrilLength*time, numberOfPlumbs))
	}

	irregularSites := repeat.FibonacciSpherePoints(tendrilCount/2, virusWidth*.5*time)
	for _, site := range irregularSites {
		start := center.Add(site)
		fields = append(fields, marching.Sphere(start, tendrilRadius*2.5, 1))
	}

	return marching.Sphere(center, virusWidth*time, 1).Combine(fields...)
}

func covidMesh(textureMap string, time float64) modeling.Mesh {
	cubesPerUnit := 10.
	canvas := marching.NewMarchingCanvas(cubesPerUnit)

	virusRadius := 2.
	canvas.AddFieldParallel(virusField(vector3.Zero[float64](), virusRadius, time))

	marchedMesh := canvas.MarchParallel(-0.1)

	if marchedMesh.PrimitiveCount() == 0 {
		return marchedMesh
	}

	uvs := make([]vector2.Float64, 0)

	return marchedMesh.
		ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
			x := (v.Length() - 1.5) / 2.1
			uvs = append(uvs, vector2.New(x, 0.5))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      10,
				SmoothingFactor: .1,
			},
			meshops.SmoothNormalsTransformer{},
		).
		SetMaterial(modeling.Material{
			Name:            "COVID",
			ColorTextureURI: &textureMap,
		})
}

func main() {
	virusColor := coloring.NewColorStack(
		coloring.NewColorStackEntry(4, 1, 1, color.RGBA{199, 195, 195, 255}),
		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{230, 50, 50, 255}),
		coloring.NewColorStackEntry(.1, 1, 1, color.RGBA{255, 112, 236, 255}),
	)

	textureMap := "covid.png"
	err := virusColor.Debug(path.Join("tmp/covid/", textureMap), 300, 100)
	if err != nil {
		panic(err)
	}

	finalMesh := covidMesh(textureMap, 1)
	err = obj.Save("tmp/covid/covid.obj", finalMesh)
	if err != nil {
		panic(err)
	}

	box := finalMesh.BoundingBox(modeling.PositionAttribute)
	size := box.Size()

	renderingMat := materials.NewLambertian(textures.NewImage(virusColor.Image(300, 100)))

	fps := 24
	duration := 3
	growFrames := fps * duration
	growInc := 1. / float64(fps*duration)

	for i := 0; i < growFrames+(fps*2); i++ {
		rand.Seed(0)
		percentComplete := math.Min(float64(i)/float64(growFrames), 1)
		mesh := covidMesh(textureMap, percentComplete)

		scene := []rendering.Hittable{}

		if mesh.PrimitiveCount() > 0 {
			rot := modeling.UnitQuaternionFromTheta((float64(i)*growInc)*math.Pi, vector3.Up[float64]())
			scene = append(scene, rendering.NewMesh(mesh.Rotate(rot), renderingMat))
		}

		origin := vector3.New(-size.X()*.61, size.Y()*0.25, -size.Z()*.61)
		lookat := vector3.Zero[float64]()
		camera := rendering.NewDefaultCamera(1, origin, lookat, 0, 0)
		imgPath := fmt.Sprintf("tmp/covid/frame_%06d.png", i)
		rendering.RenderToFile(
			25,
			200,
			800,
			scene,
			camera,
			imgPath,
			nil,
		)
		log.Println(imgPath)
	}
}
