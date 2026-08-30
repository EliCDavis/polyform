package generator_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawStringForTest(t *testing.T, v string) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestAppCommand_Sweep(t *testing.T) {
	g := graph.New(graph.Config{Name: "Test Graph", TypeFactory: &refutil.TypeFactory{}})

	messageVar := &variable.TypeVariable[string]{}
	g.NewVariable("Message", messageVar)

	g.AddProducer("output", nodes.GetNodeOutputPort[manifest.Manifest](
		&nodes.Struct[basics.TextNode]{
			Data: basics.TextNode{
				In: nodes.GetNodeOutputPort[string](messageVar.NodeReference(), "Value"),
			},
		},
		"Out",
	))

	require.NoError(t, g.SetVariantSet("sweep", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewDiscrete("Message", rawStringForTest(t, "Hello"), rawStringForTest(t, "World")),
		},
	}))

	tempDir := t.TempDir()
	outBuf := &bytes.Buffer{}
	app := generator.App{Graph: g, Out: outBuf}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "sweep", "-set", "sweep", "-folder", tempDir})

	// ASSERT =================================================================
	require.NoError(t, err)

	hello, err := os.ReadFile(filepath.Join(tempDir, "variant-0000", "output", "text.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Hello", string(hello))

	world, err := os.ReadFile(filepath.Join(tempDir, "variant-0001", "output", "text.txt"))
	require.NoError(t, err)
	assert.Equal(t, "World", string(world))

	_, err = os.Stat(filepath.Join(tempDir, "variant-0002"))
	assert.True(t, os.IsNotExist(err), "expected only 2 variants to be written")

	helloProfile, err := os.ReadFile(filepath.Join(tempDir, "variant-0000", "profile.json"))
	require.NoError(t, err)
	var decodedHelloProfile variable.Profile
	require.NoError(t, json.Unmarshal(helloProfile, &decodedHelloProfile))
	assert.JSONEq(t, `"Hello"`, string(decodedHelloProfile["Message"]))

	worldProfile, err := os.ReadFile(filepath.Join(tempDir, "variant-0001", "profile.json"))
	require.NoError(t, err)
	var decodedWorldProfile variable.Profile
	require.NoError(t, json.Unmarshal(worldProfile, &decodedWorldProfile))
	assert.JSONEq(t, `"World"`, string(decodedWorldProfile["Message"]))
}

func TestAppCommand_Sweep_OverThresholdRequiresConfirm(t *testing.T) {
	g := graph.New(graph.Config{Name: "Test Graph", TypeFactory: &refutil.TypeFactory{}})

	valueVar := &variable.TypeVariable[float64]{}
	g.NewVariable("Value", valueVar)

	g.AddProducer("output", nodes.GetNodeOutputPort[manifest.Manifest](
		&nodes.Struct[basics.TextNode]{
			Data: basics.TextNode{
				In: nodes.GetNodeOutputPort[string](&parameter.String{
					Name:         "Unused",
					CurrentValue: "unused",
				}, "Value"),
			},
		},
		"Out",
	))

	require.NoError(t, g.SetVariantSet("huge", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewNumericRange("Value", 0, 1, 1001),
		},
	}))

	tempDir := t.TempDir()
	outBuf := &bytes.Buffer{}
	app := generator.App{Graph: g, Out: outBuf}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "sweep", "-set", "huge", "-folder", tempDir})

	// ASSERT =================================================================
	assert.Error(t, err)

	_, statErr := os.Stat(filepath.Join(tempDir, "variant-0000"))
	assert.True(t, os.IsNotExist(statErr), "no variants should have been written without -confirm")
}

func TestAppCommand_Sample(t *testing.T) {
	g := graph.New(graph.Config{Name: "Test Graph", TypeFactory: &refutil.TypeFactory{}})

	messageVar := &variable.TypeVariable[string]{}
	g.NewVariable("Message", messageVar)

	g.AddProducer("output", nodes.GetNodeOutputPort[manifest.Manifest](
		&nodes.Struct[basics.TextNode]{
			Data: basics.TextNode{
				In: nodes.GetNodeOutputPort[string](messageVar.NodeReference(), "Value"),
			},
		},
		"Out",
	))

	require.NoError(t, g.SetVariantSet("sample", variant.Set{
		Dimensions: []variant.Dimension{
			variant.NewDiscrete("Message", rawStringForTest(t, "Hello"), rawStringForTest(t, "World")),
		},
	}))

	tempDir := t.TempDir()
	outBuf := &bytes.Buffer{}
	app := generator.App{Graph: g, Out: outBuf}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "sample", "-set", "sample", "-count", "3", "-seed", "42", "-folder", tempDir})

	// ASSERT =================================================================
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		folder := filepath.Join(tempDir, "variant-000"+string(rune('0'+i)), "output", "text.txt")
		contents, err := os.ReadFile(folder)
		require.NoError(t, err)
		assert.Contains(t, []string{"Hello", "World"}, string(contents))
	}
}
