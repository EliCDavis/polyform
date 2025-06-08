package meshops

import (
	"math"
	"math/rand/v2"
	"sort"

	"github.com/EliCDavis/polyform/math/morton"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type mortonIndex struct {
	OriginalPosition int
	OriginalValue    int
	MortonIndex      uint64
}

type sortByMortonIndex []mortonIndex

func (a sortByMortonIndex) Len() int           { return len(a) }
func (a sortByMortonIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByMortonIndex) Less(i, j int) bool { return a[i].MortonIndex < a[j].MortonIndex }

type MortonShuffleTransformer struct {
	// Attribute to calculate
	Attribute string

	// Size of the bins that we shuffle, if unset, defaults to 32
	BinSize int

	// Precision of the morton encoding, if unset, defaults to 10
	Resolution uint
}

func (mst MortonShuffleTransformer) attribute() string {
	return mst.Attribute
}

func MortonShuffle(mesh modeling.Mesh, attribute string, binSize int, resolution uint) modeling.Mesh {
	if err := RequireV3Attribute(mesh, attribute); err != nil {
		panic(err)
	}

	// TODO: CAN MORTON SHUFFLE EVEN WORK WITH TRIANGLE TOPOLOGIES???
	if err := RequireTopology(mesh, modeling.PointTopology); err != nil {
		panic(err)
	}

	encoder := morton.Encoder3D{
		Bounds:     mesh.BoundingBox(attribute),
		Resolution: resolution,
	}

	indices := mesh.Indices()
	mortonIndices := make([]mortonIndex, indices.Len())
	attrToShuffle := mesh.Float3Attribute(attribute)
	for i := range indices.Len() {
		index := indices.At(i)
		mortonIndices[i] = mortonIndex{
			OriginalValue:    index,
			OriginalPosition: i,
			MortonIndex:      encoder.Encode(attrToShuffle.At(index)),
		}
	}

	// Floor cause I really don't care to deal with the edgecase of the last one
	warps := math.Floor(float64(indices.Len()) / float64(binSize))

	// Sort...
	sort.Sort(sortByMortonIndex(mortonIndices))

	// Shuffle...
	tmp := make([]mortonIndex, binSize)
	rand.Shuffle(int(warps), func(i, j int) {
		iW := i * binSize
		jW := j * binSize
		copy(tmp, mortonIndices[iW:iW+binSize])
		copy(mortonIndices[i*binSize:iW+binSize], mortonIndices[jW:jW+binSize])
		copy(mortonIndices[jW:jW+binSize], tmp)
	})

	newIndices := make([]int, indices.Len())

	for i, m := range mortonIndices {
		newIndices[i] = m.OriginalValue

	}

	// mesh.SetFloat3Attribute(modeling.ColorAttribute, nil)

	return untanglePointcloud(mesh.SetFloat3Attribute(modeling.ColorAttribute, nil).SetIndices(newIndices))
}

func untanglePointcloud(in modeling.Mesh) modeling.Mesh {
	if err := RequireTopology(in, modeling.PointTopology); err != nil {
		panic(err)
	}

	// v4Data := make(map[string][]vector4.Float64)
	v3Data := make(map[string][]vector3.Float64)
	// v2Data := make(map[string][]vector2.Float64)
	// v1Data := make(map[string][]float64)

	indices := in.Indices()
	remapping := make([]int, indices.Len())
	for i := range indices.Len() {
		remapping[indices.At(i)] = i
	}

	vertexCount := in.AttributeLength()

	for _, attr := range in.Float3Attributes() {
		oldAttr := in.Float3Attribute(attr)
		newAttr := make([]vector3.Float64, vertexCount)

		for attrIndex := range oldAttr.Len() {
			newAttr[remapping[attrIndex]] = oldAttr.At(attrIndex)
		}

		v3Data[attr] = newAttr
	}

	return modeling.NewPointCloud(nil, v3Data, nil, nil)

}
