package modeling

type VertexLUT map[int]map[int]struct{}

func (vLUT VertexLUT) AddLookup(from, to int) {
	// if from == to {
	// 	panic("can't connect a vertice to itself")
	// }

	if _, ok := vLUT[from]; !ok {
		vLUT[from] = make(map[int]struct{})
	}
	vLUT[from][to] = struct{}{}
}

func (vLUT VertexLUT) RemoveLookup(from, to int) {
	delete(vLUT[from], to)
}

func (vLUT VertexLUT) Link(v1, v2 int) {
	vLUT.AddLookup(v1, v2)
	vLUT.AddLookup(v2, v1)
}

func (vLUT VertexLUT) RemoveLink(v1, v2 int) {
	vLUT.RemoveLookup(v1, v2)
	vLUT.RemoveLookup(v2, v1)
}

func (vLUT VertexLUT) Lookup(v int) map[int]struct{} {
	return vLUT[v]
}

func (vLUT VertexLUT) Count(v int) int {
	return len(vLUT[v])
}

func (vLUT VertexLUT) Remove(v int) map[int]struct{} {
	values := vLUT[v]
	delete(vLUT, v)
	return values
}

func (vLUT VertexLUT) RemoveVertex(v int) {
	for n := range vLUT[v] {
		delete(vLUT[n], v)
	}
	delete(vLUT, v)
}
