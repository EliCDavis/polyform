package potree

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"github.com/EliCDavis/polyform/math/geometry"
)

type octreeNodeHeader struct {
	Type       uint8
	ChildMask  uint8
	NumPoints  uint32
	ByteOffset uint64
	ByteSize   uint64
}

func createChildAABB(aabb geometry.AABB, index int) geometry.AABB {
	min := aabb.Min()
	max := aabb.Max()
	size := max.Sub(min).Scale(0.5)

	if (index & 0b0001) > 0 {
		min = min.SetZ(min.Z() + size.Z())
	} else {
		max = max.SetZ(max.Z() - size.Z())
	}

	if (index & 0b0010) > 0 {
		min = min.SetY(min.Y() + size.Y())
	} else {
		max = max.SetY(max.Y() - size.Y())
	}

	if (index & 0b0100) > 0 {
		min = min.SetX(min.X() + size.X())
	} else {
		max = max.SetX(max.X() - size.X())
	}

	return geometry.NewAABBFromPoints(min, max)
}

// All code derived from here
// https://github.com/potree/potree/blob/785e7a95aac3c44112c5488ee16c970fb8eeb923/src/modules/loader/2.0/OctreeLoader.js#L151

func ParseHierarchy(node *OctreeNode, buf []byte) error {
	const bytesPerNode = 22

	veiw := bytes.NewReader(buf)
	numNodes := len(buf) / bytesPerNode

	nodes := make([]*OctreeNode, numNodes)
	nodes[0] = node
	nodePos := 1
	for _, current := range nodes {
		header := octreeNodeHeader{}
		err := binary.Read(veiw, binary.LittleEndian, &header)
		if err != nil {
			return err
		}

		if current.NodeType == 2 {
			// replace proxy with real node
			current.ByteOffset = header.ByteOffset
			current.ByteSize = header.ByteSize
			current.NumPoints = header.NumPoints
		} else if header.Type == 2 {
			// load proxy
			current.HierarchyByteOffset = header.ByteOffset
			current.HierarchyByteSize = header.ByteSize
			current.NumPoints = header.NumPoints
		} else {
			// load real node
			current.ByteOffset = header.ByteOffset
			current.ByteSize = header.ByteSize
			current.NumPoints = header.NumPoints
		}

		if current.ByteSize == 0 {
			// workaround for issue #1125
			// some inner nodes erroneously report >0 points even though have 0 points
			// however, they still report a ByteSize of 0, so based on that we now set node.NumPoints to 0
			current.NumPoints = 0
		}

		current.NodeType = header.Type

		if current.NodeType == 2 {
			continue
		}

		for childIndex := 0; childIndex < 8; childIndex++ {
			childExists := ((1 << childIndex) & header.ChildMask) != 0

			if !childExists {
				continue
			}

			child := &OctreeNode{
				Name:        current.Name + strconv.Itoa(childIndex),
				BoundingBox: createChildAABB(current.BoundingBox, childIndex),
				Spacing:     current.Spacing / 2,
				Level:       current.Level + 1,
				Parent:      current,
			}

			current.Children = append(current.Children, child)
			nodes[nodePos] = child
			nodePos++
		}
	}

	return nil
}

func ParseEntireHierarchy(root *OctreeNode, buf []byte) error {
	if root.NodeType == 2 {
		start := root.HierarchyByteOffset
		scoped := buf[start : start+root.HierarchyByteSize]
		if err := ParseHierarchy(root, scoped); err != nil {
			return err
		}
	}

	for _, c := range root.Children {
		if err := ParseEntireHierarchy(c, buf); err != nil {
			return err
		}
	}

	return nil
}
