package graph_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawPayload(subGraphs string) []byte {
	return []byte(fmt.Sprintf(`{
		"buffers": [{"byteLength": 0, "uri": "data:application/octet-stream;base64,"}],
		"data": {"nodes": {}, "producers": {}, "subGraphs": %s}
	}`, subGraphs))
}

const outputOnlyNodes = `{
	"out": {"type": "` + subgraph.OutputNodeTypeKey + `", "data": {"portName": "Out", "portType": "float64"}}
}`

func placing(id string) string {
	return `{"p": {"type": "subgraph/` + id + `"}}`
}

func subGraphIDs(inst *graph.Instance) []string {
	return slices.Sorted(maps.Keys(inst.Schema().SubGraphs))
}

func out(t testing.TB, node nodes.Node) float64 {
	t.Helper()
	return nodes.GetNodeOutputPort[float64](node, "Out").Value()
}

func feed(t testing.TB, inst *graph.Instance, placementID, port string, value float64) {
	t.Helper()
	param, id, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	require.NoError(t, errOf(param.(interface{ ApplyMessage([]byte) (bool, error) }).ApplyMessage([]byte(fmt.Sprint(value)))))
	require.NoError(t, inst.ConnectNodes(id, "Value", placementID, port))
}

// nest builds a sub-graph whose Out sums the Out of one placement per inner
// id, and returns the id of that sum.
func nest(t *testing.T, inst *graph.Instance, outerID string, innerIDs ...string) string {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(outerID, outerID, ""))
	outer, err := inst.SubGraphInstance(outerID)
	require.NoError(t, err)
	_, outID, err := outer.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, outer.SetBoundaryNodeInfo(outID, "Out"))
	_, sumID, err := outer.CreateNode("Sum")
	require.NoError(t, err)
	for _, inner := range innerIDs {
		_, placementID := place(t, outer, inner)
		require.NoError(t, outer.ConnectNodes(placementID, "Out", sumID, "Values"))
	}
	require.NoError(t, outer.ConnectNodes(sumID, "Float", outID, subgraph.ValuePortName))
	return sumID
}

func TestImportRejectsAPayloadThatIsNotAGraph(t *testing.T) {
	dest := importTestInstance(t, "Dest")
	for _, payload := range []string{"", "{", "[]", `{"data": 5}`} {
		_, err := dest.ImportSubGraphDefinitions([]byte(payload), "")
		assert.Error(t, err, "payload %q", payload)
	}
	assert.Empty(t, subGraphIDs(dest))
}

func TestImportRefusesAPlacementOfASubGraphTheFileDoesNotDefine(t *testing.T) {
	dest := importTestInstance(t, "Dest")
	require.NoError(t, dest.CreateSubGraph("inner", "Theirs", ""))

	_, err := dest.ImportSubGraphDefinitions(rawPayload(`{"outer": {"name": "Outer", "nodes": `+placing("inner")+`}}`), "")

	require.Error(t, err)
	assert.Equal(t, []string{"inner"}, subGraphIDs(dest), "must not bind to the destination's own inner, nor leave outer behind")
}

func TestImportOfDefinitionsThatPlaceEachOtherLeavesNothingBehind(t *testing.T) {
	dest := importTestInstance(t, "Dest")

	_, err := dest.ImportSubGraphDefinitions(rawPayload(`{
		"a": {"name": "A", "nodes": `+placing("b")+`},
		"b": {"name": "B", "nodes": `+placing("a")+`}
	}`), "")

	require.Error(t, err)
	assert.Empty(t, subGraphIDs(dest))
}

func TestImportOfADefinitionThatPlacesItselfLeavesNothingBehind(t *testing.T) {
	dest := importTestInstance(t, "Dest")

	_, err := dest.ImportSubGraphDefinitions(rawPayload(`{"a": {"name": "A", "nodes": `+placing("a")+`}}`), "")

	require.Error(t, err)
	assert.Empty(t, subGraphIDs(dest))
}

