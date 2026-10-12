package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
)

func TestRemapZeroWidthInputIsAStep(t *testing.T) {
	remap := func(value float64, clamp bool) nodes.Node {
		return nodetest.NewNode(math.RemapNode[float64]{
			Value:  nodetest.NewPortValue(value),
			InMin:  nodetest.NewPortValue(1.),
			InMax:  nodetest.NewPortValue(1.),
			OutMin: nodetest.NewPortValue(2.),
			OutMax: nodetest.NewPortValue(10.),
			Clamp:  nodetest.NewPortValue(clamp),
		})
	}

	tests := map[string]struct {
		node       nodes.Node
		assertions []nodetest.Assertion
	}{
		"below the step, clamped":   {node: remap(0.5, true), assertions: []nodetest.Assertion{nodetest.AssertOutput("Out", 2.)}},
		"at the step, clamped":      {node: remap(1., true), assertions: []nodetest.Assertion{nodetest.AssertOutput("Out", 10.)}},
		"above the step, clamped":   {node: remap(7., true), assertions: []nodetest.Assertion{nodetest.AssertOutput("Out", 10.)}},
		"below the step, unclamped": {node: remap(0.5, false), assertions: []nodetest.Assertion{nodetest.AssertOutput("Out", 2.)}},
		"above the step, unclamped": {node: remap(7., false), assertions: []nodetest.Assertion{nodetest.AssertOutput("Out", 10.)}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, assertion := range tc.assertions {
				assertion.Assert(t, tc.node)
			}
		})
	}
}
