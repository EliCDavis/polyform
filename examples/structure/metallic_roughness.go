package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type MetallicRoughnessNode struct {
	nodes.StructData[generator.Artifact]

	Octaves          nodes.NodeOutput[int]
	MinimumRoughness nodes.NodeOutput[float64]
	MaximumRoughness nodes.NodeOutput[float64]
}

func (pnn *MetallicRoughnessNode) Out() nodes.NodeOutput[generator.Artifact] {
	return nodes.StructNodeOutput[generator.Artifact]{Definition: pnn}
}

func (pnn MetallicRoughnessNode) Process() (generator.Artifact, error) {
	dim := 256
	n := noise.NewTilingNoise(dim, 1/64., pnn.Octaves.Data())
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))

	minimumRoughness := byte(pnn.MinimumRoughness.Data() * 255)
	roughnessDelta := (pnn.MaximumRoughness.Data() * 255) - float64(minimumRoughness)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			p := roughnessDelta * n.Noise(x, y)

			img.Set(x, y, color.RGBA{
				R: 0,
				G: minimumRoughness + byte(p), //roughness (0-smooth, 1-rough)
				B: 255,                        //metallness
				A: 255,
			})
		}
	}

	return generator.ImageArtifact{
		Image: img,
	}, nil
}
