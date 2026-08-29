package variant

import (
	"encoding/json"
	"math/rand"
)

// Dimension is one variable's range of possible values.
type Dimension interface {
	Path() string
	Count() int
	Value(index int) (json.RawMessage, error)
	Random(rng *rand.Rand) (json.RawMessage, error)
}
