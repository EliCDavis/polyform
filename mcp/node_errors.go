package mcp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
)

const maxReportedNodeErrors = 20

type NodeError struct {
	Path  string `json:"path" jsonschema:"node id, prefixed by the subgraph instance ids it sits inside (Node-3/Node-12)"`
	Type  string `json:"type"`
	Port  string `json:"port"`
	Error string `json:"error"`
	Count int    `json:"count,omitempty" jsonschema:"how many nodes hit this same error, when more than one (every instance of a subgraph reports its own)"`
}

// A tool that fails returns only its error, so the node errors that likely
// caused it have to travel in the message.
func withNodeErrors(err error, found []NodeError) error {
	if err == nil || len(found) == 0 {
		return err
	}
	lines := make([]string, 0, len(found))
	for _, f := range found {
		line := fmt.Sprintf("%s (%s).%s: %s", f.Path, f.Type, f.Port, f.Error)
		if f.Count > 1 {
			line += fmt.Sprintf(" [x%d]", f.Count)
		}
		lines = append(lines, line)
	}
	return fmt.Errorf("%w\nnodes that recorded errors while running:\n%s", err, strings.Join(lines, "\n"))
}

func collectNodeErrors(inst *graph.Graph) []NodeError {
	var found []NodeError
	index := map[NodeError]int{}

	var walk func(inst *graph.Graph, prefix string)
	walk = func(inst *graph.Graph, prefix string) {
		schema := inst.Schema()
		ids := inst.NodeIds()
		sort.Strings(ids)
		for _, id := range ids {
			node := inst.Node(id)
			if sub, ok := node.(*graph.SubgraphInstanceNode); ok {
				// A copy nothing has evaluated through has no errors to report.
				if live := sub.BuiltGraph(); live != nil {
					walk(live, prefix+id+"/")
				}
				continue
			}

			outputs := node.Outputs()
			names := make([]string, 0, len(outputs))
			for name := range outputs {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				observable, ok := outputs[name].(nodes.ObservableExecution)
				if !ok {
					continue
				}
				for _, message := range observable.ExecutionReport().Errors {
					key := NodeError{Type: shortTypeKey(schema.Nodes[id].Type), Port: name, Error: message}
					if i, seen := index[key]; seen {
						found[i].Count++
						continue
					}
					index[key] = len(found)
					entry := key
					entry.Path = prefix + id
					entry.Count = 1
					found = append(found, entry)
				}
			}
		}
	}
	walk(inst, "")

	for i := range found {
		if found[i].Count == 1 {
			found[i].Count = 0
		}
	}
	if len(found) > maxReportedNodeErrors {
		found = found[:maxReportedNodeErrors]
	}
	return found
}
