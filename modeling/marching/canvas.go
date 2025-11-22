package marching

import (
	"fmt"
	"math"
	"runtime"
	"sync"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func interpolationValueFromCutoff(v1v, v2v, cutoff float64) float64 {
	return (cutoff - v1v) / (v2v - v1v)
}

func interpolateV3(v1, v2 vector3.Float64, t float64) vector3.Float64 {
	return v2.Sub(v1).Scale(t).Add(v1)
}

func interpolateV2(v1, v2 vector2.Float64, t float64) vector2.Float64 {
	return v2.Sub(v1).Scale(t).Add(v1)
}

func interpolateV1(v1, v2, t float64) float64 {
	return ((v2 - v1) * t) + v1
}

func interpolateVerts(v1, v2 vector3.Float64, v1v, v2v, cutoff float64) vector3.Float64 {
	t := interpolationValueFromCutoff(v1v, v2v, cutoff)
	// return v2.Sub(v1).Scale(t).Add(v1)
	return v1.Scale(1 - t).Add(v2.Scale(t))
}

func lookupOrAdd(data *workingData, vert vector3.Float64) int {
	distritized := modeling.Vector3ToInt(vert, 4)
	key := vector3.New(int32(distritized.X()), int32(distritized.Y()), int32(distritized.Z()))

	if foundIndex, ok := data.vertLookup[key]; ok {
		return foundIndex
	}

	index := len(data.verts)
	data.vertLookup[key] = index
	data.verts = append(data.verts, vert)
	return index
}

type MarchingDataType int64

const (
	Float1 MarchingDataType = iota
	Float2
	Float3
)

const (
	marchingSectionSize        = 100
	marchingSectionSizeSquared = marchingSectionSize * marchingSectionSize
	marchingSectionSizeCubed   = marchingSectionSize * marchingSectionSize * marchingSectionSize
)

type marchingSection struct {
	dataType  MarchingDataType
	positions map[vector3.Int]int
}

type float1MarchingSection = []float64
type float2MarchingSection = []vector2.Float64
type float3MarchingSection = []vector3.Float64

type MarchingCanvas struct {
	float1Data   []float1MarchingSection
	float2Data   []float2MarchingSection
	float3Data   []float3MarchingSection
	sections     map[string]*marchingSection
	cubesPerUnit float64
	chunkMutex   *sync.Mutex
}

func NewMarchingCanvas(cubesPerUnit float64) *MarchingCanvas {
	return &MarchingCanvas{
		float1Data:   make([]float1MarchingSection, 0),
		float2Data:   make([]float2MarchingSection, 0),
		float3Data:   make([]float3MarchingSection, 0),
		sections:     make(map[string]*marchingSection),
		cubesPerUnit: cubesPerUnit,
		chunkMutex:   &sync.Mutex{},
	}
}

func (d MarchingCanvas) index(x, y, z int) int {
	return (z * marchingSectionSizeSquared) + (y * marchingSectionSize) + x
}

func (d *MarchingCanvas) chunkIndex_atomic(section *marchingSection, vec vector3.Int) int {
	d.chunkMutex.Lock()
	defer d.chunkMutex.Unlock()
	chunkIndex, ok := section.positions[vec]
	if !ok {
		switch section.dataType {
		case Float1:
			chunkIndex = len(d.float1Data)
			d.float1Data = append(d.float1Data, make(float1MarchingSection, marchingSectionSizeCubed))

		case Float2:
			chunkIndex = len(d.float2Data)
			d.float2Data = append(d.float2Data, make(float2MarchingSection, marchingSectionSizeCubed))

		case Float3:
			chunkIndex = len(d.float3Data)
			d.float3Data = append(d.float3Data, make(float3MarchingSection, marchingSectionSizeCubed))
		}
		section.positions[vec] = chunkIndex
	}
	return chunkIndex
}

func (d MarchingCanvas) canvasPosToChunkPos(x, y, z int) vector3.Int {
	return vector3.New(
		int(math.Floor(float64(x)/marchingSectionSize)),
		int(math.Floor(float64(y)/marchingSectionSize)),
		int(math.Floor(float64(z)/marchingSectionSize)),
	)
}

// func (d *MarchingCanvas) addFloat1Value(section *marchingSection, x, y, z int, val float64) {
// 	if section.dataType != Float1 {
// 		panic(fmt.Errorf("cant add float1 to section with type of: %d", section.dataType))
// 	}

// 	chunkPos := d.canvasPosToChunkPos(x, y, z)

// 	index := d.chunkIndex_atomic(section, chunkPos)

// 	shiftedPos := vector3.New(
// 		X: x - (chunkPos.X * marchingSectionSize),
// 		Y: y - (chunkPos.Y * marchingSectionSize),
// 		Z: z - (chunkPos.Z * marchingSectionSize),
// 	}

// 	d.float1Data[index][d.index(shiftedPos.X, shiftedPos.Y, shiftedPos.Z)] += val
// }

func (d MarchingCanvas) fieldBounds(f Field) (vector3.Int, vector3.Int) {
	min := f.Domain.Min()
	max := f.Domain.Max()

	minCanvas := vector3.New(
		int(math.Floor(min.X()*d.cubesPerUnit))-1,
		int(math.Floor(min.Y()*d.cubesPerUnit))-1,
		int(math.Floor(min.Z()*d.cubesPerUnit))-1,
	)

	maxCanvas := vector3.New(
		int(math.Ceil(max.X()*d.cubesPerUnit))+1,
		int(math.Ceil(max.Y()*d.cubesPerUnit))+1,
		int(math.Ceil(max.Z()*d.cubesPerUnit))+1,
	)

	return minCanvas, maxCanvas
}

func (d MarchingCanvas) getSection(attribute string, dataType MarchingDataType) *marchingSection {
	if section, ok := d.sections[attribute]; ok {
		if section.dataType != dataType {
			panic(fmt.Errorf("field already exists with type: %d, can't add type %d", section.dataType, dataType))
		}
		return section
	}

	d.sections[attribute] = &marchingSection{
		dataType:  dataType,
		positions: make(map[vector3.Int]int),
	}

	return d.sections[attribute]
}

func (d *MarchingCanvas) addFloat1Range(section *marchingSection, chunkPos, min, max vector3.Int, function sample.Vec3ToFloat) {
	if section.dataType != Float1 {
		panic(fmt.Errorf("cant add float1 to section with type of: %d", section.dataType))
	}

	index := d.chunkIndex_atomic(section, chunkPos)
	data := d.float1Data[index]

	for z := min.Z(); z < max.Z(); z++ {
		for y := min.Y(); y < max.Y(); y++ {
			for x := min.X(); x < max.X(); x++ {
				pos := vector3.
					New(float64(x), float64(y), float64(z)).
					DivByConstant(d.cubesPerUnit)

				shiftedPos := vector3.New(
					x-(chunkPos.X()*marchingSectionSize),
					y-(chunkPos.Y()*marchingSectionSize),
					z-(chunkPos.Z()*marchingSectionSize),
				)

				data[d.index(shiftedPos.X(), shiftedPos.Y(), shiftedPos.Z())] += function(pos)
			}
		}
	}
}

func (d *MarchingCanvas) calcFloat1Range(min, max vector3.Int, function sample.Vec3ToFloat) []float64 {
	bounds := max.Sub(min)
	arr := make([]float64, bounds.X()*bounds.Y()*bounds.Z())

	i := 0

	for z := min.Z(); z < max.Z(); z++ {
		zF := float64(z) / d.cubesPerUnit
		for y := min.Y(); y < max.Y(); y++ {
			yF := float64(y) / d.cubesPerUnit
			for x := min.X(); x < max.X(); x++ {
				xF := float64(x) / d.cubesPerUnit
				arr[i] = function(vector3.New(zF, yF, xF))
				i++
			}
		}
	}
	return arr

}

func (d MarchingCanvas) chunkSectionsInRange(min, max vector3.Int) []vector3.Int {
	minChunkPos := d.canvasPosToChunkPos(min.X(), min.Y(), min.Z())
	maxChunkPos := d.canvasPosToChunkPos(max.X(), max.Y(), max.Z())

	if minChunkPos == maxChunkPos {
		return []vector3.Int{minChunkPos}
	}

	chunkRange := maxChunkPos.Sub(minChunkPos)

	allSections := make([]vector3.Int, 0)
	for x := 0; x < chunkRange.X()+1; x++ {
		for y := 0; y < chunkRange.Y()+1; y++ {
			for z := 0; z < chunkRange.Z()+1; z++ {
				allSections = append(allSections, vector3.New(
					minChunkPos.X()+x,
					minChunkPos.Y()+y,
					minChunkPos.Z()+z,
				))
			}
		}
	}
	return allSections
}

func maxInt(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (d *MarchingCanvas) AddField(field Field) {
	min, max := d.fieldBounds(field)
	chunkSections := d.chunkSectionsInRange(min, max)

	for attribute, function := range field.Float1Functions {
		section := d.getSection(attribute, Float1)

		for _, chunkPos := range chunkSections {
			canvasSpaceChunkPos := vector3.New(
				maxInt(chunkPos.X()*marchingSectionSize, min.X()),
				maxInt(chunkPos.Y()*marchingSectionSize, min.Y()),
				maxInt(chunkPos.Z()*marchingSectionSize, min.Z()),
			)
			endPos := vector3.New(
				minInt((chunkPos.X()*marchingSectionSize)+marchingSectionSize, max.X()),
				minInt((chunkPos.Y()*marchingSectionSize)+marchingSectionSize, max.Y()),
				minInt((chunkPos.Z()*marchingSectionSize)+marchingSectionSize, max.Z()),
			)
			d.addFloat1Range(section, chunkPos, canvasSpaceChunkPos, endPos, function)
		}

	}
}

func (d *MarchingCanvas) AddFieldParallel(field Field) {
	workers := runtime.NumCPU()
	if workers == 1 {
		d.AddField(field)
		return
	}

	type job struct {
		section                    *marchingSection
		chunkPos, startPos, endPos vector3.Int
		function                   sample.Vec3ToFloat
	}

	min, max := d.fieldBounds(field)
	chunkSections := d.chunkSectionsInRange(min, max)

	numJobs := len(chunkSections)
	jobs := make(chan job, numJobs)
	results := make(chan int, workers)

	for w := 0; w < workers; w++ {
		go func(jobs <-chan job, results chan<- int) {
			completed := 0
			for j := range jobs {
				d.addFloat1Range(j.section, j.chunkPos, j.startPos, j.endPos, j.function)
				completed++
			}
			results <- completed
		}(jobs, results)
	}

	for attribute, function := range field.Float1Functions {
		section := d.getSection(attribute, Float1)
		for _, chunkPos := range chunkSections {
			canvasSpaceChunkPos := vector3.New(
				maxInt(chunkPos.X()*marchingSectionSize, min.X()),
				maxInt(chunkPos.Y()*marchingSectionSize, min.Y()),
				maxInt(chunkPos.Z()*marchingSectionSize, min.Z()),
			)
			endPos := vector3.New(
				minInt((chunkPos.X()*marchingSectionSize)+marchingSectionSize, max.X()),
				minInt((chunkPos.Y()*marchingSectionSize)+marchingSectionSize, max.Y()),
				minInt((chunkPos.Z()*marchingSectionSize)+marchingSectionSize, max.Z()),
			)
			jobs <- job{
				section:  section,
				chunkPos: chunkPos,
				startPos: canvasSpaceChunkPos,
				endPos:   endPos,
				function: function,
			}
		}
	}

	close(jobs)

	for i := 0; i < workers; i++ {
		<-results
	}
}

func (d *MarchingCanvas) AddFieldParallel2(field Field) {
	type job struct {
		section                    *marchingSection
		chunkPos, startPos, endPos vector3.Int
		function                   sample.Vec3ToFloat
		data                       []float64
	}

	min, max := d.fieldBounds(field)
	chunkSections := d.chunkSectionsInRange(min, max)

	workers := runtime.NumCPU()
	numJobs := len(chunkSections)
	jobs := make(chan *job, numJobs)
	results := make(chan *job, numJobs)

	for w := 0; w < workers; w++ {
		go func(jobs <-chan *job, results chan<- *job) {
			for j := range jobs {
				j.data = d.calcFloat1Range(j.startPos, j.endPos, j.function)
				results <- j
			}
		}(jobs, results)
	}

	for attribute, function := range field.Float1Functions {
		section := d.getSection(attribute, Float1)
		for _, chunkPos := range chunkSections {
			canvasSpaceChunkPos := vector3.New(
				maxInt(chunkPos.X()*marchingSectionSize, min.X()),
				maxInt(chunkPos.Y()*marchingSectionSize, min.Y()),
				maxInt(chunkPos.Z()*marchingSectionSize, min.Z()),
			)
			endPos := vector3.New(
				minInt((chunkPos.X()*marchingSectionSize)+marchingSectionSize, max.X()),
				minInt((chunkPos.Y()*marchingSectionSize)+marchingSectionSize, max.Y()),
				minInt((chunkPos.Z()*marchingSectionSize)+marchingSectionSize, max.Z()),
			)
			jobs <- &job{
				section:  section,
				chunkPos: chunkPos,
				startPos: canvasSpaceChunkPos,
				endPos:   endPos,
				function: function,
			}

		}
	}

	close(jobs)

	for j := 0; j < numJobs; j++ {

		result := <-results
		i := 0
		chunkPos := result.chunkPos
		data := d.float1Data[d.chunkIndex_atomic(result.section, chunkPos)]
		resultData := result.data
		for z := result.startPos.Z(); z < result.endPos.Z(); z++ {
			for y := result.startPos.Y(); y < result.endPos.Y(); y++ {
				for x := result.startPos.X(); x < result.endPos.X(); x++ {

					shiftedPos := vector3.New(
						x-(chunkPos.X()*marchingSectionSize),
						y-(chunkPos.Y()*marchingSectionSize),
						z-(chunkPos.Z()*marchingSectionSize),
					)

					data[d.index(shiftedPos.X(), shiftedPos.Y(), shiftedPos.Z())] += resultData[i]
					i++
				}
			}
		}
	}
}

type workingData struct {
	tris       []int
	verts      []vector3.Float64
	vertLookup map[vector3.Int32]int
}

func (d *MarchingCanvas) marchFloat1BlockPosition(
	cutoff float64,
	meshAttribute string,
	section *marchingSection,
	blockPosition vector3.Int,
) modeling.Mesh {

	cubeDataIndexIncrements := []vector3.Int{
		vector3.New(0, 0, 0),
		vector3.New(1, 0, 0),
		vector3.New(1, 0, 1),
		vector3.New(0, 0, 1),
		vector3.New(0, 1, 0),
		vector3.New(1, 1, 0),
		vector3.New(1, 1, 1),
		vector3.New(0, 1, 1),
	}

	cubeData := make([]float1MarchingSection, 8)
	cubeDataIndexes := make([]int, 8)
	cubeCorners := make([]float64, 8)
	cubeCornersExistence := make([]bool, 8)

	// var cubeData [8]float1MarchingSection
	// var cubeDataIndexes [8]int
	// var cubeCorners [8]float64
	// var cubeCornersExistence [8]bool

	marchingWorkingData := &workingData{
		tris:       make([]int, 0),
		verts:      make([]vector3.Float64, 0),
		vertLookup: make(map[vector3.Int32]int),
	}
	blockIndex := section.positions[blockPosition]

	data := d.float1Data[blockIndex]
	offset := vector3.New(
		float64(blockPosition.X())*marchingSectionSize,
		float64(blockPosition.Y())*marchingSectionSize,
		float64(blockPosition.Z())*marchingSectionSize,
	)

	for z := 0; z < marchingSectionSize; z++ {

		zBlockPosition := blockPosition.Z()
		if z == marchingSectionSize-1 {
			zBlockPosition += 1
			nextZ := vector3.New(
				blockPosition.X(),
				blockPosition.Y(),
				zBlockPosition,
			)
			if _, ok := section.positions[nextZ]; !ok {
				continue
			}
		}

		for y := 0; y < marchingSectionSize; y++ {
			yBlockPosition := blockPosition.Y()
			if y == marchingSectionSize-1 {
				yBlockPosition += 1
				nextY := vector3.New(
					blockPosition.X(),
					yBlockPosition,
					zBlockPosition,
				)
				if _, ok := section.positions[nextY]; !ok {
					continue
				}
			}

			for x := 0; x < marchingSectionSize; x++ {
				xBlockPosition := blockPosition.X()
				if x == marchingSectionSize-1 {
					xBlockPosition += 1
				}

				cubeDataBlockPositions := []vector3.Int{
					blockPosition,
					vector3.New(xBlockPosition, blockPosition.Y(), blockPosition.Z()),
					vector3.New(xBlockPosition, blockPosition.Y(), zBlockPosition),
					vector3.New(blockPosition.X(), blockPosition.Y(), zBlockPosition),
					vector3.New(blockPosition.X(), yBlockPosition, blockPosition.Z()),
					vector3.New(xBlockPosition, yBlockPosition, blockPosition.Z()),
					vector3.New(xBlockPosition, yBlockPosition, zBlockPosition),
					vector3.New(blockPosition.X(), yBlockPosition, zBlockPosition),
				}

				cubeData[0] = data
				cubeData[1] = data
				cubeData[2] = data
				cubeData[3] = data
				cubeData[4] = data
				cubeData[5] = data
				cubeData[6] = data
				cubeData[7] = data

				cubeDataIndexes[0] = d.index(x, y, z)
				cubeDataIndexes[1] = d.index(x+1, y, z)
				cubeDataIndexes[2] = d.index(x+1, y, z+1)
				cubeDataIndexes[3] = d.index(x, y, z+1)
				cubeDataIndexes[4] = d.index(x, y+1, z)
				cubeDataIndexes[5] = d.index(x+1, y+1, z)
				cubeDataIndexes[6] = d.index(x+1, y+1, z+1)
				cubeDataIndexes[7] = d.index(x, y+1, z+1)

				allValid := true
				for i, pos := range cubeDataBlockPositions {
					if dataIndex, ok := section.positions[pos]; ok {
						cubeData[i] = d.float1Data[dataIndex]

						newIndex := vector3.New(
							x+cubeDataIndexIncrements[i].X(),
							y+cubeDataIndexIncrements[i].Y(),
							z+cubeDataIndexIncrements[i].Z(),
						)

						if pos.X() != blockPosition.X() {
							newIndex = newIndex.SetX(0)
						}
						if pos.Y() != blockPosition.Y() {
							newIndex = newIndex.SetY(0)
						}
						if pos.Z() != blockPosition.Z() {
							newIndex = newIndex.SetZ(0)
						}

						cubeDataIndexes[i] = d.index(newIndex.X(), newIndex.Y(), newIndex.Z())
					} else {
						allValid = false
						break
					}
				}
				if !allValid {
					continue
				}

				cubeCorners[0] = cubeData[0][cubeDataIndexes[0]]
				cubeCorners[1] = cubeData[1][cubeDataIndexes[1]]
				cubeCorners[2] = cubeData[2][cubeDataIndexes[2]]
				cubeCorners[3] = cubeData[3][cubeDataIndexes[3]]
				cubeCorners[4] = cubeData[4][cubeDataIndexes[4]]
				cubeCorners[5] = cubeData[5][cubeDataIndexes[5]]
				cubeCorners[6] = cubeData[6][cubeDataIndexes[6]]
				cubeCorners[7] = cubeData[7][cubeDataIndexes[7]]

				cubeCornersExistence[0] = cubeCorners[0] < cutoff
				cubeCornersExistence[1] = cubeCorners[1] < cutoff
				cubeCornersExistence[2] = cubeCorners[2] < cutoff
				cubeCornersExistence[3] = cubeCorners[3] < cutoff
				cubeCornersExistence[4] = cubeCorners[4] < cutoff
				cubeCornersExistence[5] = cubeCorners[5] < cutoff
				cubeCornersExistence[6] = cubeCorners[6] < cutoff
				cubeCornersExistence[7] = cubeCorners[7] < cutoff

				xf := float64(x)
				yf := float64(y)
				zf := float64(z)

				cubeCornerPositions := []vector3.Float64{
					vector3.New(xf, yf, zf),
					vector3.New(xf+1, yf, zf),
					vector3.New(xf+1, yf, zf+1),
					vector3.New(xf, yf, zf+1),
					vector3.New(xf, yf+1, zf),
					vector3.New(xf+1, yf+1, zf),
					vector3.New(xf+1, yf+1, zf+1),
					vector3.New(xf, yf+1, zf+1),
				}

				lookupIndex := 0
				if cubeCornersExistence[0] {
					lookupIndex |= 1
				}
				if cubeCornersExistence[1] {
					lookupIndex |= 2
				}
				if cubeCornersExistence[2] {
					lookupIndex |= 4
				}
				if cubeCornersExistence[3] {
					lookupIndex |= 8
				}
				if cubeCornersExistence[4] {
					lookupIndex |= 16
				}
				if cubeCornersExistence[5] {
					lookupIndex |= 32
				}
				if cubeCornersExistence[6] {
					lookupIndex |= 64
				}
				if cubeCornersExistence[7] {
					lookupIndex |= 128
				}

				tris := triangulation[lookupIndex]
				for i := 0; tris[i] != -1; i += 3 {
					// Get indices of corner points A and B for each of the three edges
					// of the cube that need to be joined to form the triangle.
					a0 := cornerIndexAFromEdge[tris[i]]
					b0 := cornerIndexBFromEdge[tris[i]]

					a1 := cornerIndexAFromEdge[tris[i+1]]
					b1 := cornerIndexBFromEdge[tris[i+1]]

					a2 := cornerIndexAFromEdge[tris[i+2]]
					b2 := cornerIndexBFromEdge[tris[i+2]]

					v1 := interpolateVerts(cubeCornerPositions[a0], cubeCornerPositions[b0], cubeCorners[a0], cubeCorners[b0], cutoff).Add(offset)
					v2 := interpolateVerts(cubeCornerPositions[a1], cubeCornerPositions[b1], cubeCorners[a1], cubeCorners[b1], cutoff).Add(offset)
					v3 := interpolateVerts(cubeCornerPositions[a2], cubeCornerPositions[b2], cubeCorners[a2], cubeCorners[b2], cutoff).Add(offset)

					marchingWorkingData.tris = append(
						marchingWorkingData.tris,
						lookupOrAdd(marchingWorkingData, v1),
						lookupOrAdd(marchingWorkingData, v2),
						lookupOrAdd(marchingWorkingData, v3),
					)
				}
			}
		}
	}

	return modeling.NewTriangleMesh(marchingWorkingData.tris).
		SetFloat3Data(map[string][]vector3.Float64{
			meshAttribute: marchingWorkingData.verts,
		})
}

func (d MarchingCanvas) marchFloat1(cutoff float64, meshAttribute string, section *marchingSection) modeling.Mesh {
	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)
	for blockPosition := range section.positions {
		finalMesh = finalMesh.Append(d.marchFloat1BlockPosition(cutoff, meshAttribute, section, blockPosition))
	}
	return finalMesh
}

func (d MarchingCanvas) marchFloat1Parallel(cutoff float64, meshAttribute string, section *marchingSection) modeling.Mesh {
	workers := runtime.NumCPU()

	if workers == 1 {
		return d.marchFloat1(cutoff, meshAttribute, section)
	}

	numJobs := len(section.positions)
	jobs := make(chan vector3.Int, numJobs)
	results := make(chan modeling.Mesh, numJobs)

	for w := 0; w < workers; w++ {
		go func(jobs <-chan vector3.Int, results chan<- modeling.Mesh) {
			for j := range jobs {
				results <- d.marchFloat1BlockPosition(cutoff, meshAttribute, section, j)
			}
		}(jobs, results)
	}

	for blockPosition := range section.positions {
		jobs <- blockPosition
	}
	close(jobs)

	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)
	for i := 0; i < numJobs; i++ {
		finalMesh = finalMesh.Append(<-results)
	}

	return finalMesh
}

