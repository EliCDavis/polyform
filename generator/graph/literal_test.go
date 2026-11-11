package graph_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type literalOpaque struct{ Weight float64 }

type literalVar struct{}

func (literalVar) TypeVariable() string { return "T" }

type literalSinkNode struct {
	Radius nodes.Output[float64] `default:"0.5"`
	Count  nodes.Output[int]
	Name   nodes.Output[string]          `default:"shape"`
	Offset nodes.Output[vector3.Float64] `default:"1,2,3"`
	Scale  nodes.LiftedPort[float64]     `default:"2"`
	Items  []nodes.Output[float64]
	Thing  nodes.Output[literalOpaque]
	Any    nodes.DynamicPort[literalVar]
}

func (n literalSinkNode) Out(out *nodes.StructOutput[float64]) {
	radius := nodes.TryGetOutputValue(out, n.Radius, 0.5)
	count := nodes.TryGetOutputValue(out, n.Count, 0)
	offset := nodes.TryGetOutputValue(out, n.Offset, vector3.Zero[float64]())
	out.Set(radius*float64(count) + offset.X())
}

func (n literalSinkNode) Label(out *nodes.StructOutput[string]) {
	out.Set(nodes.TryGetOutputValue(out, n.Name, ""))
}

type literalSourceNode struct{}

func (literalSourceNode) Out(out *nodes.StructOutput[float64]) { out.Set(10) }

type literalWordNode struct{}

func (literalWordNode) Out(out *nodes.StructOutput[string]) { out.Set("wired") }

const (
	literalSinkType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.literalSinkNode]"
	literalSourceType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.literalSourceNode]"
	literalWordType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.literalWordNode]"
)

func literalInstance(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[literalSinkNode]](factory)
	refutil.RegisterType[nodes.Struct[literalSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[literalWordNode]](factory)
	return graph.New(graph.Config{TypeFactory: factory, VariableFactory: variabletypes.New})
}

func literalSink(t *testing.T) (*graph.Instance, nodes.Node, string) {
	t.Helper()
	instance := literalInstance(t)
	sink, id, err := instance.CreateNode(literalSinkType)
	require.NoError(t, err)
	return instance, sink, id
}

func sinkOut(sink nodes.Node) float64 {
	return nodes.GetNodeOutputPort[float64](sink, "Out").Value()
}

func literalsOf(instance *graph.Instance, id string) map[string]string {
	shown := map[string]string{}
	for port, raw := range instance.Schema().Nodes[id].Literals {
		shown[port] = string(raw)
	}
	return shown
}

func TestANewNodeHoldsALiteralPerPrimitiveInput(t *testing.T) {
	instance, sink, id := literalSink(t)

	assert.Equal(t, map[string]string{
		"Radius": "0.5",
		"Count":  "0",
		"Name":   `"shape"`,
		"Offset": `{"x":1,"y":2,"z":3}`,
		"Scale":  "2",
	}, literalsOf(instance, id), "arrays, dynamic ports and unregistered types hold none")

	assert.Equal(t, 1., sinkOut(sink), "0.5 * 0 + offset.x")
	assert.Equal(t, "shape", nodes.GetNodeOutputPort[string](sink, "Label").Value())
	assert.Empty(t, instance.Schema().Nodes[id].AssignedInput, "a literal is not an edge")
	assert.Len(t, instance.Schema().Nodes, 1, "a literal is not a node")
}

func TestSettingALiteralChangesTheOutput(t *testing.T) {
	instance, sink, id := literalSink(t)
	before := nodes.GetNodeOutputPort[float64](sink, "Out").Version()

	require.NoError(t, instance.SetLiteral(id, "Count", []byte("4")))
	require.NoError(t, instance.SetLiteral(id, "Radius", []byte("1.5")))

	assert.Equal(t, 7., sinkOut(sink))
	assert.NotEqual(t, before, nodes.GetNodeOutputPort[float64](sink, "Out").Version())
	assert.Equal(t, "4", literalsOf(instance, id)["Count"])

	value, err := instance.LiteralValue(id, "Count")
	require.NoError(t, err)
	assert.Equal(t, "4", string(value))
}

