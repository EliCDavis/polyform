package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// RGBRange spans two hex colors, sampled the same number of times per channel.
type RGBRange struct {
	path    string
	Min     string
	Max     string
	Samples int
}

func NewRGBRange(path string, min, max string, samples int) RGBRange {
	return RGBRange{
		path:    path,
		Min:     min,
		Max:     max,
		Samples: samples,
	}
}

func (r RGBRange) Path() string { return r.path }
func (r RGBRange) Count() int {
	return axisCount(r.Samples) * axisCount(r.Samples) * axisCount(r.Samples)
}

func (r RGBRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	minR, minG, minB, err := decodeHexColor(r.Min)
	if err != nil {
		return nil, err
	}
	maxR, maxG, maxB, err := decodeHexColor(r.Max)
	if err != nil {
		return nil, err
	}

	rCount := axisCount(r.Samples)
	gCount := axisCount(r.Samples)
	ir := index % rCount
	ig := (index / rCount) % gCount
	ib := index / (rCount * gCount)
	return hexColor(
		axisValue(minR, maxR, r.Samples, ir),
		axisValue(minG, maxG, r.Samples, ig),
		axisValue(minB, maxB, r.Samples, ib),
	), nil
}

func (r RGBRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	minR, minG, minB, err := decodeHexColor(r.Min)
	if err != nil {
		return nil, err
	}
	maxR, maxG, maxB, err := decodeHexColor(r.Max)
	if err != nil {
		return nil, err
	}
	return hexColor(
		randomBetween(minR, maxR, rng),
		randomBetween(minG, maxG, rng),
		randomBetween(minB, maxB, rng),
	), nil
}

func (r RGBRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeRGBRange, rgbRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
