package graph

import (
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/nodes"
)

const enumMembersOutputPortName = "Members"

type EnumMembersNode struct {
	enumName string
	table    enum.Table
	version  int
}

func NewEnumMembers(table enum.Table, enumName string) *EnumMembersNode {
	return &EnumMembersNode{table: table, enumName: enumName}
}

func (e *EnumMembersNode) Name() string { return e.enumName + " Members" }

func (e *EnumMembersNode) Path() string { return "generator/enum" }

func (e *EnumMembersNode) Description() string {
	return "Every member of " + e.enumName + ", in declared order."
}

func (e *EnumMembersNode) Table() enum.Table { return e.table }

func (e *EnumMembersNode) SetTable(table enum.Table) {
	e.table = table
	e.version++
	nodes.Touch()
}

func (e *EnumMembersNode) Members() []enum.Choice {
	names := e.table.Names()
	choices := make([]enum.Choice, len(names))
	for i, name := range names {
		choices[i] = e.table.Choice(name)
	}
	return choices
}

func (e *EnumMembersNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		enumMembersOutputPortName: enumMembersOutput{node: e},
	}
}

func (e *EnumMembersNode) Inputs() map[string]nodes.InputPort { return nil }

func (e *EnumMembersNode) CopyStateFrom(original nodes.Node) bool {
	from, ok := original.(*EnumMembersNode)
	if !ok {
		return false
	}
	e.enumName = from.enumName
	e.table = from.table
	return true
}

type enumMembersOutput struct {
	node *EnumMembersNode
}

func (o enumMembersOutput) Value() []enum.Choice { return o.node.Members() }

func (o enumMembersOutput) Node() nodes.Node { return o.node }

func (o enumMembersOutput) Name() string { return enumMembersOutputPortName }

func (o enumMembersOutput) Version() int { return o.node.version }

func (o enumMembersOutput) Type() string { return "[]" + o.node.table.TypeName() }

func (o enumMembersOutput) BuildProxyOutput(source nodes.ProxySource) nodes.OutputPort {
	return nodes.NewProxyOutput[[]enum.Choice](source)
}
