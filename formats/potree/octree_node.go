package potree

import (
	"encoding/binary"
	"io"

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

func (on *OctreeNode) Walk(f func(o *OctreeNode) bool) {
	if traverseChildren := f(on); !traverseChildren {
		return
	}
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

func (on OctreeNode) MaxPointCount() int {
	count := int(on.NumPoints)
	for _, c := range on.Children {
		count = max(count, c.MaxPointCount())
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

func (on OctreeNode) Read(in io.ReadSeeker, buf []byte) (int, error) {
	_, err := in.Seek(int64(on.ByteOffset), io.SeekStart)
	if err != nil {
		return 0, err
	}

	return io.ReadFull(in, buf[:min(on.ByteSize, uint64(len(buf)))])
}

func (on OctreeNode) Write(out io.WriteSeeker, buf []byte) (int, error) {
	_, err := out.Seek(int64(on.ByteOffset), io.SeekStart)
	if err != nil {
		return 0, err
	}

	return out.Write(buf[:min(on.ByteSize, uint64(len(buf)))])
}

// https://github.com/potree/potree/blob/785e7a95aac3c44112c5488ee16c970fb8eeb923/src/modules/loader/2.0/DecoderWorker.js#L21

func LoadNode(on *OctreeNode, metadata *Metadata, buf []byte) modeling.Mesh {

	numPoints := on.NumPoints
	bytesPerPoint := metadata.BytesPerPoint()

	attributeOffset := 0
	endian := binary.LittleEndian

	pointcloudData := make(map[string][]vector3.Float64)

	for _, attribute := range metadata.Attributes {
		if attribute.IsPosition() {
			positionData := make([]vector3.Float64, numPoints)

			scale := vector3.New(metadata.Scale[0], metadata.Scale[1], metadata.Scale[2])
			shift := metadata.OffsetF().Sub(metadata.BoundingBox.MinF())

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

			for i := range int(numPoints) {
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

	return modeling.NewPointCloud(nil, pointcloudData, nil, nil)
}
