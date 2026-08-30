package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/vector/vector2"
)

// Vector2Range combines an independent range per axis into one vector2.Float64,
// sampled the same number of times per axis.
type Vector2Range struct {
	path    string
	Min     vector2.Float64
	Max     vector2.Float64
	Samples int
}

func NewVector2Range(path string, minX, maxX, minY, maxY float64, samples int) Vector2Range {
	return Vector2Range{
		path:    path,
		Min:     vector2.New(minX, minY),
		Max:     vector2.New(maxX, maxY),
		Samples: samples,
	}
}

func (r Vector2Range) Path() string { return r.path }
func (r Vector2Range) Count() int   { return axisCount(r.Samples) * axisCount(r.Samples) }

func (r Vector2Range) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	xCount := axisCount(r.Samples)
	ix := index % xCount
	iy := index / xCount
	return json.Marshal(vector2.New(
		axisValue(r.Min.X(), r.Max.X(), r.Samples, ix),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples, iy),
	))
}

func (r Vector2Range) Random(rng *rand.Rand) (json.RawMessage, error) {
	v := chance.NewRange2D(r.Min, r.Max, rng).Value()
	return json.Marshal(v)
}

func (r Vector2Range) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector2Range, vector2RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
