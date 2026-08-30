package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/vector/vector3"
)

// Vector3IntRange lerps between two whole-number vector3.Int values as a
// whole, sampled Samples times.
type Vector3IntRange struct {
	path    string
	Min     vector3.Int
	Max     vector3.Int
	Samples int
}

func NewVector3IntRange(path string, minX, maxX, minY, maxY, minZ, maxZ, samples int) Vector3IntRange {
	return Vector3IntRange{
		path:    path,
		Min:     vector3.New(minX, minY, minZ),
		Max:     vector3.New(maxX, maxY, maxZ),
		Samples: samples,
	}
}

func (r Vector3IntRange) Path() string { return r.path }
func (r Vector3IntRange) Count() int   { return axisCount(r.Samples) }

func (r Vector3IntRange) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	return json.Marshal(vector3.New(
		intAxisValue(r.Min.X(), r.Max.X(), r.Samples, index),
		intAxisValue(r.Min.Y(), r.Max.Y(), r.Samples, index),
		intAxisValue(r.Min.Z(), r.Max.Z(), r.Samples, index),
	))
}

func (r Vector3IntRange) Random(rng *rand.Rand) (json.RawMessage, error) {
	t := rng.Float64()
	return json.Marshal(vector3.New(
		intLerp(r.Min.X(), r.Max.X(), t),
		intLerp(r.Min.Y(), r.Max.Y(), t),
		intLerp(r.Min.Z(), r.Max.Z(), t),
	))
}

func (r Vector3IntRange) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector3IntRange, vector3IntRangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