// March is shorthand for MarchOnAttribute(modeling.PositionAttribute, cutoff)
func (d MarchingCanvas) March(cutoff float64) modeling.Mesh {
	return d.MarchOnAttribute(modeling.PositionAttribute, cutoff)
}

func (d MarchingCanvas) MarchOnAttribute(attribute string, cutoff float64) modeling.Mesh {
	for sectionAttribute, section := range d.sections {
		if section.dataType == Float1 && sectionAttribute == attribute {
			return d.marchFloat1(cutoff, sectionAttribute, section).
				Transform(
					meshops.ScaleAttribute3DTransformer{
						Amount: vector3.One[float64]().DivByConstant(d.cubesPerUnit),
					},
				).
				WeldByFloat3Attribute(attribute, 3)
		}
	}
	panic(fmt.Errorf("canvas did not contain Float1 attribute %s", attribute))
}

func (d MarchingCanvas) MarchParallel(cutoff float64) modeling.Mesh {
	return d.MarchOnAttributeParallel(modeling.PositionAttribute, cutoff)
}

func (d MarchingCanvas) MarchOnAttributeParallel(attribute string, cutoff float64) modeling.Mesh {
	for sectionAttribute, section := range d.sections {
		if section.dataType == Float1 && sectionAttribute == attribute {
			marched := d.marchFloat1Parallel(cutoff, sectionAttribute, section)
			if marched.PrimitiveCount() == 0 {
				return marched
			}
			return marched.
				Transform(
					meshops.ScaleAttribute3DTransformer{
						Amount: vector3.One[float64]().DivByConstant(d.cubesPerUnit),
					},
				).
				WeldByFloat3Attribute(attribute, 3)
		}
	}
	panic(fmt.Errorf("canvas did not contain Float1 attribute %s", attribute))
}
