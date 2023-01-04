package modeling

import (
	"fmt"
	"math"
	"runtime"
	"sync"

	"github.com/EliCDavis/vector"
)

type Mesh struct {
	v3Data    map[string][]vector.Vector3
	v2Data    map[string][]vector.Vector2
	v1Data    map[string][]float64
	indices   []int
	materials []MeshMaterial
	topology  Topology
}

func NewMesh(
	indices []int,
	v3Data map[string][]vector.Vector3,
	v2Data map[string][]vector.Vector2,
	v1Data map[string][]float64,
	materials []MeshMaterial) Mesh {
	return Mesh{
		indices:   indices,
		materials: materials,
		topology:  Triangle,
		v3Data:    v3Data,
		v2Data:    v2Data,
		v1Data:    v1Data,
	}
}

func NewPointCloud(
	indices []int,
	v3Data map[string][]vector.Vector3,
	v2Data map[string][]vector.Vector2,
	v1Data map[string][]float64,
	materials []MeshMaterial) Mesh {
	return Mesh{
		indices:   indices,
		materials: materials,
		topology:  Point,
		v3Data:    v3Data,
		v2Data:    v2Data,
		v1Data:    v1Data,
	}
}

// Creates a new triangle mesh with no vertices or attribute data
func EmptyMesh() Mesh {
	return Mesh{
		indices:   make([]int, 0),
		materials: make([]MeshMaterial, 0),
		topology:  Triangle,
		v3Data:    make(map[string][]vector.Vector3),
		v2Data:    make(map[string][]vector.Vector2),
		v1Data:    make(map[string][]float64),
	}
}

func NewTexturedMesh(
	triangles []int,
	vertices []vector.Vector3,
	normals []vector.Vector3,
	uvs []vector.Vector2,
) Mesh {
	return Mesh{
		indices: triangles,
		v3Data: map[string][]vector.Vector3{
			PositionAttribute: vertices,
			NormalAttribute:   normals,
		},
		v2Data: map[string][]vector.Vector2{
			TexCoordAttribute: uvs,
		},
		materials: []MeshMaterial{{len(triangles) / 3, nil}},
		topology:  Triangle,
	}
}

func MeshFromView(view MeshView) Mesh {
	return Mesh{
		indices:  view.Indices,
		v3Data:   view.Float3Data,
		v2Data:   view.Float2Data,
		v1Data:   view.Float1Data,
		topology: Triangle,
	}
}

// View exposes the underlying data to be modified. Using this breaks the
// immutable design of the system, but required for some mesh processing.
//
// Modifying the data stored in the mesh found here will directly update the
// mesh, and side-steps any type of validation we could have done previously.
//
// If you make changes to this view, assume the mesh and all ancestors of said
// mesh have just become garbage.
func (m Mesh) View() MeshView {
	return MeshView{
		Float3Data: m.v3Data,
		Float2Data: m.v2Data,
		Float1Data: m.v1Data,
		Indices:    m.indices,
	}
}

func (m Mesh) Materials() []MeshMaterial {
	return m.materials
}

func (m Mesh) SetMaterial(mat Material) Mesh {
	return NewMesh(m.indices, m.v3Data, m.v2Data, m.v1Data, []MeshMaterial{{NumOfTris: len(m.indices) / 3, Material: &mat}})
}

func (m Mesh) Tri(i int) Tri {
	return Tri{
		mesh:          &m,
		startingIndex: i * 3,
	}
}

func (m Mesh) TriCount() int {
	return len(m.indices) / 3
}

