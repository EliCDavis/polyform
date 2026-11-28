package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type ScaleAttribute3DTransformer struct {
	Attribute string
	Origin    vector3.Float64
	Amount    vector3.Float64
}

func (st ScaleAttribute3DTransformer) attribute() string {
	return st.Attribute
}

func (st ScaleAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttribute3D(m, attribute, st.Origin, st.Amount), nil
}

func ScaleAttribute3D(m modeling.Mesh, attribute string, origin, amount vector3.Float64) modeling.Mesh {
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = origin.Add(oldData.At(i).Sub(origin).MultByVector(amount))
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}

// ============================================================================

// ============================================================================

type ScaleAttribute2DTransformer struct {
	Attribute string
	Origin    vector2.Float64
	Amount    vector2.Float64
}

func (st ScaleAttribute2DTransformer) attribute() string {
	return st.Attribute
}

func (st ScaleAttribute2DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.TexCoordAttribute)

	if err = RequireV2Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttribute2D(m, attribute, st.Origin, st.Amount), nil
}

func ScaleAttribute2D(m modeling.Mesh, attribute string, origin, amount vector2.Float64) modeling.Mesh {
	if err := RequireV2Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float2Attribute(attribute)
	scaledData := make([]vector2.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = origin.Add(oldData.At(i).Sub(origin).MultByVector(amount))
	}

	return m.SetFloat2Attribute(attribute, scaledData)
}

func ScaleAttribute1D(m modeling.Mesh, attribute string, origin, amount float64) modeling.Mesh {
	return m.ModifyFloat1Attribute(attribute, func(_ int, v float64) float64 { return origin + (v-origin)*amount })
}

func ScaleAttribute4D(m modeling.Mesh, attribute string, origin, amount vector4.Float64) modeling.Mesh {
	return modifyFloat4(m, attribute, func(v vector4.Float64) vector4.Float64 {
		return origin.Add(v.Sub(origin).MultByVector(amount))
	})
}

func zipScale[T any](out *nodes.Lifted[modeling.Mesh], attribute string, mesh nodes.LiftedPort[modeling.Mesh], origin, amount nodes.LiftedPort[T], apply func(modeling.Mesh, string, T, T) modeling.Mesh) {
	nodes.Zip3(
		out,
		nodes.LiftedOr(mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		origin,
		amount,
		func(m modeling.Mesh, origin, amount T) modeling.Mesh {
			if m.PrimitiveCount() == 0 {
				return m
			}
			return apply(m, attribute, origin, amount)
		},
	)
}

type ScaleAttribute1DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[float64] `default:"1"`
	Origin    nodes.LiftedPort[float64]
}

func (n ScaleAttribute1DNode) Description() string {
	return "Scales one named scalar attribute of a mesh."
}

func (n ScaleAttribute1DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipScale(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Origin, n.Amount, ScaleAttribute1D)
}

type ScaleAttribute2DNode struct {
	Attribute nodes.Output[string] `default:"TexCoord"`
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[vector2.Float64] `default:"1,1"`
	Origin    nodes.LiftedPort[vector2.Float64]
}

func (n ScaleAttribute2DNode) Description() string {
	return "Scales one named 2D attribute of a mesh, per axis."
}

func (n ScaleAttribute2DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipScale(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Origin, n.Amount, ScaleAttribute2D)
}

type ScaleAttribute3DNode struct {
	Attribute nodes.Output[string] `default:"Position"`
	Mesh      nodes.Output[modeling.Mesh]
	Amount    nodes.Output[vector3.Float64] `default:"1,1,1"`
	Origin    nodes.Output[vector3.Float64]
}

func (sa3dn ScaleAttribute3DNode) Description() string {
	return "Scales one named vector attribute of a mesh, per axis."
}

func (sa3dn ScaleAttribute3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if sa3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}
	out.Set(ScaleAttribute3D(
		nodes.GetOutputValue(out, sa3dn.Mesh),
		nodes.GetOutputValue(out, sa3dn.Attribute),
		nodes.GetOutputValue(out, sa3dn.Origin),
		nodes.GetOutputValue(out, sa3dn.Amount),
	))
}

type ScaleAttribute4DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.LiftedPort[modeling.Mesh]
	Amount    nodes.LiftedPort[vector4.Float64] `default:"1,1,1,1"`
	Origin    nodes.LiftedPort[vector4.Float64]
}

func (n ScaleAttribute4DNode) Description() string {
	return "Scales one named 4D attribute of a mesh, per axis."
}

func (n ScaleAttribute4DNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	zipScale(out, nodes.GetOutputValue(out, n.Attribute), n.Mesh, n.Origin, n.Amount, ScaleAttribute4D)
}
