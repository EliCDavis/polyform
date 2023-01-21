package modeling

import "fmt"

type Topology int

const (
	TriangleTopology Topology = iota
	PointTopology
	QuadTopology
	LineTopology
	LineStripTopology
	LineLoopTopology
)

func (t Topology) String() string {
	switch t {
	case TriangleTopology:
		return "triangle"

	case PointTopology:
		return "point"

	case LineTopology:
		return "line"

	case LineStripTopology:
		return "line strip"

	case LineLoopTopology:
		return "line loop"

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

	case LineTopology, LineStripTopology, LineLoopTopology:
		return 2

	case QuadTopology:
		return 4
	}

	panic(fmt.Errorf("unimplemented topology index size case: %d", t))
}
