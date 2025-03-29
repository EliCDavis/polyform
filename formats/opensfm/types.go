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

type ReadReconstructionNode = nodes.Struct[ReadReconstructionNodeData]

type ReadReconstructionNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadReconstructionNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.PointTopology))
	}

	data, err := ReadReconstructiontData(bytes.NewReader(pn.In.Value()))

	out := nodes.NewStructOutput(data)
	out.LogError(err)

	return out
}
