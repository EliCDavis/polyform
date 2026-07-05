package register

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any {
		return subgraph.NewInputNode("", "")
	})
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any {
		return subgraph.NewOutputNode("", "")
	})

	generator.RegisterTypes(factory)

	// subgraph.RegisterInputOutputType[vector2.Float64]()
	// subgraph.RegisterInputOutputType[vector3.Float64]()
	// subgraph.RegisterInputOutputType[[]vector3.Float64]()
	// subgraph.RegisterInputOutputType[geometry.AABB]()
	// subgraph.RegisterInputOutputType[coloring.Color]()
	// subgraph.RegisterInputOutputType[image.Image]()
	// subgraph.RegisterInputOutputType[manifest.Manifest]()
	// subgraph.RegisterInputOutputType[modeling.Mesh]()
	// subgraph.RegisterInputOutputType[[]byte]()
}
