package main

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type InfoNode struct {
	nodes.StructData[string]

	Original nodes.NodeOutput[modeling.Mesh]
	Final    nodes.NodeOutput[modeling.Mesh]
}

func (in *InfoNode) Out() nodes.NodeOutput[string] {
	return &nodes.StructNodeOutput[string]{Definition: in}
}

func (in InfoNode) Process() (string, error) {
	original := in.Original.Data().AttributeLength()
	final := in.Final.Data().AttributeLength()

	return fmt.Sprintf("Points: %d / %d\nPruned: %d", final, original, original-final), nil
}