func (m Mesh) Append(other Mesh) Mesh {

	if m.topology != other.topology {
		panic(fmt.Errorf("can not combine meshes with different topologies (%s != %s)", m.topology.String(), other.topology.String()))
	}

	mAtrLength := m.AttributeLength()
	oAtrLength := other.AttributeLength()

	finalV3Data := make(map[string][]vector.Vector3)
	for atr, data := range m.v3Data {
		finalV3Data[atr] = data

		if _, ok := other.v3Data[atr]; !ok {
			for i := 0; i < oAtrLength; i++ {
				finalV3Data[atr] = append(finalV3Data[atr], vector.Vector3Zero())
			}
		}
	}

	for atr, data := range other.v3Data {
		if _, ok := finalV3Data[atr]; !ok {
			for i := 0; i < mAtrLength; i++ {
				finalV3Data[atr] = append(finalV3Data[atr], vector.Vector3Zero())
			}
		}
		finalV3Data[atr] = append(finalV3Data[atr], data...)
	}

	finalV2Data := make(map[string][]vector.Vector2)
	for atr, data := range m.v2Data {
		finalV2Data[atr] = data

		if _, ok := other.v2Data[atr]; !ok {
			for i := 0; i < oAtrLength; i++ {
				finalV2Data[atr] = append(finalV2Data[atr], vector.Vector2Zero())
			}
		}
	}

	for atr, data := range other.v2Data {
		if _, ok := finalV2Data[atr]; !ok {
			for i := 0; i < mAtrLength; i++ {
				finalV2Data[atr] = append(finalV2Data[atr], vector.Vector2Zero())
			}
		}
		finalV2Data[atr] = append(finalV2Data[atr], data...)
	}

	finalV1Data := make(map[string][]float64)
	for atr, data := range m.v1Data {
		finalV1Data[atr] = data

		if _, ok := other.v1Data[atr]; !ok {
			for i := 0; i < oAtrLength; i++ {
				finalV1Data[atr] = append(finalV1Data[atr], 0)
			}
		}
	}

	for atr, data := range other.v1Data {
		if _, ok := finalV1Data[atr]; !ok {
			for i := 0; i < mAtrLength; i++ {
				finalV1Data[atr] = append(finalV1Data[atr], 0)
			}
		}
		finalV1Data[atr] = append(finalV1Data[atr], data...)
	}

	finalTris := append(m.indices, other.indices...)
	finalMaterials := append(m.materials, other.materials...)
	for i := len(m.indices); i < len(finalTris); i++ {
		finalTris[i] += mAtrLength
	}

	return NewMesh(finalTris, finalV3Data, finalV2Data, finalV1Data, finalMaterials)
}

// Translate(v) is shorthand for TranslateAttribute3D(V, "Position")
func (m Mesh) Translate(v vector.Vector3) Mesh {
	return m.TranslateAttribute3D(PositionAttribute, v)
}

func (m Mesh) TranslateAttribute3D(attribute string, v vector.Vector3) Mesh {
	m.requireV3Attribute(attribute)
	oldData := m.v3Data[attribute]
	finalVerts := make([]vector.Vector3, len(oldData))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = oldData[i].Add(v)
	}
	return m.SetFloat3Attribute(attribute, finalVerts)
}

func (m Mesh) Rotate(q Quaternion) Mesh {
	return m.
		RotateAttribute3D(PositionAttribute, q).
		RotateAttribute3D(NormalAttribute, q)
}

func (m Mesh) RotateAttribute3D(attribute string, q Quaternion) Mesh {
	m.requireV3Attribute(attribute)

	finalMesh := m
	oldData := m.v3Data[attribute]
	finalVerts := make([]vector.Vector3, len(oldData))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = q.Rotate(oldData[i])
	}

	return finalMesh.SetFloat3Attribute(attribute, finalVerts)
}

// Scale(o, a) is shorthand for ScaleAttribute3D("Position", o, a)
func (m Mesh) Scale(origin, amount vector.Vector3) Mesh {
	return m.ScaleAttribute3D(PositionAttribute, origin, amount)
}

func (m Mesh) ScaleAttribute3D(attribute string, origin, amount vector.Vector3) Mesh {
	return m.ModifyFloat3Attribute(attribute, func(i int, v vector.Vector3) vector.Vector3 {
		return origin.Add(v.Sub(origin).MultByVector(amount))
	})
}

