package modeling

import "fmt"

type Topology int

const (
	TriangleTopology Topology = iota
	PointTopology
	LineTopology
	QuadTopology
)

func (t Topology) String() string {
	switch t {
	case TriangleTopology:
		return "triangle"

	case PointTopology:
		return "point"

	case LineTopology:
		return "line"

	case QuadTopology:
		return "quad"
	}

	panic(fmt.Errorf("unimplemented topology string case: %d", t))
}

func (t Topology) IndexSize() int {
	switch t {
	case TriangleTopology:
		return 3

	case PointTopology:
		return 1

	case LineTopology:
		return 2

	case QuadTopology:
		return 4
	}

	panic(fmt.Errorf("unimplemented topology index size case: %d", t))
}
