package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/vector/vector3"
)

// HSVRange combines a range per Hue, Saturation, Value channel into one hex color.
type HSVRange struct {
	path    string
	Min     vector3.Float64
	Max     vector3.Float64
	Samples vector3.Int
}

func NewHSVRange(path string, minH, maxH float64, samplesH int, minS, maxS float64, samplesS int, minV, maxV float64, samplesV int) HSVRange {
	return HSVRange{
		path:    path,
		Min:     vector3.New(minH, minS, minV),
		Max:     vector3.New(maxH, maxS, maxV),
		Samples: vector3.New(samplesH, samplesS, samplesV),
	}
}

func (r HSVRange) Path() string { return r.path }
func (r HSVRange) Count() int {
	return axisCount(r.Samples.X()) * axisCount(r.Samples.Y()) * axisCount(r.Samples.Z())
}

func (r HSVRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	xCount := axisCount(r.Samples.X())
	yCount := axisCount(r.Samples.Y())
	ih := index % xCount
	is := (index / xCount) % yCount
	iv := index / (xCount * yCount)
	red, green, blue := hsvToRGB(
		axisValue(r.Min.X(), r.Max.X(), r.Samples.X(), ih),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples.Y(), is),
		axisValue(r.Min.Z(), r.Max.Z(), r.Samples.Z(), iv),
	)
	return hexColor(red, green, blue), nil
}

func (r HSVRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	v := chance.NewRange3D(r.Min, r.Max, rng).Value()
	red, green, blue := hsvToRGB(v.X(), v.Y(), v.Z())
	return hexColor(red, green, blue), nil
}

func (r HSVRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeHSVRange, vector3RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
