package texturing

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
)

func OneMinus(in Texture[float64]) Texture[float64] {
	result := NewTexture[float64](in.width, in.height)
	for y := range in.height {
		for x := range in.width {
			result.Set(x, y, 1-in.Get(x, y))
		}
	}
	return result
}

type OneMinusNode struct {
	Texture nodes.Output[Texture[float64]]
}

func (n OneMinusNode) Result(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	out.Set(OneMinus(nodes.GetOutputValue(out, n.Texture)))
}

type MultiplyFloat1Node struct {
	Textures []nodes.Output[Texture[float64]]
}

func (n MultiplyFloat1Node) Result(out *nodes.StructOutput[Texture[float64]]) {
	textures := nodes.GetOutputValues(out, n.Textures)
	if len(textures) == 0 {
		return
	}

	if len(textures) == 1 {
		out.Set(textures[0])
		return
	}

	if !resolutionsMatch(textures) {
		out.CaptureError(fmt.Errorf("mismatch texture resolution"))
		return
	}

	result := NewTexture[float64](textures[0].Width(), textures[0].Height())
	for y := range result.Height() {
		for x := range result.Width() {
			v := textures[0].Get(x, y)
			for i := 1; i < len(textures); i++ {
				v *= textures[i].Get(x, y)
			}
			result.Set(x, y, v)
		}
	}

	out.Set(result)
}