func TestSettingALiteralOnANodeMadeInTheSameBatch(t *testing.T) {
	instance := literalInstance(t)

	var sink nodes.Node
	_, err := instance.Batch(func() error {
		var id string
		var err error
		if sink, id, err = instance.CreateNode(literalSinkType); err != nil {
			return err
		}
		return instance.SetLiteral(id, "Count", []byte("4"))
	})
	require.NoError(t, err)

	assert.Equal(t, 3., sinkOut(sink), "0.5 * 4 + offset.x")
}

func TestAPortWithNoLiteralSaysSo(t *testing.T) {
	instance, _, id := literalSink(t)
	assert.ErrorIs(t, instance.SetLiteral(id, "Thing", []byte("{}")), graph.ErrNoLiteral)
}

func TestSettingALiteralRefusesTheWrongShape(t *testing.T) {
	instance, _, id := literalSink(t)

	assert.Error(t, instance.SetLiteral(id, "Count", []byte(`"four"`)))
	assert.Error(t, instance.SetLiteral(id, "Items", []byte("1")), "an array input has no literal")
	assert.Error(t, instance.SetLiteral(id, "Thing", []byte("1")), "an unregistered type has no literal")
	assert.Error(t, instance.SetLiteral(id, "Nope", []byte("1")))
	assert.Error(t, instance.SetLiteral("Node-9", "Count", []byte("1")))
	assert.Equal(t, "0", literalsOf(instance, id)["Count"])
}

func TestAWireHidesTheLiteralAndUnwiringBringsItBack(t *testing.T) {
	instance, sink, id := literalSink(t)
	require.NoError(t, instance.SetLiteral(id, "Count", []byte("2")))
	require.NoError(t, instance.SetLiteral(id, "Radius", []byte("3")))
	_, source, err := instance.CreateNode(literalSourceType)
	require.NoError(t, err)

	require.NoError(t, instance.ConnectNodes(source, "Out", id, "Radius"))
	assert.Equal(t, 21., sinkOut(sink), "10 * 2 + 1")
	assert.Equal(t, "3", literalsOf(instance, id)["Radius"], "the literal keeps its value under the wire")

	_, err = instance.DeleteNodeInputConnection(id, "Radius")
	require.NoError(t, err)
	assert.Equal(t, 7., sinkOut(sink), "3 * 2 + 1")
}

func TestARefusedWireLeavesTheLiteralInPlace(t *testing.T) {
	instance, sink, id := literalSink(t)
	_, word, err := instance.CreateNode(literalWordType)
	require.NoError(t, err)

	assert.Error(t, instance.ConnectNodes(word, "Out", id, "Radius"))
	assert.Equal(t, 1., sinkOut(sink))
	assert.Equal(t, "0.5", literalsOf(instance, id)["Radius"])
}

func TestLiteralsSurviveSaveAndLoad(t *testing.T) {
	instance, _, id := literalSink(t)
	require.NoError(t, instance.SetLiteral(id, "Count", []byte("3")))
	require.NoError(t, instance.SetLiteral(id, "Name", []byte(`"saved"`)))

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)
	assert.Contains(t, string(saved), `"literals"`)

	loaded := literalInstance(t)
	require.NoError(t, loaded.ApplyAppSchema(saved))
	assert.Equal(t, "3", literalsOf(loaded, id)["Count"])
	assert.Equal(t, `"saved"`, literalsOf(loaded, id)["Name"])
	assert.Equal(t, "0.5", literalsOf(loaded, id)["Radius"], "an untouched literal is saved at its default")
	assert.Equal(t, 2.5, sinkOut(loaded.Node(id)))
}

