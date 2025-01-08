package experimental

import (
	"github.com/EliCDavis/polyform/refutil"
)

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[TRSNode](factory)
	refutil.RegisterType[ShiftNode](factory)
	refutil.RegisterType[VectorArrayNode](factory)
	refutil.RegisterType[QuaternionArrayFromThetaNode](factory)
	refutil.RegisterType[BrushedMetalNode](factory)
	refutil.RegisterType[SinNode](factory)
	refutil.RegisterType[CosNode](factory)
	refutil.RegisterType[SampleNode](factory)
	refutil.RegisterType[SeamlessPerlinNode](factory)

	return factory
}
