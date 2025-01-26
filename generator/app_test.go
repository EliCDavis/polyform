package generator_test

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/artifact/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestGetAndApplyGraph(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"
	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
			}),
		},
	}

	// ACT ====================================================================
	graphData := app.Schema()
	err := app.ApplySchema(graphData)
	graphAgain := app.Schema()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, appName, app.Name)
	assert.Equal(t, appVersion, app.Version)
	assert.Equal(t, appDescription, app.Description)
	assert.Equal(t, string(graphData), string(graphAgain))
	b := &bytes.Buffer{}
	art := app.Files[producerFileName].Value()
	err = art.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, "yee", b.String())
}

func TestAppCommand_Outline(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"

	outBuf := &bytes.Buffer{}

	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
			}),
		},

		Out: outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "outline"})
	contents, readErr := io.ReadAll(outBuf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, readErr)
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
                "description": "",
                "type": "string",
                "defaultValue": "yee",
                "currentValue": "yee"
            }
        },
        "Node-1": {
            "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/artifact.Artifact,github.com/EliCDavis/polyform/generator/artifact/basics.TextNodeData]",
            "name": "test.txt",
            "version": 0,
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
}`, string(contents))
}

func TestAppCommand_Zip(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"

	outBuf := &bytes.Buffer{}

	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
			}),
		},

		Out: outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "zip"})
	data := outBuf.Bytes()

	r, zipErr := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, zipErr)
	assert.Len(t, r.File, 1)
	assert.Equal(t, "test.txt", r.File[0].Name)

	rc, err := r.File[0].Open()
	assert.NoError(t, err)

	buf, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.Equal(t, "yee", string(buf))
}

func TestAppCommand_Swagger(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"

	outBuf := &bytes.Buffer{}

	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
				Description:  "I'm a description",
			}),
		},

		Out: outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "swagger"})
	contents, readErr := io.ReadAll(outBuf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, readErr)
	assert.Equal(t, `{
    "swagger": "2.0",
    "info": {
        "title": "Test Graph",
        "description": "Test Graph",
        "version": "Test Graph"
    },
    "paths": {
        "/producer/value/test.txt": {
            "post": {
                "summary": "",
                "description": "",
                "produces": [],
                "consumes": [
                    "application/json"
                ],
                "responses": {
                    "200": {
                        "description": "Producer Payload"
                    }
                },
                "parameters": [
                    {
                        "in": "body",
                        "name": "Request",
                        "schema": {
                            "$ref": "#/definitions/TestTxtRequest"
                        }
                    }
                ]
            }
        }
    },
    "definitions": {
        "TestTxtRequest": {
            "type": "object",
            "properties": {
                "Welp": {
                    "type": "string",
                    "description": "I'm a description"
                }
            }
        }
    }
}`, string(contents))
}

func TestAppCommand_New(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test.txt"

	outBuf := &bytes.Buffer{}

	app := generator.App{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
		Files: map[string]nodes.NodeOutput[artifact.Artifact]{
			producerFileName: basics.NewTextNode(&parameter.String{
				Name:         "Welp",
				DefaultValue: "yee",
			}),
		},

		Out: outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{
		"polyform", "new",
		"--name", "My New Graph",
		"--description", "This is just a test",
		"--version", "v1.0.2",
		"--author", "Test Runner",
	})
	contents, readErr := io.ReadAll(outBuf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, readErr)
	assert.Equal(t, `{
	"name": "My New Graph",
	"version": "v1.0.2",
	"description": "This is just a test",
	"authors": [
		{
			"name": "Test Runner"
		}
	],
	"producers": null,
	"nodes": null
}`, string(contents))
}
