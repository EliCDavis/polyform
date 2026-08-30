package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// NumericRange evenly spaces Samples values between Min and Max.
type NumericRange struct {
	path string
	axisRange
}

func NewNumericRange(path string, min, max float64, samples int) NumericRange {
	return NumericRange{path: path, axisRange: axisRange{Min: min, Max: max, Samples: samples}}
}

func (r NumericRange) Path() string { return r.path }
func (r NumericRange) Count() int   { return r.count() }

func (r NumericRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(r.value(index))
}

func (r NumericRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	return json.Marshal(r.random(rng))
}

func (r NumericRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeNumericRange, r.axisRange)
}
