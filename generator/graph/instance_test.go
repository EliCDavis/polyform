package graph_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

type TestNode = nodes.Struct[TestNodeData]

type TestNodeData struct {
	A nodes.Output[float64]
	B nodes.Output[int]
}

func (bn TestNodeData) Out() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(0.)
}

func TestBuildNodeTypeSchema(t *testing.T) {
	schema := graph.BuildNodeTypeSchema("", &TestNode{})

	assert.Equal(t, "Test", schema.DisplayName)
	assert.Equal(t, "generator/graph_test", schema.Path)

	assert.Len(t, schema.Inputs, 2)
	assert.Equal(t, "float64", schema.Inputs["A"].Type)
	assert.Equal(t, "int", schema.Inputs["B"].Type)

	assert.Len(t, schema.Outputs, 1)
	assert.Equal(t, "float64", schema.Outputs["Out"].Type)
}

func TestInstance_AddProducer_InitializeParameters_Artifacts(t *testing.T) {
	// ARRANGE ================================================================
	contentToSetViaFlag := "bruh"
	factory := &refutil.TypeFactory{}
	instance := graph.New(graph.Config{
		TypeFactory: factory,
	})
	assert.Len(t, instance.ProducerNames(), 0)
	flags := flag.NewFlagSet("set", flag.PanicOnError)

	strParam := &parameter.String{
		Name: "Welp",
		CLI: &parameter.CliConfig[string]{
			FlagName: "yeet",
			Usage:    "I'm the flag description",
		},
		DefaultValue: "yee",
		Description:  "I'm a description",
	}

	textNode := basics.TextNode{
		Data: basics.TextNodeData{
			In: nodes.GetNodeOutputPort[string](strParam, "Value"),
		},
	}

	// ACT ====================================================================
	instance.AddProducer("test.txt", nodes.GetNodeOutputPort[manifest.Manifest](&textNode, "Out"))
	producerNames := instance.ProducerNames()
	instance.InitializeParameters(flags)
	assert.NoError(t, flags.Parse([]string{"-yeet", contentToSetViaFlag}))
	textManifest := instance.Manifest("test.txt")

	buf := &bytes.Buffer{}
	assert.NoError(t, textManifest.Entries[textManifest.Main].Artifact.Write(buf))

	instanceSchema := instance.Schema()
	instanceSchemaData, err := json.MarshalIndent(instanceSchema, "", "\t")
	assert.NoError(t, err)

	appSchema, err := instance.EncodeToAppSchema()
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.Len(t, producerNames, 1)
	assert.Len(t, textManifest.Entries, 1)
	assert.Equal(t, "test.txt", producerNames[0])
	assert.Equal(t, contentToSetViaFlag, buf.String())
	assert.Equal(t, `{
	"producers": {
		"test.txt": {
			"nodeID": "Node-1",
			"port": "Out"
		}
	},
	"nodes": {
		"Node-0": {
			"type": "github.com/EliCDavis/polyform/generator/parameter.Value[string]",
			"name": "Welp",
			"assignedInput": {},
			"output": {
				"Value": {
					"version": 0
				}
			},
			"parameter": {
				"name": "Welp",
				"description": "I'm a description",
				"type": "string",
				"defaultValue": "yee",
				"currentValue": "bruh"
			}
		},
		"Node-1": {
			"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/manifest/basics.TextNodeData]",
			"name": "test.txt",
			"assignedInput": {
				"In": {
					"id": "Node-0",
					"port": "Value"
				}
			},
			"output": {
				"Out": {
					"version": 0
				}
			}
		}
	},
	"notes": null,
	"variables": {
		"variables": {},
		"subgroups": {}
	}
}`, string(instanceSchemaData))

	assert.Equal(t, `{
	"buffers": [
		{
			"byteLength": 0,
			"uri": "data:application/octet-stream;base64,"
		}
	],
	"data": {
		"nodes": {
			"Node-0": {
				"type": "github.com/EliCDavis/polyform/generator/parameter.Value[string]",
				"data": {
					"name": "Welp",
					"description": "I'm a description",
					"currentValue": "bruh",
					"defaultValue": "yee",
					"cli": {
						"flagName": "yeet",
						"usage": "I'm the flag description"
					}
				}
			},
			"Node-1": {
				"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/manifest/basics.TextNodeData]",
				"assignedInput": {
					"In": {
						"id": "Node-0",
						"port": "Value"
					}
				}
			}
		},
		"producers": {
			"test.txt": {
				"nodeID": "Node-1",
				"port": "Out"
			}
		},
		"variables": {
			"subgroups": {},
			"variables": {}
		}
	}
}`, string(appSchema))
}
