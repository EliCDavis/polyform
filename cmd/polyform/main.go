package main

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/nodes"
)

func main() {
	app := generator.App{
		Name:        "Polyform",
		Version:     "0.0.1",
		Description: "",
		Producers: map[string]nodes.NodeOutput[artifact.Artifact]{
			"model.glb": &artifact.GltfArtifact{},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
