package spz

import (
	"bytes"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[ReadNode](factory)
	generator.RegisterTypes(factory)
}

type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	Data nodes.Output[[]byte]
}

func (gad ReadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if gad.Data == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}

	data := gad.Data.Value()
	if len(data) == 0 {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}

	cloud, err := Read(bytes.NewReader(data))
	if err != nil {
		out := nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
		out.CaptureError(err)
		return out
	}

	return nodes.NewStructOutput(cloud.Mesh)
}