func (m Mesh) BoundingBox(atr string) AABB {
	m.requireV3Attribute(atr)
	oldData := m.v3Data[atr]

	min := vector.NewVector3(math.Inf(1), math.Inf(1), math.Inf(1))
	max := vector.NewVector3(math.Inf(-1), math.Inf(-1), math.Inf(-1))
	for _, v := range oldData {
		min = min.SetX(math.Min(v.X(), min.X()))
		min = min.SetY(math.Min(v.Y(), min.Y()))
		min = min.SetZ(math.Min(v.Z(), min.Z()))

		max = max.SetX(math.Max(v.X(), max.X()))
		max = max.SetY(math.Max(v.Y(), max.Y()))
		max = max.SetZ(math.Max(v.Z(), max.Z()))
	}

	center := max.Sub(min).DivByConstant(2).Add(min)

	return NewAABB(center, max.Sub(min))
}

func (m Mesh) CenterFloat3Attribute(atr string) Mesh {
	m.requireV3Attribute(atr)
	oldData := m.v3Data[atr]
	modified := make([]vector.Vector3, len(oldData))

	min := vector.NewVector3(math.Inf(1), math.Inf(1), math.Inf(1))
	max := vector.NewVector3(math.Inf(-1), math.Inf(-1), math.Inf(-1))
	for _, v := range oldData {
		min = min.SetX(math.Min(v.X(), min.X()))
		min = min.SetY(math.Min(v.Y(), min.Y()))
		min = min.SetZ(math.Min(v.Z(), min.Z()))

		max = max.SetX(math.Max(v.X(), max.X()))
		max = max.SetY(math.Max(v.Y(), max.Y()))
		max = max.SetZ(math.Max(v.Z(), max.Z()))
	}

	center := max.Sub(min).DivByConstant(2).Add(min)
	for i, v := range oldData {
		modified[i] = v.Sub(center)
	}

	return m.SetFloat3Attribute(atr, modified)
}

