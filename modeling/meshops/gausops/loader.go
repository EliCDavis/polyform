package gausops

import (
	"bufio"
	"bytes"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type LoaderNode = nodes.StructNode[modeling.Mesh, LoaderNodeData]

type LoaderNodeData struct {
	Data nodes.NodeOutput[[]byte]
}

func (pn LoaderNodeData) Process() (modeling.Mesh, error) {
	bufReader := bufio.NewReader(bytes.NewReader(pn.Data.Value()))

	header, err := ply.ReadHeader(bufReader)
	if err != nil {
		return modeling.EmptyPointcloud(), err
	}

	reader := header.BuildReader(bufReader)
	plyMesh, err := reader.ReadMesh(ply.GuassianSplatVertexAttributes)
	if err != nil {
		return modeling.EmptyPointcloud(), err
	}
	return *plyMesh, err
}
