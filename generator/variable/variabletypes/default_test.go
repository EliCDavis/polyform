package variabletypes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DefaultWeave int

const (
	DefaultPlain DefaultWeave = iota
	DefaultTwill
)

var defaultWeaves = enum.Define(enum.Value(DefaultPlain, "Plain", ""), enum.Value(DefaultTwill, "Twill", ""))

func TestKeyOfPortType(t *testing.T) {
	cases := map[string]string{
		"float64": "float64",
		"github.com/EliCDavis/vector/vector3.Vector[float64]":                                                           "vector3.Vector[float64]",
		"github.com/EliCDavis/polyform/drawing/coloring.Color":                                                          "coloring.Color",
		"github.com/EliCDavis/polyform/drawing/coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]": "coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]",
		"github.com/EliCDavis/polyform/formats/gltf.AlphaMode":                                                          "gltf.AlphaMode",
		"enum/weave": "enum/weave",
	}
	for portType, key := range cases {
		assert.Equal(t, key, variabletypes.KeyOfPortType(portType), portType)
	}
}

func TestParseDefaultAppliesToItsVariable(t *testing.T) {
	cases := []struct {
		key, text, json string
	}{
		{"float64", "0.5", "0.5"},
		{"float64", "1.", "1"},
		{"int", "64", "64"},
		{"bool", "true", "true"},
		{"string", "hello world", `"hello world"`},
		{"coloring.Color", "#ff8800", `"#ff8800"`},
		{"coloring.Color", "#ff8800cc", `"#ff8800cc"`},
		{"coloring.Color", "#f80", `"#f80"`},
		{"vector2.Vector[float64]", "1,2", `{"x":1,"y":2}`},
		{"vector2.Vector[int]", "1, 2", `{"x":1,"y":2}`},
		{"vector3.Vector[float64]", "1,2.5,3", `{"x":1,"y":2.5,"z":3}`},
		{"vector3.Vector[int]", "256,256,256", `{"x":256,"y":256,"z":256}`},
		{"vector4.Vector[float64]", "1,2,3,4", `{"w":4,"x":1,"y":2,"z":3}`},
		{"quaternion.Quaternion", "0,0,0,1", `{"w":1,"x":0,"y":0,"z":0}`},
		{"geometry.AABB", "0,0,0;10,10,10", `{"center":{"x":0,"y":0,"z":0},"extents":{"x":5,"y":5,"z":5}}`},
		{"geometry.AABB", "1,2,3; 2,4,6", `{"center":{"x":1,"y":2,"z":3},"extents":{"x":1,"y":2,"z":3}}`},
	}
	for _, c := range cases {
		parsed, err := variabletypes.ParseDefault(c.key, c.text)
		require.NoError(t, err, "%s %q", c.key, c.text)
		assert.JSONEq(t, c.json, string(parsed), "%s %q", c.key, c.text)

		v, err := variabletypes.New(c.key)
		require.NoError(t, err, c.key)
		_, err = v.ApplyMessage(parsed)
		assert.NoError(t, err, "%s %q", c.key, c.text)
	}
}

func TestParseDefaultIsCaseInsensitiveOnTheKey(t *testing.T) {
	parsed, err := variabletypes.ParseDefault("Vector3.Vector[Float64]", "1,2,3")
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":1,"y":2,"z":3}`, string(parsed))
}

func TestParseDefaultRefuses(t *testing.T) {
	cases := []struct{ key, text string }{
		{"float64", "abc"},
		{"int", "1.5"},
		{"bool", "yes"},
		{"coloring.Color", "ff8800"},
		{"coloring.Color", "#ff888"},
		{"coloring.Color", "#gg8800"},
		{"vector3.Vector[float64]", "1,2"},
		{"vector3.Vector[int]", "1.5,2,3"},
		{"quaternion.Quaternion", "0,0,0"},
		{"trs.TRS", "1,2,3"},
		{"geometry.AABB", "1"},
		{"geometry.AABB", "0,0,0;1,1"},
		{"[]float64", "1,2"},
		{"image.Image", "x"},
		{"no.such.type", "1"},
	}
	for _, c := range cases {
		_, err := variabletypes.ParseDefault(c.key, c.text)
		assert.Error(t, err, "%s %q", c.key, c.text)
	}
}

func TestParseDefaultEnumByMemberName(t *testing.T) {
	variabletypes.RegisterEnum(defaultWeaves)

	parsed, err := variabletypes.ParseDefault("variabletypes_test.DefaultWeave", "Twill")
	require.NoError(t, err)
	assert.JSONEq(t, `"Twill"`, string(parsed))

	v, err := variabletypes.New("variabletypes_test.DefaultWeave")
	require.NoError(t, err)
	_, err = v.ApplyMessage(parsed)
	assert.NoError(t, err)

	unknown, err := variabletypes.ParseDefault("variabletypes_test.DefaultWeave", "Satin")
	require.NoError(t, err)
	_, err = v.ApplyMessage(unknown)
	assert.Error(t, err, "a member the set lacks is caught when applied")

	graphEnum, err := variabletypes.ParseDefault("enum/weave", "Twill")
	require.NoError(t, err)
	assert.JSONEq(t, `"Twill"`, string(graphEnum))
}
