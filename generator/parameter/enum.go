package parameter

import (
	"encoding/json"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func DefineEnum[E comparable](entries ...enum.Entry[E]) enum.Set[E] {
	set := enum.Define(entries...)
	factory := &refutil.TypeFactory{}
	refutil.RegisterTypeWithBuilder(factory, func() Enum[E] { return Enum[E]{set: set} })
	refutil.RegisterTypeWithBuilder(factory, func() enum.Members[E] { return enum.NewMembers(set) })
	registry.Register(factory)
	variabletypes.RegisterEnum(set)
	return set
}

type Enum[E comparable] struct {
	Name        string
	Description string

	set          enum.Set[E]
	version      int
	CurrentValue E
}

type enumGraphSchema struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	CurrentValue string `json:"currentValue"`
}

func (en *Enum[E]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		valueOutputPortName: enumOutput[E]{node: en},
	}
}

func (en Enum[E]) Inputs() map[string]nodes.InputPort {
	return nil
}

func (en *Enum[E]) SetName(name string) {
	en.Name = name
}

func (en *Enum[E]) SetDescription(description string) {
	en.Description = description
}

func (en *Enum[E]) DisplayName() string {
	return en.Name
}

func (en *Enum[E]) Value() E {
	return en.CurrentValue
}

func (en *Enum[E]) ApplyMessage(msg []byte) (bool, error) {
	var name string
	if err := json.Unmarshal(msg, &name); err != nil {
		return false, err
	}
	val, err := en.set.Parse(name)
	if err != nil {
		return false, err
	}
	en.version++
	nodes.Touch()
	en.CurrentValue = val
	return true, nil
}

func (en Enum[E]) ToMessage() []byte {
	data, err := json.Marshal(en.set.Name(en.CurrentValue))
	if err != nil {
		panic(err)
	}
	return data
}

func (en *Enum[E]) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(enumGraphSchema{
		Name:         en.Name,
		Description:  en.Description,
		CurrentValue: en.set.Name(en.CurrentValue),
	})
}

func (en *Enum[E]) FromJSON(decoder jbtf.Decoder, body []byte) error {
	gn := enumGraphSchema{}
	if err := json.Unmarshal(body, &gn); err != nil {
		return err
	}
	val, err := en.set.Parse(gn.CurrentValue)
	if err != nil {
		return err
	}
	en.Name = gn.Name
	en.Description = gn.Description
	en.CurrentValue = val
	return nil
}

func (en *Enum[E]) CopyStateFrom(original nodes.Node) bool {
	from, ok := original.(*Enum[E])
	if !ok {
		return false
	}
	en.Name = from.Name
	en.Description = from.Description
	en.CurrentValue = from.CurrentValue
	return true
}

func (en *Enum[E]) Schema() schema.Parameter {
	return schema.EnumParameter{
		ParameterBase: schema.ParameterBase{
			Name:        en.Name,
			Description: en.Description,
			Type:        refutil.TypeResolution{}.Resolve(*new(E)),
		},
		CurrentValue: en.set.Name(en.CurrentValue),
	}
}

type enumOutput[E comparable] struct {
	node *Enum[E]
}

func (o enumOutput[E]) Value() E { return o.node.CurrentValue }

func (o enumOutput[E]) Node() nodes.Node { return o.node }

func (o enumOutput[E]) Name() string { return valueOutputPortName }

func (o enumOutput[E]) Version() int { return o.node.version }

func (o enumOutput[E]) BuildProxyOutput(source nodes.ProxySource) nodes.OutputPort {
	return nodes.NewProxyOutput[E](source)
}

func (o enumOutput[E]) Type() string {
	return refutil.TypeResolution{IncludePackage: true}.Resolve(*new(E))
}
