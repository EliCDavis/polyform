package gausops

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sh"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/utils"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func RotateSH(m modeling.Mesh, amount quaternion.Quaternion) modeling.Mesh {
	count := modeling.SHCoefficientCount(m)
	if count == 0 {
		return m
	}

	rotation := sh.NewRotation(amount)
	bands := make([]*iter.ArrayIterator[vector3.Float64], count)
	rotated := make([][]vector3.Float64, count)
	for j := range bands {
		bands[j] = m.Float3Attribute(modeling.SHAttribute(j))
		rotated[j] = make([]vector3.Float64, bands[j].Len())
	}

	utils.ParallelRange(bands[0].Len(), func(start, end int) {
		coefficients := make([]vector3.Float64, count)
		for i := start; i < end; i++ {
			for j := range bands {
				coefficients[j] = bands[j].At(i)
			}
			rotation.Rotate(coefficients)
			for j := range rotated {
				rotated[j][i] = coefficients[j]
			}
		}
	})

	for j := range rotated {
		m = m.SetFloat3Attribute(modeling.SHAttribute(j), rotated[j])
	}
	return m
}

func Rotate(m modeling.Mesh, amount quaternion.Quaternion) modeling.Mesh {
	if m.HasFloat3Attribute(modeling.PositionAttribute) {
		m = meshops.RotateAttribute3D(m, modeling.PositionAttribute, amount)
	}
	if m.HasFloat4Attribute(modeling.RotationAttribute) {
		m = RotateAttribute(m, modeling.RotationAttribute, amount)
	}
	return RotateSH(m, amount)
}

type RotateNode struct {
	Mesh   nodes.Output[modeling.Mesh]
	Amount nodes.Output[quaternion.Quaternion]
}

func (rn RotateNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if rn.Mesh == nil {
		out.Set(modeling.EmptyPointcloud())
		return
	}

	mesh := nodes.GetOutputValue(out, rn.Mesh)
	amount := nodes.TryGetOutputValue(out, rn.Amount, quaternion.Identity())
	out.Set(Rotate(mesh, amount))
}

// https://github.com/aras-p/UnityGaussianSplatting/blob/ff268cfc6e12b4db80e2b1e9f14f7e31a68a8e25/package/Shaders/SplatUtilities.compute#L548
func RotateAttribute(m modeling.Mesh, attribute string, amount quaternion.Quaternion) modeling.Mesh {
	amount = amount.Normalize()
	oldData := m.Float4Attribute(attribute)
	rotatedData := make([]vector4.Float64, oldData.Len())
	utils.ParallelRange(oldData.Len(), func(start, end int) {
		for i := start; i < end; i++ {
			old := oldData.At(i)
			rot := amount.Multiply(quaternion.New(vector3.New(old.Y(), old.Z(), old.W()), old.X()))
			rotatedData[i] = vector4.New(rot.W(), rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z())
		}
	})

	return m.SetFloat4Attribute(attribute, rotatedData)
}

type RotateAttributeNode struct {
	Mesh      nodes.Output[modeling.Mesh]
	Attribute nodes.Output[string] `default:"Rotation"`
	Amount    nodes.Output[quaternion.Quaternion]
}

func (rand RotateAttributeNode) Description() string {
	return "Rotates a gaussian splat attribute."
}

func (rand RotateAttributeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if rand.Mesh == nil {
		out.Set(modeling.EmptyPointcloud())
		return
	}

	mesh := nodes.GetOutputValue(out, rand.Mesh)
	amt := nodes.GetOutputValue(out, rand.Amount)
	attr := nodes.GetOutputValue(out, rand.Attribute)
	out.Set(RotateAttribute(mesh, attr, amt))
}
