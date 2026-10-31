package variable

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type EnumVariable struct {
	table   enum.Table
	value   string
	version int
	info    Info
}

func NewEnumVariable(table enum.Table) *EnumVariable {
	return &EnumVariable{table: table, value: table.Names()[0]}
}

func (ev *EnumVariable) Table() enum.Table { return ev.table }

func (ev *EnumVariable) SetTable(table enum.Table) {
	ev.table = table
	ev.version++
	nodes.Touch()
}

func (ev *EnumVariable) GetValue() string { return ev.value }

func (ev *EnumVariable) Choice() enum.Choice { return ev.table.Choice(ev.value) }

func (ev *EnumVariable) SetValue(name string) error {
	canonical, err := ev.table.Canonical(name)
	if err != nil {
		return err
	}
	ev.value = canonical
	ev.version++
	nodes.Touch()
	return nil
}

func (ev *EnumVariable) Version() int { return ev.version }

func (ev *EnumVariable) Info() Info { return ev.info }

func (ev *EnumVariable) setInfo(i Info) error {
	if ev.info != nil {
		return fmt.Errorf("already assigned info")
	}
	ev.info = i
	return nil
}

func (ev *EnumVariable) currentValue() any { return ev.Choice() }

func (ev *EnumVariable) currentVersion() int { return ev.version }

func (ev *EnumVariable) ApplyMessage(msg []byte) (bool, error) {
	var name string
	if err := json.Unmarshal(msg, &name); err != nil {
		return false, err
	}
	return true, ev.SetValue(name)
}

func (ev *EnumVariable) applyProfile(profile json.RawMessage) error {
	_, err := ev.ApplyMessage(profile)
	return err
}

func (ev *EnumVariable) getProfile() json.RawMessage { return ev.ToMessage() }

func (ev EnumVariable) ToMessage() []byte {
	data, err := json.Marshal(ev.value)
	if err != nil {
		panic(err)
	}
	return data
}

func (ev *EnumVariable) NodeReference() nodes.Node {
	return &enumReferenceNode{variable: ev}
}

func (ev EnumVariable) runtimeSchema() schema.Variable {
	v := schema.Variable{
		Type:  ev.table.TypeName(),
		Value: ev.value,
	}
	if ev.info != nil {
		v.Description = ev.info.Description()
	}
	return v
}

type enumVariableSchema struct {
	variableSchemaBase
	Value string `json:"value"`
}

func (ev EnumVariable) toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(enumVariableSchema{
		variableSchemaBase: variableSchemaBase{Type: ev.table.TypeName()},
		Value:              ev.value,
	})
}

func (ev *EnumVariable) fromPersistantJSON(decoder jbtf.Decoder, body []byte) error {
	var saved enumVariableSchema
	if err := json.Unmarshal(body, &saved); err != nil {
		return err
	}
	canonical, err := ev.table.Canonical(saved.Value)
	if err != nil {
		return err
	}
	ev.value = canonical
	return nil
}

func (ev *EnumVariable) SwaggerProperty() swagger.Property {
	prop := swagger.Property{Type: swagger.StringPropertyType}
	for _, name := range ev.table.Names() {
		prop.Enum = append(prop.Enum, name)
	}
	if ev.info != nil && ev.info.Description() != "" {
		prop.Description = ev.info.Description()
	}
	return prop
}

type enumReferenceNode struct {
	variable *EnumVariable
}

func (n *enumReferenceNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		variableReferenceNodeOutputPortName: &enumReferencePort{node: n},
	}
}

func (n *enumReferenceNode) Reference() Variable { return n.variable }

func (n *enumReferenceNode) Inputs() map[string]nodes.InputPort { return nil }

func (n *enumReferenceNode) Name() string { return n.variable.Info().Name() }

func (n *enumReferenceNode) Description() string { return n.variable.Info().Description() }

type enumReferencePort struct {
	node *enumReferenceNode
}

func (p *enumReferencePort) Node() nodes.Node { return p.node }

func (p *enumReferencePort) Name() string { return variableReferenceNodeOutputPortName }

func (p *enumReferencePort) Version() int { return p.node.variable.version }

func (p *enumReferencePort) Value() enum.Choice { return p.node.variable.Choice() }

func (p *enumReferencePort) Type() string { return p.node.variable.table.TypeName() }

func (p *enumReferencePort) BuildProxyOutput(source nodes.ProxySource) nodes.OutputPort {
	return nodes.NewProxyOutput[enum.Choice](source)
}
