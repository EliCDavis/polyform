package opensfm

import (
	"bytes"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ReadReconstructionNode](factory)

	generator.RegisterTypes(factory)
}

type ReadReconstructionNode = nodes.Struct[modeling.Mesh, ReadReconstructionNodeData]

type ReadReconstructionNodeData struct {
	In nodes.NodeOutput[[]byte]
}

func (pn ReadReconstructionNodeData) Process() (modeling.Mesh, error) {
	if pn.In == nil {
		return modeling.EmptyMesh(modeling.PointTopology), nil
	}
	return ReadReconstructiontData(bytes.NewReader(pn.In.Value()))
}
