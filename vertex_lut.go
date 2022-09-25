package mesh

type VertexLUT map[int]map[int]struct{}

func (vlut VertexLUT) AddLookup(from, to int) {
	if _, ok := vlut[from]; !ok {
		vlut[from] = make(map[int]struct{})
	}
	vlut[from][to] = struct{}{}
}

func (vlut VertexLUT) Lookup(v int) map[int]struct{} {
	return vlut[v]
}

func (vlut VertexLUT) Count(v int) int {
	return len(vlut[v])
}
