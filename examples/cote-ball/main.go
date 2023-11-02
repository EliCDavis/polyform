package main

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func main() {

	app := &generator.App{
		Name:        "Cote Ball",
		Description: "Creating Tilable UV texture",
		Version:     "1.0.0",
		Authors: []generator.Author{
			{
				Name: "Eli Davis",
			},
		},
		Generator: &generator.Generator{
			Parameters: &generator.GroupParameter{
				Parameters: []generator.Parameter{
					&generator.GroupParameter{
						Name: "Noise",
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Negative Value Color",
								DefaultValue: coloring.WebColor{R: 0, G: 0, B: 0, A: 255},
							},
							&generator.ColorParameter{
								Name:         "Positive Value Color",
								DefaultValue: coloring.WebColor{R: 255, G: 255, B: 255, A: 255},
							},
							&generator.IntParameter{
								Name:         "Tiles",
								DefaultValue: 3,
							},
							&generator.FloatParameter{
								Name:         "Frequency",
								DefaultValue: 2,
							},
						},
					},
					&generator.GroupParameter{
						Name: "Sphere",
						Parameters: []generator.Parameter{
							&generator.IntParameter{
								Name:         "Rows",
								DefaultValue: 20,
							},
							&generator.IntParameter{
								Name:         "Columns",
								DefaultValue: 20,
							},
							&generator.FloatParameter{
								Name:         "Radius",
								DefaultValue: 1,
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"texture.png": func(c *generator.Context) (generator.Artifact, error) {
					dim := 1024
					img := image.NewRGBA(image.Rect(0, 0, dim, dim))

					textureParams := c.Parameters.Group("Noise")
					tiles := float64(textureParams.Int("Tiles"))
					frequency := textureParams.Float64("Frequency")
					negativeColor := textureParams.Color("Negative Value Color")
					positiveColor := textureParams.Color("Positive Value Color")

					nR, nG, nB, _ := negativeColor.RGBA()
					pR, pG, pB, _ := positiveColor.RGBA()

					rRange := float64(pR>>8) - float64(nR>>8)
					gRange := float64(pG>>8) - float64(nG>>8)
					bRange := float64(pB>>8) - float64(nB>>8)

					for x := 0; x < dim; x++ {
						xDim := (float64(x) / float64(dim)) * tiles
						xRot := xDim * math.Pi * 2.

						for y := 0; y < dim; y++ {
							yDim := (float64(y) / float64(dim)) * tiles
							yRot := yDim * math.Pi * 2.

							// A regular sphere
							rot1 := modeling.UnitQuaternionFromTheta(yRot, vector3.Up[float64]())
							rot2 := modeling.UnitQuaternionFromTheta(xRot, vector3.Forward[float64]())
							final := rot1.Rotate(rot2.Rotate(vector3.Right[float64]()))

							p := noise.Perlin3D(final.Scale(frequency))

							r := uint32(float64(nR) + (rRange * p))
							g := uint32(float64(nG) + (gRange * p))
							b := uint32(float64(nB) + (bRange * p))

							img.Set(x, y, color.RGBA{
								R: byte(r), // byte(len * 255),
								G: byte(g),
								B: byte(b),
								A: 255,
							})
						}
					}
					return &generator.ImageArtifact{Image: img}, nil
				},
				"sphere.glb": func(c *generator.Context) (generator.Artifact, error) {
					sphereParams := c.Parameters.Group("Sphere")

					sphere := primitives.UVSphere(
						sphereParams.Float64("Radius"),
						sphereParams.Int("Rows"),
						sphereParams.Int("Columns"),
					)
					verts := sphere.Float3Attribute(modeling.PositionAttribute)

					uvs := make([]vector2.Float64, verts.Len())
					for i := 0; i < verts.Len(); i++ {
						v := verts.At(i)

						xz := v.XZ().Normalized()
						xzTheta := math.Atan2(xz.Y(), xz.X())

						xy := v.XY().Normalized()
						xyTheta := math.Atan2(xy.Y(), xy.X())

						uvs[i] = vector2.New(xzTheta, xyTheta).
							Scale(1. / (2. * math.Pi)).
							Add(vector2.One[float64]()).
							Scale(0.5)
					}

					mappedSphere := sphere.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "UV Sphere",
									Mesh: mappedSphere,
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorTexture: &gltf.PolyformTexture{
												URI: "texture.png",
											},
											BaseColorFactor: color.White,
										},
									},
								},
							},
						},
					}, nil
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}

}
