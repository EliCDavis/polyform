package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/vector/vector3"
)

// RGBRange combines a range per R, G, B channel into one hex color.
type RGBRange struct {
	path    string
	Min     vector3.Float64
	Max     vector3.Float64
	Samples vector3.Int
}

func NewRGBRange(path string, minR, maxR float64, samplesR int, minG, maxG float64, samplesG int, minB, maxB float64, samplesB int) RGBRange {
	return RGBRange{
		path:    path,
		Min:     vector3.New(minR, minG, minB),
		Max:     vector3.New(maxR, maxG, maxB),
		Samples: vector3.New(samplesR, samplesG, samplesB),
	}
}

func (r RGBRange) Path() string { return r.path }
func (r RGBRange) Count() int {
	return axisCount(r.Samples.X()) * axisCount(r.Samples.Y()) * axisCount(r.Samples.Z())
}

func (r RGBRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	xCount := axisCount(r.Samples.X())
	yCount := axisCount(r.Samples.Y())
	ir := index % xCount
	ig := (index / xCount) % yCount
	ib := index / (xCount * yCount)
	return hexColor(
		axisValue(r.Min.X(), r.Max.X(), r.Samples.X(), ir),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples.Y(), ig),
		axisValue(r.Min.Z(), r.Max.Z(), r.Samples.Z(), ib),
	), nil
}

func (r RGBRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	v := chance.NewRange3D(r.Min, r.Max, rng).Value()
	return hexColor(v.X(), v.Y(), v.Z()), nil
}

func (r RGBRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeRGBRange, vector3RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
