package animation

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/vector/vector3"
)

type skeletonJoint struct {
	path             string
	weight           float64
	worldPosition    vector3.Float64
	relativePosition vector3.Float64
	up, forward      vector3.Float64
	relativeMatrix   mat.Matrix4x4
	children         []int
}

type Skeleton struct {
	joints   []skeletonJoint
	jointLUT map[string]int // mapping of path of joints to index in array
}

func (s Skeleton) JointCount() int {
	return len(s.joints)
}

func (s Skeleton) Lookup(name string) int {
	if index, ok := s.jointLUT[name]; ok {
		return index
	}
	panic(fmt.Errorf("skeleton did not contain a joint with the path: %s", name))
}

func (s Skeleton) Children(index int) []int {
	return s.joints[index].children
}

func (s Skeleton) ClosestJoints(point vector3.Float64, pointToConsider int) []int {
	queue := make(minJointValPriorityQueue, 0)

	for i, n := range s.joints {
		dist := n.worldPosition.DistanceSquared(point)
		// if queue.Len() < maxPointsToConsider {
		heap.Push(&queue, jointValItem{
			val:   dist,
			joint: i,
		})
		// }
	}

	size := pointToConsider
	if queue.Len() < pointToConsider {
		size = queue.Len()
	}

	joints := make([]int, size)
	for i := 0; i < size; i++ {
		item := heap.Pop(&queue).(jointValItem)
		joints[i] = item.joint
	}
	return joints
}

func (s Skeleton) WorldPosition(index int) vector3.Float64 {
	return s.joints[index].worldPosition
}

func (s Skeleton) RelativeMatrix(index int) mat.Matrix4x4 {
	// j := s.joints[index]
	// return mat.MatFromDirs(j.up, j.forward, j.relativePosition)
	return s.joints[index].relativeMatrix
}

func (s Skeleton) RelativePosition(index int) vector3.Float64 {
	return s.joints[index].relativePosition
}

func (s Skeleton) Heat(index int) float64 {
	return s.joints[index].weight
}

func (s Skeleton) InverseBindMatrix(index int) mat.Matrix4x4 {
	j := s.joints[index]
	return mat.MatFromDirs(j.up, j.forward, j.worldPosition).Inverse()
}

func flattenJoints(index int, curPath string, root Joint, parent *Joint) []skeletonJoint {
	if root.name == "" {
		panic("joint name can not be empty")
	}

	if strings.Contains(root.name, "/") {
		panic(fmt.Errorf("joint name '%s' can not contain the character '/'", root.name))
	}

	combinedName := root.name
	if curPath != "" {
		combinedName = fmt.Sprintf("%s/%s", curPath, root.name)
	}

	parentMat := mat.Identity()
	parentPos := vector3.Zero[float64]()
	if parent != nil {
		parentMat = parent.Matrix()
		parentPos = parent.worldPosition
	}

	flattened := make([]skeletonJoint, 1)

	flattened[0] = skeletonJoint{
		path: combinedName,
		// relativePosition: root.worldPosition.Sub(parentPos),
		relativeMatrix: parentMat.
			Inverse().
			Multiply(root.Matrix()),
		worldPosition:    root.worldPosition,
		relativePosition: root.worldPosition.Sub(parentPos),
		up:               root.up,
		forward:          root.forward,
		children:         make([]int, 0),
		weight:           root.weight,
	}

	offset := index + 1
	for _, child := range root.children {
		flattened[0].children = append(flattened[0].children, offset)
		nodes := flattenJoints(offset, combinedName, child, &root)
		offset += len(nodes)
		flattened = append(flattened, nodes...)
	}

	return flattened
}

func NewSkeleton(root Joint) Skeleton {
	nodes := flattenJoints(0, "", root, nil)

	lut := make(map[string]int)
	for i, n := range nodes {
		if _, ok := lut[n.path]; ok {
			panic(fmt.Errorf("skeleton requires unique names for joints that share the same parent, found duplicate %s", n.path))
		}
		lut[n.path] = i
	}

	return Skeleton{
		joints:   nodes,
		jointLUT: lut,
	}
}
