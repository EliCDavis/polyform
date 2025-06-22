package meshops

import (
	"sort"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/morton"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type MortonSortTransformer struct {
	// Attribute to calculate
	Attribute string

	// Precision of the morton encoding, if unset, defaults to 10
	Resolution uint
}

func (mst MortonSortTransformer) attribute() string {
	return mst.Attribute
}

func MortonSortIndices(indices *iter.ArrayIterator[int], positions *iter.ArrayIterator[vector3.Float64], resolution uint) []int {
	encoder := morton.Encoder3D{
		Bounds:     geometry.NewAABBFromIter(positions),
		Resolution: resolution,
	}

	mortonIndices := make([]mortonIndex, indices.Len())
	for i := range indices.Len() {
		index := indices.At(i)
		mortonIndices[i] = mortonIndex{
			OriginalValue:    index,
			OriginalPosition: i,
			MortonIndex:      encoder.Encode(positions.At(index)),
		}
	}

	// Sort...
	sort.Sort(sortByMortonIndex(mortonIndices))

	newIndices := make([]int, indices.Len())

	for i, m := range mortonIndices {
		newIndices[i] = m.OriginalValue
	}

	return newIndices
}

func MortonSort(mesh modeling.Mesh, attribute string, resolution uint) modeling.Mesh {
	if err := RequireV3Attribute(mesh, attribute); err != nil {
		panic(err)
	}

	// TODO: CAN MORTON SHUFFLE EVEN WORK WITH TRIANGLE TOPOLOGIES???
	if err := RequireTopology(mesh, modeling.PointTopology); err != nil {
		panic(err)
	}

	newIndices := MortonSortIndices(mesh.Indices(), mesh.Float3Attribute(attribute), resolution)

	return untanglePointcloud(mesh.SetIndices(newIndices))
}
