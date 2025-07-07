package main

import (
	"log"
	"os"

	"github.com/EliCDavis/polyform/generator"

	// Import these so they register their nodes with the generator
	_ "github.com/EliCDavis/polyform/drawing/texturing/normals"

	_ "github.com/EliCDavis/polyform/formats/colmap"
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/formats/obj"
	_ "github.com/EliCDavis/polyform/formats/opensfm"
	_ "github.com/EliCDavis/polyform/formats/ply"
	_ "github.com/EliCDavis/polyform/formats/splat"
	_ "github.com/EliCDavis/polyform/formats/spz"
	_ "github.com/EliCDavis/polyform/formats/stl"

	_ "github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/schema"

	_ "github.com/EliCDavis/polyform/math"
	_ "github.com/EliCDavis/polyform/math/colors"
	_ "github.com/EliCDavis/polyform/math/constant"
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

	_ "github.com/EliCDavis/polyform/nodes/experimental"
)

func main() {
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

		Out: os.Stdout,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
