package enum_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestMembersListsAGoEnumInDeclaredOrder(t *testing.T) {
	node := enum.NewMembers(compressions)
	port := nodes.GetNodeOutputPort[[]Compression](&node, "Members")
	assert.Equal(t, []Compression{Fast, Balanced, Smallest}, port.Value())
	assert.Equal(t, "[]github.com/EliCDavis/polyform/generator/enum_test.Compression", port.(nodes.Typed).Type())
	assert.Equal(t, "Compression Members", node.Name())
}
