package parameter

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

// ============================================================================

type ParameterNodeOutput[T any] struct {
	Val *Value[T]
}

func (sno ParameterNodeOutput[T]) Value() T {
	return sno.Val.Value()
}

func (sno ParameterNodeOutput[T]) Node() nodes.Node {
	return sno.Val
}

func (sno ParameterNodeOutput[T]) Port() string {
	return "Out"
}

// ============================================================================

type ValueSchema[T any] struct {
	generator.ParameterSchemaBase
	DefaultValue T `json:"defaultValue"`
	CurrentValue T `json:"currentValue"`
}

// ============================================================================

type CliConfig[T any] struct {
	FlagName string `json:"flagName"`
	Usage    string `json:"usage"`
	// Default  T      `json:"default"`
	value *T
}

// ============================================================================

type parameterNodeGraphSchema[T any] struct {
	Name         string        `json:"name"`
	CurrentValue T             `json:"currentValue"`
	DefaultValue T             `json:"defaultValue"`
	CLI          *CliConfig[T] `json:"cli"`
}

// ============================================================================

// Common types for shorthand purposes
type Float64 = Value[float64]
type Float32 = Value[float32]
type Int = Value[int]
type String = Value[string]
type Bool = Value[bool]
type Vector3 = Value[vector3.Float64]
type AABB = Value[geometry.AABB]
type Color = Value[coloring.WebColor]

type Value[T any] struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	DefaultValue T             `json:"defaultValue"`
	CLI          *CliConfig[T] `json:"cli"`

	subs           []nodes.Alertable
	version        int
	appliedProfile *T
}

func (in *Value[T]) Port() string {
	return "Out"
}

func (vn Value[T]) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
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

	// if pn.appliedProfile != nil && val == *pn.appliedProfile {
	// 	return false, nil
	// }

	pn.version++
	pn.appliedProfile = &val

	for _, s := range pn.subs {
		s.Alert(pn.version, nodes.Processed)
	}

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
	if pn.appliedProfile != nil {
		return *pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil {
		return *pn.CLI.value
	}
	return pn.DefaultValue
}

// CUSTOM JTF Serialization ===================================================

func (pn *Value[T]) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(parameterNodeGraphSchema[T]{
		Name:         pn.Name,
		CurrentValue: pn.Value(),
		DefaultValue: pn.DefaultValue,
		CLI:          pn.CLI,
	})
}

func (pn *Value[T]) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
	gn := parameterNodeGraphSchema[T]{}
	err = json.Unmarshal(body, &gn)
	if err != nil {
		return
	}

	pn.Name = gn.Name
	pn.DefaultValue = gn.DefaultValue
	pn.CLI = gn.CLI
	pn.appliedProfile = &gn.CurrentValue
	return
}

// ============================================================================

func (pn *Value[T]) Schema() generator.ParameterSchema {
	return ValueSchema[T]{
		ParameterSchemaBase: generator.ParameterSchemaBase{
			Name: pn.Name,
			Type: fmt.Sprintf("%T", *new(T)),
		},
		DefaultValue: pn.DefaultValue,
		CurrentValue: pn.Value(),
	}
}

func (pn *Value[T]) AddSubscription(a nodes.Alertable) {
	if pn.subs == nil {
		pn.subs = make([]nodes.Alertable, 0, 1)
	}

	pn.subs = append(pn.subs, a)
}

func (pn *Value[T]) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *Value[T]) State() nodes.NodeState {
	return nodes.Processed
}

func (tn *Value[T]) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			// Name: "Out",
			Type: refutil.GetTypeWithPackage(new(T)),
			NodeOutput: ParameterNodeOutput[T]{
				Val: tn,
			},
		},
	}
}

func (tn *Value[T]) Out() nodes.NodeOutput[T] {
	return ParameterNodeOutput[T]{
		Val: tn,
	}
}

func (in *Value[T]) Node() nodes.Node {
	return in
}

func (tn Value[T]) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn Value[T]) Version() int {
	return pn.version
}

func (pn Value[T]) InitializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	switch cli := any(pn.CLI).(type) {
	case *CliConfig[string]:
		cli.value = set.String(cli.FlagName, (any(pn.DefaultValue)).(string), cli.Usage)

	case *CliConfig[float64]:
		cli.value = set.Float64(cli.FlagName, (any(pn.DefaultValue)).(float64), cli.Usage)

	case *CliConfig[bool]:
		cli.value = set.Bool(cli.FlagName, (any(pn.DefaultValue)).(bool), cli.Usage)

	case *CliConfig[int]:
		cli.value = set.Int(cli.FlagName, (any(pn.DefaultValue)).(int), cli.Usage)

	case *CliConfig[int64]:
		cli.value = set.Int64(cli.FlagName, (any(pn.DefaultValue)).(int64), cli.Usage)
	default:
		panic(fmt.Errorf("parameter node %s has a type that can not be initialized on the command line. Please open up a issue on github.com/EliCDavis/polyform", pn.DisplayName()))
	}
}

func (pn Value[T]) SwaggerProperty() swagger.Property {
	prop := swagger.Property{
		Description: pn.Description,
	}
	switch any(pn).(type) {
	case Value[string]:
		prop.Type = swagger.StringPropertyType

	case Value[time.Time]:
		prop.Type = swagger.StringPropertyType
		prop.Format = swagger.DateTimePropertyFormat

	case Value[float64]:
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.DoublePropertyFormat

	case Value[float32]:
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.FloatPropertyFormat

	case Value[bool]:
		prop.Type = swagger.BooleanPropertyType

	case Value[int]:
		prop.Type = swagger.IntegerPropertyType

	case Value[int64]:
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int64PropertyFormat

	case Value[int32]:
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int32PropertyFormat

	case Value[vector3.Float64]:
		prop.Ref = "#/definitions/Vector3"

	case Value[geometry.AABB]:
		prop.Ref = "#/definitions/AABB"

	default:
		panic(fmt.Errorf("parameter node %s has a type that can not be converted to a swagger property. Please open up a issue on github.com/EliCDavis/polyform", pn.DisplayName()))
	}
	return prop
}
