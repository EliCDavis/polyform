package modeling

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type Mesh struct {
	v4Data    map[string][]vector4.Float64
	v3Data    map[string][]vector3.Float64
	v2Data    map[string][]vector2.Float64
	v1Data    map[string][]float64
	indices   []int
	materials []MeshMaterial
	topology  Topology
}

// New Mesh creates a new mesh with the specified topology with all empty
// attribute data arrays stripped.
func NewMesh(topo Topology, indices []int) Mesh {
	return Mesh{
		indices:   indices,
		materials: nil,
		topology:  topo,
		v1Data:    make(map[string][]float64),
		v2Data:    make(map[string][]vector2.Float64),
		v3Data:    make(map[string][]vector3.Float64),
		v4Data:    make(map[string][]vector4.Float64),
	}
}

// NewTriangleMesh creates a new triangle mesh with all empty attribute data
// arrays stripped.
func NewTriangleMesh(indices []int) Mesh {
	return Mesh{
		indices:   indices,
		materials: nil,
		topology:  TriangleTopology,
		v1Data:    make(map[string][]float64),
		v2Data:    make(map[string][]vector2.Float64),
		v3Data:    make(map[string][]vector3.Float64),
		v4Data:    make(map[string][]vector4.Float64),
	}
}

func newImpliedIndicesMesh(
	topo Topology,
	v1Data map[string][]float64,
	v2Data map[string][]vector2.Float64,
	v3Data map[string][]vector3.Float64,
	v4Data map[string][]vector4.Float64,
	materials []MeshMaterial,
) Mesh {
	attributeCount := 0

	cleanedV4Data := make(map[string][]vector4.Float64)
	for key, vals := range v4Data {
		if len(vals) == 0 {
			continue
		}
		cleanedV4Data[key] = vals
		attributeCount = len(vals)
	}

	cleanedV3Data := make(map[string][]vector3.Float64)
	for key, vals := range v3Data {
		if len(vals) == 0 {
			continue
		}
		cleanedV3Data[key] = vals
		attributeCount = len(vals)
	}

	cleanedV2Data := make(map[string][]vector2.Float64)
	for key, vals := range v2Data {
		if len(vals) == 0 {
			continue
		}
		cleanedV2Data[key] = vals
		attributeCount = len(vals)
	}

	cleanedV1Data := make(map[string][]float64)
	for key, vals := range v1Data {
		if len(vals) == 0 {
			continue
		}
		cleanedV1Data[key] = vals
		attributeCount = len(vals)
	}

	if topo == LineStripTopology && attributeCount == 1 {
		panic(fmt.Errorf("invalid attribute count for line strip mesh"))
	}

	indices := make([]int, attributeCount)
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}

	return Mesh{
		indices:   indices,
		materials: materials,
		topology:  topo,
		v4Data:    cleanedV4Data,
		v3Data:    cleanedV3Data,
		v2Data:    cleanedV2Data,
		v1Data:    cleanedV1Data,
	}
}

func NewLineStripMesh(
	v3Data map[string][]vector3.Float64,
	v2Data map[string][]vector2.Float64,
	v1Data map[string][]float64,
	materials []MeshMaterial,
) Mesh {
	return newImpliedIndicesMesh(LineStripTopology, v1Data, v2Data, v3Data, nil, materials)
}

func NewPointCloud(
	v3Data map[string][]vector3.Float64,
	v2Data map[string][]vector2.Float64,
	v1Data map[string][]float64,
	materials []MeshMaterial,
) Mesh {
	return newImpliedIndicesMesh(PointTopology, v1Data, v2Data, v3Data, nil, materials)
}

// Creates a new triangle mesh with no vertices or attribute data
func EmptyMesh(topo Topology) Mesh {
	return Mesh{
		indices:   make([]int, 0),
		materials: make([]MeshMaterial, 0),
		topology:  topo,
		v4Data:    make(map[string][]vector4.Float64),
		v3Data:    make(map[string][]vector3.Float64),
		v2Data:    make(map[string][]vector2.Float64),
		v1Data:    make(map[string][]float64),
	}
}

func (m Mesh) Topology() Topology {
	return m.topology
}

func (m Mesh) ToPointCloud() Mesh {
	if m.topology == PointTopology {
		return m
	}

	indices := make([]int, m.AttributeLength())
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}

	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   indices,
		materials: m.materials,
		topology:  PointTopology,
	}
}