func TestImportOfAVariableTheFileDoesNotDefineNamesItAndLeavesNothingBehind(t *testing.T) {
	dest := importTestInstance(t, "Dest")
	floatVariable(t, dest, "Ghost", 1)

	_, err := dest.ImportSubGraphDefinitions(rawPayload(`{
		"fine": {"name": "Fine", "nodes": `+outputOnlyNodes+`},
		"haunted": {"name": "Haunted", "nodes": {"v": {"type": "Ghost", "variable": "Ghost"}}}
	}`), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `"Ghost"`)
	assert.Empty(t, subGraphIDs(dest))
}

func TestImportGivesAnEmptyIDAName(t *testing.T) {
	dest := importTestInstance(t, "Dest")

	result, err := dest.ImportSubGraphDefinitions(rawPayload(`{"": {"name": "Nameless", "nodes": `+outputOnlyNodes+`}}`), "")

	require.NoError(t, err)
	require.Len(t, result.Imported, 1)
	assert.NotEmpty(t, result.Imported[0].ID)
	assert.Equal(t, []string{result.Imported[0].ID}, subGraphIDs(dest))
}

func TestImportingAGraphIntoItselfCopiesWithoutDisturbingTheOriginals(t *testing.T) {
	inst := importTestInstance(t, "Self")
	floatVariable(t, inst, "Size", 7)
	readsVariable(t, inst, "inner", "Size")
	nest(t, inst, "outer", "inner")
	original, _ := place(t, inst, "outer")

	result, err := inst.ImportSubGraphDefinitions(encode(t, inst), "")
	require.NoError(t, err)

	assert.Equal(t, []string{"inner", "inner_2", "outer", "outer_2"}, subGraphIDs(inst))
	assert.Equal(t, 7.0, out(t, original))
	copied, _ := place(t, inst, "outer_2")
	assert.Equal(t, 7.0, out(t, copied))
	for _, imported := range result.Imported {
		assert.NotEmpty(t, imported.OriginalID, imported.ID)
	}
}

func TestImportingTwiceWiresEachCopyToItsOwnNestedCopy(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")
	nest(t, src, "outer", "inner")
	payload := encode(t, src)

	dest := importTestInstance(t, "Dest")
	_, err := dest.ImportSubGraphDefinitions(payload, "")
	require.NoError(t, err)
	_, err = dest.ImportSubGraphDefinitions(payload, "")
	require.NoError(t, err)

	second, err := dest.SubGraphInstance("outer_2")
	require.NoError(t, err)
	var placed []string
	for _, node := range second.Schema().Nodes {
		if subgraph.IsRuntimeNodeType(node.Type) {
			placed = append(placed, node.Type)
		}
	}
	assert.Equal(t, []string{subgraph.RuntimeTypePath("inner_2")}, placed)
}

func TestAVariableReadThreeLevelsDownIsOneInputAtTheTop(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "leaf", "Size")
	nest(t, src, "mid", "leaf")
	nest(t, src, "top", "mid")

	dest := importTestInstance(t, "Dest")
	_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	top, topID := place(t, dest, "top")
	assert.Equal(t, 7.0, out(t, top))
	assert.Equal(t, []string{"Size"}, inputLabels(top))
	feed(t, dest, topID, "Size", 3)
	assert.Equal(t, 3.0, out(t, top))
}

func TestTwoPlacedSubGraphsReadingOneVariableShareTheirPlacersInput(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "left", "Size")
	readsVariable(t, src, "right", "Size")
	nest(t, src, "both", "left", "right")

	dest := importTestInstance(t, "Dest")
	_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	both, bothID := place(t, dest, "both")
	assert.Len(t, both.Inputs(), 1)
	assert.Equal(t, 14.0, out(t, both))
	feed(t, dest, bothID, "Size", 1)
	assert.Equal(t, 2.0, out(t, both))
}

