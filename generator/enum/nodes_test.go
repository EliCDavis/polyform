package enum_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemberNameAndIndexOfAGoEnum(t *testing.T) {
	name := nodes.GetNodeOutputPort[string](&nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{
		Value: nodes.ConstOutput[Compression]{Val: Smallest},
	}}, "Out")
	assert.Equal(t, "Smallest", name.Value())

	index := nodes.GetNodeOutputPort[int](&nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{
		Value: nodes.ConstOutput[Compression]{Val: Smallest},
	}}, "Out")
	assert.Equal(t, 2, index.Value())
}

func TestMemberNameAndIndexOfAGraphEnum(t *testing.T) {
	table, err := enum.NewTable("enum/WindowStyle", []enum.Member{{Name: "Casement"}, {Name: "Sash"}, {Name: "Fixed"}})
	require.NoError(t, err)

	name := nodes.GetNodeOutputPort[string](&nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{
		Value: nodes.ConstOutput[enum.Choice]{Val: table.Choice("sash")},
	}}, "Out")
	assert.Equal(t, "sash", name.Value())

	index := nodes.GetNodeOutputPort[int](&nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{
		Value: nodes.ConstOutput[enum.Choice]{Val: table.Choice("Sash")},
	}}, "Out")
	assert.Equal(t, 1, index.Value())
}

func TestMemberNameAndIndexOfAListOfAGoEnum(t *testing.T) {
	list := nodes.ConstOutput[[]Compression]{Val: []Compression{Smallest, Fast}}

	names := nodes.GetNodeOutputPort[[]string](&nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{Value: list}}, "Out")
	assert.Equal(t, []string{"Smallest", "Fast"}, names.Value())

	indices := nodes.GetNodeOutputPort[[]int](&nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{Value: list}}, "Out")
	assert.Equal(t, []int{2, 0}, indices.Value())
}

func TestMemberNameAndIndexOfAListOfAGraphEnum(t *testing.T) {
	table, err := enum.NewTable("enum/WindowStyle", []enum.Member{{Name: "Casement"}, {Name: "Sash"}, {Name: "Fixed"}})
	require.NoError(t, err)
	list := nodes.ConstOutput[[]enum.Choice]{Val: []enum.Choice{table.Choice("Fixed"), table.Choice("Casement")}}

	names := nodes.GetNodeOutputPort[[]string](&nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{Value: list}}, "Out")
	assert.Equal(t, []string{"Fixed", "Casement"}, names.Value())

	indices := nodes.GetNodeOutputPort[[]int](&nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{Value: list}}, "Out")
	assert.Equal(t, []int{2, 0}, indices.Value())
}

func TestMemberIndexOfAnEmptyListIsAnEmptyList(t *testing.T) {
	indices := nodes.GetNodeOutputPort[[]int](&nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{
		Value: nodes.ConstOutput[[]Compression]{Val: []Compression{}},
	}}, "Out")
	assert.Empty(t, indices.Value())
}

func TestMemberIndexOfAListOfNonEnumsIsAnError(t *testing.T) {
	node := &nodes.Struct[enum.MemberIndexNode]{Data: enum.MemberIndexNode{Value: nodes.ConstOutput[[]float64]{Val: []float64{1}}}}
	assert.Equal(t, []int{0}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
	assert.Equal(t, []string{"float64 is not an enum"}, node.Outputs()["Out"].(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestMemberNameOfANonEnumIsAnError(t *testing.T) {
	node := &nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{Value: nodes.ConstOutput[float64]{Val: 3}}}
	assert.Equal(t, "", nodes.GetNodeOutputPort[string](node, "Out").Value())
	assert.Equal(t, []string{"float64 is not an enum"}, node.Outputs()["Out"].(nodes.ObservableExecution).ExecutionReport().Errors)
}
