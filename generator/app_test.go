package generator_test

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func buildTextArifact(p *parameter.String) nodes.Output[manifest.Manifest] {
	return nodes.GetNodeOutputPort[manifest.Manifest](
		&nodes.Struct[basics.TextNode]{
			Data: basics.TextNode{
				In: nodes.GetNodeOutputPort[string](p, "Value"),
			},
		},
		"Out",
	)
}

func TestGetAndApplyGraph(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	app := generator.App{
		Graph: graph.New(graph.Config{
			Name:        appName,
			Version:     appVersion,
			Description: appDescription,
		}),
	}

	// ACT ====================================================================
	graphData := app.Schema()
	err := app.ApplySchema(graphData)
	graphAgain := app.Schema()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, string(graphData), string(graphAgain))
}

func TestAppCommand_Zip(t *testing.T) {
	appName := "Test Graph"
	appVersion := "Test Graph"
	appDescription := "Test Graph"
	producerFileName := "test"

	outBuf := &bytes.Buffer{}

	g := graph.New(graph.Config{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
	})

	g.AddProducer(producerFileName, buildTextArifact(&parameter.String{
		Name:         "Welp",
		CurrentValue: "yee",
	}))

	app := generator.App{
		Graph: g,
		Out:   outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{"polyform", "zip"})
	data := outBuf.Bytes()

	r, zipErr := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, zipErr)
	assert.Len(t, r.File, 1)
	assert.Equal(t, "test/text.txt", r.File[0].Name)

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
	producerFileName := "test"

	outBuf := &bytes.Buffer{}

	g := graph.New(graph.Config{
		Name:        appName,
		Version:     appVersion,
		Description: appDescription,
	})

	g.AddProducer(producerFileName, buildTextArifact(&parameter.String{
		Name:         "Welp",
		CurrentValue: "yee",
		Description:  "I'm a description",
	}))

	app := generator.App{
		Graph: g,
		Out:   outBuf,
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
        "/manifest/test/Out": {
            "post": {
                "summary": "",
                "description": "",
                "produces": [
                    "application/json"
                ],
                "consumes": [
                    "application/json"
                ],
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "$ref": "#/definitions/CreateManifestResponse"
                        }
                    }
                },
                "parameters": [
                    {
                        "in": "body",
                        "name": "Request",
                        "schema": {
                            "$ref": "#/definitions/VariableProfile"
                        }
                    }
                ]
            }
        },
        "/manifest/{id}/{entry}": {
            "get": {
                "summary": "",
                "description": "",
                "produces": [
                    "application/octet-stream"
                ],
                "consumes": null,
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "type": "file"
                        }
                    }
                },
                "parameters": [
                    {
                        "in": "path",
                        "name": "id",
                        "description": "ID of the produced manifest to query",
                        "required": true,
                        "type": "string"
                    },
                    {
                        "in": "path",
                        "name": "entry",
                        "description": "Entry in the produced manifest to retrieve",
                        "required": true,
                        "type": "string"
                    }
                ]
            }
        },
        "/profile": {
            "get": {
                "summary": "",
                "description": "",
                "produces": [
                    "application/json"
                ],
                "consumes": null,
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                },
                "parameters": null
            }
        }
    },
    "definitions": {
        "CreateManifestResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "manifest": {
                    "$ref": "#/definitions/Manifest"
                }
            }
        },
        "Entries": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/Entry"
            }
        },
        "Entry": {
            "type": "object",
            "properties": {
                "metadata": {
                    "$ref": "#/definitions/Metadata"
                }
            }
        },
        "Manifest": {
            "type": "object",
            "properties": {
                "entries": {
                    "$ref": "#/definitions/Entries"
                },
                "main": {
                    "type": "string"
                }
            }
        },
        "Metadata": {
            "type": "object",
            "additionalProperties": true
        },
        "VariableProfile": {
            "type": "object"
        }
    }
}`, string(contents))
}

func TestAppCommand_New(t *testing.T) {
	outBuf := &bytes.Buffer{}

	app := generator.App{
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
	"nodes": null,
	"variables": {
		"variables": null,
		"subgroups": null
	}
}`, string(contents))
}

func TestAppCommand_Help(t *testing.T) {
	outBuf := &bytes.Buffer{}

	app := generator.App{
		Name:        "Test App",
		Version:     "test",
		Description: "This is just a test app",
		Authors: []schema.Author{
			{
				Name: "Test Runner",
				ContactInfo: []schema.AuthorContact{
					{
						Medium: "package",
						Value:  "testing",
					},
				},
			},
		},
		Out: outBuf,
	}

	// ACT ====================================================================
	err := app.Run([]string{
		"polyform", "help",
	})
	contents, readErr := io.ReadAll(outBuf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, readErr)
	assert.Equal(t, `Test App test
    This is just a test app

AUTHORS:
    Test Runner
        package - testing
        
COMMANDS:
    New: new 
        Create a new graph
    Generate: generate gen 
        Runs all producers the graph has defined and saves it to the file system
    Edit: edit 
        Starts an http server and hosts a webplayer for editing the execution graph
    Serve: serve 
        Starts a 'production' server meant for taking requests for executing a certain graph
    Zip: zip z 
        Runs all producers defined and writes it to a zip file
    Mermaid: mermaid 
        Create a mermaid flow chart for a specific producer
    Documentation: documentation docs 
        Create a document describing all available nodes
    Swagger: swagger 
        Create a swagger 2.0 file
    Outline: outline 
        outline the data embedded in a graph
    Help: help h 
        
    `, string(contents))
}
