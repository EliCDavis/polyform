package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/generator/persistence"
)

// RGBRange lerps between two colors as a whole, sampled Samples times.
type RGBRange struct {
	path    string
	Min     persistence.WebColor
	Max     persistence.WebColor
	Samples int
}

func NewRGBRange(path string, min, max persistence.WebColor, samples int) RGBRange {
	return RGBRange{
		path:    path,
		Min:     min,
		Max:     max,
		Samples: samples,
	}
}

func (r RGBRange) Path() string { return r.path }
func (r RGBRange) Count() int   { return axisCount(r.Samples) }

func (r RGBRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return hexColor(
		axisValue(byteToUnit(r.Min.R), byteToUnit(r.Max.R), r.Samples, index),
		axisValue(byteToUnit(r.Min.G), byteToUnit(r.Max.G), r.Samples, index),
		axisValue(byteToUnit(r.Min.B), byteToUnit(r.Max.B), r.Samples, index),
	), nil
}

func (r RGBRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	return hexColor(
		lerp(byteToUnit(r.Min.R), byteToUnit(r.Max.R), t),
		lerp(byteToUnit(r.Min.G), byteToUnit(r.Max.G), t),
		lerp(byteToUnit(r.Min.B), byteToUnit(r.Max.B), t),
	), nil
}

func (r RGBRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeRGBRange, rgbRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