func TestAPlacerThatAlsoReadsTheVariableItselfGetsOneInput(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")
	sumID := nest(t, src, "outer", "inner")
	outer, err := src.SubGraphInstance("outer")
	require.NoError(t, err)
	_, refID, err := outer.CreateNode("Size")
	require.NoError(t, err)
	require.NoError(t, outer.ConnectNodes(refID, "Value", sumID, "Values"))
	original, _ := place(t, src, "outer")
	require.Equal(t, 14.0, out(t, original))

	dest := importTestInstance(t, "Dest")
	_, err = dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	placement, placementID := place(t, dest, "outer")
	assert.Len(t, placement.Inputs(), 1)
	assert.Equal(t, 14.0, out(t, placement))
	feed(t, dest, placementID, "Size", 2)
	assert.Equal(t, 4.0, out(t, placement))
}

func TestVariablesSharingANameInDifferentGroupsBecomeSeparateInputs(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Wood/Size", 2)
	floatVariable(t, src, "Metal/Size", 5)
	readsVariable(t, src, "mixed", "Wood/Size", "Metal/Size")

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	ports := map[string]string{}
	for _, input := range result.Imported[0].Inputs {
		ports[input.Variable] = input.Port
	}
	assert.ElementsMatch(t, []string{"Size", "Size_2"}, slices.Collect(maps.Values(ports)))

	placement, placementID := place(t, dest, "mixed")
	assert.Equal(t, 7.0, out(t, placement))
	feed(t, dest, placementID, ports["Wood/Size"], 10)
	assert.Equal(t, 15.0, out(t, placement), "only the wood input changed")
}

func TestAVariableWithNoValueNodeStillBecomesAnInputWithoutADefault(t *testing.T) {
	src := importTestInstance(t, "Source")
	require.NoError(t, errOf(src.NewVariable("Flag", newBoolVariable(true))))
	require.NoError(t, src.CreateSubGraph("flagged", "Flagged", ""))
	def, err := src.SubGraphInstance("flagged")
	require.NoError(t, err)
	_, refID, err := def.CreateNode("Flag")
	require.NoError(t, err)
	_, outID, err := def.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "bool")
	require.NoError(t, err)
	require.NoError(t, def.SetBoundaryNodeInfo(outID, "Out"))
	require.NoError(t, def.ConnectNodes(refID, "Value", outID, subgraph.ValuePortName))

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	assert.Equal(t, []graph.VariableInput{{Variable: "Flag", Port: "Flag", HasDefault: false}}, result.Imported[0].Inputs)
	placement, _ := place(t, dest, "flagged")
	assert.Contains(t, inputLabels(placement), "Flag")
}

func TestAVariableReadOnlyByTheRootIsNotLifted(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	_, _, err := src.CreateNode("Size")
	require.NoError(t, err)
	readsVariable(t, src, "plain")

	dest := importTestInstance(t, "Dest")
	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	assert.Empty(t, result.Imported[0].Inputs)
	assert.Empty(t, dest.Schema().Nodes)
}

func TestALiftedSubGraphSurvivesASaveAndLoadOfTheGraphItWasImportedInto(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")
	nest(t, src, "outer", "inner")

	dest := importTestInstance(t, "Dest")
	_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	reloaded := importTestInstance(t, "Reloaded")
	require.NoError(t, reloaded.ApplyAppSchema(encode(t, dest)))
	placement, placementID := place(t, reloaded, "outer")
	assert.Equal(t, 7.0, out(t, placement))
	feed(t, reloaded, placementID, "Size", 4)
	assert.Equal(t, 4.0, out(t, placement))
}

func TestImportLeavesTheDestinationsOwnPlacementsWorking(t *testing.T) {
	dest := importTestInstance(t, "Dest")
	floatVariable(t, dest, "Size", 100)
	readsVariable(t, dest, "inner", "Size")
	existing, _ := place(t, dest, "inner")

	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")
	_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	assert.Equal(t, 100.0, out(t, existing))
	assert.Empty(t, mustInputs(t, dest, "inner"), "the destination's inner still reads its variable directly")
	copied, _ := place(t, dest, "inner_2")
	assert.Equal(t, 7.0, out(t, copied))
}

