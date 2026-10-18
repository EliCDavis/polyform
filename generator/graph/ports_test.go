package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/require"
)

func portLabel(port nodes.Port) string {
	if l, ok := port.(nodes.Labeled); ok {
		return l.Label()
	}
	return port.Name()
}

func inputLabels(node nodes.Node) []string {
	var labels []string
	for _, port := range node.Inputs() {
		labels = append(labels, portLabel(port))
	}
	return labels
}

func inputByLabel(t *testing.T, node nodes.Node, label string) nodes.InputPort {
	t.Helper()
	for _, port := range node.Inputs() {
		if portLabel(port) == label {
			return port
		}
	}
	require.FailNowf(t, "no input", "no input labeled %q", label)
	return nil
}

func outputByLabel(t *testing.T, node nodes.Node, label string) nodes.OutputPort {
	t.Helper()
	for _, port := range node.Outputs() {
		if portLabel(port) == label {
			return port
		}
	}
	require.FailNowf(t, "no output", "no output labeled %q", label)
	return nil
}
