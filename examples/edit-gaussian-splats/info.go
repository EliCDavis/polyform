package main

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type InfoNode = nodes.Struct[InfoNodeData]

type InfoNodeData struct {
	Original nodes.Output[modeling.Mesh]
	Final    nodes.Output[modeling.Mesh]
}

func (in InfoNodeData) Out() nodes.StructOutput[string] {
	if in.Original == nil || in.Final == nil {
		return nodes.NewStructOutput("")
	}

	original := in.Original.Value().AttributeLength()
	final := in.Final.Value().AttributeLength()

	return nodes.NewStructOutput(fmt.Sprintf("Points: %d / %d\nPruned: %d", final, original, original-final))
}
