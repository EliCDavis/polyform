package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/colors"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func closestTimeOnMultiLineSegment(point vector3.Float64, multiLine []vector3.Float64, totalLength float64) float64 {
	if len(multiLine) < 2 {
		panic("line segment required 2 or more points")
	}

	minDist := math.MaxFloat64

	closestTime := 0.
	lengthTraversed := 0.
	for i := 1; i < len(multiLine); i++ {
		line := geometry.NewLine3D(multiLine[i-1], multiLine[i])
		lineLength := line.Length()
		dist := line.ClosestPointOnLine(point).Distance(point)
		if dist < minDist {
			minDist = dist
			closestTime = (lengthTraversed + (lineLength * line.ClosestTimeOnLine(point))) / totalLength
		}
		lengthTraversed += lineLength
	}

	return closestTime
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

type Albedo struct {
	Positive nodes.Node[color.Color]
	Negative nodes.Node[color.Color]
}

func (an *Albedo) Process() (image.Image, error) {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	nR, nG, nB, _ := an.Negative.Data().RGBA()
	pR, pG, pB, _ := an.Positive.Data().RGBA()

	rRange := float64(pR>>8) - float64(nR>>8)
	gRange := float64(pG>>8) - float64(nG>>8)
	bRange := float64(pB>>8) - float64(nB>>8)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 0.5) + 0.5

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
	return img, nil
}

type StemNormalImage struct {
	NumberOfLines nodes.Node[int]
}

