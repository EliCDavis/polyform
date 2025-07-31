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

	out := nodes.StructOutput[modeling.Mesh]{}

	data, err := ReadSparsePointData(bytes.NewReader(nodes.GetOutputValue(&out, pn.In)))
	out.Set(data)
	out.CaptureError(err)
	return out
}
