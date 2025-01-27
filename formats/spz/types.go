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

type ReadNode = nodes.Struct[modeling.Mesh, ReadNodeData]

type ReadNodeData struct {
	Data nodes.NodeOutput[[]byte]
}

func (gad ReadNodeData) Process() (modeling.Mesh, error) {
	if gad.Data == nil {
		return modeling.EmptyMesh(modeling.PointTopology), nil
	}

	data := gad.Data.Value()
	if len(data) == 0 {
		return modeling.EmptyMesh(modeling.PointTopology), nil
	}

	cloud, err := Read(bytes.NewReader(data))
	if err != nil {
		return modeling.EmptyMesh(modeling.PointTopology), err
	}

	return cloud.Mesh, nil
}