func (sni StemNormalImage) Process() (generator.Artifact, error) {
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

	numLines := sni.NumberOfLines.Data()

	spacing := float64(dim) / float64(numLines)
	halfSpacing := float64(spacing) / 2.

	segments := 8
	yInc := float64(dim) / float64(segments)
	halfYInc := yInc / 2.

	for i := 0; i < numLines; i++ {
		dir := normals.Subtractive
		if rand.Float64() > 0.5 {
			dir = normals.Additive
		}

		startX := (float64(i) * spacing) + (spacing / 2)
		width := 10 + (rand.Float64() * 20)

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

	return &generator.ImageArtifact{Image: img}, nil
}

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

func jitterPositions(pos []vector3.Float64, amplitude, frequency float64) []vector3.Float64 {
	return vector3.Array[float64](pos).
		Modify(func(v vector3.Float64) vector3.Float64 {
			return vector3.New(
				noise.Perlin1D((v.X()*frequency)+0),
				noise.Perlin1D((v.Y()*frequency)+100),
				noise.Perlin1D((v.Z()*frequency)+200),
			).Scale(amplitude).Add(v)
		})
}

type LaplacianSmoothTransformerParams struct {
	Attribute       nodes.Node[string]
	Iterations      nodes.Node[int]
	SmoothingFactor nodes.Node[float64]
	Mesh            nodes.Node[modeling.Mesh]
}

func LaplacianSmoothingNode(
	mesh nodes.Node[modeling.Mesh],
	iterations nodes.Node[int],
	factor nodes.Node[float64],
) *nodes.TransformerNode[LaplacianSmoothTransformerParams, modeling.Mesh] {
	return nodes.Transformer(
		"Laplacian Smoothing",
		LaplacianSmoothTransformerParams{
			Attribute:       nodes.Input[string](modeling.PositionAttribute),
			Iterations:      iterations,
			SmoothingFactor: factor,
			Mesh:            mesh,
		},
		func(in LaplacianSmoothTransformerParams) (modeling.Mesh, error) {
			return meshops.LaplacianSmooth(
				in.Mesh.Data(),
				in.Attribute.Data(),
				in.Iterations.Data(),
				in.SmoothingFactor.Data(),
			), nil
		},
	)
}

type MeshNodeParams struct {
	Mesh nodes.Node[modeling.Mesh]
}

func SmoothNormalsNode(
	mesh nodes.Node[modeling.Mesh],
) *nodes.TransformerNode[MeshNodeParams, modeling.Mesh] {
	return nodes.Transformer(
		"Smooth Normals",
		MeshNodeParams{
			Mesh: mesh,
		},
		func(in MeshNodeParams) (modeling.Mesh, error) {
			return meshops.SmoothNormals(
				in.Mesh.Data(),
			), nil
		},
	)
}

func newPumpkinMesh(imageField nodes.Node[[][]float64], topDip nodes.Node[float64]) nodes.Node[modeling.Mesh] {

	type PumpkinParams struct {
		CubersPerUnit, MaxWidth, TopDip, DistanceFromCenter, WedgeLineRadius nodes.Node[float64]
		Sides                                                                nodes.Node[int]
		ImageField                                                           nodes.Node[[][]float64]
		UseImageField                                                        nodes.Node[bool]
	}

	pumpkinParams := PumpkinParams{
		CubersPerUnit: &generator.ParameterNode[float64]{
			Name:         "Pumpkin Resolution",
			DefaultValue: 20,
		},
		MaxWidth: &generator.ParameterNode[float64]{
			Name:         "Max Width",
			DefaultValue: .3,
		},
		TopDip: topDip,
		DistanceFromCenter: &generator.ParameterNode[float64]{
			Name:         "Wedge Spacing",
			DefaultValue: .1,
		},
		WedgeLineRadius: &generator.ParameterNode[float64]{
			Name:         "Wedge Radius",
			DefaultValue: .3,
		},
		Sides: &generator.ParameterNode[int]{
			Name:         "Wedges",
			DefaultValue: 10,
		},
		UseImageField: &generator.ParameterNode[bool]{
			Name:         "Carve",
			DefaultValue: true,
		},
		ImageField: imageField,
	}

	pumpkinMeshNode := nodes.Transformer("Pumpkin Mesh", pumpkinParams, func(in PumpkinParams) (modeling.Mesh, error) {
		canvas := marching.NewMarchingCanvas(in.CubersPerUnit.Data())

		distanceFromCenter := in.DistanceFromCenter.Data()
		maxWidth := in.MaxWidth.Data()
		topDip := in.TopDip.Data()
		wedgeLineRadius := in.WedgeLineRadius.Data()
		sides := in.Sides.Data()
		useImageField := in.UseImageField.Data()
		imageField := in.ImageField.Data()

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

		addFieldStart := time.Now()
		canvas.AddField(pumpkinField)
		// canvas.AddFieldParallel(pumpkinField)
		log.Printf("time to add field: %s", time.Since(addFieldStart))

		marchStart := time.Now()
		log.Println("starting march...")
		// mesh := canvas.MarchParallel(0)
		mesh := canvas.March(0)
		// mesh := pumpkinField.March(modeling.PositionAttribute, cubersPerUnit, 0)
		log.Printf("time to march: %s", time.Since(marchStart))
		return mesh, nil
	})

	smoothedMeshNode := SmoothNormalsNode(
		LaplacianSmoothingNode(
			pumpkinMeshNode,
			&generator.ParameterNode[int]{
				Name:         "Smoothing Iterations",
				DefaultValue: 20,
			},
			&generator.ParameterNode[float64]{
				Name:         "Smoothing Factor",
				DefaultValue: .1,
			},
		),
	)

	// METHOD 1 ===============================================================
	// Works okay, issues from the dip of the top of the pumpkin causing the
	// texture to reverse directions
	// pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
	// newUVs := make([]vector2.Float64, pumpkinVerts.Len())
	// for i := 0; i < pumpkinVerts.Len(); i++ {
	// 	vert := pumpkinVerts.At(i)
	// 	xzPos := vert.XZ()
	// 	xzTheta := math.Atan2(xzPos.Y(), xzPos.X())
	// 	newUVs[i] = vector2.New(xzTheta/(math.Pi*2), vert.Y())
	// }

	// METHOD 2 ===============================================================
	return nodes.Transformer[nodes.Node[modeling.Mesh], modeling.Mesh](
		"Spherical UV Mapping",
		smoothedMeshNode,
		func(in nodes.Node[modeling.Mesh]) (modeling.Mesh, error) {
			mesh := in.Data()
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
		},
	)

}

func pumpkinStemMesh(topDip nodes.Node[float64]) nodes.Node[gltf.PolyformModel] {
	// maxWidth := stemParams.Float64("Base Width")
	// minWidth := stemParams.Float64("Tip Width")
	// length := stemParams.Float64("Length")
	// tipOffset := stemParams.Float64("Tip Offset")

	type StemParams struct {
		StemResolution nodes.Node[float64]
		TopDip         nodes.Node[float64]
	}

	params := StemParams{
		StemResolution: &generator.ParameterNode[float64]{
			Name:         "Stem Resolution",
			DefaultValue: 100,
		},
		TopDip: topDip,
	}

	return nodes.Transformer("Stem Mesh", params, func(in StemParams) (gltf.PolyformModel, error) {

		stemCanvas := marching.NewMarchingCanvas(in.StemResolution.Data())

		// stemCanvas.AddFieldParallel(marching.VarryingThicknessLine([]sdf.LinePoint{
		// 	{Point: vector3.New(0., 0., 0.), Radius: maxWidth},
		// 	{Point: vector3.New(0., length*.8, 0.), Radius: minWidth},
		// 	{Point: vector3.New(tipOffset, length, 0.), Radius: minWidth},
		// }, 1))

		sides := 6

		fields := make([]marching.Field, 0)
		angleInc := (math.Pi * 2.) / float64(sides)

		topPoint := 0.2

		fields = append(fields, marching.Line(
			vector3.New(0., 0.05, 0.),
			vector3.New(0., topPoint*.95, 0.),
			0.02,
			1,
		))

		for i := 0; i < sides; i++ {
			rot := quaternion.FromTheta(angleInc*float64(i), vector3.Up[float64]())

			rotatedPoints := rot.RotateArray([]vector3.Float64{
				vector3.New(.15, 0.08, -.025+(rand.Float64()*.05)),
				vector3.New(.05, 0.05, 0.),
				vector3.New(.03, topPoint, 0.),
			})

			fields = append(
				fields,
				marching.VarryingThicknessLine(
					[]sdf.LinePoint{
						{
							Point:  rotatedPoints[0],
							Radius: 0.01 + (rand.Float64() * 0.005),
						},
						{
							Point:  rotatedPoints[1],
							Radius: 0.02 + (rand.Float64() * 0.02),
						},
						{
							Point:  rotatedPoints[2],
							Radius: 0.02 + (rand.Float64() * 0.01),
						},
					},
					1,
				// start,
				// start.Add(vector3.Up[float64]().Scale(0.2)),
				// 0.03,
				// 1,
				),
			)
		}
		stemCanvas.AddFieldParallel(marching.CombineFields(fields...))

		mesh := stemCanvas.
			MarchParallel(0).
			Transform(
				meshops.LaplacianSmoothTransformer{
					Iterations:      20,
					SmoothingFactor: 0.1,
				},
				meshops.TranslateAttribute3DTransformer{
					Amount: vector3.New(0., 1-in.TopDip.Data()+0.055, 0.),
				},
				meshops.SmoothNormalsTransformer{},
			)

		pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
		newUVs := make([]vector2.Float64, pumpkinVerts.Len())
		center := vector3.New(0., 0.5, 0.)
		up := vector3.Up[float64]()
		for i := 0; i < pumpkinVerts.Len(); i++ {
			vert := pumpkinVerts.At(i)

			xzTheta := math.Atan2(vert.Z(), vert.X())
			xzTheta = math.Abs(xzTheta) // Avoid the UV seam

			dir := vert.Sub(center)
			angle := math.Acos(dir.Dot(up) / (dir.Length() * up.Length()))

			newUVs[i] = vector2.New(xzTheta/(math.Pi*2), angle)
		}

		return gltf.PolyformModel{
			Name: "Stem",
			Mesh: mesh.SetFloat2Attribute(modeling.TexCoordAttribute, newUVs),
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					BaseColorTexture: &gltf.PolyformTexture{
						URI: "Texturing/stem.png",
					},
					MetallicRoughnessTexture: &gltf.PolyformTexture{
						URI: "Texturing/stem-roughness.png",
					},
				},
				NormalTexture: &gltf.PolyformNormal{
					PolyformTexture: gltf.PolyformTexture{
						URI: "Texturing/stem-normal.png",
					},
				},
			},
		}, nil
	})

}

