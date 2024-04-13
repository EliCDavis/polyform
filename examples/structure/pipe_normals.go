package main

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type PipeNormalsNode = nodes.StructNode[image.Image, PipeNormalsNodeData]

type PipeNormalsNodeData struct {
	LineCount nodes.NodeOutput[int]
	LineWidth nodes.NodeOutput[float64]

	BoltCount  nodes.NodeOutput[int]
	BoltRadius nodes.NodeOutput[float64]

	BlurIterations nodes.NodeOutput[int]
}

func (pnn PipeNormalsNodeData) Process() (image.Image, error) {
	dim := 256
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(256, 1/64., 3)

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

	img = texturing.ToNormal(img)

	bolts := pnn.BoltCount.Value()
	boltRadius := pnn.BoltRadius.Value()
	boltInc := float64(dim) / float64(bolts)
	halfBoltInc := boltInc / 2

	lines := pnn.LineCount.Value()
	lineWidth := pnn.LineWidth.Value()
	lineInc := float64(dim) / float64(lines)
	halfLineInc := lineInc / 2
	for lineIndex := 0; lineIndex < lines; lineIndex++ {
		l := (lineInc * float64(lineIndex)) + halfLineInc
		normals.Line{
			Start:           vector2.New(0., l),
			End:             vector2.New(float64(dim), l),
			Width:           lineWidth,
			NormalDirection: normals.Additive,
		}.Round(img)

		for boltIndex := 0; boltIndex < bolts; boltIndex++ {
			b := (boltInc * float64(boltIndex)) + halfBoltInc

			normals.Sphere{
				Center:    vector2.New(b, l-lineWidth-boltRadius-(boltRadius)),
				Radius:    boltRadius,
				Direction: normals.Additive,
			}.Draw(img)
		}
	}

	return texturing.BoxBlurNTimes(img, pnn.BlurIterations.Value()), nil
}
