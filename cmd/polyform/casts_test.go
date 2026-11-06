package main

import (
	"image"
	"testing"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	texturePkg = "github.com/EliCDavis/polyform/drawing/texturing.Texture"
	colorType  = texturePkg + "[github.com/EliCDavis/polyform/drawing/coloring.Color]"
	unormType  = "github.com/EliCDavis/polyform/math/norm.UFloat64"
	snormType  = "github.com/EliCDavis/polyform/math/norm.SFloat64"
)

func castTo[T any](t *testing.T, port nodes.OutputPort, to string) T {
	t.Helper()
	out, ok := casts().Output(port, to)
	require.True(t, ok, "no cast to %s", to)
	return out.(nodes.Output[T]).Value()
}

func castTexture[From, To any](t *testing.T, in texturing.Texture[From], to string) texturing.Texture[To] {
	t.Helper()
	return castTo[texturing.Texture[To]](t, nodes.ConstOutput[texturing.Texture[From]]{Val: in}, to)
}

func TestBoolCastsBothWays(t *testing.T) {
	assert.Equal(t, 1., castTo[float64](t, nodes.ConstOutput[bool]{Val: true}, "float64"))
	assert.Equal(t, 0, castTo[int](t, nodes.ConstOutput[bool]{Val: false}, "int"))
	assert.False(t, castTo[bool](t, nodes.ConstOutput[float64]{Val: 0}, "bool"))
	assert.True(t, castTo[bool](t, nodes.ConstOutput[float64]{Val: -0.2}, "bool"))
	assert.True(t, castTo[bool](t, nodes.ConstOutput[int]{Val: 2}, "bool"))
	assert.Equal(t, []bool{false, true}, castTo[[]bool](t, nodes.ConstOutput[[]int]{Val: []int{0, 5}}, "[]bool"))
}

func TestFloatToIntRounds(t *testing.T) {
	assert.Equal(t, []int{1, 3, -2}, castTo[[]int](t, nodes.ConstOutput[[]float64]{Val: []float64{0.6, 2.5, -1.6}}, "[]int"))
}

func TestNumberFillsEveryVectorComponent(t *testing.T) {
	assert.Equal(t, vector2.Fill(1.5), castTo[vector2.Float64](t, nodes.ConstOutput[float64]{Val: 1.5}, "github.com/EliCDavis/vector/vector2.Vector[float64]"))
	assert.Equal(t, vector3.Fill(2), castTo[vector3.Int](t, nodes.ConstOutput[int]{Val: 2}, "github.com/EliCDavis/vector/vector3.Vector[int]"))
	assert.Equal(t, vector4.Fill(0.5), castTo[vector4.Float64](t, nodes.ConstOutput[float64]{Val: 0.5}, "github.com/EliCDavis/vector/vector4.Vector[float64]"))
	assert.Equal(t,
		[]vector3.Float64{vector3.Fill(1.), vector3.Fill(2.)},
		castTo[[]vector3.Float64](t, nodes.ConstOutput[[]float64]{Val: []float64{1, 2}}, "[]github.com/EliCDavis/vector/vector3.Vector[float64]"))
}

func TestVector2WidensWithZeroZ(t *testing.T) {
	assert.Equal(t, vector3.New(1., 2., 0.), castTo[vector3.Float64](t, nodes.ConstOutput[vector2.Float64]{Val: vector2.New(1., 2.)}, "github.com/EliCDavis/vector/vector3.Vector[float64]"))
	assert.Equal(t, vector3.New(3, 4, 0), castTo[vector3.Int](t, nodes.ConstOutput[vector2.Int]{Val: vector2.New(3, 4)}, "github.com/EliCDavis/vector/vector3.Vector[int]"))

	_, narrows := casts().Output(nodes.ConstOutput[vector3.Float64]{Val: vector3.One[float64]()}, "github.com/EliCDavis/vector/vector2.Vector[float64]")
	assert.False(t, narrows)
}

func TestNormCastsClamp(t *testing.T) {
	assert.Equal(t, 1., castTo[norm.UFloat64](t, nodes.ConstOutput[float64]{Val: 1.5}, unormType).Value())
	assert.Equal(t,
		[]norm.SFloat64{-1, 0.5},
		castTo[[]norm.SFloat64](t, nodes.ConstOutput[[]float64]{Val: []float64{-1.5, 0.5}}, "[]"+snormType))
	assert.Equal(t, 1., castTo[float64](t, nodes.ConstOutput[norm.UFloat64]{Val: 1}, "float64"))
}

func TestIntsAndUNormsCastToNorms(t *testing.T) {
	assert.Equal(t, 1., castTo[norm.UFloat64](t, nodes.ConstOutput[int]{Val: 2}, unormType).Value())
	assert.Equal(t, -1., castTo[norm.SFloat64](t, nodes.ConstOutput[int]{Val: -3}, snormType).Value())
	assert.Equal(t, 0.3, castTo[norm.SFloat64](t, nodes.ConstOutput[norm.UFloat64]{Val: 0.3}, snormType).Value())
}

