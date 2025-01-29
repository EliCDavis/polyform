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

type ReadPointsNode = nodes.Struct[modeling.Mesh, ReadPointsNodeData]

type ReadPointsNodeData struct {
	In nodes.NodeOutput[[]byte]
}

func (pn ReadPointsNodeData) Process() (modeling.Mesh, error) {
	if pn.In == nil {
		return modeling.EmptyMesh(modeling.PointTopology), nil
	}
	return ReadSparsePointData(bytes.NewReader(pn.In.Value()))
}