func imageToEdgeData(srcNode nodes.Node[image.Image], fillValueNode nodes.Node[float64]) nodes.Node[[][]float64] {

	type ImageToEdgeDataOarams struct {
		SrcImage  nodes.Node[image.Image]
		FillValue nodes.Node[float64]
	}

	imageToEdgeParams := ImageToEdgeDataOarams{
		SrcImage:  srcNode,
		FillValue: fillValueNode,
	}

	return nodes.Transformer("Edge Detection", imageToEdgeParams, func(in ImageToEdgeDataOarams) ([][]float64, error) {
		src := in.SrcImage.Data()
		imageData := make([][]float64, src.Bounds().Dx())
		fillValue := in.FillValue.Data()
		for i := 0; i < len(imageData); i++ {
			imageData[i] = make([]float64, src.Bounds().Dy())
		}

		texturing.Convolve(src, func(x, y int, kernel []color.Color) {
			if texturing.SimpleEdgeTest(kernel) {
				imageData[x][y] = 0
				return
			}

			if colors.RedEqual(kernel[4], 255) {
				imageData[x][y] = -fillValue
			} else {
				imageData[x][y] = fillValue
			}
		})

		return imageData, nil
	})
}

func loadImage(imageData []byte) (image.Image, error) {
	imgBuf := bytes.NewBuffer(imageData)
	img, _, err := image.Decode(imgBuf)
	return img, err
}

