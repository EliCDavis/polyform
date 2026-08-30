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

// HSVRange combines a range per Hue, Saturation, Value channel into one hex
// color, sampled the same number of times per channel.
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
func (r HSVRange) Count() int {
	return axisCount(r.Samples) * axisCount(r.Samples) * axisCount(r.Samples)
}

func (r HSVRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	hCount := axisCount(r.Samples)
	sCount := axisCount(r.Samples)
	ih := index % hCount
	is := (index / hCount) % sCount
	iv := index / (hCount * sCount)
	red, green, blue := hsvToRGB(
		axisValue(r.Min.H, r.Max.H, r.Samples, ih),
		axisValue(r.Min.S, r.Max.S, r.Samples, is),
		axisValue(r.Min.V, r.Max.V, r.Samples, iv),
	)
	return hexColor(red, green, blue), nil
}

func (r HSVRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	red, green, blue := hsvToRGB(
		randomBetween(r.Min.H, r.Max.H, rng),
		randomBetween(r.Min.S, r.Max.S, rng),
		randomBetween(r.Min.V, r.Max.V, rng),
	)
	return hexColor(red, green, blue), nil
}

func (r HSVRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeHSVRange, hsvRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
