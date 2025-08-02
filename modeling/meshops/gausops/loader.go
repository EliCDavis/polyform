package gausops

import (
	"bufio"
	"bytes"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type LoaderNode struct {
	Data nodes.Output[[]byte]
}

func (pn LoaderNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	bufReader := bufio.NewReader(bytes.NewReader(nodes.TryGetOutputValue(out, pn.Data, nil)))

	splat, err := ply.ReadMesh(bufReader)
	if err != nil {
		out.Set(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return
	}

	out.Set(*splat)
}

type SpzLoaderNode struct {
	Data nodes.Output[[]byte]
}

func (pn SpzLoaderNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	bufReader := bufio.NewReader(bytes.NewReader(nodes.TryGetOutputValue(out, pn.Data, nil)))

	header, err := spz.Read(bufReader)
	if err != nil {
		out.Set(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return
	}

	out.Set(header.Mesh)
}
