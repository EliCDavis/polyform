package gausops

import (
	"image"
	"math"

	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ColorGradingLutTransformer struct {
	Attribute string
	LUT       image.Image
}

func (cglt ColorGradingLutTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(cglt.Attribute, modeling.FDCAttribute)

	if err = meshops.RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return ColorGradingLut(m, cglt.LUT, attribute), nil
}

func ColorGradingLut(m modeling.Mesh, lut image.Image, attr string) modeling.Mesh {
	check(meshops.RequireV3Attribute(m, attr))

	width := float64(lut.Bounds().Dx())
	height := float64(lut.Bounds().Dy())

	cellWidth := width / 16

	fdcColors := m.Float3Attribute(attr)
	newColor := make([]vector3.Float64, fdcColors.Len())
	for i := 0; i < fdcColors.Len(); i++ {
		fdc := fdcColors.At(i)

		col := fdc.Scale(splat.SH_C0).
			Add(vector3.Fill(0.5)).
			Clamp(0, 1)

		px := lut.At(
			int((math.Floor(col.Z()*15)*cellWidth)+(col.X()*(cellWidth-1))),
			int(col.Y()*(height-1)),
		)

		r, g, b, _ := px.RGBA()

		newColor[i] = vector3.New(int(r>>8), int(g>>8), int(b>>8)).
			ToFloat64().
			DivByConstant(255).
			Sub(vector3.Fill(0.5)).
			DivByConstant(splat.SH_C0)
	}

	return m.SetFloat3Attribute(attr, newColor)
}

type ColorGradingLutNode = nodes.StructNode[modeling.Mesh, ColorGradingLutNodeData]

type ColorGradingLutNodeData struct {
	Mesh      nodes.NodeOutput[modeling.Mesh]
	Attribute nodes.NodeOutput[string]
	LUT       nodes.NodeOutput[image.Image]
}

func (ca3dn ColorGradingLutNodeData) Process() (modeling.Mesh, error) {
	attr := modeling.FDCAttribute

	if ca3dn.Attribute != nil {
		attr = ca3dn.Attribute.Value()
	}

	lut := ca3dn.LUT
	if lut == nil {
		return ca3dn.Mesh.Value(), nil
	}

	img := lut.Value()
	if img == nil {
		return ca3dn.Mesh.Value(), nil
	}

	return ColorGradingLut(ca3dn.Mesh.Value(), lut.Value(), attr), nil
}
