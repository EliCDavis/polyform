package generator

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
)

type ParameterNodeSchema[T any] struct {
	ParameterSchemaBase
	DefaultValue T `json:"defaultValue"`
	CurrentValue T `json:"currentValue"`
}

type CliParameterNodeConfig[T any] struct {
	FlagName string
	Usage    string
	value    *T
}

type ParameterNode[T any] struct {
	Name         string
	DefaultValue T
	CLI          *CliParameterNodeConfig[T]

	subs           []nodes.Alertable
	version        int
	appliedProfile *T
}

func (in *ParameterNode[T]) Node() nodes.Node {
	return in
}

func (pn *ParameterNode[T]) DisplayName() string {
	return pn.Name
}

func (pn *ParameterNode[T]) ApplyJsonMessage(msg json.RawMessage) (bool, error) {
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

func (pn *ParameterNode[T]) Data() T {
	if pn.appliedProfile != nil {
		return *pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil {
		return *pn.CLI.value
	}
	return pn.DefaultValue
}

func (pn *ParameterNode[T]) Schema() ParameterSchema {
	return ParameterNodeSchema[T]{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: pn.Name,
			Type: fmt.Sprintf("%T", *new(T)),
		},
		DefaultValue: pn.DefaultValue,
		CurrentValue: pn.Data(),
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

func (pn ParameterNode[T]) Version() int {
	return pn.version
}

func (pn ParameterNode[T]) initializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}

	panic("parameter node initializeForCLI is unimplemented")
}
