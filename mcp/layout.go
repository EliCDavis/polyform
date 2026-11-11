package mcp

import (
	"sort"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
)

const (
	layoutColumnWidth = 380.
	layoutRowHeight   = 260.
)

func nodePosition(inst *graph.Graph, id string) (x, y float64, ok bool) {
	meta, _ := inst.Metadata("nodes." + id).(map[string]any)
	position, _ := meta["position"].(map[string]any)
	x, xOk := position["x"].(float64)
	y, yOk := position["y"].(float64)
	return x, y, xOk && yOk
}

// layoutUnplaced gives every node without a position one, in columns by
// the longest wired path feeding it, to the right of anything already
// placed. Nodes that have a position keep it.
func layoutUnplaced(inst *graph.Graph) int {
	sch := inst.Schema()
	unplaced := map[string]bool{}
	startX := 0.
	for id := range sch.Nodes {
		if x, _, ok := nodePosition(inst, id); ok {
			startX = max(startX, x+layoutColumnWidth)
		} else {
			unplaced[id] = true
		}
	}
	if len(unplaced) == 0 {
		return 0
	}

	ranks := map[string]int{}
	var rankOf func(id string, seen map[string]bool) int
	rankOf = func(id string, seen map[string]bool) int {
		if r, ok := ranks[id]; ok {
			return r
		}
		if seen[id] {
			return 0
		}
		seen[id] = true
		r := 0
		for _, ref := range sch.Nodes[id].AssignedInput {
			if _, exists := sch.Nodes[ref.NodeId]; exists {
				r = max(r, rankOf(ref.NodeId, seen)+1)
			}
		}
		ranks[id] = r
		return r
	}

	columns := map[int][]string{}
	for id := range unplaced {
		r := rankOf(id, map[string]bool{})
		columns[r] = append(columns[r], id)
	}
	for _, ids := range columns {
		sort.Strings(ids)
	}

	placed := map[string]float64{}
	ranksInOrder := make([]int, 0, len(columns))
	for r := range columns {
		ranksInOrder = append(ranksInOrder, r)
	}
	sort.Ints(ranksInOrder)
	for _, r := range ranksInOrder {
		ids := columns[r]
		sort.SliceStable(ids, func(i, j int) bool {
			return feederY(sch.Nodes[ids[i]].AssignedInput, placed) < feederY(sch.Nodes[ids[j]].AssignedInput, placed)
		})
		for row, id := range ids {
			y := float64(row) * layoutRowHeight
			placed[id] = y
			inst.SetMetadata("nodes."+id+".position", map[string]any{"x": startX + float64(r)*layoutColumnWidth, "y": y})
		}
	}
	return len(unplaced)
}

func feederY(inputs map[string]schema.PortReference, placed map[string]float64) float64 {
	sum, n := 0., 0.
	for _, ref := range inputs {
		if y, ok := placed[ref.NodeId]; ok {
			sum += y
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

func (s *Server) layoutAllScopes() {
	layoutUnplaced(s.graph.Graph)
	for id := range s.graph.Schema().SubGraphs {
		if child, err := s.graph.SubGraphInstance(id); err == nil {
			layoutUnplaced(child)
		}
	}
}
