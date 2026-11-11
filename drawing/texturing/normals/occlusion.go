package normals

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func wrap(i, n int) int {
	return ((i % n) + n) % n
}

// Neighbors wrap around the edges, since a height map built for a tiling texture has none.
func OcclusionFromHeightmap(heightmap HeightMap, radius int, strength float64) texturing.Texture[float64] {
	width, height := heightmap.Width(), heightmap.Height()
	occlusion := texturing.Empty[float64](width, height)
	if radius < 1 || width == 0 || height == 0 {
		occlusion.Fill(1)
		return occlusion
	}

	offsets := make([]vector2.Int, 0, radius*8)
	for r := 1; r <= radius; r++ {
		for d := range 8 {
			angle := (float64(d) + 0.5*float64(r%2)) * math.Pi / 4
			sin, cos := math.Sincos(angle)
			offsets = append(offsets, vector2.New(cos, sin).Scale(float64(r)).RoundToInt())
		}
	}

	occlusion.MutateParallel(func(x, y int, _ float64) float64 {
		h := heightmap.Get(x, y)
		above := 0.
		for _, o := range offsets {
			above += math.Max(heightmap.Get(wrap(x+o.X(), width), wrap(y+o.Y(), height))-h, 0)
		}
		return math.Min(math.Max(1-strength*above/float64(len(offsets)), 0), 1)
	})
	return occlusion
}

type OcclusionFromHeightMapNode struct {
	In       nodes.Output[HeightMap]
	Radius   nodes.Output[int]     `description:"How far around each pixel, in pixels, to look for higher ground." default:"4"`
	Strength nodes.Output[float64] `description:"How dark a crevice gets for how much higher its surroundings are." default:"2"`
}

func (n OcclusionFromHeightMapNode) Description() string {
	return "Darkens the low points of a height map that higher surroundings shade: 1 in the open, toward 0 in crevices. Feeds a material's occlusion texture, or darkens a color texture."
}

func (n OcclusionFromHeightMapNode) occlusion(out nodes.ExecutionRecorder) texturing.Texture[float64] {
	return OcclusionFromHeightmap(
		nodes.GetOutputValue(out, n.In),
		nodes.GetOutputValue(out, n.Radius),
		nodes.GetOutputValue(out, n.Strength),
	)
}

func (n OcclusionFromHeightMapNode) Occlusion(out *nodes.StructOutput[texturing.Texture[float64]]) {
	if n.In == nil {
		return
	}
	out.Set(n.occlusion(out))
}

func (n OcclusionFromHeightMapNode) OcclusionImage(out *nodes.StructOutput[image.Image]) {
	if n.In == nil {
		return
	}
	out.Set(n.occlusion(out).ToImage(func(v float64) color.Color {
		g := uint8(math.Round(v * 255))
		return color.RGBA{R: g, G: g, B: g, A: 255}
	}))
}
