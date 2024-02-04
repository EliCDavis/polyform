package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type NormalImage struct {
	NumberOfLines nodes.Node[int]
	NumberOfWarts nodes.Node[int]
}

func (ni NormalImage) Process() (generator.Artifact, error) {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			// p := n.Noise(vector2.New(xDim*10, yDim*10), 100)
			p := (val * 128) + 128

			img.Set(x, y, color.RGBA{
				R: byte(p), // byte(len * 255),
				G: byte(p),
				B: byte(p),
				A: 255,
			})
		}
	}

	img = texturing.ToNormal(img)

	numLines := ni.NumberOfLines.Data()

	spacing := float64(dim) / float64(numLines)
	halfSpacing := float64(spacing) / 2.

	segments := 8
	yInc := float64(dim) / float64(segments)
	halfYInc := yInc / 2.

	for i := 0; i < numLines; i++ {
		dir := normals.Subtractive
		if rand.Float64() > 0.75 {
			dir = normals.Additive
		}

		startX := (float64(i) * spacing) + (spacing / 2)
		width := 4 + (rand.Float64() * 10)

		start := vector2.New(startX, 0)
		for seg := 0; seg < segments-1; seg++ {
			end := vector2.New(
				startX-(halfSpacing/2)+(rand.Float64()*halfSpacing),
				start.Y()+halfYInc+(yInc*rand.Float64()),
			)
			normals.Line{
				Start:           start,
				End:             end,
				Width:           width,
				NormalDirection: dir,
			}.Round(img)
			start = end
		}

		normals.Line{
			Start:           start,
			End:             vector2.New(startX, float64(dim)),
			Width:           width,
			NormalDirection: dir,
		}.Round(img)

	}

	numWarts := ni.NumberOfWarts.Data()
	wartSizeRange := vector2.New(8., 20.)
	for i := 0; i < numWarts; i++ {
		normals.Sphere{
			Center: vector2.New(
				float64(dim)*rand.Float64(),
				float64(dim)*rand.Float64(),
			),
			Radius: ((wartSizeRange.Y() - wartSizeRange.X()) * rand.Float64()) + wartSizeRange.X(),
		}.Draw(img)
	}

	// return img
	return generator.ImageArtifact{Image: texturing.BoxBlurNTimes(img, 10)}, nil
}

type PumpkinField struct {
	MaxWidth, TopDip, DistanceFromCenter, WedgeLineRadius nodes.Node[float64]
	Sides                                                 nodes.Node[int]
	ImageField                                            nodes.Node[[][]float64]
	UseImageField                                         nodes.Node[bool]
}

func (pf PumpkinField) Process() (marching.Field, error) {
	distanceFromCenter := pf.DistanceFromCenter.Data()
	maxWidth := pf.MaxWidth.Data()
	topDip := pf.TopDip.Data()
	wedgeLineRadius := pf.WedgeLineRadius.Data()
	sides := pf.Sides.Data()
	useImageField := pf.UseImageField.Data()
	imageField := pf.ImageField.Data()

	outerPoints := []vector3.Float64{
		vector3.New(0., .3, distanceFromCenter),
		vector3.New(0., .25, distanceFromCenter+(maxWidth*0.5)),
		vector3.New(0., 0.5, distanceFromCenter+maxWidth),
		vector3.New(0., .8, distanceFromCenter+(maxWidth*0.75)),
		vector3.New(0., 1-topDip, distanceFromCenter),
	}

	pointsBoundsLower, pointsBoundsHigher := vector3.Float64Array(outerPoints).Bounds()
	boundsCenter := pointsBoundsHigher.Midpoint(pointsBoundsLower)
	innerPoints := vector3.Float64Array(outerPoints).
		Add(boundsCenter.Scale(-1)).
		Scale(0.3).
		Add(boundsCenter)

	fields := make([]marching.Field, 0)
	angleInc := (math.Pi * 2.) / float64(sides)
	for i := 0; i < sides; i++ {
		rot := quaternion.FromTheta(angleInc*float64(i), vector3.Up[float64]())

		rotatedOuterPoints := jitterPositions(rot.RotateArray(outerPoints), .05, 10)

		outer := []sdf.LinePoint{
			{Point: rotatedOuterPoints[0], Radius: 0.33 * wedgeLineRadius * (.9 + (rand.Float64() * 0.2))},
			{Point: rotatedOuterPoints[1], Radius: 0.33 * wedgeLineRadius * (.9 + (rand.Float64() * 0.2))},
			{Point: rotatedOuterPoints[2], Radius: 1.00 * wedgeLineRadius * (.9 + (rand.Float64() * 0.2))},
			{Point: rotatedOuterPoints[3], Radius: 0.66 * wedgeLineRadius * (.9 + (rand.Float64() * 0.2))},
			{Point: rotatedOuterPoints[4], Radius: 0.33 * wedgeLineRadius * (.9 + (rand.Float64() * 0.2))},
		}

		inner := []sdf.LinePoint{
			{Point: rot.Rotate(innerPoints[0]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[1]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[2]), Radius: 1.00 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[3]), Radius: 0.66 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[4]), Radius: 0.33 * wedgeLineRadius},
		}

		if useImageField {
			fields = append(fields, marching.Subtract(marching.VarryingThicknessLine(outer, 1), marching.VarryingThicknessLine(inner, 2)))

		} else {
			fields = append(fields, marching.VarryingThicknessLine(outer, 1))
		}
	}

	allFields := marching.CombineFields(fields...)

	pumpkinField := allFields
	if useImageField {
		pumpkinField = marching.Subtract(
			allFields,
			marching.Field{
				Domain: allFields.Domain,
				Float1Functions: map[string]sample.Vec3ToFloat{
					modeling.PositionAttribute: func(f vector3.Float64) float64 {

						pixel := f.XY().
							Scale(float64(len(imageField)) * 2).
							RoundToInt().
							Sub(vector2.New(-len(imageField)/2, int(float64(len(imageField))*0.75)))

						if pixel.X() < 0 || pixel.X() >= len(imageField) {
							return 10
						}

						if pixel.Y() < 0 || pixel.Y() >= len(imageField) {
							return 10
						}

						if f.Z() < .2 {
							return 10
						}

						return -imageField[pixel.X()][len(imageField)-1-pixel.Y()]
					},
				},
			},
		)
	}

	return pumpkinField, nil
}