func heatPropegate(data nodes.Node[[][]float64], iterations nodes.Node[int], decay nodes.Node[float64]) nodes.Node[[][]float64] {

	type HeatPropogateParams struct {
		Data       nodes.Node[[][]float64]
		Iterations nodes.Node[int]
		Decay      nodes.Node[float64]
	}

	params := HeatPropogateParams{
		Data:       data,
		Iterations: iterations,
		Decay:      decay,
	}

	return nodes.Transformer("Heat Propogate", params, func(in HeatPropogateParams) ([][]float64, error) {

		originalData := in.Data.Data()
		iterations := in.Iterations.Data()
		decay := in.Decay.Data()

		data := make([][]float64, len(originalData))
		tempData := make([][]float64, len(data))
		for r := 0; r < len(tempData); r++ {
			data[r] = make([]float64, len(data[r]))
			for x, v := range originalData[r] {
				data[r][x] = v
			}
			tempData[r] = make([]float64, len(data[r]))
		}

		for i := 0; i < iterations; i++ {
			toConvole := data
			toStore := tempData
			if i%2 == 1 {
				toConvole = tempData
				toStore = data
			}
			texturing.ConvolveArray[float64](toConvole, func(x, y int, kernel []float64) {
				if toConvole[x][y] == 0 {
					return
				}
				total := kernel[0] + kernel[1] + kernel[2] + kernel[3] + kernel[5] + kernel[6] + kernel[7] + kernel[8]
				toStore[x][y] = (total / 8) * decay
			})
		}

		if iterations%2 == 1 {
			return tempData, nil
		}
		return data, nil
	})

}

func debugPropegation(data [][]float64, filename string) error {
	dst := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: len(data), Y: len(data[0])}})

	max := -math.MaxFloat64
	min := math.MaxFloat64
	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			max = math.Max(max, row[y])
			min = math.Min(min, row[y])
		}
	}

	delta := max - min

	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			val := row[y] / delta
			if val > 0 {
				dst.SetRGBA(x, y, color.RGBA{R: byte(val * 255), G: 0, B: 0, A: 255})
			} else {
				dst.SetRGBA(x, y, color.RGBA{R: 0, G: byte(val * -255), B: 0, A: 255})
			}
		}
	}

	imgFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer imgFile.Close()
	return png.Encode(imgFile, dst)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Returns +-1
