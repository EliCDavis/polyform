package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/vector/vector3"
)

// Vector3Range lerps between Min and Max as a whole, sampled Samples times.
type Vector3Range struct {
	path    string
	Min     vector3.Float64
	Max     vector3.Float64
	Samples int
}

func NewVector3Range(path string, minX, maxX, minY, maxY, minZ, maxZ float64, samples int) Vector3Range {
	return Vector3Range{
		path:    path,
		Min:     vector3.New(minX, minY, minZ),
		Max:     vector3.New(maxX, maxY, maxZ),
		Samples: samples,
	}
}

func (r Vector3Range) Path() string { return r.path }
func (r Vector3Range) Count() int   { return axisCount(r.Samples) }

func (r Vector3Range) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(vector3.New(
		axisValue(r.Min.X(), r.Max.X(), r.Samples, index),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples, index),
		axisValue(r.Min.Z(), r.Max.Z(), r.Samples, index),
	))
}

func (r Vector3Range) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	return json.Marshal(vector3.New(
		lerp(r.Min.X(), r.Max.X(), t),
		lerp(r.Min.Y(), r.Max.Y(), t),
		lerp(r.Min.Z(), r.Max.Z(), t),
	))
}

func (r Vector3Range) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector3Range, vector3RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