type SphericalUVMapping struct {
	Mesh nodes.Node[modeling.Mesh]
}

func (sm SphericalUVMapping) Process() (modeling.Mesh, error) {
	mesh := sm.Mesh.Data()
	pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
	newUVs := make([]vector2.Float64, pumpkinVerts.Len())
	center := vector3.New(0., 0.5, 0.)
	up := vector3.Up[float64]()
	for i := 0; i < pumpkinVerts.Len(); i++ {
		vert := pumpkinVerts.At(i)

		xzTheta := math.Atan2(vert.Z(), vert.X()) * 4
		xzTheta = math.Abs(xzTheta) // Avoid the UV seam

		dir := vert.Sub(center)
		angle := math.Acos(dir.Dot(up) / (dir.Length() * up.Length()))

		newUVs[i] = vector2.New(xzTheta/(math.Pi*2), angle)
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, newUVs), nil
}

type PumpkinGLBArtifact struct {
	PumpkinBody nodes.Node[modeling.Mesh]
	PumpkinStem nodes.Node[gltf.PolyformModel]
	LightColor  nodes.Node[color.Color]
}

func (pga PumpkinGLBArtifact) Process() (generator.Artifact, error) {
	return &generator.GltfArtifact{
		Scene: gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{
					Name: "Pumpkin",
					Mesh: pga.PumpkinBody.Data(),
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							BaseColorTexture: &gltf.PolyformTexture{
								URI: "Texturing/pumpkin.png", //"uvMap.png",
								// URI: "uvMap.png", //"uvMap.png",
								Sampler: &gltf.Sampler{
									WrapS: gltf.SamplerWrap_REPEAT,
									WrapT: gltf.SamplerWrap_REPEAT,
								},
							},
							MetallicRoughnessTexture: &gltf.PolyformTexture{
								URI: "Texturing/roughness.png",
							},
							// BaseColorFactor: texturingParams.Color("Base Color"),
							// MetallicFactor:  1,
							// RoughnessFactor: 0,
						},
						NormalTexture: &gltf.PolyformNormal{
							PolyformTexture: gltf.PolyformTexture{
								URI: "Texturing/normal.png",
							},
						},
						Extensions: []gltf.MaterialExtension{
							// gltf.PolyformMaterialsUnlit{},
						},
					},
				},
				pga.PumpkinStem.Data(),
			},
			Lights: []gltf.KHR_LightsPunctual{
				{
					Type:     gltf.KHR_LightsPunctualType_Point,
					Position: vector3.New(0., 0.5, 0.),
					Color:    pga.LightColor.Data(),
				},
			},
		},
	}, nil
}

type MetalRoughness struct {
	Roughness nodes.Node[float64]
}

func (mr MetalRoughness) Process() (generator.Artifact, error) {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 128) + 128

			p = 255 - (p * mr.Roughness.Data())

			img.Set(x, y, color.RGBA{
				R: 0, // byte(len * 255),
				G: byte(p),
				B: 0,
				A: 255,
			})
		}
	}
	return &generator.ImageArtifact{Image: img}, nil
}
