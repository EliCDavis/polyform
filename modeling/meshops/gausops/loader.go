package gausops

import (
	"bufio"
	"bytes"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type LoaderNode = nodes.Struct[LoaderNodeData]

type LoaderNodeData struct {
	Data nodes.Output[[]byte]
}

func (pn LoaderNodeData) Out(out *nodes.StructOutput[modeling.Mesh]) {
	bufReader := bufio.NewReader(bytes.NewReader(nodes.TryGetOutputValue(out, pn.Data, nil)))

	splat, err := ply.ReadMesh(bufReader)
	if err != nil {
		out.Set(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return
	}

	out.Set(*splat)
}

type SpzLoaderNode = nodes.Struct[SpzLoaderNodeData]

type SpzLoaderNodeData struct {
	Data nodes.Output[[]byte]
}

func (pn SpzLoaderNodeData) Out(out *nodes.StructOutput[modeling.Mesh]) {
	bufReader := bufio.NewReader(bytes.NewReader(nodes.TryGetOutputValue(out, pn.Data, nil)))

	header, err := spz.Read(bufReader)
	if err != nil {
		out.Set(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return
	}

	out.Set(header.Mesh)
}
