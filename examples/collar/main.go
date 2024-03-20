package main

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

const mrTexturePath = "collar_mr.png"
const collarAlbedoPath = "collar_albedo.png"

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

type ConeNode struct {
	nodes.StructData[modeling.Mesh]

	Height nodes.NodeOutput[float64]
	Radius nodes.NodeOutput[float64]
	Sides  nodes.NodeOutput[int]
}

func (r ConeNode) Process() (modeling.Mesh, error) {
	sides := r.Sides.Data()
	if sides < 3 {
		panic("can not make cone with less that 3 sides")
	}

	radius := r.Radius.Data()
	height := r.Height.Data()

	verts := repeat.CirclePoints(sides, radius)
	lastVert := len(verts)
	verts = append(verts, vector3.New(0., height, 0.))
	uvs := make([]vector2.Float64, len(verts))
	uvs[len(uvs)-1] = vector2.One[float64]()

	tris := make([]int, 0, sides*3)
	for i := 0; i < sides; i++ {
		tris = append(tris, i, lastVert, i+1)
	}
	tris[len(tris)-1] = 0

	return modeling.NewMesh(modeling.TriangleTopology, tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs).
		Rotate(quaternion.FromTheta(math.Pi/2, vector3.Right[float64]())), nil
}

func (r *ConeNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: r}
}

type CollarNode struct {
	nodes.StructData[modeling.Mesh]

	Height     nodes.NodeOutput[float64]
	Thickness  nodes.NodeOutput[float64]
	Resolution nodes.NodeOutput[int]
	Radius     nodes.NodeOutput[float64]
}

func (cn CollarNode) Process() (modeling.Mesh, error) {
	collarHeight := cn.Height.Data()
	collarThickness := cn.Thickness.Data()
	collarRadius := cn.Radius.Data()
	collarResolution := repeat.CirclePoints(cn.Resolution.Data(), collarRadius)

	collarUVs := make([]vector2.Float64, 0)
	return extrude.ClosedShape(SquarePoints(collarThickness, collarHeight), collarResolution).
		Transform(
			meshops.SmoothNormalsTransformer{},
		).
		ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
			xy := v.XZ().Normalized()
			angle := math.Atan2(xy.Y(), xy.X()) * 4
			height := (v.Y() + (collarHeight / 2)) / collarHeight
			collarUVs = append(collarUVs, vector2.New(angle, height))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, collarUVs), nil
}

func (cn *CollarNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: cn}
}

type GlbArtifactNode struct {
	nodes.StructData[generator.Artifact]

	Collar     nodes.NodeOutput[modeling.Mesh]
	Spikes     nodes.NodeOutput[modeling.Mesh]
	SpikeColor nodes.NodeOutput[coloring.WebColor]
}

func (gan *GlbArtifactNode) Out() nodes.NodeOutput[generator.Artifact] {
	return &nodes.StructNodeOutput[generator.Artifact]{Definition: gan}
}

func (gan GlbArtifactNode) Process() (generator.Artifact, error) {
	scene := gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "Collar",
				Mesh: gan.Collar.Data(),
				Material: &gltf.PolyformMaterial{
					Name: "Collar",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorTexture: &gltf.PolyformTexture{
							URI: collarAlbedoPath,
						},
					},
				},
			},
			{
				Name: "Spikes",
				Mesh: gan.Spikes.Data(),
				Material: &gltf.PolyformMaterial{
					Name: "Spikes",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: gan.SpikeColor.Data(),
						MetallicFactor:  1,
						MetallicRoughnessTexture: &gltf.PolyformTexture{
							URI: mrTexturePath,
						},
					},
				},
			},
		},
	}

	return generator.GltfArtifact{
		Scene: scene,
	}, nil
}

type CollarAlbedoTextureNode struct {
	nodes.StructData[image.Image]

	BaseColor   nodes.NodeOutput[coloring.WebColor]
	StitchColor nodes.NodeOutput[coloring.WebColor]
}

func (catn *CollarAlbedoTextureNode) Out() nodes.NodeOutput[image.Image] {
	return &nodes.StructNodeOutput[image.Image]{Definition: catn}
}

func (catn CollarAlbedoTextureNode) Process() (image.Image, error) {
	ctx := gg.NewContext(256, 256)
	ctx.SetColor(catn.BaseColor.Data())
	ctx.DrawRectangle(0, 0, 256, 256)
	ctx.Fill()

	ctx.SetColor(catn.StitchColor.Data())
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

	return ctx.Image(), nil
}

func mrTexture() image.Image {
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

	return ctx.Image()
}

func main() {

	collarRadius := &generator.ParameterNode[float64]{Name: "Collar/Radius", DefaultValue: 1}

	spikeRing := repeat.CircleNode{
		Mesh: (&meshops.SmoothNormalsNode{
			Mesh: (&ConeNode{
				Height: &generator.ParameterNode[float64]{Name: "Spike/Height", DefaultValue: .2},
				Radius: &generator.ParameterNode[float64]{Name: "Spike/Radius", DefaultValue: .1},
				Sides:  &generator.ParameterNode[int]{Name: "Spike/Resolution", DefaultValue: 30},
			}).Out(),
		}).SmoothedMesh(),
		Radius: collarRadius,
		Times:  &generator.ParameterNode[int]{Name: "Spike/Count", DefaultValue: 20},
	}

	collar := CollarNode{
		Height:     &generator.ParameterNode[float64]{Name: "Collar/Height", DefaultValue: .2},
		Thickness:  &generator.ParameterNode[float64]{Name: "Collar/Thickness", DefaultValue: .1},
		Resolution: &generator.ParameterNode[int]{Name: "Collar/Resolution", DefaultValue: 30},
		Radius:     collarRadius,
	}

	gltfNode := GlbArtifactNode{
		Collar: collar.Out(),
		Spikes: spikeRing.Out(),
		SpikeColor: &generator.ParameterNode[coloring.WebColor]{
			Name:         "Spike/Color",
			DefaultValue: coloring.WebColor{244, 244, 244, 255},
		},
	}

	app := generator.App{
		Name:        "Spiked Collar Demo",
		Version:     "0.0.1",
		Description: "Small demo that let's you edit a spiked collar",
		Authors: []generator.Author{
			{
				Name: "Eli Davis",
				ContactInfo: []generator.AuthorContact{
					{
						Medium: "twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.glb":    gltfNode.Out(),
			mrTexturePath: generator.NewImageArtifactNode(nodes.FuncValue(mrTexture)),
			collarAlbedoPath: generator.NewImageArtifactNode((&CollarAlbedoTextureNode{
				BaseColor: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Collar/Base Color",
					DefaultValue: coloring.WebColor{46, 46, 46, 255},
				},
				StitchColor: &generator.ParameterNode[coloring.WebColor]{
					Name:         "Collar/Stitch Color",
					DefaultValue: coloring.WebColor{10, 10, 10, 255},
				},
			}).Out()),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}

}
