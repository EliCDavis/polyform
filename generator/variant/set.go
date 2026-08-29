package variant

import (
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/generator/variable"
)

// Set is a collection of Dimensions to sweep or sample combinations from.
type Set struct {
	Dimensions []Dimension
}

// TotalCombinations is the product of every Dimension's Count().
func (s Set) TotalCombinations() int {
	total := 1
	for _, d := range s.Dimensions {
		total *= d.Count()
	}
	return total
}

// Sweep produces every combination of every Dimension's values.
func (s Set) Sweep() ([]variable.Profile, error) {
	total := s.TotalCombinations()
	profiles := make([]variable.Profile, total)

	for c := range total {
		profile := make(variable.Profile, len(s.Dimensions))
		remaining := c
		for _, d := range s.Dimensions {
			count := d.Count()
			index := remaining % count
			remaining /= count

			val, err := d.Value(index)
			if err != nil {
				return nil, fmt.Errorf("dimension %q: %w", d.Path(), err)
			}
			profile[d.Path()] = val
		}
		profiles[c] = profile
	}

	return profiles, nil
}

// Sample produces n combinations, each dimension drawn at random.
func (s Set) Sample(n int, rng *rand.Rand) ([]variable.Profile, error) {
	if n < 0 {
		return nil, fmt.Errorf("sample count must be non-negative, got %d", n)
	}
	if rng == nil {
		return nil, fmt.Errorf("rng must not be nil")
	}

	profiles := make([]variable.Profile, n)
	for i := range n {
		profile := make(variable.Profile, len(s.Dimensions))
		for _, d := range s.Dimensions {
			val, err := d.Random(rng)
			if err != nil {
				return nil, fmt.Errorf("dimension %q: %w", d.Path(), err)
			}
			profile[d.Path()] = val
		}
		profiles[i] = profile
	}

	return profiles, nil
}
