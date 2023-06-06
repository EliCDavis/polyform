package meshops

import (
	"errors"
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	ErrRequireTriangleTopology  = errors.New("mesh is required to have a triangle topology")
	ErrRequireLineTopology      = errors.New("mesh is required to have a line topology")
	ErrRequirePointTopology     = errors.New("mesh is required to have a point topology")
	ErrRequireDifferentTopology = errors.New("mesh does not have required topology")
)

func requireTopology(m modeling.Mesh, topo modeling.Topology) error {
	if m.Topology() == topo {
		return nil
	}

	switch topo {
	case modeling.TriangleTopology:
		return ErrRequireTriangleTopology

	case modeling.LineTopology:
		return ErrRequireLineTopology

	case modeling.PointTopology:
		return ErrRequirePointTopology
	}

	return ErrRequireDifferentTopology
}

func requireV3Attribute(m modeling.Mesh, attr string) error {
	if m.HasFloat3Attribute(attr) {
		return nil
	}
	return fmt.Errorf("mesh is required to have the vector3 attribute: '%s'", attr)
}
