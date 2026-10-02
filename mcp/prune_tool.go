package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type PruneOrphansInput struct {
	Scope  string   `json:"scope,omitempty" jsonschema:"subgraph id to prune; omit for the root graph. 'all' prunes the root and every subgraph definition."`
	Apply  bool     `json:"apply,omitempty" jsonschema:"false (default) only reports what would be deleted; true deletes it"`
	Except []string `json:"except,omitempty" jsonschema:"node ids to keep even though nothing reads them"`
}

type PrunedNode struct {
	Scope string `json:"scope"`
	Id    string `json:"id"`
	Type  string `json:"type"`
}

type PruneOrphansOutput struct {
	Orphans []PrunedNode `json:"orphans" jsonschema:"nodes whose outputs nothing reads, including whole dead chains behind them; deleted when apply was true"`
	Applied bool         `json:"applied"`
}

// orphansIn lists nodes in inst that feed nothing, iterating so a chain
// whose only consumer was itself dead is listed whole. Producers'
// sources, subgraph boundary nodes of either direction, and anything in
// keep survive.
func orphansIn(inst *graph.Instance, keep map[string]bool) []string {
	sch := inst.Schema()
	protected := map[string]bool{}
	for _, p := range sch.Producers {
		protected[p.NodeID] = true
	}
	for id := range keep {
		protected[id] = true
	}
	// Both kinds of boundary node are the subgraph's public interface, not
	// working nodes. An input boundary nothing reads yet is still a
	// declared port: deleting it drops every instance's connection into
	// that port and leaves the outline describing an input that no longer
	// exists.
	for id, n := range sch.Nodes {
		if n.SubGraphOutputBoundary != nil || n.SubGraphInputBoundary != nil {
			protected[id] = true
		}
	}

	dead := map[string]bool{}
	for {
		consumed := map[string]bool{}
		for id, n := range sch.Nodes {
			if dead[id] {
				continue
			}
			for _, ref := range n.AssignedInput {
				consumed[ref.NodeId] = true
			}
		}
		grew := false
		for id := range sch.Nodes {
			if dead[id] || protected[id] || consumed[id] {
				continue
			}
			dead[id] = true
			grew = true
		}
		if !grew {
			break
		}
	}

	ids := make([]string, 0, len(dead))
	for id := range dead {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *Server) pruneOrphans(ctx context.Context, req *mcpsdk.CallToolRequest, in PruneOrphansInput) (*mcpsdk.CallToolResult, PruneOrphansOutput, error) {
	var out PruneOrphansOutput
	var err error
	s.atomic(&err, func() error {
		scopes := map[string]*graph.Instance{}
		if in.Scope == "all" {
			scopes["root"] = s.graph
			for id := range s.graph.Schema().SubGraphs {
				child, e := s.graph.SubGraphInstance(id)
				if e != nil {
					return e
				}
				scopes[id] = child
			}
		} else {
			inst, e := s.resolveScope(in.Scope)
			if e != nil {
				return e
			}
			name := in.Scope
			if name == "" {
				name = "root"
			}
			scopes[name] = inst
		}

		keep := map[string]bool{}
		for _, id := range in.Except {
			keep[id] = true
		}

		names := make([]string, 0, len(scopes))
		for name := range scopes {
			names = append(names, name)
		}
		sort.Strings(names)

		out.Orphans = []PrunedNode{}
		for _, name := range names {
			inst := scopes[name]
			sch := inst.Schema()
			for _, id := range orphansIn(inst, keep) {
				typeKey := sch.Nodes[id].Type
				if inst.Node(id) != nil {
					if _, isBoundary := subgraph.IsBoundaryNode(inst.Node(id)); isBoundary {
						typeKey = "boundary " + typeKey
					}
				}
				out.Orphans = append(out.Orphans, PrunedNode{Scope: name, Id: id, Type: shortTypeKey(typeKey)})
				if in.Apply {
					inst.DeleteNodeById(id)
				}
			}
		}
		out.Applied = in.Apply
		return nil
	})
	return nil, out, err
}

func (s *Server) registerPruneTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "prune_orphans",
		Description: fmt.Sprintf("List (or with apply: true, delete) every node whose outputs nothing reads - literals left behind by a rewire, variable references that lost their consumer, a primitive that was replaced - and the whole dead chain behind each one. Producer sources and subgraph boundary ports (inputs and outputs alike, since both are the subgraph's public interface) are never touched. Run it before save_graph; a graph is not done while it carries nodes that compute nothing. Scope '%s' covers the root and every subgraph.", "all"),
	}, s.pruneOrphans)
}
