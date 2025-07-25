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

func (pn LoaderNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	bufReader := bufio.NewReader(bytes.NewReader(pn.Data.Value()))

	splat, err := ply.ReadMesh(bufReader)
	if err != nil {
		out := nodes.NewStructOutput(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return out
	}

	return nodes.NewStructOutput(*splat)
}

type SpzLoaderNode = nodes.Struct[SpzLoaderNodeData]

type SpzLoaderNodeData struct {
	Data nodes.Output[[]byte]
}

func (pn SpzLoaderNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	bufReader := bufio.NewReader(bytes.NewReader(pn.Data.Value()))

	header, err := spz.Read(bufReader)
	if err != nil {
		out := nodes.NewStructOutput(modeling.EmptyPointcloud())
		out.CaptureError(err)
		return out
	}

	return nodes.NewStructOutput(header.Mesh)
}
