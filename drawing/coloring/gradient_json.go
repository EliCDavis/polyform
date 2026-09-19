package coloring

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type jsonGradientKey[T any] struct {
	Time  float64 `json:"time"`
	Value T       `json:"value"`
}

type jsonGradient[T any] struct {
	Keys []jsonGradientKey[T] `json:"keys"`
}

// Keys are the normalized ones, so times come back in 0 to 1 whatever they
// were built from.
func (g Gradient[T]) Keys() []GradientKey[T] {
	keys := make([]GradientKey[T], len(g.keys))
	for i, k := range g.keys {
		keys[i] = GradientKey[T]{Time: k.Time, Value: k.Color}
	}
	return keys
}

func (g Gradient[T]) MarshalJSON() ([]byte, error) {
	keys := make([]jsonGradientKey[T], len(g.keys))
	for i, k := range g.keys {
		keys[i] = jsonGradientKey[T]{Time: k.Time, Value: k.Color}
	}
	return json.Marshal(jsonGradient[T]{Keys: keys})
}

func (g *Gradient[T]) UnmarshalJSON(data []byte) error {
	space, err := spaceFor[T]()
	if err != nil {
		return err
	}
	var parsed jsonGradient[T]
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	keys := make([]GradientKey[T], len(parsed.Keys))
	for i, k := range parsed.Keys {
		keys[i] = GradientKey[T]{Time: k.Time, Value: k.Value}
	}
	*g = NewGradient(space, keys...)
	return nil
}

func spaceFor[T any]() (vector.Space[T], error) {
	var zero T
	var space any
	switch any(zero).(type) {
	case float64:
		space = vector1.Space[float64]{}
	case vector2.Float64:
		space = vector2.Space[float64]{}
	case vector3.Float64:
		space = vector3.Space[float64]{}
	case vector4.Float64:
		space = vector4.Space[float64]{}
	case Color:
		space = Space{}
	default:
		return nil, fmt.Errorf("no gradient space for %T", zero)
	}
	return space.(vector.Space[T]), nil
}
