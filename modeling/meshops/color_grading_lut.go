package meshops

import (
	"image"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type ColorGradingLutTransformer struct {
	Attribute string

	LUT image.Image
}

func (cglt ColorGradingLutTransformer) attribute() string {
	return cglt.Attribute
}

func (cglt ColorGradingLutTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(cglt, modeling.ColorAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return ColorGradingLut(m, cglt.LUT, attribute), nil
}

func ColorGradingLut(m modeling.Mesh, lut image.Image, attr string) modeling.Mesh {
	check(RequireV3Attribute(m, attr))

	width := float64(lut.Bounds().Dx())
	height := float64(lut.Bounds().Dy())

	cellWidth := width / 16

	oldColors := m.Float3Attribute(attr)
	newColor := make([]vector3.Float64, oldColors.Len())
	for i := 0; i < oldColors.Len(); i++ {
		old := oldColors.At(i)

		col := lut.At(
			int((math.Floor(old.Z()*15)*cellWidth)+(old.X()*(cellWidth-1))),
			int(old.Y()*(height-1)),
		)

		r, g, b, _ := col.RGBA()

		newColor[i] = vector3.New(int(r>>8), int(g>>8), int(b>>8)).
			ToFloat64().
			DivByConstant(255)
	}

	return m.SetFloat3Attribute(attr, newColor)
}
