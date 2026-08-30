package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// IntRange evenly spaces Samples whole-number values between Min and Max.
type IntRange struct {
	path    string
	Min     int
	Max     int
	Samples int
}

func NewIntRange(path string, min, max, samples int) IntRange {
	return IntRange{path: path, Min: min, Max: max, Samples: samples}
}

func (r IntRange) Path() string { return r.path }
func (r IntRange) Count() int   { return axisCount(r.Samples) }

func (r IntRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(intAxisValue(r.Min, r.Max, r.Samples, index))
}

func (r IntRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	return json.Marshal(intLerp(r.Min, r.Max, rng.Float64()))
}

func (r IntRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeIntRange, intRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
