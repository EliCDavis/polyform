package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/refutil"
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

func TestARenamedVariableReadInsideASubGraphSurvivesSaveAndLoad(t *testing.T) {
	newInstance := func() *graph.Instance {
		return graph.New(graph.Config{TypeFactory: &refutil.TypeFactory{}, VariableFactory: variabletypes.New})
	}
	inst := newInstance()
	require.NoError(t, errOf(inst.NewVariable("Wave", &variable.TypeVariable[float64]{})))
	require.NoError(t, inst.CreateSubGraph("fox", "Fox", ""))
	fox, err := inst.SubGraphInstance("fox")
	require.NoError(t, err)
	_, reference, err := fox.CreateNode("Wave")
	require.NoError(t, err)

	require.NoError(t, inst.SetVariableInfo("Wave", "Arm Raise", ""))
	_, _, err = inst.CreateNode("Arm Raise")
	require.NoError(t, err, "the new path is a node type")

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	reloaded := newInstance()
	require.NoError(t, reloaded.ApplyAppSchema(saved))
	assert.Equal(t, "Arm Raise", reloaded.Schema().SubGraphs["fox"].Nodes[reference].VariablePath)
}
