package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type MetallicRoughnessNode = nodes.Struct[MetallicRoughnessNodeData]

type MetallicRoughnessNodeData struct {
	Octaves          nodes.Output[int]
	MinimumRoughness nodes.Output[float64]
	MaximumRoughness nodes.Output[float64]
}

func (pnn MetallicRoughnessNodeData) Out() nodes.StructOutput[artifact.Artifact] {
	dim := 256
	n := noise.NewTilingNoise(dim, 1/64., pnn.Octaves.Value())
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))

	minimumRoughness := byte(pnn.MinimumRoughness.Value() * 255)
	roughnessDelta := (pnn.MaximumRoughness.Value() * 255) - float64(minimumRoughness)

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

	return basics.Image{
		Image: img,
	}, nil
}
