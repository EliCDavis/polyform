package register

import (
	"image"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
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

	subgraph.RegisterInputOutputType[vector2.Float64]()
	subgraph.RegisterInputOutputType[vector3.Float64]()
	subgraph.RegisterInputOutputType[[]vector3.Float64]()
	subgraph.RegisterInputOutputType[geometry.AABB]()
	subgraph.RegisterInputOutputType[coloring.Color]()
	subgraph.RegisterInputOutputType[image.Image]()
	subgraph.RegisterInputOutputType[manifest.Manifest]()
	subgraph.RegisterInputOutputType[modeling.Mesh]()
	subgraph.RegisterInputOutputType[[]byte]()
}
