package meshops

import (
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func TranslateAttribute3DByPerlinNoise(m modeling.Mesh, attribute string, frequency, amplitude, shift vector3.Float64) modeling.Mesh {
	data := m.Float3Attribute(attribute)
	out := make([]vector3.Float64, data.Len())

	for i := range data.Len() {
		v := data.At(i)

		px := vector3.New(
			noise.Perlin1D((v.X()*frequency.X())+shift.X()),
			noise.Perlin1D((v.Y()*frequency.X())+shift.X()),
			noise.Perlin1D((v.Z()*frequency.X())+shift.X()),
		)

		py := vector3.New(
			noise.Perlin1D((v.X()*frequency.Y())+shift.Y()),
			noise.Perlin1D((v.Y()*frequency.Y())+shift.Y()),
			noise.Perlin1D((v.Z()*frequency.Y())+shift.Y()),
		)

		pz := vector3.New(
			noise.Perlin1D((v.X()*frequency.Z())+shift.Z()),
			noise.Perlin1D((v.Y()*frequency.Z())+shift.Z()),
			noise.Perlin1D((v.Z()*frequency.Z())+shift.Z()),
		)

		out[i] = v.Add(vector3.New(
			((px.Y()+px.Z())/2.)*amplitude.X(),
			((py.X()+py.Z())/2.)*amplitude.Y(),
			((pz.X()+pz.Y())/2.)*amplitude.Z(),
		))
	}

	return m.SetFloat3Attribute(attribute, out)
}

type TranslateAttribute3DByPerlinNoiseTransformer struct {
	Attribute                   string
	Frequency, Amplitude, Shift vector3.Float64
}

func (tabp TranslateAttribute3DByPerlinNoiseTransformer) attribute() string {
	return tabp.Attribute
}

func (tabp TranslateAttribute3DByPerlinNoiseTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(tabp, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return TranslateAttribute3DByPerlinNoise(m, attribute, tabp.Frequency, tabp.Amplitude, tabp.Shift), nil
}

type TranslateAttributeByPerlinNoise3DNode struct {
	Attribute nodes.Output[string]
	Mesh      nodes.Output[modeling.Mesh]
	Frequency nodes.Output[vector3.Float64]
	Amplitude nodes.Output[vector3.Float64]
	Shift     nodes.Output[vector3.Float64]
}

func (ta3dn TranslateAttributeByPerlinNoise3DNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if ta3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	out.Set(TranslateAttribute3DByPerlinNoise(
		nodes.GetOutputValue(out, ta3dn.Mesh),
		nodes.TryGetOutputValue(out, ta3dn.Attribute, modeling.PositionAttribute),
		nodes.TryGetOutputValue(out, ta3dn.Frequency, vector3.One[float64]()),
		nodes.TryGetOutputValue(out, ta3dn.Amplitude, vector3.One[float64]()),
		nodes.TryGetOutputValue(out, ta3dn.Shift, vector3.Zero[float64]()),
	))
}
