package potree

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type MetadataHierarchy struct {
	FirstChunkSize uint64 `json:"firstChunkSize"`
	StepSize       int    `json:"stepSize"`
	Depth          int    `json:"depth"`
}

type MetadataBounds struct {
	Min []float64 `json:"min"`
	Max []float64 `json:"max"`
}

func (mb MetadataBounds) MinF() vector3.Float64 {
	return vector3.New(mb.Min[0], mb.Min[1], mb.Min[2])
}

func (mb MetadataBounds) MaxF() vector3.Float64 {
	return vector3.New(mb.Max[0], mb.Max[1], mb.Max[2])
}

type Metadata struct {
	Version     string            `json:"version"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Points      int64             `json:"points"`
	Projection  string            `json:"projection"`
	Hierarchy   MetadataHierarchy `json:"hierarchy"`
	Offset      []float64         `json:"offset"`
	Scale       []float64         `json:"scale"`
	Spacing     float64           `json:"spacing"`
	BoundingBox MetadataBounds    `json:"boundingBox"`
	Encoding    string            `json:"encoding"`
	Attributes  []Attribute       `json:"attributes"`
}

func LoadMetadata(filepath string) (*Metadata, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadMetadata(f)
}

func ReadMetadata(in io.Reader) (*Metadata, error) {
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	m := &Metadata{}
	return m, json.Unmarshal(data, m)
}

func (m Metadata) OffsetF() vector3.Float64 {
	return vector3.New(m.Offset[0], m.Offset[1], m.Offset[2])
}

func (m Metadata) BytesPerPoint() int {
	count := 0
	for _, attr := range m.Attributes {
		count += attr.Size
	}
	return count
}

func (m Metadata) AttributeOffset(attribute string) int {
	count := 0
	for _, attr := range m.Attributes {
		if attr.Name == attribute {
			return count
		}
		count += attr.Size
	}
	return -1
}

func (m Metadata) LoadHierarchy(filepath string) (*OctreeNode, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return m.ReadHierarchy(bufio.NewReader(f))
}

func (m Metadata) ReadHierarchy(in io.Reader) (*OctreeNode, error) {
	// offset := m.OffsetF()
	root := &OctreeNode{
		Name: "r",
		BoundingBox: geometry.NewAABBFromPoints(
			m.BoundingBox.MinF(),
			m.BoundingBox.MaxF(),
		),
		Level:               0,
		NodeType:            2,
		HierarchyByteOffset: 0,
		HierarchyByteSize:   m.Hierarchy.FirstChunkSize,
		Spacing:             m.Spacing,
		ByteOffset:          0,
	}

	buf, err := io.ReadAll(in)
	if err != nil {
		return root, err
	}

	return root, ParseEntireHierarchy(root, buf)
}
