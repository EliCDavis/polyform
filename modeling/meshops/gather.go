package meshops

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/utils"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func SelectIndices(count int, keep func(i int) bool) []int {
	kept := make([]bool, count)
	utils.ParallelRange(count, func(start, end int) {
		for i := start; i < end; i++ {
			kept[i] = keep(i)
		}
	})

	n := 0
	for _, k := range kept {
		if k {
			n++
		}
	}

	indices := make([]int, 0, n)
	for i, k := range kept {
		if k {
			indices = append(indices, i)
		}
	}
	return indices
}

func gather[T any](data *iter.ArrayIterator[T], indices []int) []T {
	out := make([]T, len(indices))
	utils.ParallelRange(len(indices), func(start, end int) {
		for k := start; k < end; k++ {
			out[k] = data.At(indices[k])
		}
	})
	return out
}

func gatherAttributes[T any](attributes []string, indices []int, retrieve func(string) *iter.ArrayIterator[T]) map[string][]T {
	gathered := make(map[string][]T, len(attributes))
	if len(indices) == 0 {
		return gathered
	}
	for _, attribute := range attributes {
		gathered[attribute] = gather(retrieve(attribute), indices)
	}
	return gathered
}

func gatherVertices(m modeling.Mesh, indices []int) (
	map[string][]vector4.Float64,
	map[string][]vector3.Float64,
	map[string][]vector2.Float64,
	map[string][]float64,
) {
	return gatherAttributes(m.Float4Attributes(), indices, m.Float4Attribute),
		gatherAttributes(m.Float3Attributes(), indices, m.Float3Attribute),
		gatherAttributes(m.Float2Attributes(), indices, m.Float2Attribute),
		gatherAttributes(m.Float1Attributes(), indices, m.Float1Attribute)
}

func GatherPoints(m modeling.Mesh, indices []int) modeling.Mesh {
	check(RequireTopology(m, modeling.PointTopology))
	v4, v3, v2, v1 := gatherVertices(m, indices)
	return modeling.NewPointCloud(v4, v3, v2, v1)
}