func (m Mesh) SetIndices(indices []int) Mesh {
	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) Indices() iter.ArrayIterator[int] {
	return iter.Array(m.indices)
}

func (m Mesh) Transform(ops ...Transformer) Mesh {
	final := m
	for _, transformer := range ops {
		next, err := transformer.Transform(final)
		if err != nil {
			panic(err)
		}
		final = next
	}
	return final
}

func (m Mesh) Float4Attributes() []string {
	attributes := make([]string, 0, len(m.v4Data))

	for atr := range m.v4Data {
		attributes = append(attributes, atr)
	}

	sort.Strings(attributes)

	return attributes
}

func (m Mesh) Float3Attributes() []string {
	attributes := make([]string, 0, len(m.v3Data))

	for atr := range m.v3Data {
		attributes = append(attributes, atr)
	}

	sort.Strings(attributes)

	return attributes
}

func (m Mesh) Float2Attributes() []string {
	attributes := make([]string, 0, len(m.v2Data))

	for atr := range m.v2Data {
		attributes = append(attributes, atr)
	}

	sort.Strings(attributes)

	return attributes
}

func (m Mesh) Float1Attributes() []string {
	attributes := make([]string, 0, len(m.v1Data))

	for atr := range m.v1Data {
		attributes = append(attributes, atr)
	}

	sort.Strings(attributes)

	return attributes
}

func (m Mesh) Materials() []MeshMaterial {
	return m.materials
}

func (m Mesh) SetMaterial(mat Material) Mesh {
	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   m.indices,
		materials: []MeshMaterial{{PrimitiveCount: len(m.indices) / m.topology.IndexSize(), Material: &mat}},
		topology:  m.topology,
	}
}

func (m Mesh) SetMaterials(mat []MeshMaterial) Mesh {
	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   m.indices,
		materials: mat,
		topology:  m.topology,
	}
}

func (m Mesh) Tri(i int) Tri {
	return Tri{
		mesh:          &m,
		startingIndex: i * 3,
	}
}

func (m Mesh) LineStrip(i int) Line {
	return Line{
		mesh:          &m,
		startingIndex: i,
	}
}

func (m Mesh) PrimitiveCount() int {
	switch m.topology {
	case QuadTopology, TriangleTopology:
		return len(m.indices) / m.topology.IndexSize()

	case PointTopology, LineLoopTopology:
		return len(m.indices)

	case LineTopology, LineStripTopology:
		return len(m.indices) - 1
	}

	panic(fmt.Errorf("unimplemented topology: %s", m.topology.String()))
}

func appendData[T any](a, b map[string][]T, aLen, bLen int, nilVal func() T) map[string][]T {
	finalData := make(map[string][]T)

	for atr, data := range a {
		finalData[atr] = data

		if _, ok := b[atr]; !ok {
			for i := 0; i < bLen; i++ {
				finalData[atr] = append(finalData[atr], nilVal())
			}
		}
	}

	for atr, data := range b {
		if _, ok := finalData[atr]; !ok {
			for i := 0; i < aLen; i++ {
				finalData[atr] = append(finalData[atr], nilVal())
			}
		}
		finalData[atr] = append(finalData[atr], data...)
	}
	return finalData
}

func (m Mesh) Append(other Mesh) Mesh {
	if m.topology != other.topology {
		panic(fmt.Errorf("can not combine meshes with different topologies (%s != %s)", m.topology.String(), other.topology.String()))
	}

	mAtrLength := m.AttributeLength()
	oAtrLength := other.AttributeLength()

	finalV1Data := appendData(m.v1Data, other.v1Data, mAtrLength, oAtrLength, func() float64 { return 0 })
	finalV2Data := appendData(m.v2Data, other.v2Data, mAtrLength, oAtrLength, func() vector2.Vector[float64] { return vector2.Zero[float64]() })
	finalV3Data := appendData(m.v3Data, other.v3Data, mAtrLength, oAtrLength, func() vector3.Vector[float64] { return vector3.Zero[float64]() })
	finalV4Data := appendData(m.v4Data, other.v4Data, mAtrLength, oAtrLength, func() vector4.Vector[float64] { return vector4.Zero[float64]() })

	finalTris := append(m.indices, other.indices...)
	finalMaterials := append(m.materials, other.materials...)
	for i := len(m.indices); i < len(finalTris); i++ {
		finalTris[i] += mAtrLength
	}

	return Mesh{
		v1Data:    finalV1Data,
		v2Data:    finalV2Data,
		v3Data:    finalV3Data,
		v4Data:    finalV4Data,
		materials: finalMaterials,
		indices:   finalTris,
		topology:  m.topology,
	}
}

