package meshops_test

import (
	"testing"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func attributeMesh() modeling.Mesh {
	return modeling.NewPointCloud(nil, nil, nil, nil).
		SetFloat1Attribute("w", []float64{1, 3}).
		SetFloat2Attribute("uv", []vector2.Float64{vector2.New(1., 2.), vector2.New(3., 4.)}).
		SetFloat4Attribute("q", []vector4.Float64{vector4.New(1., 2., 3., 4.), vector4.New(5., 6., 7., 8.)})
}

func TestTranslateAttributeInOneTwoAndFourDimensions(t *testing.T) {
	m := attributeMesh()
	assert.Equal(t, []float64{1.5, 3.5}, iter.ReadFull(meshops.TranslateAttribute1D(m, "w", 0.5).Float1Attribute("w")))
	assert.Equal(t,
		[]vector2.Float64{vector2.New(2., 2.), vector2.New(4., 4.)},
		iter.ReadFull(meshops.TranslateAttribute2D(m, "uv", vector2.New(1., 0.)).Float2Attribute("uv")),
	)
	assert.Equal(t,
		[]vector4.Float64{vector4.New(1., 2., 3., 5.), vector4.New(5., 6., 7., 9.)},
		iter.ReadFull(meshops.TranslateAttribute4D(m, "q", vector4.New(0., 0., 0., 1.)).Float4Attribute("q")),
	)
}

func TestScaleAttributeInOneTwoAndFourDimensionsAboutAnOrigin(t *testing.T) {
	m := attributeMesh()
	assert.Equal(t, []float64{1, 5}, iter.ReadFull(meshops.ScaleAttribute1D(m, "w", 1, 2).Float1Attribute("w")))
	assert.Equal(t,
		[]vector2.Float64{vector2.New(1., 2.), vector2.New(5., 2.)},
		iter.ReadFull(meshops.ScaleAttribute2D(m, "uv", vector2.New(1., 2.), vector2.New(2., 0.)).Float2Attribute("uv")),
	)
	assert.Equal(t,
		[]vector4.Float64{vector4.New(2., 4., 6., 8.), vector4.New(10., 12., 14., 16.)},
		iter.ReadFull(meshops.ScaleAttribute4D(m, "q", vector4.Zero[float64](), vector4.Fill(2.)).Float4Attribute("q")),
	)
}
