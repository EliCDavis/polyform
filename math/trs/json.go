package trs

import (
	"encoding/json"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
)

type jsonTRS struct {
	Position vector3.Float64       `json:"position"`
	Rotation quaternion.Quaternion `json:"rotation"`
	Scale    vector3.Float64       `json:"scale"`
}

func (trs TRS) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTRS{trs.position, trs.rotation, trs.scale})
}

// Fields left out read as the identity, so "{}" is a valid transform.
func (trs *TRS) UnmarshalJSON(data []byte) error {
	parsed := jsonTRS{Rotation: quaternion.Identity(), Scale: vector3.One[float64]()}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*trs = New(parsed.Position, parsed.Rotation, parsed.Scale)
	return nil
}
