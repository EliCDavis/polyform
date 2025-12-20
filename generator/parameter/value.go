package parameter

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// ============================================================================

const valueOutputPortName = "Value"

type parameterNodeOutput[T any] struct {
	Val *Value[T]
}

func (sno parameterNodeOutput[T]) Value() T {
	return sno.Val.Value()
}

func (sno parameterNodeOutput[T]) Node() nodes.Node {
	return sno.Val
}

func (sno parameterNodeOutput[T]) Name() string {
	return valueOutputPortName
}

func (sno parameterNodeOutput[T]) Version() int {
	return sno.Val.version
}

func (sno parameterNodeOutput[T]) Type() string {
	resolver := refutil.TypeResolution{
		IncludePackage: true,
		IncludePointer: false,
	}

	return resolver.Resolve(new(T))
}

// ============================================================================

type ValueSchema[T any] struct {
	schema.ParameterBase
	CurrentValue T `json:"currentValue"`
}

// ============================================================================

type parameterNodeGraphSchema[T any] struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	CurrentValue T      `json:"currentValue"`
}

// ============================================================================

// Common types for shorthand purposes
type Float64 = Value[float64]
type Float32 = Value[float32]
type Int = Value[int]
type String = Value[string]
type Bool = Value[bool]
type Vector2 = Value[vector2.Float64]
type Vector3 = Value[vector3.Float64]
type Vector3Array = Value[[]vector3.Float64]
type AABB = Value[geometry.AABB]
type Color = Value[coloring.Color]

type Value[T any] struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	version      int
	CurrentValue T
}

func (tn *Value[T]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		valueOutputPortName: parameterNodeOutput[T]{Val: tn},
	}
}

func (tn Value[T]) Inputs() map[string]nodes.InputPort {
	return nil
}

func (in *Value[T]) SetName(name string) {
	in.Name = name
}

func (in *Value[T]) SetDescription(description string) {
	in.Description = description
}

func (pn *Value[T]) DisplayName() string {
	return pn.Name
}

func (pn *Value[T]) ApplyMessage(msg []byte) (bool, error) {
	var val T
	err := json.Unmarshal(msg, &val)
	if err != nil {
		return false, err
	}

	pn.version++
	pn.CurrentValue = val

	return true, nil
}

func (pn Value[T]) ToMessage() []byte {
	data, err := json.Marshal(pn.Value())
	if err != nil {
		panic(err)
	}
	return data
}

func (pn *Value[T]) Value() T {
	return pn.CurrentValue
}

// CUSTOM JTF Serialization ===================================================

func (pn *Value[T]) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(parameterNodeGraphSchema[T]{
		Name:         pn.Name,
		Description:  pn.Description,
		CurrentValue: pn.Value(),
	})
}

func (pn *Value[T]) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
	gn := parameterNodeGraphSchema[T]{}
	err = json.Unmarshal(body, &gn)
	if err != nil {
		return
	}

	pn.Name = gn.Name
	pn.Description = gn.Description
	pn.CurrentValue = gn.CurrentValue
	return
}

// ============================================================================

func (pn *Value[T]) Schema() schema.Parameter {
	return ValueSchema[T]{
		ParameterBase: schema.ParameterBase{
			Name:        pn.Name,
			Description: pn.Description,
			Type:        fmt.Sprintf("%T", *new(T)),
		},
		CurrentValue: pn.Value(),
	}
}
