package variant

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const (
	typeDiscrete     = "discrete"
	typeNumericRange = "numericRange"
	typeVector2Range = "vector2Range"
	typeVector3Range = "vector3Range"
	typeRGBRange     = "rgbRange"
	typeHSVRange     = "hsvRange"
	typeCombination  = "combination"
)

// dimensionEnvelope pairs a Dimension's type with its encoded data.
type dimensionEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func marshalDimension(dimensionType string, data any) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(dimensionEnvelope{Type: dimensionType, Data: encoded})
}

type discreteJSON struct {
	Values []json.RawMessage `json:"values"`
}

type vector2RangeJSON struct {
	Min     vector2.Float64 `json:"min"`
	Max     vector2.Float64 `json:"max"`
	Samples vector2.Int     `json:"samples"`
}

type vector3RangeJSON struct {
	Min     vector3.Float64 `json:"min"`
	Max     vector3.Float64 `json:"max"`
	Samples vector3.Int     `json:"samples"`
}

type combinationJSON struct {
	Dimensions []json.RawMessage `json:"dimensions"`
}

// UnmarshalDimension decodes a persisted Dimension for the given path.
func UnmarshalDimension(path string, data []byte) (Dimension, error) {
	var envelope dimensionEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	switch envelope.Type {
	case typeDiscrete:
		var d discreteJSON
		if err := json.Unmarshal(envelope.Data, &d); err != nil {
			return nil, err
		}
		return Discrete{path: path, Values: d.Values}, nil

	case typeNumericRange:
		var r axisRange
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return NumericRange{path: path, axisRange: r}, nil

	case typeVector2Range:
		var r vector2RangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return Vector2Range{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeVector3Range:
		var r vector3RangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return Vector3Range{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeRGBRange:
		var r vector3RangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return RGBRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeHSVRange:
		var r vector3RangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return HSVRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeCombination:
		var c combinationJSON
		if err := json.Unmarshal(envelope.Data, &c); err != nil {
			return nil, err
		}
		dimensions := make([]Dimension, 0, len(c.Dimensions))
		for _, raw := range c.Dimensions {
			dim, err := UnmarshalDimension(path, raw)
			if err != nil {
				return nil, err
			}
			dimensions = append(dimensions, dim)
		}
		return Combination{path: path, Dimensions: dimensions}, nil

	default:
		return nil, fmt.Errorf("unknown variant dimension type %q", envelope.Type)
	}
}
