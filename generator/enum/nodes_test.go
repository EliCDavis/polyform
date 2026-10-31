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

func TestMemberNameOfANonEnumIsAnError(t *testing.T) {
	node := &nodes.Struct[enum.MemberNameNode]{Data: enum.MemberNameNode{Value: nodes.ConstOutput[float64]{Val: 3}}}
	assert.Equal(t, "", nodes.GetNodeOutputPort[string](node, "Out").Value())
	assert.Equal(t, []string{"float64 is not an enum"}, node.Outputs()["Out"].(nodes.ObservableExecution).ExecutionReport().Errors)
}