func savedNodeIn(t *testing.T, document any, id string) map[string]any {
	t.Helper()
	var found map[string]any
	var walk func(v any)
	walk = func(v any) {
		m, ok := v.(map[string]any)
		if !ok {
			return
		}
		if nodes, ok := m["nodes"].(map[string]any); ok {
			if node, ok := nodes[id].(map[string]any); ok {
				found = node
				return
			}
		}
		for _, child := range m {
			walk(child)
		}
	}
	walk(document)
	require.NotNil(t, found, "saved node %s", id)
	return found
}

func TestAMissingSavedLiteralTakesTheDefault(t *testing.T) {
	instance, _, id := literalSink(t)
	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	var app map[string]any
	require.NoError(t, json.Unmarshal(saved, &app))
	delete(savedNodeIn(t, app, id), "literals")
	stripped, err := json.Marshal(app)
	require.NoError(t, err)

	loaded := literalInstance(t)
	require.NoError(t, loaded.ApplyAppSchema(stripped))
	assert.Equal(t, "0.5", literalsOf(loaded, id)["Radius"])
}

func TestASavedLiteralOfTheWrongShapeFailsTheLoad(t *testing.T) {
	instance, _, id := literalSink(t)
	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	var app map[string]any
	require.NoError(t, json.Unmarshal(saved, &app))
	savedNodeIn(t, app, id)["literals"].(map[string]any)["Count"] = "three"
	broken, err := json.Marshal(app)
	require.NoError(t, err)

	loaded := literalInstance(t)
	assert.ErrorContains(t, loaded.ApplyAppSchema(broken), "Count")
}

func TestALiteralEditIsUndoable(t *testing.T) {
	instance, sink, id := literalSink(t)

	require.NoError(t, instance.History().Transact("count", func() error {
		return instance.SetLiteral(id, "Count", []byte("5"))
	}))
	assert.Equal(t, 3.5, sinkOut(sink))

	_, err := instance.History().Undo()
	require.NoError(t, err)
	assert.Equal(t, 1., sinkOut(instance.Node(id)))

	_, err = instance.History().Redo()
	require.NoError(t, err)
	assert.Equal(t, 3.5, sinkOut(instance.Node(id)))
}

func TestDeletingANodeDropsItsLiterals(t *testing.T) {
	instance, _, id := literalSink(t)
	_, err := instance.DeleteNodeById(id)
	require.NoError(t, err)

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)
	assert.NotContains(t, string(saved), "literals")
}

func TestWithoutAVariableFactoryThereAreNoLiterals(t *testing.T) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[literalSinkNode]](factory)
	instance := graph.New(graph.Config{TypeFactory: factory})

	sink, id, err := instance.CreateNode(literalSinkType)
	require.NoError(t, err)
	assert.Empty(t, literalsOf(instance, id))
	assert.Equal(t, 1., sinkOut(sink), "the node fills its own tag defaults: 0.5 * 0 + offset.x")
	assert.Empty(t, instance.Schema().Nodes[id].AssignedInput, "a self-filled default is not an edge")
}

func TestNodeTypeSchemaNamesTheLiteralInputs(t *testing.T) {
	instance := literalInstance(t)

	for _, typeSchema := range instance.BuildSchemaForAllNodeTypes() {
		if typeSchema.Type != literalSinkType {
			continue
		}
		assert.True(t, typeSchema.Inputs["Radius"].Literal)
		assert.Equal(t, "0.5", string(typeSchema.Inputs["Radius"].Default))
		assert.True(t, typeSchema.Inputs["Count"].Literal)
		assert.Equal(t, "0", string(typeSchema.Inputs["Count"].Default))
		assert.True(t, typeSchema.Inputs["Scale"].Literal)
		assert.False(t, typeSchema.Inputs["Items"].Literal)
		assert.False(t, typeSchema.Inputs["Thing"].Literal)
		assert.False(t, typeSchema.Inputs["Any"].Literal)
		return
	}
	t.Fatal("sink type not listed")
}
