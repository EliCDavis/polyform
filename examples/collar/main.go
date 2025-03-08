package main

import (
	"image"
	"image/color"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
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

type ConeNode = nodes.Struct[ConeNodeData]

type ConeNodeData struct {
	Height nodes.Output[float64]
	Radius nodes.Output[float64]
	Sides  nodes.Output[int]
}

func (r ConeNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	sides := r.Sides.Value()
	if sides < 3 {
		panic("can not make cone with less that 3 sides")
	}

	radius := r.Radius.Value()
	height := r.Height.Value()

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

type CollarNode = nodes.Struct[CollarNodeData]

type CollarNodeData struct {
	Height     nodes.Output[float64]
	Thickness  nodes.Output[float64]
	Resolution nodes.Output[int]
	Radius     nodes.Output[float64]
}

func (cn CollarNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	collarHeight := cn.Height.Value()
	collarThickness := cn.Thickness.Value()
	collarRadius := cn.Radius.Value()
	collarResolution := repeat.CirclePoints(cn.Resolution.Value(), collarRadius)

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

type GlbArtifactNode = nodes.Struct[GlbArtifactNodeData]

type GlbArtifactNodeData struct {
	Collar         nodes.Output[modeling.Mesh]
	Spike          nodes.Output[modeling.Mesh]
	SpikePositions nodes.Output[[]trs.TRS]
	SpikeColor     nodes.Output[coloring.WebColor]
}

func (gan GlbArtifactNodeData) Out() nodes.StructOutput[artifact.Artifact] {
	collar := gan.Collar.Value()
	spikes := gan.Spike.Value()
	scene := gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "Collar",
				Mesh: &collar,
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
				Name:         "Spikes",
				Mesh:         &spikes,
				GpuInstances: gan.SpikePositions.Value(),
				Material: &gltf.PolyformMaterial{
					Name: "Spikes",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: gan.SpikeColor.Value(),
						MetallicRoughnessTexture: &gltf.PolyformTexture{
							URI: mrTexturePath,
						},
					},
				},
			},
		},
	}

	return gltf.Artifact{
		Scene: scene,
	}, nil
}

type CollarAlbedoTextureNode = nodes.Struct[CollarAlbedoTextureNodeData]

type CollarAlbedoTextureNodeData struct {
	BaseColor   nodes.Output[coloring.WebColor]
	StitchColor nodes.Output[coloring.WebColor]
}

func (catn CollarAlbedoTextureNodeData) Out() nodes.StructOutput[image.Image] {
	ctx := gg.NewContext(256, 256)
	ctx.SetColor(catn.BaseColor.Value())
	ctx.DrawRectangle(0, 0, 256, 256)
	ctx.Fill()

	ctx.SetColor(catn.StitchColor.Value())
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
	collarRadius := &parameter.Float64{Name: "Collar/Radius", DefaultValue: 1}

	spike := &meshops.SmoothNormalsNode{
		Data: meshops.SmoothNormalsNodeData{
			Mesh: &ConeNode{
				Data: ConeNodeData{
					Height: &parameter.Float64{Name: "Spike/Height", DefaultValue: .2},
					Radius: &parameter.Float64{Name: "Spike/Radius", DefaultValue: .1},
					Sides:  &parameter.Int{Name: "Spike/Resolution", DefaultValue: 30},
				},
			},
		},
	}

	collar := &CollarNode{
		Data: CollarNodeData{
			Height:     &parameter.Float64{Name: "Collar/Height", DefaultValue: .2},
			Thickness:  &parameter.Float64{Name: "Collar/Thickness", DefaultValue: .1},
			Resolution: &parameter.Int{Name: "Collar/Resolution", DefaultValue: 30},
			Radius:     collarRadius,
		},
	}

	gltfNode := &GlbArtifactNode{
		Data: GlbArtifactNodeData{
			Collar: collar.Out(),
			Spike:  spike,
			SpikePositions: &repeat.CircleNode{
				Data: repeat.CircleNodeData{
					Radius: collarRadius,
					Times:  &parameter.Int{Name: "Spike/Count", DefaultValue: 20},
				},
			},
			SpikeColor: &parameter.Color{
				Name:         "Spike/Color",
				DefaultValue: coloring.WebColor{244, 244, 244, 255},
			},
		},
	}

	app := generator.App{
		Name:        "Spiked Collar Demo",
		Version:     "0.0.1",
		Description: "Small demo that let's you edit a spiked collar",
		Authors: []schema.Author{
			{
				Name: "Eli Davis",
				ContactInfo: []schema.AuthorContact{
					{
						Medium: "twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		Files: map[string]nodes.Output[artifact.Artifact]{
			"mesh.glb":    gltfNode.Out(),
			mrTexturePath: basics.NewImageNode(nodes.FuncValue(mrTexture)),
			collarAlbedoPath: basics.NewImageNode(&CollarAlbedoTextureNode{
				Data: CollarAlbedoTextureNodeData{
					BaseColor: &parameter.Color{
						Name:         "Collar/Base Color",
						DefaultValue: coloring.WebColor{46, 46, 46, 255},
					},
					StitchColor: &parameter.Color{
						Name:         "Collar/Stitch Color",
						DefaultValue: coloring.WebColor{10, 10, 10, 255},
					},
				},
			}),
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