func (m Mesh) Rotate(q Quaternion) Mesh {
	m.requireV3Attribute(PositionAttribute)

	finalMesh := m
	oldData := m.v3Data[PositionAttribute]
	finalVerts := make([]vector3.Float64, len(oldData))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = q.Rotate(oldData[i])
	}

	return finalMesh.SetFloat3Attribute(PositionAttribute, finalVerts)
}

func (m Mesh) Scale(amount vector3.Float64) Mesh {
	m.requireV3Attribute(PositionAttribute)
	return m.ModifyFloat3Attribute(PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
		return v.MultByVector(amount)
	})
}

func (m Mesh) BoundingBox(atr string) geometry.AABB {
	m.requireV3Attribute(atr)
	return geometry.NewAABBFromPoints(m.v3Data[atr]...)
}

func (m Mesh) scanTrisPrimitives(start, size int, f func(i int, p Primitive)) {
	for i := start; i < size; i++ {
		f(i, m.Tri(i))
	}
}

func (m Mesh) scanPointPrimitives(start, size int, f func(i int, p Primitive)) {
	for i := start; i < size; i++ {
		f(i, &Point{
			mesh:  &m,
			index: i,
		})
	}
}

func (m Mesh) scanLinePrimitives(start, size int, f func(i int, p Primitive)) {
	for i := start; i < size; i++ {
		f(i, &Line{
			mesh:          &m,
			startingIndex: i,
		})
	}
}

func (m Mesh) ScanPrimitives(f func(i int, p Primitive)) Mesh {
	switch m.topology {
	case TriangleTopology:
		m.scanTrisPrimitives(0, m.PrimitiveCount(), f)

	case PointTopology:
		m.scanPointPrimitives(0, m.PrimitiveCount(), f)

	case LineStripTopology:
		m.scanLinePrimitives(0, m.PrimitiveCount(), f)

	default:
		panic(fmt.Errorf("unimplemented topology: %s", m.topology.String()))
	}
	return m
}

func (m Mesh) ScanPrimitivesParallel(f func(i int, p Primitive)) Mesh {
	return m.ScanPrimitivesParallelWithPoolSize(runtime.NumCPU(), f)
}

