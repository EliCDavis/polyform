package graph_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"
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
	schema := graph.BuildNodeTypeSchema(&TestNode{})

	assert.Equal(t, "TestNodeData", schema.DisplayName)
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
	instance := graph.New(factory)
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
	instance.AddProducer("test.txt", nodes.GetNodeOutputPort[artifact.Artifact](&textNode, "Out"))
	producerNames := instance.ProducerNames()
	instance.InitializeParameters(flags)
	assert.NoError(t, flags.Parse([]string{"-yeet", contentToSetViaFlag}))
	textArtifact := instance.Artifact("test.txt")
	buf := &bytes.Buffer{}
	assert.NoError(t, textArtifact.Write(buf))

	instanceSchema := instance.Schema()
	instanceSchemaData, err := json.MarshalIndent(instanceSchema, "", "\t")
	assert.NoError(t, err)

	appSchema := &schema.App{}
	encoder := &jbtf.Encoder{}
	instance.EncodeToAppSchema(appSchema, encoder)
	appSchemaData, err := encoder.ToPgtf(appSchema)
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.Len(t, producerNames, 1)
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
			"version": 0,
			"dependencies": [],
			"parameter": {
				"name": "Welp",
				"description": "I'm a description",
				"type": "string",
				"defaultValue": "yee",
				"currentValue": "bruh"
			}
		},
		"Node-1": {
			"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/artifact.Artifact,github.com/EliCDavis/polyform/generator/artifact/basics.TextNodeData]",
			"name": "test.txt",
			"version": 1,
			"dependencies": [
				{
					"dependencyID": "Node-0",
					"dependencyPort": "Out",
					"name": "In"
				}
			]
		}
	},
	"types": [
		{
			"displayName": "parameter.Value[string]",
			"info": "",
			"type": "github.com/EliCDavis/polyform/generator/parameter.Value[string]",
			"path": "generator/parameter",
			"outputs": [
				{
					"name": "Out",
					"type": "string"
				}
			],
			"parameter": {
				"name": "",
				"description": "",
				"type": "string",
				"defaultValue": "",
				"currentValue": ""
			}
		},
		{
			"displayName": "TextNodeData",
			"info": "",
			"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/artifact.Artifact,github.com/EliCDavis/polyform/generator/artifact/basics.TextNodeData]",
			"path": "generator/artifact/basics",
			"outputs": [
				{
					"name": "Out",
					"type": "github.com/EliCDavis/polyform/generator/artifact.Artifact"
				}
			],
			"inputs": {
				"In": {
					"type": "string",
					"isArray": false
				}
			}
		}
	],
	"notes": null
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
				"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/artifact.Artifact,github.com/EliCDavis/polyform/generator/artifact/basics.TextNodeData]",
				"dependencies": [
					{
						"dependencyID": "Node-0",
						"dependencyPort": "Out",
						"name": "In"
					}
				]
			}
		},
		"producers": {
			"test.txt": {
				"nodeID": "Node-1",
				"port": "Out"
			}
		}
	}
}`, string(appSchemaData))
}
