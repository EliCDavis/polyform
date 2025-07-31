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

type ColorGradingLutNode = nodes.Struct[ColorGradingLutNodeData]

type ColorGradingLutNodeData struct {
	Mesh      nodes.Output[modeling.Mesh]
	Attribute nodes.Output[string]
	LUT       nodes.Output[image.Image]
}

func (ca3dn ColorGradingLutNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	out := nodes.StructOutput[modeling.Mesh]{}

	if ca3dn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.PointTopology))
		return out
	}

	mesh := nodes.GetOutputValue(out, ca3dn.Mesh)
	img := nodes.TryGetOutputValue(&out, ca3dn.LUT, nil)
	if img == nil {
		out.Set(mesh)
		return out
	}

	attr := nodes.TryGetOutputValue(&out, ca3dn.Attribute, modeling.FDCAttribute)
	out.Set(ColorGradingLut(mesh, img, attr))
	return out
}
