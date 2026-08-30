package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// HSVChannels is a Hue (degrees), Saturation (0-1), Value (0-1) triple.
type HSVChannels struct {
	H float64 `json:"h"`
	S float64 `json:"s"`
	V float64 `json:"v"`
}

// HSVRange lerps between two Hue, Saturation, Value triples as a whole into
// one hex color, sampled Samples times.
type HSVRange struct {
	path    string
	Min     HSVChannels
	Max     HSVChannels
	Samples int
}

func NewHSVRange(path string, minH, maxH, minS, maxS, minV, maxV float64, samples int) HSVRange {
	return HSVRange{
		path:    path,
		Min:     HSVChannels{H: minH, S: minS, V: minV},
		Max:     HSVChannels{H: maxH, S: maxS, V: maxV},
		Samples: samples,
	}
}

func (r HSVRange) Path() string { return r.path }
func (r HSVRange) Count() int   { return axisCount(r.Samples) }

func (r HSVRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	red, green, blue := hsvToRGB(
		axisValue(r.Min.H, r.Max.H, r.Samples, index),
		axisValue(r.Min.S, r.Max.S, r.Samples, index),
		axisValue(r.Min.V, r.Max.V, r.Samples, index),
	)
	return hexColor(red, green, blue), nil
}

func (r HSVRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	red, green, blue := hsvToRGB(
		lerp(r.Min.H, r.Max.H, t),
		lerp(r.Min.S, r.Max.S, t),
		lerp(r.Min.V, r.Max.V, t),
	)
	return hexColor(red, green, blue), nil
}

func (r HSVRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeHSVRange, hsvRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
