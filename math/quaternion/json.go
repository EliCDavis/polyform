package quaternion

import (
	"encoding/json"
	"math"

	"github.com/EliCDavis/vector/vector3"
)

type jsonQuaternion struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

func (q Quaternion) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonQuaternion{q.v.X(), q.v.Y(), q.v.Z(), q.w})
}

// Fields left out read as the identity, so "{}" is a valid rotation.
// Hand-typed components like 0.7071068 are renormalized: Rotate assumes
// unit length, and the residual skews a box's quads enough for CSG to
// reject them as non-planar.
func (q *Quaternion) UnmarshalJSON(data []byte) error {
	parsed := jsonQuaternion{W: 1}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*q = New(vector3.New(parsed.X, parsed.Y, parsed.Z), parsed.W)
	if length := q.Dot(*q); length > 0 && math.Abs(length-1) > 1e-12 {
		*q = q.Normalize()
	}
	return nil
}
