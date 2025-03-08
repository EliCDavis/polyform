package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
)

type PerlinNoiseNormalsNode = nodes.Struct[PerlinNoiseNormalsNodeData]

type PerlinNoiseNormalsNodeData struct {
	Octaves        nodes.Output[int]
	BlurIterations nodes.Output[int]
}

func (pnn PerlinNoiseNormalsNodeData) Out() nodes.StructOutput[image.Image] {
	dim := 256
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))

	octaves := 3
	if pnn.Octaves != nil {
		octaves = pnn.Octaves.Value()
	}

	n := noise.NewTilingNoise(256, 1/64., octaves)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 128) + 128

			img.Set(x, y, color.RGBA{
				R: byte(p), // byte(len * 255),
				G: byte(p),
				B: byte(p),
				A: 255,
			})
		}
	}

	blurIterations := 5
	if pnn.BlurIterations != nil {
		blurIterations = pnn.BlurIterations.Value()
	}

	return texturing.BoxBlurNTimes(texturing.ToNormal(img), blurIterations), nil
}
