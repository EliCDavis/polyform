package marching

import (
	"log"
	"math"
	"time"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

// Queue represents a single instance of the queue data structure.
type Queue[V any] struct {
	buf               []*V
	head, tail, count int
}

// New constructs and returns a new Queue.
func New[V any]() *Queue[V] {
	return &Queue[V]{
		buf: make([]*V, minQueueLen),
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *Queue[V]) Length() int {
	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[V]) resize() {
	newBuf := make([]*V, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Add puts an element on the end of the queue.
func (q *Queue[V]) Add(elem V) {
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = &elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *Queue[V]) Peek() V {
	if q.count <= 0 {
		panic("queue: Peek() called on empty queue")
	}
	return *(q.buf[q.head])
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *Queue[V]) Get(i int) V {
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		panic("queue: Get() called with index out of range")
	}
	// bitwise modulus
	return *(q.buf[(q.head+i)&(len(q.buf)-1)])
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue[V]) Remove() V {
	if q.count <= 0 {
		panic("queue: Remove() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return *ret
}

func toInt32(v vector3.Int) vector3.Int32 {
	return vector3.New(int32(v.X()), int32(v.Y()), int32(v.Z()))
}

type AabbInt struct {
	Min vector3.Int32
	Max vector3.Int32
}

func (a AabbInt) Center() vector3.Int32 {
	return a.Max.Add(a.Min).Scale(0.5)
}

func (a AabbInt) Size() vector3.Int32 {
	return a.Max.Sub(a.Min)
}

func splitAABB(b AabbInt, results []AabbInt) {
	min := b.Min
	max := b.Max
	size := b.Size()

	// center ‘split plane’
	cx := min.X() + size.X()/2
	cy := min.Y() + size.Y()/2
	cz := min.Z() + size.Z()/2

	// z–
	results[0] = AabbInt{Min: min, Max: vector3.New(cx, cy, cz)}
	results[1] = AabbInt{Min: vector3.New(cx+1, min.Y(), min.Z()), Max: vector3.New(max.X(), cy, cz)}
	results[2] = AabbInt{Min: vector3.New(min.X(), cy+1, min.Z()), Max: vector3.New(cx, max.Y(), cz)}
	results[3] = AabbInt{Min: vector3.New(cx+1, cy+1, min.Z()), Max: vector3.New(max.X(), max.Y(), cz)}

	// z+
	results[4] = AabbInt{Min: vector3.New(min.X(), min.Y(), cz+1), Max: vector3.New(cx, cy, max.Z())}
	results[5] = AabbInt{Min: vector3.New(cx+1, min.Y(), cz+1), Max: vector3.New(max.X(), cy, max.Z())}
	results[6] = AabbInt{Min: vector3.New(min.X(), cy+1, cz+1), Max: vector3.New(cx, max.Y(), max.Z())}
	results[7] = AabbInt{Min: vector3.New(cx+1, cy+1, cz+1), Max: max}
}

func marchCracked(field sample.Vec3ToFloat, initialBounds geometry.AABB, cubeSize, surface float64, res map[vector3.Int32]float64) {
	splitResults := make([]AabbInt, 8)

	stack := New[AabbInt]()
	stack.Add(AabbInt{
		Min: toInt32(initialBounds.Min().DivByConstant(cubeSize).RoundToInt()),
		Max: toInt32(initialBounds.Max().DivByConstant(cubeSize).RoundToInt()),
	})

	cachedResult := 0
	for stack.Length() > 0 {
		bounds := stack.Remove()
		center := bounds.Center()

		fieldResult, ok := res[center]
		if !ok {
			recentered := center.ToFloat64().Scale(cubeSize)
			fieldResult = field(recentered) - surface
		} else {
			cachedResult++
		}

		// realSize := bounds.Size().ToFloat64().Scale(cubeSize)
		// diagonal := realSize.Length()

		// // TODO: WE THIS IS OUR BIGGEST SPEEDUP, FIGURE OUT HOW TO PRUNE HARDER
		// // The closest surface is not within the bounds
		// if math.Abs(fieldResult) > (diagonal/2)+(cubeSize*2) {
		// 	continue
		// }

		if !ok {
			res[center] = fieldResult
		}

		size := bounds.Size()
		if size.X() <= 1 && size.Y() <= 1 && size.Z() <= 1 {
			continue
		}

		splitAABB(bounds, splitResults)
		for _, child := range splitResults {
			// if child.Min == child.Max && child.Min == center {
			// 	continue
			// }

			s := child.Size()
			if s.X() <= 0 || s.Y() <= 0 || s.Z() <= 0 {
				continue
			}
			stack.Add(child)
		}

		// halfSize := toInt32(size.ToFloat64().Scale(0.5).FloorToInt())
		// otherHalf := size.Sub(halfSize)
		// if otherHalf.X() == halfSize.X() {
		// 	otherHalf = otherHalf.AddX(1)
		// }

		// if otherHalf.Y() == halfSize.Y() {
		// 	otherHalf = otherHalf.AddY(1)
		// }
		// newMin := bounds.Min

		// stack.Add(AabbInt{Min: newMin, Max: center})
		// stack.Add(AabbInt{Min: newMin.AddX(otherHalf.X()), Max: center.SetX(bounds.Max.X())})
		// stack.Add(AabbInt{Min: newMin.AddY(otherHalf.Y()), Max: center.SetY(bounds.Max.Y())})

		// stack.Add(AabbInt{
		// 	Min: newMin.AddY(otherHalf.Y()).AddX(otherHalf.X()),
		// 	Max: center.SetY(bounds.Max.Y()).SetX(bounds.Max.X()),
		// })

		// newMin = newMin.SetZ(center.Z() + 1)
		// center = center.SetZ(bounds.Max.Z())

		// stack.Add(AabbInt{Min: newMin, Max: center})
		// stack.Add(AabbInt{Min: newMin.AddX(otherHalf.X()), Max: center.SetX(bounds.Max.X())})
		// stack.Add(AabbInt{Min: newMin.AddY(otherHalf.Y()), Max: center.SetY(bounds.Max.Y())})

		// stack.Add(AabbInt{
		// 	Min: newMin.AddY(otherHalf.Y()).AddX(otherHalf.X()),
		// 	Max: center.SetY(bounds.Max.Y()).SetX(bounds.Max.X()),
		// })

		// qs := halfSize.Scale(0.5)

		// stack.Add(geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), -qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), -qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), -qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), qs.Z())), halfSize))
		// stack.Add(geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), -qs.Z())), halfSize))

	}

	log.Printf("Cached Results: %d\n", cachedResult)
	log.Printf("Total Results: %d\n", len(res))
}

func marchRecurse(field sample.Vec3ToFloat, bounds geometry.AABB, cubeSize, surface float64, res map[vector3.Int32]float64) {

	center := bounds.Center()
	centerIndex := center.DivByConstant(cubeSize).RoundToInt()
	recentered := centerIndex.ToFloat64().Scale(cubeSize)

	fieldResult := field(recentered) - surface
	size := bounds.Size()
	diagonal := size.Length()

	// TODO: WE THIS IS OUR BIGGEST SPEEDUP, FIGURE OUT HOW TO PRUNE HARDER
	// The closest surface is not within the bounds
	if math.Abs(fieldResult) > (diagonal/2)+(cubeSize)+center.Distance(recentered) {
		return
	}

	res[vector3.New(int32(centerIndex.X()), int32(centerIndex.Y()), int32(centerIndex.Z()))] = fieldResult
	if size.MaxComponent() < cubeSize {
		return
	}

	halfSize := size.Scale(0.5)
	qs := halfSize.Scale(0.5)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
}

func dedup(data *workingData, vert vector3.Float64, size float64) int {
	distritized := modeling.Vector3ToInt(vert, 4)
	key := vector3.New(int32(distritized.X()), int32(distritized.Y()), int32(distritized.Z()))

	if foundIndex, ok := data.vertLookup[key]; ok {
		return foundIndex
	}

	index := len(data.verts)
	data.vertLookup[key] = index
	data.verts = append(data.verts, vert.Scale(size))
	return index
}

func March(field sample.Vec3ToFloat, domain geometry.AABB, cubeSize, surface float64) modeling.Mesh {
	results := make(map[vector3.Int32]float64)
	sdfCompute := time.Now()
	marchRecurse(field, domain, cubeSize, surface, results)
	log.Printf("Time To Compute SDFs %s", time.Since(sdfCompute))

	marchCompute := time.Now()
	marchingWorkingData := &workingData{
		tris:       make([]int, 0),
		verts:      make([]vector3.Float64, 0),
		vertLookup: make(map[vector3.Int32]int),
	}

	cubeCorners := make([]float64, 8)
	cubeCornerPositions := make([]vector3.Float64, 8)
	for key, nnn := range results {
		cubeCorners[0] = nnn

		var ok bool
		cubeCorners[1], ok = results[key.AddX(1)]
		if !ok {
			continue
		}
		cubeCorners[2], ok = results[key.Add(vector3.New(int32(1), 0, 1))]
		if !ok {
			continue
		}
		cubeCorners[3], ok = results[key.AddZ(1)]
		if !ok {
			continue
		}
		cubeCorners[4], ok = results[key.AddY(1)]
		if !ok {
			continue
		}
		cubeCorners[5], ok = results[key.Add(vector3.New(int32(1), 1, 0))]
		if !ok {
			continue
		}
		cubeCorners[6], ok = results[key.Add(vector3.New(int32(1), 1, 1))]
		if !ok {
			continue
		}
		cubeCorners[7], ok = results[key.Add(vector3.New(int32(0), 1, 1))]
		if !ok {
			continue
		}

		lookupIndex := 0
		if cubeCorners[0] < 0 {
			lookupIndex |= 1
		}
		if cubeCorners[1] < 0 {
			lookupIndex |= 2
		}
		if cubeCorners[2] < 0 {
			lookupIndex |= 4
		}
		if cubeCorners[3] < 0 {
			lookupIndex |= 8
		}
		if cubeCorners[4] < 0 {
			lookupIndex |= 16
		}
		if cubeCorners[5] < 0 {
			lookupIndex |= 32
		}
		if cubeCorners[6] < 0 {
			lookupIndex |= 64
		}
		if cubeCorners[7] < 0 {
			lookupIndex |= 128
		}

		if lookupIndex == 0 || lookupIndex == 255 {
			continue
		}

		xf := float64(key.X())
		yf := float64(key.Y())
		zf := float64(key.Z())

		cubeCornerPositions[0] = vector3.New(xf, yf, zf)
		cubeCornerPositions[1] = vector3.New(xf+1, yf, zf)
		cubeCornerPositions[2] = vector3.New(xf+1, yf, zf+1)
		cubeCornerPositions[3] = vector3.New(xf, yf, zf+1)
		cubeCornerPositions[4] = vector3.New(xf, yf+1, zf)
		cubeCornerPositions[5] = vector3.New(xf+1, yf+1, zf)
		cubeCornerPositions[6] = vector3.New(xf+1, yf+1, zf+1)
		cubeCornerPositions[7] = vector3.New(xf, yf+1, zf+1)

		tris := triangulation[lookupIndex]
		for i := 0; i < len(tris); i += 3 {
			// Get indices of corner points A and B for each of the three edges
			// of the cube that need to be joined to form the triangle.
			a0 := cornerIndexAFromEdge[tris[i]]
			b0 := cornerIndexBFromEdge[tris[i]]

			a1 := cornerIndexAFromEdge[tris[i+1]]
			b1 := cornerIndexBFromEdge[tris[i+1]]

			a2 := cornerIndexAFromEdge[tris[i+2]]
			b2 := cornerIndexBFromEdge[tris[i+2]]

			v1 := interpolateVerts(cubeCornerPositions[a0], cubeCornerPositions[b0], cubeCorners[a0], cubeCorners[b0], 0)
			v2 := interpolateVerts(cubeCornerPositions[a1], cubeCornerPositions[b1], cubeCorners[a1], cubeCorners[b1], 0)
			v3 := interpolateVerts(cubeCornerPositions[a2], cubeCornerPositions[b2], cubeCorners[a2], cubeCorners[b2], 0)

			marchingWorkingData.tris = append(
				marchingWorkingData.tris,
				dedup(marchingWorkingData, v1, cubeSize),
				dedup(marchingWorkingData, v2, cubeSize),
				dedup(marchingWorkingData, v3, cubeSize),
			)
		}
	}

	m := modeling.NewMesh(modeling.TriangleTopology, marchingWorkingData.tris).
		SetFloat3Attribute(modeling.PositionAttribute, marchingWorkingData.verts)

	if len(marchingWorkingData.tris) == 0 {
		return m
	}

	log.Printf("Time To March Mesh %s", time.Since(marchCompute))

	return meshops.RemoveNullFaces3D(m, modeling.PositionAttribute, 0)
}
