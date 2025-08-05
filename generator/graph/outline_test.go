package graph_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

func TestOutline_NoVariablesOrProfiles(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	instance := graph.New(graph.Config{
		TypeFactory: factory,
	})

	instance.SetDetails(graph.Details{
		Name:        "My Name",
		Version:     "test version",
		Description: "A good description",
	})

	strParam := &parameter.String{
		Name:         "Welp",
		Description:  "I'm a description",
		CurrentValue: "bruh",
	}

	textNode := nodes.Struct[basics.TextNode]{
		Data: basics.TextNode{
			In: nodes.GetNodeOutputPort[string](strParam, "Value"),
		},
	}

	instance.AddProducer("test.txt", nodes.GetNodeOutputPort[manifest.Manifest](&textNode, "Out"))

	out := &bytes.Buffer{}

	// ACT ====================================================================
	err := graph.WriteOutline(instance, out)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `# My Name

test version

A good description

## Variables

(none)

## Profiles

(none)

`, out.String())
}