func (m Mesh) ScanPrimitivesParallelWithPoolSize(size int, f func(i int, p Primitive)) Mesh {
	if size < 1 {
		panic(fmt.Errorf("unable to scan primitives, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ScanPrimitives(f)
	}

	var wg sync.WaitGroup

	totalWork := m.PrimitiveCount()
	workSize := int(math.Floor(float64(totalWork) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = totalWork - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			switch m.topology {
			case TriangleTopology:
				m.scanTrisPrimitives(start, size, f)

			case PointTopology:
				m.scanPointPrimitives(start, size, f)

			case LineStripTopology:
				m.scanLinePrimitives(start, size, f)

			default:
				panic(fmt.Errorf("unimplemented topology: %s", m.topology.String()))
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ScanFloat3Attribute(atr string, f func(i int, v vector3.Float64)) Mesh {
	m.requireV3Attribute(atr)

	data := m.v3Data[atr]
	for i, v := range data {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat4Attribute(atr string, f func(i int, v vector4.Float64)) Mesh {
	m.requireV4Attribute(atr)

	data := m.v4Data[atr]
	for i, v := range data {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat3AttributeParallel(atr string, f func(i int, v vector3.Float64)) Mesh {
	return m.ScanFloat3AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ScanFloat3AttributeParallelWithPoolSize(atr string, size int, f func(i int, v vector3.Float64)) Mesh {
	m.requireV3Attribute(atr)

	if size < 1 {
		panic(fmt.Errorf("unable to scan float3, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ScanFloat3Attribute(atr, f)
	}

	var wg sync.WaitGroup

	data := m.v3Data[atr]
	workSize := int(math.Floor(float64(len(data)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(data) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				f(i, data[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ScanFloat2Attribute(atr string, f func(i int, v vector2.Float64)) Mesh {
	m.requireV2Attribute(atr)

	data := m.v2Data[atr]
	for i, v := range data {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat2AttributeParallel(atr string, f func(i int, v vector2.Float64)) Mesh {
	return m.ScanFloat2AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ScanFloat2AttributeParallelWithPoolSize(atr string, size int, f func(i int, v vector2.Float64)) Mesh {
	m.requireV2Attribute(atr)

	if size < 1 {
		panic(fmt.Errorf("unable to scan float2, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ScanFloat2Attribute(atr, f)
	}

	var wg sync.WaitGroup

	data := m.v2Data[atr]
	workSize := int(math.Floor(float64(len(data)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(data) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				f(i, data[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ScanFloat1Attribute(atr string, f func(i int, v float64)) Mesh {
	m.requireV1Attribute(atr)

	data := m.v1Data[atr]
	for i, v := range data {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat1AttributeParallel(atr string, f func(i int, v float64)) Mesh {
	return m.ScanFloat1AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ScanFloat1AttributeParallelWithPoolSize(atr string, size int, f func(i int, v float64)) Mesh {
	m.requireV1Attribute(atr)

	if size < 1 {
		panic(fmt.Errorf("unable to scan float1, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ScanFloat1Attribute(atr, f)
	}

	var wg sync.WaitGroup

	data := m.v1Data[atr]
	workSize := int(math.Floor(float64(len(data)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(data) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				f(i, data[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ModifyFloat3Attribute(atr string, f func(i int, v vector3.Float64) vector3.Float64) Mesh {
	m.requireV3Attribute(atr)
	oldData := m.v3Data[atr]
	modified := make([]vector3.Float64, len(oldData))

	for i, v := range oldData {
		modified[i] = f(i, v)
	}

	return m.SetFloat3Attribute(atr, modified)
}

func (m Mesh) ModifyFloat3AttributeParallel(atr string, f func(i int, v vector3.Float64) vector3.Float64) Mesh {
	return m.ModifyFloat3AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ModifyFloat3AttributeParallelWithPoolSize(atr string, size int, f func(i int, v vector3.Float64) vector3.Float64) Mesh {
	m.requireV3Attribute(atr)

	if size < 1 {
		panic(fmt.Errorf("unable to modify float3, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ModifyFloat3Attribute(atr, f)
	}

	oldData := m.v3Data[atr]
	modified := make([]vector3.Float64, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(oldData)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(oldData) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				modified[i] = f(i, oldData[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m.SetFloat3Attribute(atr, modified)
}

func (m Mesh) ModifyFloat2Attribute(atr string, f func(i int, v vector2.Float64) vector2.Float64) Mesh {
	m.requireV2Attribute(atr)
	oldData := m.v2Data[atr]
	modified := make([]vector2.Float64, len(oldData))

	for i, v := range oldData {
		modified[i] = f(i, v)
	}

	return m.SetFloat2Attribute(atr, modified)
}

func (m Mesh) ModifyFloat2AttributeParallel(atr string, f func(i int, v vector2.Float64) vector2.Float64) Mesh {
	return m.ModifyFloat2AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ModifyFloat2AttributeParallelWithPoolSize(atr string, size int, f func(i int, v vector2.Float64) vector2.Float64) Mesh {
	m.requireV2Attribute(atr)

	if size < 1 {
		panic(fmt.Errorf("unable to modify float2, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ModifyFloat2Attribute(atr, f)
	}

	oldData := m.v2Data[atr]
	modified := make([]vector2.Float64, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(oldData)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(oldData) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				modified[i] = f(i, oldData[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m.SetFloat2Attribute(atr, modified)
}

func (m Mesh) ModifyFloat1Attribute(atr string, f func(i int, v float64) float64) Mesh {
	m.requireV1Attribute(atr)
	oldData := m.v1Data[atr]
	modified := make([]float64, len(oldData))

	for i, v := range oldData {
		modified[i] = f(i, v)
	}

	return m.SetFloat1Attribute(atr, modified)
}

func (m Mesh) ModifyFloat1AttributeParallel(atr string, f func(i int, v float64) float64) Mesh {
	return m.ModifyFloat1AttributeParallelWithPoolSize(atr, runtime.NumCPU(), f)
}

func (m Mesh) ModifyFloat1AttributeParallelWithPoolSize(atr string, size int, f func(i int, v float64) float64) Mesh {
	m.requireV1Attribute(atr)
	if size < 1 {
		panic(fmt.Errorf("unable to modify float1, invalid worker pool size: %d", size))
	}

	if size == 1 {
		return m.ModifyFloat1Attribute(atr, f)
	}

	oldData := m.v1Data[atr]
	modified := make([]float64, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(oldData)) / float64(size)))
	for i := 0; i < size; i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == size-1 {
			jobSize = len(oldData) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			end := start + size
			for i := start; i < end; i++ {
				modified[i] = f(i, oldData[i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m.SetFloat1Attribute(atr, modified)
}

func (m Mesh) WeldByFloat3Attribute(attribute string, decimalPlace int) Mesh {
	m.requireV3Attribute(attribute)
	m.requireTopology(TriangleTopology)

	// =================== Finding unique vertices ============================
	vertILU := make(map[VectorInt]int)
	vertIToOriginalLU := make(map[int]int)

	// Mapping from rounded vector to whether or not it get's used by a triangle
	// in the resulting mesh
	vertLUUsed := make(map[VectorInt]bool)

	// count of unique vertices once rounded
	uniqueVertCount := 0

	data := m.v3Data[attribute]
	for vi, v := range data {
		vInt := Vector3ToInt(v, decimalPlace)

		if _, ok := vertILU[vInt]; !ok {
			vertILU[vInt] = uniqueVertCount
			vertLUUsed[vInt] = false
			vertIToOriginalLU[uniqueVertCount] = vi
			uniqueVertCount++
		}
	}

	// Building tris from unique vertices
	newTris := make([]int, 0)
	for triI := 0; triI < len(m.indices); triI += 3 {
		v1 := Vector3ToInt(data[m.indices[triI+0]], decimalPlace)
		v2 := Vector3ToInt(data[m.indices[triI+1]], decimalPlace)
		v3 := Vector3ToInt(data[m.indices[triI+2]], decimalPlace)

		if v1 == v2 {
			continue
		}

		if v1 == v3 {
			continue
		}

		if v2 == v3 {
			continue
		}

		vertLUUsed[v1] = true
		vertLUUsed[v2] = true
		vertLUUsed[v3] = true
		newTris = append(newTris, vertILU[v1], vertILU[v2], vertILU[v3])
	}

	finalV4Data := make(map[string][]vector4.Float64)
	for key := range m.v4Data {
		finalV4Data[key] = make([]vector4.Float64, 0)
	}

	finalV3Data := make(map[string][]vector3.Float64)
	for key := range m.v3Data {
		finalV3Data[key] = make([]vector3.Float64, 0)
	}

	finalV2Data := make(map[string][]vector2.Float64)
	for key := range m.v2Data {
		finalV2Data[key] = make([]vector2.Float64, 0)
	}

	finalV1Data := make(map[string][]float64)
	for key := range m.v1Data {
		finalV1Data[key] = make([]float64, 0)
	}

	shiftBy := make([]int, uniqueVertCount)
	curShift := 0
	for vertIndex := 0; vertIndex < uniqueVertCount; vertIndex++ {

		originalIndex := vertIToOriginalLU[vertIndex]
		v := data[originalIndex]
		vi := Vector3ToInt(v, decimalPlace)
		if vertLUUsed[vi] {
			for key, vals := range m.v4Data {
				finalV4Data[key] = append(finalV4Data[key], vals[originalIndex])
			}

			for key, vals := range m.v3Data {
				finalV3Data[key] = append(finalV3Data[key], vals[originalIndex])
			}

			for key, vals := range m.v2Data {
				finalV2Data[key] = append(finalV2Data[key], vals[originalIndex])
			}

			for key, vals := range m.v1Data {
				finalV1Data[key] = append(finalV1Data[key], vals[originalIndex])
			}

		} else {
			// Not used, need to shift triangles who's points point to vertices that come after this unsed one
			curShift++
		}
		shiftBy[vertIndex] = curShift
	}

	// Shift all the triangles appropriately since we just removed a bunch of vertices no longer used
	for triI := 0; triI < len(newTris); triI++ {
		newTris[triI] -= shiftBy[newTris[triI]]
	}

	return Mesh{
		indices:   newTris,
		v4Data:    finalV4Data,
		v3Data:    finalV3Data,
		v2Data:    finalV2Data,
		v1Data:    finalV1Data,
		materials: nil, // TODO: Figure out the new tri counts for the materials. Util then clear em
		topology:  m.topology,
	}
}

func (m Mesh) VertexNeighborTable() VertexLUT {
	table := VertexLUT{}

	switch m.topology {
	case TriangleTopology:
		for triI := 0; triI < len(m.indices); triI += 3 {
			p1 := m.indices[triI]
			p2 := m.indices[triI+1]
			p3 := m.indices[triI+2]

			table.Link(p1, p2)
			table.Link(p2, p3)
			table.Link(p1, p3)
		}

	case LineStripTopology:
		for i := 1; i < len(m.indices); i++ {
			table.Link(m.indices[i-1], m.indices[i])
		}

	case LineTopology:
		for i := 1; i < len(m.indices); i += 2 {
			table.Link(m.indices[i-1], m.indices[i])
		}

	case LineLoopTopology:
		for i := 1; i < len(m.indices); i++ {
			table.Link(m.indices[i-1], m.indices[i])
		}
		table.Link(m.indices[0], m.indices[len(m.indices)-1])

	default:
		panic(fmt.Errorf("unimplemented topology for vertex LUT: %s", m.topology.String()))
	}

	return table
}

func (m Mesh) requireTopology(t Topology) {
	if m.topology != t {
		panic(fmt.Errorf("can not perform operation for a mesh with a topology of %s, requires %s topology", m.topology.String(), t.String()))
	}
}

func (m Mesh) Float4Attribute(attr string) iter.ArrayIterator[vector4.Float64] {
	m.requireV4Attribute(attr)
	return iter.Array(m.v4Data[attr])
}

func (m Mesh) SetFloat4Attribute(attr string, data []vector4.Float64) Mesh {
	finalV4Data := make(map[string][]vector4.Float64)
	for key, val := range m.v4Data {
		finalV4Data[key] = val
	}
	finalV4Data[attr] = data

	if len(data) == 0 {
		delete(finalV4Data, attr)
	}

	return Mesh{
		v4Data:    finalV4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) SetFloat4Data(data map[string][]vector4.Float64) Mesh {
	return Mesh{
		v1Data:    m.v1Data,
		v2Data:    m.v2Data,
		v3Data:    m.v3Data,
		v4Data:    data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) Float3Attribute(attr string) iter.ArrayIterator[vector3.Float64] {
	m.requireV3Attribute(attr)
	return iter.Array(m.v3Data[attr])
}

func (m Mesh) SetFloat3Attribute(attr string, data []vector3.Float64) Mesh {
	finalV3Data := make(map[string][]vector3.Float64)
	for key, val := range m.v3Data {
		finalV3Data[key] = val
	}
	finalV3Data[attr] = data

	if len(data) == 0 {
		delete(finalV3Data, attr)
	}

	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    finalV3Data,
		v2Data:    m.v2Data,
		v1Data:    m.v1Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) SetFloat3Data(data map[string][]vector3.Float64) Mesh {
	return Mesh{
		v1Data:    m.v1Data,
		v2Data:    m.v2Data,
		v3Data:    data,
		v4Data:    m.v4Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) CopyFloat4Attribute(src Mesh, attr string) Mesh {
	return m.SetFloat4Attribute(attr, src.v4Data[attr])
}

func (m Mesh) CopyFloat3Attribute(src Mesh, attr string) Mesh {
	return m.SetFloat3Attribute(attr, src.v3Data[attr])
}

func (m Mesh) CopyFloat2Attribute(src Mesh, attr string) Mesh {
	return m.SetFloat2Attribute(attr, src.v2Data[attr])
}

func (m Mesh) CopyFloat1Attribute(src Mesh, attr string) Mesh {
	return m.SetFloat1Attribute(attr, src.v1Data[attr])
}

func (m Mesh) Float2Attribute(attr string) iter.ArrayIterator[vector2.Float64] {
	m.requireV2Attribute(attr)
	return iter.Array(m.v2Data[attr])
}

func (m Mesh) SetFloat2Attribute(attr string, data []vector2.Float64) Mesh {
	finalV2Data := make(map[string][]vector2.Float64)
	for key, val := range m.v2Data {
		finalV2Data[key] = val
	}
	finalV2Data[attr] = data

	if len(data) == 0 {
		delete(finalV2Data, attr)
	}

	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    finalV2Data,
		v1Data:    m.v1Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) SetFloat2Data(data map[string][]vector2.Float64) Mesh {
	return Mesh{
		v1Data:    m.v1Data,
		v2Data:    data,
		v3Data:    m.v3Data,
		v4Data:    m.v4Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) Float1Attribute(attr string) iter.ArrayIterator[float64] {
	m.requireV1Attribute(attr)
	return iter.Array(m.v1Data[attr])
}

func (m Mesh) SetFloat1Attribute(attr string, data []float64) Mesh {
	finalV1Data := make(map[string][]float64)
	for key, val := range m.v1Data {
		finalV1Data[key] = val
	}
	finalV1Data[attr] = data

	if len(data) == 0 {
		delete(finalV1Data, attr)
	}

	return Mesh{
		v4Data:    m.v4Data,
		v3Data:    m.v3Data,
		v2Data:    m.v2Data,
		v1Data:    finalV1Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) SetFloat1Data(data map[string][]float64) Mesh {
	return Mesh{
		v1Data:    data,
		v2Data:    m.v2Data,
		v3Data:    m.v3Data,
		v4Data:    m.v4Data,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) ClearAttributeData() Mesh {
	return Mesh{
		v1Data:    nil,
		v2Data:    nil,
		v3Data:    nil,
		v4Data:    nil,
		indices:   m.indices,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) HasVertexAttribute(atr string) bool {
	if m.HasFloat4Attribute(atr) {
		return true
	}

	if m.HasFloat3Attribute(atr) {
		return true
	}

	if m.HasFloat2Attribute(atr) {
		return true
	}

	if m.HasFloat1Attribute(atr) {
		return true
	}

	return false
}

func (m Mesh) HasFloat4Attribute(atr string) bool {
	for v4Atr := range m.v4Data {
		if v4Atr == atr {
			return true
		}
	}

	return false
}

func (m Mesh) HasFloat3Attribute(atr string) bool {
	for v3Atr := range m.v3Data {
		if v3Atr == atr {
			return true
		}
	}

	return false
}

func (m Mesh) HasFloat2Attribute(atr string) bool {
	for v2Atr := range m.v2Data {
		if v2Atr == atr {
			return true
		}
	}

	return false
}

func (m Mesh) HasFloat1Attribute(atr string) bool {
	for v1Atr := range m.v1Data {
		if v1Atr == atr {
			return true
		}
	}

	return false
}

func (m Mesh) requireV4Attribute(atr string) {
	if !m.HasFloat4Attribute(atr) {
		panic(fmt.Errorf("can not perform operation for a mesh without the attribute '%s'", atr))
	}
}

func (m Mesh) requireV3Attribute(atr string) {
	if !m.HasFloat3Attribute(atr) {
		panic(fmt.Errorf("can not perform operation for a mesh without the attribute '%s'", atr))
	}
}

func (m Mesh) requireV2Attribute(atr string) {
	if !m.HasFloat2Attribute(atr) {
		panic(fmt.Errorf("can not perform operation for a mesh without the attribute '%s'", atr))
	}
}

func (m Mesh) requireV1Attribute(atr string) {
	if !m.HasFloat1Attribute(atr) {
		panic(fmt.Errorf("can not perform operation for a mesh without the attribute '%s'", atr))
	}
}

func (m Mesh) AttributeLength() int {
	for _, v := range m.v3Data {
		return len(v)
	}
	for _, v := range m.v2Data {
		return len(v)
	}
	for _, v := range m.v1Data {
		return len(v)
	}
	return 0
}

func (m Mesh) Translate(v vector3.Float64) Mesh {
	m.requireV3Attribute(PositionAttribute)
	oldData := m.v3Data[PositionAttribute]
	finalVerts := make([]vector3.Float64, len(oldData))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = oldData[i].Add(v)
	}
	return m.SetFloat3Attribute(PositionAttribute, finalVerts)
}

func (m Mesh) OctTree() *trees.OctTree {
	treeDepth := trees.OctreeDepthFromCount(m.PrimitiveCount())
	return m.OctTreeWithAttributeAndDepth(PositionAttribute, treeDepth)
}

func (m Mesh) OctTreeDepth(depth int) *trees.OctTree {
	return m.OctTreeWithAttributeAndDepth(PositionAttribute, depth)
}

func (m Mesh) OctTreeWithAttributeAndDepth(atr string, depth int) *trees.OctTree {
	primitives := make([]trees.Element, m.PrimitiveCount())

	m.ScanPrimitives(func(i int, p Primitive) {
		primitives[i] = p.Scope(atr)
	})

	return trees.NewOctreeWithDepth(primitives, depth)
}