func signNotZero(v vector2.Float64) vector2.Float64 {
	x := 1.0
	if v.X() < 0.0 {
		x = -1.0
	}

	y := 1.0
	if v.Y() < 0.0 {
		y = -1.0
	}

	return vector2.New(x, y)
}

func multVect(a, b vector2.Float64) vector2.Float64 {
	return vector2.New(
		a.X()*b.X(),
		a.Y()*b.Y(),
	)
}

// FromOctUV converts a 2D octahedron UV coordinate to a point on a 3D sphere.
func FromOctUV(e vector2.Float64) vector3.Float64 {
	// vec3 v = vec3(e.xy, 1.0 - abs(e.x) - abs(e.y));
	v := vector3.New(e.X(), e.Y(), 1.0-math.Abs(e.X())-math.Abs(e.Y()))

	// if (v.z < 0) v.xy = (1.0 - abs(v.yx)) * signNotZero(v.xy);
	if v.Z() < 0 {
		n := multVect(vector2.New(1.0-math.Abs(v.Y()), 1.0-math.Abs(v.X())), signNotZero(vector2.New(v.X(), v.Y())))
		v = v.SetX(n.X()).SetY(n.Y())
	}

	return v.Normalized()
}

//go:embed face.png
var facePNG []byte

func ImageArtifactNode(imageNode nodes.Node[image.Image]) nodes.Node[generator.Artifact] {
	return nodes.Transformer("Image Artifact", imageNode, func(i nodes.Node[image.Image]) (generator.Artifact, error) {
		return &generator.ImageArtifact{Image: i.Data()}, nil
	})
}

type StemRoughness struct {
	Dimensions nodes.Node[int]
	Roughness  nodes.Node[float64]
}

func (sr StemRoughness) Process() (generator.Artifact, error) {
	dim := sr.Dimensions.Data()
	stemRoughnessImage := image.NewRGBA(image.Rect(0, 0, dim, dim))

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			stemRoughnessImage.Set(x, y, color.RGBA{
				R: 0, // byte(len * 255),
				G: byte(255 * sr.Roughness.Data()),
				B: 0,
				A: 255,
			})
		}
	}

	return &generator.ImageArtifact{Image: stemRoughnessImage}, nil
}