func TestTheGroupFollowsARenameAndAnInfoEdit(t *testing.T) {
	src := importTestInstance(t, "Library")
	readsVariable(t, src, "piece")
	dest := importTestInstance(t, "Dest")
	_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
	require.NoError(t, err)

	require.NoError(t, dest.RenameSubGraph("piece", "part"))
	require.NoError(t, dest.SetSubGraphInfo("part", "Part", "renamed", "Library"))

	assert.Equal(t, "Library", dest.Schema().SubGraphs["part"].Group)
}

func TestAnUnnamedSourceImportedWithoutAGroupStaysUngrouped(t *testing.T) {
	src := importTestInstance(t, "")
	readsVariable(t, src, "piece")
	dest := importTestInstance(t, "Dest")

	result, err := dest.ImportSubGraphDefinitions(encode(t, src), "")

	require.NoError(t, err)
	assert.Empty(t, result.Group)
	assert.Empty(t, dest.Schema().SubGraphs["piece"].Group)
}

func TestImportCanBeUndone(t *testing.T) {
	src := importTestInstance(t, "Source")
	floatVariable(t, src, "Size", 7)
	readsVariable(t, src, "inner", "Size")
	dest := importTestInstance(t, "Dest")
	readsVariable(t, dest, "mine")

	require.NoError(t, dest.History().Transact("import", func() error {
		_, err := dest.ImportSubGraphDefinitions(encode(t, src), "")
		return err
	}))
	require.Equal(t, []string{"inner", "mine"}, subGraphIDs(dest))

	require.NoError(t, errOf(dest.History().Undo()))
	assert.Equal(t, []string{"mine"}, subGraphIDs(dest))
	require.NoError(t, errOf(dest.History().Redo()))
	assert.Equal(t, []string{"inner", "mine"}, subGraphIDs(dest))
	placement, _ := place(t, dest, "inner")
	assert.Equal(t, 7.0, out(t, placement))
}

func mustInputs(t *testing.T, inst *graph.Instance, id string) []string {
	t.Helper()
	ports, err := inst.CollectBoundaryPorts(id)
	require.NoError(t, err)
	var inputs []string
	for _, port := range ports {
		if port.Kind == graph.BoundaryPortKindInput {
			inputs = append(inputs, port.Name)
		}
	}
	return inputs
}

func TestImportOfAnUnknownNodeTypeFailsWithoutPanickingOrLeftovers(t *testing.T) {
	dest := importTestInstance(t, "Dest")

	_, err := dest.ImportSubGraphDefinitions(rawPayload(`{
		"fine": {"name": "Fine", "nodes": `+outputOnlyNodes+`},
		"odd": {"name": "Odd", "nodes": {"x": {"type": "no.such.Node"}}}
	}`), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no.such.Node")
	assert.Empty(t, subGraphIDs(dest))
}

func TestALoadThatFailsPartwayLeavesTheOpenGraphAsItWas(t *testing.T) {
	for name, payload := range map[string][]byte{
		"unknown node type": rawPayload(`{"odd": {"name": "Odd", "nodes": {"x": {"type": "no.such.Node"}}}}`),
		"unknown variable":  rawPayload(`{"odd": {"name": "Odd", "nodes": {"v": {"type": "Ghost", "variable": "Ghost"}}}}`),
	} {
		t.Run(name, func(t *testing.T) {
			inst := importTestInstance(t, "Open")
			floatVariable(t, inst, "Size", 7)
			readsVariable(t, inst, "mine", "Size")
			before := string(encode(t, inst))

			require.Error(t, inst.ApplyAppSchema(payload))

			assert.Equal(t, before, string(encode(t, inst)))
			placement, _ := place(t, inst, "mine")
			assert.Equal(t, 7.0, out(t, placement))
		})
	}
}

func newBoolVariable(value bool) *variable.TypeVariable[bool] {
	v := &variable.TypeVariable[bool]{}
	v.SetValue(value)
	return v
}
