package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"

	// Import these so they register their nodes with the generator
	"github.com/EliCDavis/polyform/drawing/coloring"
	_ "github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	_ "github.com/EliCDavis/polyform/drawing/texturing"
	_ "github.com/EliCDavis/polyform/drawing/texturing/normals"

	_ "github.com/EliCDavis/polyform/formats/colmap"
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/formats/obj"
	_ "github.com/EliCDavis/polyform/formats/opensfm"
	_ "github.com/EliCDavis/polyform/formats/ply"
	_ "github.com/EliCDavis/polyform/formats/splat"
	_ "github.com/EliCDavis/polyform/formats/spz"
	_ "github.com/EliCDavis/polyform/formats/stl"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/generator/variable"

	_ "github.com/EliCDavis/polyform/math"
	_ "github.com/EliCDavis/polyform/math/constant"
	"github.com/EliCDavis/polyform/math/geometry"
	_ "github.com/EliCDavis/polyform/math/geometry"
	_ "github.com/EliCDavis/polyform/math/noise"
	_ "github.com/EliCDavis/polyform/math/quaternion"
	_ "github.com/EliCDavis/polyform/math/trig"
	_ "github.com/EliCDavis/polyform/math/trs"
	_ "github.com/EliCDavis/polyform/math/unit"
	_ "github.com/EliCDavis/polyform/math/vector2"
	_ "github.com/EliCDavis/polyform/math/vector3"

	_ "github.com/EliCDavis/polyform/modeling"
	_ "github.com/EliCDavis/polyform/modeling/extrude"
	_ "github.com/EliCDavis/polyform/modeling/meshops"
	_ "github.com/EliCDavis/polyform/modeling/meshops/gausops"
	_ "github.com/EliCDavis/polyform/modeling/primitives"
	_ "github.com/EliCDavis/polyform/modeling/repeat"
	_ "github.com/EliCDavis/polyform/modeling/triangulation"
	_ "github.com/EliCDavis/polyform/modeling/voxelize"

	_ "github.com/EliCDavis/polyform/nodes/experimental"
)

func main() {
	outputSerialization := &serialize.TypeSwitch[manifest.Entry]{}

	serialize.Register(outputSerialization, func(i image.Image) manifest.Entry {
		return manifest.Entry{Artifact: basics.Image{Image: i}}
	})

	serialize.Register(outputSerialization, func(i texturing.Texture[float64]) manifest.Entry {
		return manifest.Entry{Artifact: texturing.Artifact[float64]{
			Texture: i,
			Conversion: func(v float64) color.Color {
				return color.RGBA{R: byte(v * 255), G: byte(v * 255), B: byte(v * 255), A: 255}
			},
		}}
	})

	serialize.Register(outputSerialization, func(i texturing.Texture[bool]) manifest.Entry {
		return manifest.Entry{Artifact: texturing.Artifact[bool]{
			Texture: i,
			Conversion: func(b bool) color.Color {
				var v byte = 0
				if b {
					v = 255
				}
				return color.RGBA{R: v, G: v, B: v, A: 255}
			},
		}}
	})

	serialize.Register(outputSerialization, func(i texturing.Texture[color.Color]) manifest.Entry {
		return manifest.Entry{Artifact: texturing.Artifact[color.Color]{
			Texture: i,
			Conversion: func(b color.Color) color.Color {
				if b == nil {
					return color.RGBA{R: 0, G: 0, B: 0, A: 255}
				}
				return b
			},
		}}
	})

	app := generator.App{
		Name:        "Polyform",
		Description: "Immutable mesh processing pipelines",
		Authors: []schema.Author{
			{
				Name: "Eli C Davis",
				ContactInfo: []schema.AuthorContact{
					{Medium: "bsky.app", Value: "@elicdavis.bsky.social"},
					{Medium: "github.com", Value: "EliCDavis"},
				},
			},
		},
		NodeOutputSerialization: outputSerialization,
		VariableFactory: func(variableType string) (variable.Variable, error) {
			switch strings.ToLower(variableType) {
			case "float64":
				return &variable.TypeVariable[float64]{}, nil

			case "string":
				return &variable.TypeVariable[string]{}, nil

			case "int":
				return &variable.TypeVariable[int]{}, nil

			case "bool":
				return &variable.TypeVariable[bool]{}, nil

			case "vector2.vector[float64]":
				return &variable.TypeVariable[vector2.Float64]{}, nil

			case "vector2.vector[int]":
				return &variable.TypeVariable[vector2.Int]{}, nil

			case "vector3.vector[float64]":
				return &variable.TypeVariable[vector3.Float64]{}, nil

			case "vector3.vector[int]":
				return &variable.TypeVariable[vector3.Int]{}, nil

			case "[]vector3.vector[float64]":
				return &variable.TypeVariable[[]vector3.Float64]{}, nil

			case "geometry.aabb":
				return &variable.TypeVariable[geometry.AABB]{}, nil

			case "coloring.webcolor":
				return &variable.TypeVariable[coloring.WebColor]{}, nil

			case "image.image":
				return &variable.ImageVariable{}, nil

			case "file":
				return &variable.FileVariable{}, nil

			default:
				return nil, fmt.Errorf("unrecognized variable type: %q", variableType)
			}
		},

		Out: os.Stdout,
		Err: os.Stderr,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
