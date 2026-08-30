package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// Combination aggregates several Dimensions for the same variable.
// Count is their sum, not their product.
type Combination struct {
	path       string
	Dimensions []Dimension
}

func NewCombination(path string, dimensions ...Dimension) Combination {
	return Combination{path: path, Dimensions: dimensions}
}

func (c Combination) Path() string { return c.path }

func (c Combination) Count() int {
	total := 0
	for _, d := range c.Dimensions {
		total += d.Count()
	}
	return total
}

func (c Combination) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= c.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, c.Count())
	}
	for _, d := range c.Dimensions {
		if index < d.Count() {
			return d.Value(index)
		}
		index -= d.Count()
	}
	return nil, fmt.Errorf("index out of range")
}

func (c Combination) Random(rng *rand.Rand) (json.RawMessage, error) {
	total := c.Count()
	if total == 0 {
		return nil, fmt.Errorf("no values to choose from")
	}
	return c.Value(rng.Intn(total))
}

func (c Combination) MarshalJSON() ([]byte, error) {
	encoded := make([]json.RawMessage, len(c.Dimensions))
	for i, d := range c.Dimensions {
		raw, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		encoded[i] = raw
	}
	return marshalDimension(typeCombination, combinationJSON{Dimensions: encoded})
}
