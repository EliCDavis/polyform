package modeling

import "sort"

func (m Mesh) directedEdges() map[[2]int]int {
	directed := make(map[[2]int]int, len(m.indices))
	for i := 0; i+3 <= len(m.indices); i += 3 {
		a, b, c := m.indices[i], m.indices[i+1], m.indices[i+2]
		directed[[2]int{a, b}]++
		directed[[2]int{b, c}]++
		directed[[2]int{c, a}]++
	}
	return directed
}

func sortEdges(edges [][2]int) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i][0] != edges[j][0] {
			return edges[i][0] < edges[j][0]
		}
		return edges[i][1] < edges[j][1]
	})
}

// BoundaryEdges are the directed edges without a matching edge the other way.
// Each runs the way its face winds it. None means the surface is closed.
func (m Mesh) BoundaryEdges() [][2]int {
	if m.topology != TriangleTopology {
		return nil
	}

	directed := m.directedEdges()
	unpaired := make([][2]int, 0)
	for edge, count := range directed {
		if directed[[2]int{edge[1], edge[0]}] != count {
			unpaired = append(unpaired, edge)
		}
	}
	sortEdges(unpaired)
	return unpaired
}

// BoundaryLoops returns each hole's rim as a ring of vertices, wound the way
// the surrounding faces are. Rims that share a vertex are skipped.
func (m Mesh) BoundaryLoops() [][]int {
	if m.topology != TriangleTopology {
		return nil
	}

	directed := m.directedEdges()

	next := make(map[int][]int)
	incoming := make(map[int]int)
	starts := make([]int, 0)
	for edge := range directed {
		if directed[[2]int{edge[1], edge[0]}] > 0 {
			continue
		}
		if _, seen := next[edge[0]]; !seen {
			starts = append(starts, edge[0])
		}
		next[edge[0]] = append(next[edge[0]], edge[1])
		incoming[edge[1]]++
	}
	sort.Ints(starts)

	loops := make([][]int, 0)
	visited := make(map[int]bool)

	for _, start := range starts {
		if visited[start] {
			continue
		}

		loop := make([]int, 0)
		at := start
		simple := true
		for {
			if len(next[at]) != 1 || incoming[at] != 1 {
				simple = false
			}
			visited[at] = true
			loop = append(loop, at)

			if !simple {
				break
			}
			at = next[at][0]
			if at == start {
				break
			}
			if visited[at] {
				simple = false
				break
			}
		}

		if simple && len(loop) >= 3 {
			loops = append(loops, loop)
		}
	}

	return loops
}
