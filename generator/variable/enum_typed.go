package variable

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type EnumTypeVariable[E comparable] struct {
	set     enum.Set[E]
	value   E
	version int
	info    Info
}

func NewEnumTypeVariable[E comparable](set enum.Set[E]) *EnumTypeVariable[E] {
	return &EnumTypeVariable[E]{set: set, value: set.Values()[0]}
}

func (ev *EnumTypeVariable[E]) SetValue(v E) {
	ev.value = v
	ev.version++
	nodes.Touch()
}

func (ev *EnumTypeVariable[E]) GetValue() E { return ev.value }

func (ev *EnumTypeVariable[E]) Version() int { return ev.version }

func (ev *EnumTypeVariable[E]) Info() Info { return ev.info }

func (ev *EnumTypeVariable[E]) setInfo(i Info) error {
	if ev.info != nil {
		return fmt.Errorf("already assigned info")
	}
	ev.info = i
	return nil
}

func (ev *EnumTypeVariable[E]) currentValue() any { return ev.value }

func (ev *EnumTypeVariable[E]) currentVersion() int { return ev.version }

func (ev *EnumTypeVariable[E]) typeName() string {
	return refutil.TypeResolution{}.Resolve(*new(E))
}

func (ev *EnumTypeVariable[E]) ApplyMessage(msg []byte) (bool, error) {
	var name string
	if err := json.Unmarshal(msg, &name); err != nil {
		return false, err
	}
	val, err := ev.set.Parse(name)
	if err != nil {
		return false, err
	}
	ev.SetValue(val)
	return true, nil
}

func (ev *EnumTypeVariable[E]) applyProfile(profile json.RawMessage) error {
	_, err := ev.ApplyMessage(profile)
	return err
}

func (ev *EnumTypeVariable[E]) getProfile() json.RawMessage { return ev.ToMessage() }

func (ev EnumTypeVariable[E]) ToMessage() []byte {
	data, err := json.Marshal(ev.set.Name(ev.value))
	if err != nil {
		panic(err)
	}
	return data
}

func (ev *EnumTypeVariable[E]) NodeReference() nodes.Node {
	return &VariableReferenceNode[E]{variable: ev}
}

func (ev EnumTypeVariable[E]) runtimeSchema() schema.Variable {
	v := schema.Variable{
		Type:  ev.typeName(),
		Value: ev.set.Name(ev.value),
	}
	if ev.info != nil {
		v.Description = ev.info.Description()
	}
	return v
}

func (ev EnumTypeVariable[E]) MarshalJSON() ([]byte, error) {
	return json.Marshal(enumVariableSchema{
		variableSchemaBase: variableSchemaBase{Type: ev.typeName()},
		Value:              ev.set.Name(ev.value),
	})
}

func (ev EnumTypeVariable[E]) toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(ev)
}

func (ev *EnumTypeVariable[E]) fromPersistantJSON(decoder jbtf.Decoder, body []byte) error {
	var saved enumVariableSchema
	if err := json.Unmarshal(body, &saved); err != nil {
		return err
	}
	val, err := ev.set.Parse(saved.Value)
	if err != nil {
		return err
	}
	ev.value = val
	return nil
}

func (ev *EnumTypeVariable[E]) SwaggerProperty() swagger.Property {
	prop := swagger.Property{Type: swagger.StringPropertyType}
	for _, name := range ev.set.Names() {
		prop.Enum = append(prop.Enum, name)
	}
	if ev.info != nil && ev.info.Description() != "" {
		prop.Description = ev.info.Description()
	}
	return prop
}