func main() {
	maxHeatNode := &generator.ParameterNode[float64]{
		Name:         "Max Heat",
		DefaultValue: 100.,
	}
	img, err := loadImage(facePNG)
	check(err)

	imgData := heatPropegate(
		imageToEdgeData(nodes.Input(img), maxHeatNode),
		&generator.ParameterNode[int]{
			Name:         "Iterations",
			DefaultValue: 250,
		},
		&generator.ParameterNode[float64]{
			Name:         "Decay",
			DefaultValue: 0.9999,
		},
	)
	// check(debugPropegation(imgData, "debug.png"))

	topDip := &generator.ParameterNode[float64]{
		Name:         "Top Dip",
		DefaultValue: .2,
	}

	pumpkinMesh := newPumpkinMesh(
		imgData,
		topDip,
	)

	type PumpkinGlbArtifactParams struct {
		PumpkinBody nodes.Node[modeling.Mesh]
		PumpkinStem nodes.Node[gltf.PolyformModel]
		LightColor  nodes.Node[color.Color]
	}

	pumpkinGlbArtifactParams := PumpkinGlbArtifactParams{
		PumpkinBody: pumpkinMesh,
		LightColor: &generator.ParameterNode[color.Color]{
			Name:         "Light Color",
			DefaultValue: coloring.WebColor{R: 0xf4, G: 0xf5, B: 0xad, A: 255},
		},
		PumpkinStem: pumpkinStemMesh(topDip),
	}

	pumpkinGlbArtifact := nodes.Transformer(
		"Pumpkin Glb Scene",
		pumpkinGlbArtifactParams,
		func(in PumpkinGlbArtifactParams) (generator.Artifact, error) {
			return &generator.GltfArtifact{
				Scene: gltf.PolyformScene{
					Models: []gltf.PolyformModel{
						{
							Name: "Pumpkin",
							Mesh: in.PumpkinBody.Data(),
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
						in.PumpkinStem.Data(),
					},
					Lights: []gltf.KHR_LightsPunctual{
						{
							Type:     gltf.KHR_LightsPunctualType_Point,
							Position: vector3.New(0., 0.5, 0.),
							Color:    in.LightColor.Data(),
						},
					},
				},
			}, nil
		},
	)

	textureDimensions := &generator.ParameterNode[int]{
		Name:         "Texture Dimension",
		DefaultValue: 10124,
	}

	app := generator.App{
		Name:        "Pumpkin",
		Version:     "1.0.0",
		Description: "Making a pumpkin for Haloween",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
				ContactInfo: []generator.AuthorContact{
					{
						Medium: "Twitter",
						Value:  "@EliCDavis",
					},
				},
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Near:  2,
				Far:   10,
				Color: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x4f, G: 0x6d, B: 0x55, A: 255},
			Background: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Producers: map[string]nodes.Node[generator.Artifact]{
			"pumpkin.glb": pumpkinGlbArtifact,
			"Texturing/pumpkin.png": ImageArtifactNode(nodes.Struct(&Albedo{
				Positive: &generator.ParameterNode[color.Color]{
					Name:         "Base Color",
					DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
				},
				Negative: &generator.ParameterNode[color.Color]{
					Name:         "Negative Color",
					DefaultValue: coloring.WebColor{R: 0xf7, G: 0x71, B: 0x02, A: 255},
				},
			})),
			"Texturing/stem.png": ImageArtifactNode(nodes.Struct(&Albedo{
				Positive: &generator.ParameterNode[color.Color]{
					Name:         "Stem Base Color",
					DefaultValue: coloring.WebColor{R: 0xce, G: 0xa2, B: 0x7e, A: 255},
				},
				Negative: &generator.ParameterNode[color.Color]{
					Name:         "Stem Negative Color",
					DefaultValue: coloring.WebColor{R: 0x7d, G: 0x53, B: 0x2c, A: 255},
				},
			})),
			"Texturing/normal.png": nodes.Struct(&NormalImage{
				NumberOfLines: &generator.ParameterNode[int]{
					Name:         "Number of Lines",
					DefaultValue: 20,
				},
				NumberOfWarts: &generator.ParameterNode[int]{
					Name:         "Number of Warts",
					DefaultValue: 50,
				},
			}),
			"Texturing/stem-normal.png": nodes.Struct(&StemNormalImage{
				NumberOfLines: &generator.ParameterNode[int]{
					Name:         "Stem Normal Line Count",
					DefaultValue: 30,
				},
			}),
			"Texturing/roughness.png": nodes.Struct(MetalRoughness{
				Roughness: &generator.ParameterNode[float64]{
					Name:         "Pumpkin Roughness",
					DefaultValue: 0.75,
				},
			}),
			"Texturing/stem-roughness.png": nodes.Struct(StemRoughness{
				Dimensions: textureDimensions,
				Roughness:  nodes.Input(0.78),
			}),
			// "uvMap.png": nodes.InputFromFunc(func() generator.Artifact {
			// 	img := texturing.DebugUVTexture{
			// 		ImageResolution:      1024,
			// 		BoardResolution:      10,
			// 		NegativeCheckerColor: color.RGBA{0, 0, 0, 255},

			// 		PositiveCheckerColor: color.RGBA{255, 0, 0, 255},
			// 		XColorScale:          color.RGBA{0, 255, 0, 255},
			// 		YColorScale:          color.RGBA{0, 0, 255, 255},
			// 	}.Image()
			// 	return &generator.ImageArtifact{Image: img}
			// }),
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
