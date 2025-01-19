package main

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/nodes"

	// Import these so they register their nodes with the generator
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/formats/ply"
	_ "github.com/EliCDavis/polyform/generator/artifact/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	_ "github.com/EliCDavis/polyform/modeling/extrude"
	_ "github.com/EliCDavis/polyform/modeling/meshops"
	_ "github.com/EliCDavis/polyform/modeling/meshops/gausops"
	_ "github.com/EliCDavis/polyform/modeling/primitives"
	_ "github.com/EliCDavis/polyform/modeling/repeat"
	_ "github.com/EliCDavis/polyform/nodes/experimental"
)

func main() {
	app := generator.App{
		Name:        "Polyform",
		Version:     "0.0.1",
		Description: "",
		Producers:   map[string]nodes.NodeOutput[artifact.Artifact]{},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
