package main

import (
	"image/color"
	"math"
	"os"
	"path"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func SquarePoints(width, height float64) []vector2.Float64 {
	halfWidth := width / 2
	halfHeight := height / 2
	return []vector2.Float64{
		vector2.New(-halfWidth, halfHeight),
		vector2.New(-halfWidth, -halfHeight),
		vector2.New(halfWidth, -halfHeight),
		vector2.New(halfWidth, halfHeight),
	}
}

func Cone(height, radius float64, sides int) modeling.Mesh {
	if sides < 3 {
		panic("can not make cone with less that 3 sides")
	}

	verts := repeat.CirclePoints(sides, radius)
	lastVert := len(verts)
	verts = append(verts, vector3.New(0, height, 0))
	uvs := make([]vector2.Float64, len(verts))
	uvs[len(uvs)-1] = vector2.One[float64]()

	tris := make([]int, 0, sides*3)
	for i := 0; i < sides; i++ {
		tris = append(tris, i, lastVert, i+1)
	}
	tris[len(tris)-1] = 0

	return modeling.NewMesh(modeling.TriangleTopology, tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func SpikeRing(spikes int, ringRadius, spikeHeight, spikeRadius float64, spikeSides int) modeling.Mesh {
	cone := Cone(spikeHeight, spikeRadius, spikeSides).
		Rotate(modeling.UnitQuaternionFromTheta(math.Pi/2, vector3.Right[float64]()))
	return repeat.Circle(cone, spikes, ringRadius)
}

func CollarAlbedoTexture(imgPath string) error {
	baseColor := color.RGBA{46, 46, 46, 255}
	stitchColor := color.RGBA{10, 10, 10, 255}

	ctx := gg.NewContext(256, 256)
	ctx.SetColor(baseColor)
	ctx.DrawRectangle(0, 0, 256, 256)
	ctx.Fill()

	ctx.SetColor(stitchColor)
	ctx.SetLineWidth(8)
	lineLength := 20
	lineGap := 12
	stiches := 8
	// ctx.DrawLine(10, 10, 30, 10)
	for i := 0; i < stiches; i++ {
		startStitch := float64((lineGap / 2) + (lineLength * i) + (lineGap * i))
		stopStitch := float64((lineGap / 2) + (lineLength * (i + 1)) + (lineGap * i))
		ctx.DrawLine(startStitch, 20, stopStitch, 20)
		ctx.DrawLine(startStitch, 235, stopStitch, 235)
	}
	ctx.Stroke()

	err := os.MkdirAll(path.Dir(imgPath), os.ModeDir)
	if err != nil {
		return err
	}

	return ctx.SavePNG(imgPath)
}

func texture(imgPath string) error {
	ctx := gg.NewContext(2, 2)
	ctx.SetColor(color.RGBA{
		R: 0,
		G: 0, // Roughness
		B: 0, // Metal - 0 = metal
		A: 255,
	})
	ctx.SetPixel(0, 0)
	ctx.SetPixel(1, 0)
	ctx.SetPixel(0, 1)
	ctx.SetPixel(1, 1)

	err := os.MkdirAll(path.Dir(imgPath), os.ModeDir)
	if err != nil {
		return err
	}

	return ctx.SavePNG(imgPath)
}

func main() {
	collarRadius := 1.
	collarHeight := 0.2
	collarThickness := 0.1
	collarResolution := repeat.CirclePoints(30, collarRadius)

	spikeCount := 20
	spikeHeight := 0.2

	collarUVs := make([]vector2.Float64, 0)
	collar := extrude.ClosedShape(SquarePoints(collarThickness, collarHeight), collarResolution).
		Transform(
			meshops.SmoothNormalsTransformer{},
		).
		ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
			xy := v.XZ().Normalized()
			angle := math.Atan2(xy.Y(), xy.X()) * 4
			height := (v.Y() + (collarHeight / 2)) / collarHeight
			collarUVs = append(collarUVs, vector2.New(angle, height))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, collarUVs)

	ouputDir := "tmp/examples/collar"
	mrTex := "collar_mr.png"
	collarAlbedo := "collar_albedo.png"

	CollarAlbedoTexture(path.Join(ouputDir, collarAlbedo))
	texture(path.Join(ouputDir, mrTex))
	err := gltf.SaveBinary(path.Join(ouputDir, "collar.glb"), gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "Collar",
				Mesh: collar,
				Material: &gltf.PolyformMaterial{
					Name: "Collar",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorTexture: &gltf.PolyformTexture{
							URI: collarAlbedo,
						},
					},
				},
			},
			{
				Name: "Spikes",
				Mesh: SpikeRing(
					spikeCount,
					collarRadius+(collarThickness/2.)-0.02, // -0.02 to set it in to the collar
					spikeHeight,
					0.05,
					20,
				).Transform(
					meshops.SmoothNormalsTransformer{},
				),
				Material: &gltf.PolyformMaterial{
					Name: "Spikes",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{220, 220, 200, 255},
						MetallicFactor:  1,
						MetallicRoughnessTexture: &gltf.PolyformTexture{
							URI: mrTex,
						},
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}
