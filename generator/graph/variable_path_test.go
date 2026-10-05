package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaTellsSameNamedVariablesApartByPath(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, errOf(inst.NewVariable("Copper/Ground Pour", &variable.TypeVariable[bool]{})))
	require.NoError(t, errOf(inst.NewVariable("Colors/Ground Pour", &variable.TypeVariable[coloring.Color]{})))

	_, shown, err := inst.CreateNode("Copper/Ground Pour")
	require.NoError(t, err)
	_, color, err := inst.CreateNode("Colors/Ground Pour")
	require.NoError(t, err)

	nodes := inst.Schema().Nodes
	assert.Equal(t, nodes[shown].Name, nodes[color].Name)
	assert.Equal(t, "Copper/Ground Pour", nodes[shown].VariablePath)
	assert.Equal(t, "Colors/Ground Pour", nodes[color].VariablePath)
}
