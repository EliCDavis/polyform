package graph

import (
	"encoding/json"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

const EnumParameterTypePrefix = "enum-parameter/"

func IsEnumParameterType(typeName string) bool {
	return len(typeName) > len(EnumParameterTypePrefix) && typeName[:len(EnumParameterTypePrefix)] == EnumParameterTypePrefix
}

func EnumParameterTypeID(typeName string) string {
	if !IsEnumParameterType(typeName) {
		return ""
	}
	return typeName[len(EnumParameterTypePrefix):]
}

func EnumParameterTypePath(id string) string {
	return EnumParameterTypePrefix + id
}

const enumParameterOutputPortName = "Value"

type EnumParameterNode struct {
	title       string
	description string
	enumName    string

	table   enum.Table
	value   string
	version int
}

func NewEnumParameter(table enum.Table, enumName string) *EnumParameterNode {
	return &EnumParameterNode{table: table, enumName: enumName, value: table.Names()[0]}
}

func (e *EnumParameterNode) Name() string { return e.enumName + " Parameter" }

func (e *EnumParameterNode) Path() string { return "generator/parameter" }

func (e *EnumParameterNode) Table() enum.Table { return e.table }

func (e *EnumParameterNode) SetTable(table enum.Table) {
	e.table = table
	e.version++
	nodes.Touch()
}

func (e *EnumParameterNode) Value() string { return e.value }

func (e *EnumParameterNode) Choice() enum.Choice { return e.table.Choice(e.value) }

func (e *EnumParameterNode) SetValue(name string) error {
	canonical, err := e.table.Canonical(name)
	if err != nil {
		return err
	}
	e.value = canonical
	e.version++
	nodes.Touch()
	return nil
}

func (e *EnumParameterNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		enumParameterOutputPortName: enumParameterOutput{param: e},
	}
}

func (e *EnumParameterNode) Inputs() map[string]nodes.InputPort { return nil }

func (e *EnumParameterNode) SetName(name string) { e.title = name }

func (e *EnumParameterNode) SetDescription(description string) { e.description = description }

func (e *EnumParameterNode) DisplayName() string { return e.title }

func (e *EnumParameterNode) ApplyMessage(msg []byte) (bool, error) {
	var name string
	if err := json.Unmarshal(msg, &name); err != nil {
		return false, err
	}
	return true, e.SetValue(name)
}

func (e *EnumParameterNode) ToMessage() []byte {
	data, err := json.Marshal(e.value)
	if err != nil {
		panic(err)
	}
	return data
}

type enumParameterGraphSchema struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	CurrentValue string `json:"currentValue"`
}

func (e *EnumParameterNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(enumParameterGraphSchema{
		Name:         e.title,
		Description:  e.description,
		CurrentValue: e.value,
	})
}

func (e *EnumParameterNode) FromJSON(decoder jbtf.Decoder, body []byte) error {
	var saved enumParameterGraphSchema
	if err := json.Unmarshal(body, &saved); err != nil {
		return err
	}
	e.title = saved.Name
	e.description = saved.Description
	return e.SetValue(saved.CurrentValue)
}

func (e *EnumParameterNode) CopyStateFrom(original nodes.Node) bool {
	from, ok := original.(*EnumParameterNode)
	if !ok {
		return false
	}
	e.title = from.title
	e.description = from.description
	e.enumName = from.enumName
	e.table = from.table
	e.value = from.value
	return true
}

func (e *EnumParameterNode) Schema() schema.Parameter {
	return schema.EnumParameter{
		ParameterBase: schema.ParameterBase{
			Name:        e.title,
			Description: e.description,
			Type:        e.table.TypeName(),
		},
		CurrentValue: e.value,
	}
}

type enumParameterOutput struct {
	param *EnumParameterNode
}

func (o enumParameterOutput) Value() enum.Choice { return o.param.Choice() }

func (o enumParameterOutput) Node() nodes.Node { return o.param }

func (o enumParameterOutput) Name() string { return enumParameterOutputPortName }

func (o enumParameterOutput) Version() int { return o.param.version }

func (o enumParameterOutput) Type() string { return o.param.table.TypeName() }

func (o enumParameterOutput) BuildProxyOutput(source nodes.ProxySource) nodes.OutputPort {
	return nodes.NewProxyOutput[enum.Choice](source)
}
