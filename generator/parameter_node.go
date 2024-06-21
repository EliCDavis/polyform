package generator

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

// ============================================================================

type ParameterNodeOutput[T any] struct {
	Val *ParameterNode[T]
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

type ParameterNodeSchema[T any] struct {
	ParameterSchemaBase
	DefaultValue T `json:"defaultValue"`
	CurrentValue T `json:"currentValue"`
}

// ============================================================================

type CliParameterNodeConfig[T any] struct {
	FlagName string `json:"flagName"`
	Usage    string `json:"usage"`
	// Default  T      `json:"default"`
	value *T
}

// ============================================================================

type parameterNodeGraphSchema[T any] struct {
	Name         string                     `json:"name"`
	CurrentValue T                          `json:"currentValue"`
	DefaultValue T                          `json:"defaultValue"`
	CLI          *CliParameterNodeConfig[T] `json:"cli"`
}

// ============================================================================

type ParameterNode[T any] struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	DefaultValue T                          `json:"defaultValue"`
	CLI          *CliParameterNodeConfig[T] `json:"cli"`

	subs           []nodes.Alertable
	version        int
	appliedProfile *T
}

func (in *ParameterNode[T]) Port() string {
	return "Out"
}

func (vn ParameterNode[T]) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
}

func (pn *ParameterNode[T]) DisplayName() string {
	return pn.Name
}

func (pn *ParameterNode[T]) ApplyMessage(msg []byte) (bool, error) {
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

func (pn ParameterNode[T]) ToMessage() []byte {
	data, err := json.Marshal(pn.Value())
	if err != nil {
		panic(err)
	}
	return data
}

func (pn *ParameterNode[T]) Value() T {
	if pn.appliedProfile != nil {
		return *pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil {
		return *pn.CLI.value
	}
	return pn.DefaultValue
}

// CUSTOM JTF Serialization ===================================================

func (pn *ParameterNode[T]) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(parameterNodeGraphSchema[T]{
		Name:         pn.Name,
		CurrentValue: pn.Value(),
		DefaultValue: pn.DefaultValue,
		CLI:          pn.CLI,
	})
}

func (pn *ParameterNode[T]) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
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

func (pn *ParameterNode[T]) Schema() ParameterSchema {
	return ParameterNodeSchema[T]{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: pn.Name,
			Type: fmt.Sprintf("%T", *new(T)),
		},
		DefaultValue: pn.DefaultValue,
		CurrentValue: pn.Value(),
	}
}

func (pn *ParameterNode[T]) AddSubscription(a nodes.Alertable) {
	if pn.subs == nil {
		pn.subs = make([]nodes.Alertable, 0, 1)
	}

	pn.subs = append(pn.subs, a)
}

func (pn *ParameterNode[T]) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *ParameterNode[T]) State() nodes.NodeState {
	return nodes.Processed
}

func (tn *ParameterNode[T]) Outputs() []nodes.Output {
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

func (tn *ParameterNode[T]) Out() nodes.NodeOutput[T] {
	return ParameterNodeOutput[T]{
		Val: tn,
	}
}

func (in *ParameterNode[T]) Node() nodes.Node {
	return in
}

func (tn ParameterNode[T]) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn ParameterNode[T]) Version() int {
	return pn.version
}

func (pn ParameterNode[T]) initializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	switch cli := any(pn.CLI).(type) {
	case *CliParameterNodeConfig[string]:
		cli.value = set.String(cli.FlagName, (any(pn.DefaultValue)).(string), cli.Usage)

	case *CliParameterNodeConfig[float64]:
		cli.value = set.Float64(cli.FlagName, (any(pn.DefaultValue)).(float64), cli.Usage)

	case *CliParameterNodeConfig[bool]:
		cli.value = set.Bool(cli.FlagName, (any(pn.DefaultValue)).(bool), cli.Usage)

	case *CliParameterNodeConfig[int]:
		cli.value = set.Int(cli.FlagName, (any(pn.DefaultValue)).(int), cli.Usage)

	case *CliParameterNodeConfig[int64]:
		cli.value = set.Int64(cli.FlagName, (any(pn.DefaultValue)).(int64), cli.Usage)
	default:
		panic(fmt.Errorf("parameter node %s has a type that can not be initialized on the command line. Please open up a issue on github.com/EliCDavis/polyform", pn.DisplayName()))
	}
}

func (pn ParameterNode[T]) SwaggerProperty() swagger.Property {
	prop := swagger.Property{
		Description: pn.Description,
	}
	switch any(pn).(type) {
	case ParameterNode[string]:
		prop.Type = swagger.StringPropertyType

	case ParameterNode[time.Time]:
		prop.Type = swagger.StringPropertyType
		prop.Format = swagger.DateTimePropertyFormat

	case ParameterNode[float64]:
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.DoublePropertyFormat

	case ParameterNode[float32]:
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.FloatPropertyFormat

	case ParameterNode[bool]:
		prop.Type = swagger.BooleanPropertyType

	case ParameterNode[int]:
		prop.Type = swagger.IntegerPropertyType

	case ParameterNode[int64]:
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int64PropertyFormat

	case ParameterNode[int32]:
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int32PropertyFormat

	case ParameterNode[vector3.Float64]:
		prop.Ref = "#/definitions/Vector3"

	case ParameterNode[geometry.AABB]:
		prop.Ref = "#/definitions/AABB"

	default:
		panic(fmt.Errorf("parameter node %s has a type that can not be converted to a swagger property. Please open up a issue on github.com/EliCDavis/polyform", pn.DisplayName()))
	}
	return prop
}
