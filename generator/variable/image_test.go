package variable_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyImageSurvivesAProfileRoundTrip(t *testing.T) {
	system := variable.NewSystem()
	img := &variable.ImageVariable{}
	require.NoError(t, system.Add("texture", img))

	assert.Empty(t, img.ToMessage())

	profile := system.GetProfile()
	require.NoError(t, system.ApplyProfile(profile))
	assert.Nil(t, img.GetValue())
}

func TestImageSurvivesAProfileRoundTrip(t *testing.T) {
	system := variable.NewSystem()
	img := &variable.ImageVariable{}
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.Set(1, 0, color.RGBA{R: 255, A: 255})
	img.SetValue(src)
	require.NoError(t, system.Add("texture", img))

	profile := system.GetProfile()
	img.SetValue(nil)
	require.NoError(t, system.ApplyProfile(profile))

	require.NotNil(t, img.GetValue())
	r, _, _, _ := img.GetValue().At(1, 0).RGBA()
	assert.Equal(t, uint32(0xffff), r)
}
