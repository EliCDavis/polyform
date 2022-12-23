package modeling

import "fmt"

type Topology int

const (
	Triangle Topology = iota
	Point
	Line
	Quad
)

func (t Topology) String() string {
	switch t {
	case Triangle:
		return "triangle"

	case Point:
		return "point"

	case Line:
		return "line"

	case Quad:
		return "quad"
	}

	panic(fmt.Errorf("unimplemented topology string case: %d", t))
}

func (t Topology) IndexSize() int {
	switch t {
	case Triangle:
		return 3

	case Point:
		return 1

	case Line:
		return 2

	case Quad:
		return 4
	}

	panic(fmt.Errorf("unimplemented topology index size case: %d", t))
}