func (m Mesh) ScanFloat3Attribute(atr string, f func(i int, v vector.Vector3)) Mesh {
	m.requireV3Attribute(atr)

	for i, v := range m.v3Data[atr] {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat3AttributeParallel(atr string, f func(i int, v vector.Vector3)) Mesh {
	m.requireV3Attribute(atr)

	if len(m.v3Data[atr]) <= runtime.NumCPU()*4 || runtime.NumCPU() == 1 {
		return m.ScanFloat3Attribute(atr, f)
	}

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v3Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v3Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				f(i, m.v3Data[atr][i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ScanFloat2Attribute(atr string, f func(i int, v vector.Vector2)) Mesh {
	m.requireV2Attribute(atr)

	for i, v := range m.v2Data[atr] {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat2AttributeParallel(atr string, f func(i int, v vector.Vector2)) Mesh {
	m.requireV2Attribute(atr)

	if len(m.v2Data[atr]) <= runtime.NumCPU()*4 || runtime.NumCPU() == 1 {
		return m.ScanFloat2Attribute(atr, f)
	}

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v2Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v2Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				f(i, m.v2Data[atr][i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ScanFloat1Attribute(atr string, f func(i int, v float64)) Mesh {
	m.requireV1Attribute(atr)

	for i, v := range m.v1Data[atr] {
		f(i, v)
	}

	return m
}

func (m Mesh) ScanFloat1AttributeParallel(atr string, f func(i int, v float64)) Mesh {
	m.requireV1Attribute(atr)

	if len(m.v1Data[atr]) <= runtime.NumCPU()*4 || runtime.NumCPU() == 1 {
		return m.ScanFloat1Attribute(atr, f)
	}

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v1Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v1Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				f(i, m.v1Data[atr][i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m
}

func (m Mesh) ModifyFloat3Attribute(atr string, f func(i int, v vector.Vector3) vector.Vector3) Mesh {
	m.requireV3Attribute(atr)
	oldData := m.v3Data[atr]
	modified := make([]vector.Vector3, len(oldData))

	for i, v := range oldData {
		modified[i] = f(i, v)
	}

	return m.SetFloat3Attribute(atr, modified)
}

func (m Mesh) ModifyFloat3AttributeParallel(atr string, f func(i int, v vector.Vector3) vector.Vector3) Mesh {
	m.requireV3Attribute(atr)
	if len(m.v3Data[atr]) < runtime.NumCPU() || runtime.NumCPU() == 1 {
		return m.ModifyFloat3Attribute(atr, f)
	}

	oldData := m.v3Data[atr]
	modified := make([]vector.Vector3, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v3Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v3Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				modified[i] = f(i, m.v3Data[atr][i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m.SetFloat3Attribute(atr, modified)
}

func (m Mesh) ModifyFloat2Attribute(atr string, f func(i int, v vector.Vector2) vector.Vector2) Mesh {
	m.requireV2Attribute(atr)
	oldData := m.v2Data[atr]
	modified := make([]vector.Vector2, len(oldData))

	for i, v := range oldData {
		modified[i] = f(i, v)
	}

	return m.SetFloat2Attribute(atr, modified)
}

func (m Mesh) ModifyFloat2AttributeParallel(atr string, f func(i int, v vector.Vector2) vector.Vector2) Mesh {
	m.requireV2Attribute(atr)
	if len(m.v2Data[atr]) < runtime.NumCPU() || runtime.NumCPU() == 1 {
		return m.ModifyFloat2Attribute(atr, f)
	}

	oldData := m.v2Data[atr]
	modified := make([]vector.Vector2, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v2Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v2Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				modified[i] = f(i, m.v2Data[atr][i])
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
	m.requireV1Attribute(atr)
	if len(m.v1Data[atr]) < runtime.NumCPU() || runtime.NumCPU() == 1 {
		return m.ModifyFloat1Attribute(atr, f)
	}

	oldData := m.v1Data[atr]
	modified := make([]float64, len(oldData))

	var wg sync.WaitGroup

	workSize := int(math.Floor(float64(len(m.v1Data[atr])) / float64(runtime.NumCPU())))
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		jobSize := workSize

		// Make sure to clean up potential last cell due to rounding error of
		// division of number of CPUs
		if i == runtime.NumCPU()-1 {
			jobSize = len(m.v1Data[atr]) - (workSize * i)
		}

		go func(start, size int) {
			defer wg.Done()
			for i := start; i < start+size; i++ {
				modified[i] = f(i, m.v1Data[atr][i])
			}
		}(workSize*i, jobSize)
	}

	wg.Wait()

	return m.SetFloat1Attribute(atr, modified)
}

func (m Mesh) CalculateFlatNormals() Mesh {
	m.requireTopology(Triangle)
	m.requireV3Attribute(PositionAttribute)

	vertices := m.v3Data[PositionAttribute]
	normals := make([]vector.Vector3, len(vertices))
	for i := range normals {
		normals[i] = vector.Vector3One()
	}

	tris := m.indices
	for triIndex := 0; triIndex < len(tris); triIndex += 3 {
		p1 := tris[triIndex]
		p2 := tris[triIndex+1]
		p3 := tris[triIndex+2]
		// normalize(cross(B-A, C-A))
		normalized := vertices[p2].Sub(vertices[p1]).Cross(vertices[p3].Sub(vertices[p1])).Normalized()
		normals[p1] = normalized
		normals[p2] = normalized
		normals[p3] = normalized
	}

	for i, n := range normals {
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(NormalAttribute, normals)
}

func (m Mesh) WeldByFloat3Attribute(attribute string, decimalPlace int) Mesh {
	m.requireV3Attribute(attribute)
	m.requireTopology(Triangle)

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

	finalV3Data := make(map[string][]vector.Vector3)
	for key := range m.v3Data {
		finalV3Data[key] = make([]vector.Vector3, 0)
	}

	finalV2Data := make(map[string][]vector.Vector2)
	for key := range m.v2Data {
		finalV2Data[key] = make([]vector.Vector2, 0)
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
		v3Data:    finalV3Data,
		v2Data:    finalV2Data,
		v1Data:    finalV1Data,
		materials: nil, // TODO: Figure out the new tri counts for the materials. Util then clear em
		topology:  m.topology,
	}
}

func (m Mesh) VertexNeighborTable() VertexLUT {
	table := VertexLUT{}
	for triI := 0; triI < len(m.indices); triI += 3 {
		p1 := m.indices[triI]
		p2 := m.indices[triI+1]
		p3 := m.indices[triI+2]

		table.AddLookup(p1, p2)
		table.AddLookup(p1, p3)

		table.AddLookup(p2, p1)
		table.AddLookup(p2, p3)

		table.AddLookup(p3, p1)
		table.AddLookup(p3, p2)
	}
	return table
}

func (m Mesh) requireTopology(t Topology) {
	if m.topology != t {
		panic(fmt.Errorf("can not perform operation for a mesh with a topology of %s, requires %s topology", m.topology.String(), t.String()))
	}
}

func (m Mesh) SetFloat3Attribute(atr string, data []vector.Vector3) Mesh {
	finalV3Data := make(map[string][]vector.Vector3)
	for key, val := range m.v3Data {
		finalV3Data[key] = val
	}
	finalV3Data[atr] = data
	return NewMesh(
		m.indices,
		finalV3Data,
		m.v2Data,
		m.v1Data,
		m.materials,
	)
}
func (m Mesh) SetFloat2Attribute(atr string, data []vector.Vector2) Mesh {
	finalV2Data := make(map[string][]vector.Vector2)
	for key, val := range m.v2Data {
		finalV2Data[key] = val
	}
	finalV2Data[atr] = data
	return NewMesh(
		m.indices,
		m.v3Data,
		finalV2Data,
		m.v1Data,
		m.materials,
	)
}

func (m Mesh) SetFloat1Attribute(atr string, data []float64) Mesh {
	finalV1Data := make(map[string][]float64)
	for key, val := range m.v1Data {
		finalV1Data[key] = val
	}
	finalV1Data[atr] = data
	return NewMesh(
		m.indices,
		m.v3Data,
		m.v2Data,
		finalV1Data,
		m.materials,
	)
}

func (m Mesh) HasVertexAttribute(atr string) bool {
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

func (m Mesh) SmoothLaplacian(iterations int, smoothingFactor float64) Mesh {
	m.requireTopology(Triangle)
	m.requireV3Attribute(PositionAttribute)

	lut := m.VertexNeighborTable()

	oldVertices := m.v3Data[PositionAttribute]
	vertices := make([]vector.Vector3, len(oldVertices))
	for i := range vertices {
		vertices[i] = oldVertices[i]
	}

	for i := 0; i < iterations; i++ {
		for vi, vertex := range vertices {
			vs := vector.Vector3Zero()

			for vn := range lut.Lookup(vi) {
				vs = vs.Add(vertices[vn])
			}

			vertices[vi] = vertex.Add(
				vs.
					DivByConstant(float64(lut.Count(vi))).
					Sub(vertex).
					MultByConstant(smoothingFactor))
		}
	}

	return m.SetFloat3Attribute(PositionAttribute, vertices)
}

func (m Mesh) CalculateSmoothNormals() Mesh {
	m.requireTopology(Triangle)
	m.requireV3Attribute(PositionAttribute)

	vertices := m.v3Data[PositionAttribute]
	normals := make([]vector.Vector3, len(vertices))
	for i := range normals {
		normals[i] = vector.Vector3Zero()
	}

	tris := m.indices
	for triIndex := 0; triIndex < len(tris); triIndex += 3 {
		p1 := tris[triIndex]
		p2 := tris[triIndex+1]
		p3 := tris[triIndex+2]
		// normalize(cross(B-A, C-A))
		normalized := vertices[p2].Sub(vertices[p1]).Cross(vertices[p3].Sub(vertices[p1]))

		// This occurs whenever the given tri is actually just a line
		if math.IsNaN(normalized.X()) {
			continue
		}

		normals[p1] = normals[p1].Add(normalized)
		normals[p2] = normals[p2].Add(normalized)
		normals[p3] = normals[p3].Add(normalized)
	}

	for i, n := range normals {
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(NormalAttribute, normals)
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

func (m Mesh) RemoveUnusedIndices() Mesh {
	finalTris := make([]int, len(m.indices))
	finalV3Data := make(map[string][]vector.Vector3)
	finalV2Data := make(map[string][]vector.Vector2)
	finalV1Data := make(map[string][]float64)

	used := make([]bool, m.AttributeLength())
	for _, t := range m.indices {
		used[t] = true
	}

	shiftBy := make([]int, m.AttributeLength())
	skipped := 0
	for i := range shiftBy {
		if !used[i] {
			skipped++
		}
		shiftBy[i] = skipped
	}

	for atr, vals := range m.v3Data {
		finalAtrVals := make([]vector.Vector3, 0)
		for i, v := range vals {
			if used[i] {
				finalAtrVals = append(finalAtrVals, v)
			}
		}
		finalV3Data[atr] = finalAtrVals
	}

	for atr, vals := range m.v2Data {
		finalAtrVals := make([]vector.Vector2, 0)
		for i, v := range vals {
			if used[i] {
				finalAtrVals = append(finalAtrVals, v)
			}
		}
		finalV2Data[atr] = finalAtrVals
	}

	for atr, vals := range m.v1Data {
		finalAtrVals := make([]float64, 0)
		for i, v := range vals {
			if used[i] {
				finalAtrVals = append(finalAtrVals, v)
			}
		}
		finalV1Data[atr] = finalAtrVals
	}

	for triI := 0; triI < len(finalTris); triI++ {
		finalTris[triI] = m.indices[triI] - shiftBy[m.indices[triI]]
	}

	return Mesh{
		indices:   finalTris,
		v3Data:    finalV3Data,
		v2Data:    finalV2Data,
		v1Data:    finalV1Data,
		materials: m.materials,
		topology:  m.topology,
	}
}

// SplitOnUniqueMaterials generates a mesh per material,
func (m Mesh) SplitOnUniqueMaterials() []Mesh {
	if len(m.materials) < 2 {
		return []Mesh{m}
	}

	workingMeshes := make(map[*Material]*Mesh)

	curMatIndex := 0
	trisFromOtherMats := 0

	workingMeshes[m.materials[curMatIndex].Material] = &Mesh{
		v3Data: m.v3Data,
		v2Data: m.v2Data,
		v1Data: m.v1Data,
		materials: []MeshMaterial{
			{
				NumOfTris: 0,
				Material:  m.materials[curMatIndex].Material,
			},
		},
	}

	for triStart := 0; triStart < len(m.indices); triStart += 3 {
		if m.materials[curMatIndex].NumOfTris+trisFromOtherMats <= triStart/3 {
			trisFromOtherMats += m.materials[curMatIndex].NumOfTris
			curMatIndex++
			if _, ok := workingMeshes[m.materials[curMatIndex].Material]; !ok {
				workingMeshes[m.materials[curMatIndex].Material] = &Mesh{
					v3Data: m.v3Data,
					v2Data: m.v2Data,
					v1Data: m.v1Data,
					materials: []MeshMaterial{
						{
							NumOfTris: 0,
							Material:  m.materials[curMatIndex].Material,
						},
					},
					topology: m.topology,
				}
			}
		}
		mesh := workingMeshes[m.materials[curMatIndex].Material]
		mesh.indices = append(
			mesh.indices,
			m.indices[triStart],
			m.indices[triStart+1],
			m.indices[triStart+2],
		)
		mesh.materials[0].NumOfTris += 1
	}

	finalMeshes := make([]Mesh, 0, len(workingMeshes))
	for _, m := range workingMeshes {
		finalMeshes = append(finalMeshes, m.RemoveUnusedIndices())
	}
	return finalMeshes
}
