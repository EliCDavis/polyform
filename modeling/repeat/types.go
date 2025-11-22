package repeat

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[MeshNode]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode]](factory)
	refutil.RegisterType[nodes.Struct[SplineNode]](factory)
	refutil.RegisterType[nodes.Struct[LineNode]](factory)
	refutil.RegisterType[nodes.Struct[FibonacciSphereNode]](factory)
	refutil.RegisterType[nodes.Struct[FibonacciSpiralNode]](factory)
	refutil.RegisterType[nodes.Struct[TRSNode]](factory)
	refutil.RegisterType[nodes.Struct[TransformationNode]](factory)
	refutil.RegisterType[nodes.Struct[SampleMeshSurfaceNode]](factory)
	refutil.RegisterType[nodes.Struct[polygonNode]](factory)
	refutil.RegisterType[nodes.Struct[GridNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomPointsInSphereNode]](factory)
	generator.RegisterTypes(factory)
}
