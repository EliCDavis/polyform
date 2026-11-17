package mcp

import (
	"fmt"
	"math"
	"sort"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// TextureStats is what a texture reads as in numbers. A texture has no
// useful JSON form - its fields are unexported, so it marshalled to {} -
// and every question asked of one during a build ("did the camo actually
// cover anything?") is answered by its distribution rather than a picture.
type TextureStats struct {
	Width   int            `json:"width"`
	Height  int            `json:"height"`
	Pixels  int            `json:"pixels"`
	Channel []ChannelStats `json:"channels" jsonschema:"one entry per component; a float texture has one, a color has r/g/b"`
}

type ChannelStats struct {
	Name      string  `json:"name"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Mean      float64 `json:"mean"`
	Median    float64 `json:"median"`
	Histogram []int   `json:"histogram" jsonschema:"10 equal buckets spanning min..max, so a pattern that covers little of the texture shows as a spike at one end"`
}

func channelStats(name string, values []float64) ChannelStats {
	stats := ChannelStats{Name: name, Histogram: make([]int, 10)}
	if len(values) == 0 {
		return stats
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	stats.Min, stats.Max = sorted[0], sorted[len(sorted)-1]
	stats.Median = sorted[len(sorted)/2]

	total := 0.
	for _, v := range values {
		total += v
	}
	stats.Mean = total / float64(len(values))

	span := stats.Max - stats.Min
	if span == 0 {
		stats.Histogram[0] = len(values)
		return stats
	}
	for _, v := range values {
		bucket := int(((v - stats.Min) / span) * float64(len(stats.Histogram)))
		bucket = min(max(bucket, 0), len(stats.Histogram)-1)
		stats.Histogram[bucket]++
	}
	return stats
}

// textureStats reports on the texture instantiations a graph can actually
// produce, returning nil for anything that is not one.
func textureStats(value any) *TextureStats {
	switch t := value.(type) {
	case texturing.Texture[float64]:
		return gather(t.Width(), t.Height(), map[string][]float64{
			"value": scanTexture(t, func(v float64) []float64 { return []float64{v} })[0],
		}, []string{"value"})

	case texturing.Texture[coloring.Color]:
		cols := scanTexture(t, func(c coloring.Color) []float64 {
			return []float64{float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255, float64(c.A) / 255}
		})
		return gather(t.Width(), t.Height(), map[string][]float64{
			"r": cols[0], "g": cols[1], "b": cols[2], "a": cols[3],
		}, []string{"r", "g", "b", "a"})

	case texturing.Texture[vector2.Float64]:
		cols := scanTexture(t, func(v vector2.Float64) []float64 { return []float64{v.X(), v.Y()} })
		return gather(t.Width(), t.Height(), map[string][]float64{"x": cols[0], "y": cols[1]}, []string{"x", "y"})

	case texturing.Texture[vector3.Float64]:
		cols := scanTexture(t, func(v vector3.Float64) []float64 { return []float64{v.X(), v.Y(), v.Z()} })
		return gather(t.Width(), t.Height(), map[string][]float64{
			"x": cols[0], "y": cols[1], "z": cols[2],
		}, []string{"x", "y", "z"})
	}
	return nil
}

func scanTexture[T any](t texturing.Texture[T], split func(T) []float64) [][]float64 {
	var columns [][]float64
	for y := range t.Height() {
		for x := range t.Width() {
			parts := split(t.Get(x, y))
			if columns == nil {
				columns = make([][]float64, len(parts))
				for i := range columns {
					columns[i] = make([]float64, 0, t.Pixels())
				}
			}
			for i, p := range parts {
				columns[i] = append(columns[i], p)
			}
		}
	}
	if columns == nil {
		columns = [][]float64{{}}
	}
	return columns
}

func gather(width, height int, columns map[string][]float64, order []string) *TextureStats {
	stats := &TextureStats{Width: width, Height: height, Pixels: width * height}
	for _, name := range order {
		stats.Channel = append(stats.Channel, channelStats(name, columns[name]))
	}
	return stats
}

func (t TextureStats) summary() string {
	if len(t.Channel) == 0 {
		return fmt.Sprintf("an empty %dx%d texture", t.Width, t.Height)
	}
	first := t.Channel[0]
	if math.IsNaN(first.Mean) {
		return fmt.Sprintf("a %dx%d texture carrying NaN", t.Width, t.Height)
	}
	return fmt.Sprintf("a %dx%d texture; %s runs %.4g to %.4g, mean %.4g",
		t.Width, t.Height, first.Name, first.Min, first.Max, first.Mean)
}
