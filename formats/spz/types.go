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

func (gad ReadNodeData) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.PointTopology))
	if gad.Data == nil {
		return
	}

	data := nodes.GetOutputValue(out, gad.Data)
	if len(data) == 0 {
		return
	}

	cloud, err := Read(bytes.NewReader(data))
	if err != nil {
		out.CaptureError(err)
		return
	}

	out.Set(cloud.Mesh)
}