func TestNormsWithNoValueCastAsZero(t *testing.T) {
	assert.Equal(t, 0., castTo[float64](t, nodes.ConstOutput[norm.UFloat64]{}, "float64"))
	assert.Equal(t, 0., castTo[float64](t, nodes.ConstOutput[norm.SFloat64]{}, "float64"))
	assert.Equal(t, 0., castTo[norm.SFloat64](t, nodes.ConstOutput[norm.UFloat64]{}, snormType).Value())
}

func TestFloatTextureCastsToGreyscale(t *testing.T) {
	out := castTexture[float64, coloring.Color](t, texturing.FromArray([]float64{0, 0.25}, 2, 1), colorType)

	assert.Equal(t, coloring.Color{R: 0, G: 0, B: 0, A: 1}, out.Get(0, 0))
	assert.Equal(t, coloring.Color{R: 0.25, G: 0.25, B: 0.25, A: 1}, out.Get(1, 0))
}

func TestMaskTextureCastsToBlackAndWhite(t *testing.T) {
	out := castTexture[bool, coloring.Color](t, texturing.FromArray([]bool{true, false}, 2, 1), colorType)

	assert.Equal(t, coloring.Color{R: 1, G: 1, B: 1, A: 1}, out.Get(0, 0))
	assert.Equal(t, coloring.Color{A: 1}, out.Get(1, 0))
}

func TestColorTextureAndImageCastBothWays(t *testing.T) {
	table := casts()
	tex := texturing.FromArray([]coloring.Color{{R: 2, G: 0.5, B: -1, A: 1}}, 1, 1)

	img, ok := table.Output(nodes.ConstOutput[texturing.Texture[coloring.Color]]{Val: tex}, "image.Image")
	require.True(t, ok)
	r, g, b, a := img.(nodes.Output[image.Image]).Value().At(0, 0).RGBA()
	assert.Equal(t, []uint32{255, 127, 0, 255}, []uint32{r >> 8, g >> 8, b >> 8, a >> 8})

	back, ok := table.Output(img, colorType)
	require.True(t, ok)
	assert.Equal(t, coloring.Color{R: 1, G: 127. / 255, B: 0, A: 1}, back.(nodes.Output[texturing.Texture[coloring.Color]]).Value().Get(0, 0))
}

type errorRecorder struct{ errors []error }

func (r *errorRecorder) CaptureTiming(string, time.Duration) {}

func (r *errorRecorder) CaptureError(err error) {
	if err != nil {
		r.errors = append(r.errors, err)
	}
}

func TestMissingImageIsReportedByTheNodeReadingIt(t *testing.T) {
	out, ok := casts().Output(nodes.ConstOutput[image.Image]{}, colorType)
	require.True(t, ok)

	reader := &errorRecorder{}
	nodes.GetOutputValue(reader, out.(nodes.Output[texturing.Texture[coloring.Color]]))

	assert.Len(t, reader.errors, 1)
}

func TestMaskTextureCastsBothWays(t *testing.T) {
	floats := castTexture[bool, float64](t, texturing.FromArray([]bool{true, false}, 2, 1), texturePkg+"[float64]")
	assert.Equal(t, 1., floats.Get(0, 0))
	assert.Equal(t, 0., floats.Get(1, 0))

	mask := castTexture[float64, bool](t, texturing.FromArray([]float64{0, 0.3}, 2, 1), texturePkg+"[bool]")
	assert.False(t, mask.Get(0, 0))
	assert.True(t, mask.Get(1, 0))
}

func TestColoredFieldCastsToItsDistance(t *testing.T) {
	field := sdf.ColoredField{
		Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1),
		Color:    sdf.ConstantColor(coloring.Color{R: 1, A: 1}),
	}

	got := castTo[sample.Vec3ToFloat](t, nodes.ConstOutput[sdf.ColoredField]{Val: field}, "github.com/EliCDavis/polyform/math/sample.Vec3ToFloat")
	assert.InDelta(t, -1.0, got(vector3.New(0., 0., 0.)), 1e-9)
}

func TestMeshCastsIntoTheManifestModels(t *testing.T) {
	models := (&nodes.Struct[gltf.ManifestNode]{}).Inputs()["Models"].(nodes.Typed).Type()
	mesh := primitives.UVSphere(1, 4, 4)

	model := castTo[*gltf.PolyformModel](t, nodes.ConstOutput[modeling.Mesh]{Val: mesh}, models)

	assert.Equal(t, "Mesh", model.Name)
	assert.Equal(t, mesh.PrimitiveCount(), model.Mesh.PrimitiveCount())
	assert.Nil(t, model.Material)
}
