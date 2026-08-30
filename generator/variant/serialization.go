package variant

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const (
	typeDiscrete        = "discrete"
	typeNumericRange    = "numericRange"
	typeIntRange        = "intRange"
	typeVector2Range    = "vector2Range"
	typeVector3Range    = "vector3Range"
	typeVector2IntRange = "vector2IntRange"
	typeVector3IntRange = "vector3IntRange"
	typeRGBRange        = "rgbRange"
	typeHSVRange        = "hsvRange"
	typeCombination     = "combination"
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

type rangeJSON[V any] struct {
	Min     V   `json:"min"`
	Max     V   `json:"max"`
	Samples int `json:"samples"`
}

type intRangeJSON = rangeJSON[int]
type vector2RangeJSON = rangeJSON[vector2.Float64]
type vector3RangeJSON = rangeJSON[vector3.Float64]
type vector2IntRangeJSON = rangeJSON[vector2.Int]
type vector3IntRangeJSON = rangeJSON[vector3.Int]
type rgbRangeJSON = rangeJSON[persistence.WebColor]
type hsvRangeJSON = rangeJSON[HSVChannels]

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

	case typeIntRange:
		var r intRangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return IntRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

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

	case typeVector2IntRange:
		var r vector2IntRangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return Vector2IntRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeVector3IntRange:
		var r vector3IntRangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return Vector3IntRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeRGBRange:
		var r rgbRangeJSON
		if err := json.Unmarshal(envelope.Data, &r); err != nil {
			return nil, err
		}
		return RGBRange{path: path, Min: r.Min, Max: r.Max, Samples: r.Samples}, nil

	case typeHSVRange:
		var r hsvRangeJSON
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
