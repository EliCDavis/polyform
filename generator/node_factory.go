package generator

import (
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes/experimental"
	"github.com/EliCDavis/polyform/refutil"
)

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}
	return factory.Combine(
		meshops.Nodes(),
		gausops.Nodes(),
		parameter.Nodes(),
		experimental.Nodes(),
		artifact.Nodes(),
		repeat.Nodes(),
		primitives.Nodes(),
		extrude.Nodes(),
	)
}
