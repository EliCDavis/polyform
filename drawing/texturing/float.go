package texturing

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
)

type OneMinusNode struct {
	Texture nodes.Output[Texture[float64]]
}

func (n OneMinusNode) Result(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}

	tex := nodes.GetOutputValue(out, n.Texture)
	result := NewTexture[float64](tex.width, tex.height)

	for y := range tex.height {
		for x := range tex.width {
			result.Set(x, y, 1-tex.Get(x, y))
		}
	}

	out.Set(result)
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
