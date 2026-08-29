package variant

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/vector/vector3"
)

// Vector3Range combines an independent range per axis into one vector3.Float64.
type Vector3Range struct {
	path    string
	Min     vector3.Float64
	Max     vector3.Float64
	Samples vector3.Int
}

func NewVector3Range(path string, minX, maxX float64, samplesX int, minY, maxY float64, samplesY int, minZ, maxZ float64, samplesZ int) Vector3Range {
	return Vector3Range{
		path:    path,
		Min:     vector3.New(minX, minY, minZ),
		Max:     vector3.New(maxX, maxY, maxZ),
		Samples: vector3.New(samplesX, samplesY, samplesZ),
	}
}

func (r Vector3Range) Path() string { return r.path }
func (r Vector3Range) Count() int {
	return axisCount(r.Samples.X()) * axisCount(r.Samples.Y()) * axisCount(r.Samples.Z())
}

func (r Vector3Range) Value(index int) (json.RawMessage, error) {
	if index < 0 || index >= r.Count() {
		return nil, fmt.Errorf("index %d out of range [0,%d)", index, r.Count())
	}
	xCount := axisCount(r.Samples.X())
	yCount := axisCount(r.Samples.Y())
	ix := index % xCount
	iy := (index / xCount) % yCount
	iz := index / (xCount * yCount)
	return json.Marshal(vector3.New(
		axisValue(r.Min.X(), r.Max.X(), r.Samples.X(), ix),
		axisValue(r.Min.Y(), r.Max.Y(), r.Samples.Y(), iy),
		axisValue(r.Min.Z(), r.Max.Z(), r.Samples.Z(), iz),
	))
}

func (r Vector3Range) Random(rng *rand.Rand) (json.RawMessage, error) {
	v := chance.NewRange3D(r.Min, r.Max, rng).Value()
	return json.Marshal(v)
}

func (r Vector3Range) MarshalJSON() ([]byte, error) {
	return marshalDimension(typeVector3Range, vector3RangeJSON{Min: r.Min, Max: r.Max, Samples: r.Samples})
}
