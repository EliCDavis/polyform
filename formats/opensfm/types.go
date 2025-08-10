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

	refutil.RegisterType[nodes.Struct[ReadReconstructionNode]](factory)

	generator.RegisterTypes(factory)
}

type ReadReconstructionNode struct {
	In nodes.Output[[]byte]
}

func (pn ReadReconstructionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if pn.In == nil {
		out.Set(modeling.EmptyMesh(modeling.PointTopology))
		return
	}

	data, err := ReadReconstructiontData(bytes.NewReader(nodes.GetOutputValue(out, pn.In)))
	out.Set(data)
	out.CaptureError(err)
}
