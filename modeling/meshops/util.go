package meshops

import (
	"errors"
	"fmt"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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

func RequireTopology(m modeling.Mesh, topo modeling.Topology) error {
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

func RequireV4Attribute(m modeling.Mesh, attr string) error {
	if m.HasFloat4Attribute(attr) {
		return nil
	}
	return fmt.Errorf("mesh is required to have the vector4 attribute: '%s'", attr)
}

func RequireV3Attribute(m modeling.Mesh, attr string) error {
	if m.HasFloat3Attribute(attr) {
		return nil
	}
	return fmt.Errorf("mesh is required to have the vector3 attribute: '%s'", attr)
}

func RequireV2Attribute(m modeling.Mesh, attr string) error {
	if m.HasFloat2Attribute(attr) {
		return nil
	}
	return fmt.Errorf("mesh is required to have the vector2 attribute: '%s'", attr)
}

func RequireV1Attribute(m modeling.Mesh, attr string) error {
	if m.HasFloat1Attribute(attr) {
		return nil
	}
	return fmt.Errorf("mesh is required to have the vector1 attribute: '%s'", attr)
}

func readAllFloatXData[T any](attrs []string, reader func(string) *iter.ArrayIterator[T]) map[string][]T {
	data := make(map[string][]T)
	for _, attr := range attrs {
		attrData := reader(attr)
		data[attr] = iter.ReadFull[T](attrData)
	}
	return data
}

func readAllFloat4Data(m modeling.Mesh) map[string][]vector4.Float64 {
	return readAllFloatXData(m.Float4Attributes(), func(s string) *iter.ArrayIterator[vector4.Float64] { return m.Float4Attribute(s) })
}

func readAllFloat3Data(m modeling.Mesh) map[string][]vector3.Float64 {
	return readAllFloatXData(m.Float3Attributes(), func(s string) *iter.ArrayIterator[vector3.Float64] { return m.Float3Attribute(s) })
}

func readAllFloat2Data(m modeling.Mesh) map[string][]vector2.Float64 {
	return readAllFloatXData(m.Float2Attributes(), func(s string) *iter.ArrayIterator[vector2.Float64] { return m.Float2Attribute(s) })
}

func readAllFloat1Data(m modeling.Mesh) map[string][]float64 {
	return readAllFloatXData(m.Float1Attributes(), func(s string) *iter.ArrayIterator[float64] { return m.Float1Attribute(s) })
}
