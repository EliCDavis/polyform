package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
)

// Vector2Range lerps between Min and Max as a whole, sampled Samples times.
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
func (r Vector2Range) Count() int   { return axisCount(r.Samples) }

func (r Vector2Range) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(vector2.New(
		axisValue(r.Min.X(), r.Max.X(), r.Samples, index),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples, index),
	))
}

func (r Vector2Range) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	return json.Marshal(vector2.New(
		lerp(r.Min.X(), r.Max.X(), t),
		lerp(r.Min.Y(), r.Max.Y(), t),
	))
}

func (r Vector2Range) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector2Range, vector2RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
