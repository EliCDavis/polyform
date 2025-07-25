package colmap

import (
	"bytes"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ReadPointsNode](factory)

	generator.RegisterTypes(factory)
}

type ReadPointsNode = nodes.Struct[ReadPointsNodeData]

type ReadPointsNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadPointsNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}

	data, err := ReadSparsePointData(bytes.NewReader(pn.In.Value()))
	out := nodes.NewStructOutput(data)
	out.CaptureError(err)
	return out
}
