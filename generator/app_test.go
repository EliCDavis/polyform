package generator_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type TestNode = nodes.Struct[float64, TestNodeData]

type TestNodeData struct {
	A nodes.NodeOutput[float64]
	B nodes.NodeOutput[int]
}

func (bn TestNodeData) Process() (float64, error) {
	return 0, nil
}

func TestBuildNodeTypeSchema(t *testing.T) {
	schema := generator.BuildNodeTypeSchema(&TestNode{})

	assert.Equal(t, "TestNodeData", schema.DisplayName)
	assert.Equal(t, "generator_test", schema.Path)

	assert.Len(t, schema.Inputs, 2)
	assert.Equal(t, "float64", schema.Inputs["A"].Type)
	assert.Equal(t, "int", schema.Inputs["B"].Type)

	assert.Len(t, schema.Outputs, 1)
	assert.Equal(t, "float64", schema.Outputs[0].Type)
	assert.Equal(t, "Out", schema.Outputs[0].Name)
}

func TestGetAndApplyGraph(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"
	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Producers: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
			}),
		},
	}

	app.SetupProducers()

	// ACT ====================================================================
	graphData := app.Graph()
	err := app.ApplyGraph(graphData)
	graphAgain := app.Graph()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, appName, app.Name)
	assert.Equal(t, appVersion, app.Version)
	assert.Equal(t, appDescription, app.Description)
	assert.Equal(t, string(graphData), string(graphAgain))
	b := &bytes.Buffer{}
	art := app.Producers[producerFileName].Value()
	err = art.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, "yee", b.String())
}
