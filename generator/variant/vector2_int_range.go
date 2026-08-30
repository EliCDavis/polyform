package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
)

// Vector2IntRange lerps between two whole-number vector2.Int values as a
// whole, sampled Samples times.
type Vector2IntRange struct {
	path    string
	Min     vector2.Int
	Max     vector2.Int
	Samples int
}

func NewVector2IntRange(path string, minX, maxX, minY, maxY, samples int) Vector2IntRange {
	return Vector2IntRange{
		path:    path,
		Min:     vector2.New(minX, minY),
		Max:     vector2.New(maxX, maxY),
		Samples: samples,
	}
}

func (r Vector2IntRange) Path() string { return r.path }
func (r Vector2IntRange) Count() int   { return axisCount(r.Samples) }

func (r Vector2IntRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(vector2.New(
		intAxisValue(r.Min.X(), r.Max.X(), r.Samples, index),
		intAxisValue(r.Min.Y(), r.Max.Y(), r.Samples, index),
	))
}

func (r Vector2IntRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	return json.Marshal(vector2.New(
		intLerp(r.Min.X(), r.Max.X(), t),
		intLerp(r.Min.Y(), r.Max.Y(), t),
	))
}

func (r Vector2IntRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector2IntRange, vector2IntRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
