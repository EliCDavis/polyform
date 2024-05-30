package potree

import (
	"encoding/binary"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type OctreeNode struct {
	Name        string
	Spacing     float64
	Level       int
	BoundingBox geometry.AABB
	Parent      *OctreeNode `json:"-"`
	Children    []*OctreeNode

	NodeType   uint8
	ChildMask  uint8
	NumPoints  uint32
	ByteOffset uint64
	ByteSize   uint64

	HierarchyByteOffset uint64
	HierarchyByteSize   uint64
}

func (on *OctreeNode) Walk(f func(o *OctreeNode)) {
	f(on)
	for _, c := range on.Children {
		c.Walk(f)
	}
}

func (on OctreeNode) Height() int {
	if len(on.Children) == 0 {
		return 0
	}

	height := 0
	for _, c := range on.Children {
		height = max(height, c.Height())
	}
	return height + 1
}

func (on OctreeNode) PointCount() uint64 {
	count := uint64(on.NumPoints)
	for _, c := range on.Children {
		count += c.PointCount()
	}
	return count
}

func (on OctreeNode) DescendentCount() int {
	if len(on.Children) == 0 {
		return 0
	}

	count := len(on.Children)
	for _, c := range on.Children {
		count += c.DescendentCount()
	}
	return count
}

func LoadNodePositionDataIntoArray(m *Metadata, buf []byte, positions []vector3.Float64) {
	attributeOffset := 0

	for _, attribute := range m.Attributes {
		if attribute.IsPosition() {

			endian := binary.LittleEndian
			bytesPerPoint := m.BytesPerPoint()
			scale := vector3.New(m.Scale[0], m.Scale[1], m.Scale[2])
			shift := vector3.New(m.Offset[0], m.Offset[1], m.Offset[2]).
				Sub(m.BoundingBox.MinF())

			pointOffset := attributeOffset
			for i := 0; i < len(positions); i++ {
				positions[i] = vector3.
					New(
						int(endian.Uint32(buf[pointOffset:])),
						int(endian.Uint32(buf[pointOffset+4:])),
						int(endian.Uint32(buf[pointOffset+8:])),
					).
					ToFloat64().
					MultByVector(scale).
					Add(shift)
				pointOffset += bytesPerPoint
			}
			return
		}
		attributeOffset += attribute.Size
	}
}

func LoadNodeColorDataIntoArray(m *Metadata, buf []byte, colors []vector3.Float64) {
	attributeOffset := 0

	for _, attribute := range m.Attributes {
		if attribute.IsColor() {
			endian := binary.LittleEndian
			bytesPerPoint := m.BytesPerPoint()

			pointOffset := attributeOffset
			for i := 0; i < len(colors); i++ {
				col := vector3.
					New(
						int(endian.Uint16(buf[pointOffset:])),
						int(endian.Uint16(buf[pointOffset+2:])),
						int(endian.Uint16(buf[pointOffset+4:])),
					).
					ToFloat64()

				if col.X() > 255 || col.Y() > 255 || col.Z() > 255 {
					col = col.DivByConstant(256)
				}

				colors[i] = col.DivByConstant(255)
				pointOffset += bytesPerPoint
			}
			return
		}
		attributeOffset += attribute.Size
	}
}

// https://github.com/potree/potree/blob/785e7a95aac3c44112c5488ee16c970fb8eeb923/src/modules/loader/2.0/DecoderWorker.js#L21

func LoadNode(on *OctreeNode, m *Metadata, buf []byte) modeling.Mesh {

	numPoints := on.NumPoints
	bytesPerPoint := m.BytesPerPoint()

	attributeOffset := 0
	endian := binary.LittleEndian

	pointcloudData := make(map[string][]vector3.Float64)

	for _, attribute := range m.Attributes {
		if attribute.IsPosition() {
			positionData := make([]vector3.Float64, numPoints)

			scale := vector3.New(m.Scale[0], m.Scale[1], m.Scale[2])
			shift := vector3.New(m.Offset[0], m.Offset[1], m.Offset[2]).
				Sub(m.BoundingBox.MinF())

			for i := 0; i < int(numPoints); i++ {
				pointOffset := (i * bytesPerPoint) + attributeOffset
				positionData[i] = vector3.
					New(
						int(endian.Uint32(buf[pointOffset:])),
						int(endian.Uint32(buf[pointOffset+4:])),
						int(endian.Uint32(buf[pointOffset+8:])),
					).
					ToFloat64().
					MultByVector(scale).
					Add(shift)
			}

			pointcloudData[modeling.PositionAttribute] = positionData
		}

		if attribute.IsColor() {
			colorData := make([]vector3.Float64, numPoints)

			for i := 0; i < int(numPoints); i++ {
				pointOffset := (i * bytesPerPoint) + attributeOffset
				col := vector3.
					New(
						int(endian.Uint16(buf[pointOffset:])),
						int(endian.Uint16(buf[pointOffset+2:])),
						int(endian.Uint16(buf[pointOffset+4:])),
					).
					ToFloat64()

				if col.X() > 255 || col.Y() > 255 || col.Z() > 255 {
					col = col.DivByConstant(256)
				}

				colorData[i] = col.DivByConstant(255)
			}

			pointcloudData[modeling.ColorAttribute] = colorData
		}

		attributeOffset += attribute.Size
	}

	return modeling.NewPointCloud(nil, pointcloudData, nil, nil, nil)
}
