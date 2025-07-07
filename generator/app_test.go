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
		&basics.TextNode{
			Data: basics.TextNodeData{
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
	producerFileName := "test.txt"

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
        "ManifestRequest": {
            "type": "object",
            "properties": {}
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
    Zip: zip z 
        Runs all producers defined and writes it to a zip file 
    Mermaid: mermaid 
        Create a mermaid flow chart for a specific producer 
    Documentation: documentation docs 
        Create a document describing all available nodes 
    Swagger: swagger 
        Create a swagger 2.0 file 
    Help: help h 
         
    `, string(contents))
}
