package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// Discrete is a list of specific values to choose from.
type Discrete struct {
	path   string
	Values []json.RawMessage
}

func NewDiscrete(path string, values ...json.RawMessage) Discrete {
	return Discrete{path: path, Values: values}
}

func (d Discrete) Path() string { return d.path }
func (d Discrete) Count() int   { return len(d.Values) }

func (d Discrete) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= len(d.Values) {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, len(d.Values))
	}
	return d.Values[index], nil
}

func (d Discrete) Random(rng *rand.Rand) (json.RawMessage, error) {
	if len(d.Values) == 0 {
		return nil, fmt.Errorf("no values to choose from")
	}
	return d.Values[rng.Intn(len(d.Values))], nil
}

func (d Discrete) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeDiscrete, discreteJSON{Values: d.Values})
}
