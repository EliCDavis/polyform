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

func metalRoughness() image.Image {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			p := (val * 128) + 128

			p = 255 - (p * 0.75)

			img.Set(x, y, color.RGBA{
				R: 0, // byte(len * 255),
				G: byte(p),
				B: 0,
				A: 255,
			})
		}
	}
	return img
}

func albedo(positiveColor, negativeColor color.Color) image.Image {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	nR, nG, nB, _ := negativeColor.RGBA()
	pR, pG, pB, _ := positiveColor.RGBA()

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
	return img
}

func stemNormalImage() image.Image {
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

	numLines := 30

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

	return img
}

func normalImage() image.Image {
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

	numLines := 20

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

	numWarts := 50
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
	return texturing.BoxBlurNTimes(img, 10)
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

func newPumpkinMesh(
	cubersPerUnit float64,
	maxWidth, topDip, distanceFromCenter, wedgeLineRadius float64,
	sides int,
	imageField [][]float64,
	useImageField bool,
) modeling.Mesh {
	canvas := marching.NewMarchingCanvas(cubersPerUnit)

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

	mesh = mesh.Transform(
		meshops.LaplacianSmoothTransformer{
			Iterations:      20,
			SmoothingFactor: 0.1,
		},
		meshops.SmoothNormalsTransformer{},
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
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, newUVs)
}

func pumpkinStemMesh(stemParams *generator.GroupParameter, topDip float64) gltf.PolyformModel {
	// maxWidth := stemParams.Float64("Base Width")
	// minWidth := stemParams.Float64("Tip Width")
	// length := stemParams.Float64("Length")
	// tipOffset := stemParams.Float64("Tip Offset")

	stemCanvas := marching.NewMarchingCanvas(stemParams.Float64("Cubes Per Unit"))

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
				Amount: vector3.New(0., 1-topDip+0.055, 0.),
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
	}
}

func imageToEdgeData(src image.Image, fillValue float64) [][]float64 {
	imageData := make([][]float64, src.Bounds().Dx())
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

	return imageData
}

func loadImage(imageData []byte) (image.Image, error) {
	imgBuf := bytes.NewBuffer(imageData)
	img, _, err := image.Decode(imgBuf)
	return img, err
}

func heatPropegate(data [][]float64, iterations int, decay float64) [][]float64 {
	tempData := make([][]float64, len(data))
	for r := 0; r < len(tempData); r++ {
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
		return tempData
	}
	return data
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

func main() {

	maxHeat := 100.
	img, err := loadImage(facePNG)
	check(err)
	imgData := imageToEdgeData(img, maxHeat)
	imgData = heatPropegate(imgData, 250, 0.9999)
	check(debugPropegation(imgData, "debug.png"))

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
		Generator: &generator.Generator{
			SubGenerators: map[string]*generator.Generator{
				"Texturing": {
					Parameters: &generator.GroupParameter{
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Base Color",
								DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
							},
							&generator.ColorParameter{
								Name:         "Negative Color",
								DefaultValue: coloring.WebColor{R: 0xf7, G: 0x71, B: 0x02, A: 255},
							},

							&generator.ColorParameter{
								Name:         "Stem Base Color",
								DefaultValue: coloring.WebColor{R: 0xce, G: 0xa2, B: 0x7e, A: 255},
							},
							&generator.ColorParameter{
								Name:         "Stem Negative Color",
								DefaultValue: coloring.WebColor{R: 0x7d, G: 0x53, B: 0x2c, A: 255},
							},
						},
					},
					Producers: map[string]generator.Producer{
						"pumpkin.png": func(c *generator.Context) (generator.Artifact, error) {
							return generator.ImageArtifact{
								Image: albedo(
									c.Parameters.Color("Base Color"),
									c.Parameters.Color("Negative Color"),
								),
							}, nil
						},
						"stem.png": func(c *generator.Context) (generator.Artifact, error) {
							return generator.ImageArtifact{
								Image: albedo(
									c.Parameters.Color("Stem Base Color"),
									c.Parameters.Color("Stem Negative Color"),
								),
							}, nil
						},
						"normal.png": func(c *generator.Context) (generator.Artifact, error) {
							return &generator.ImageArtifact{Image: normalImage()}, nil
						},
						"stem-normal.png": func(c *generator.Context) (generator.Artifact, error) {
							return &generator.ImageArtifact{Image: stemNormalImage()}, nil
						},
						"roughness.png": func(c *generator.Context) (generator.Artifact, error) {
							return &generator.ImageArtifact{Image: metalRoughness()}, nil
						},
						"stem-roughness.png": func(c *generator.Context) (generator.Artifact, error) {
							dim := 1024
							img := image.NewRGBA(image.Rect(0, 0, dim, dim))
							// normals.Fill(img)

							for x := 0; x < dim; x++ {
								for y := 0; y < dim; y++ {
									img.Set(x, y, color.RGBA{
										R: 0, // byte(len * 255),
										G: byte(200),
										B: 0,
										A: 255,
									})
								}
							}

							return &generator.ImageArtifact{Image: img}, nil
						},
					},
				},
			},
			Parameters: &generator.GroupParameter{
				Name: "Pumpkin",
				Parameters: []generator.Parameter{
					&generator.FloatParameter{
						Name:         "Pumpkin Cubes Per Unit",
						DefaultValue: 20,
					},

					&generator.IntParameter{
						Name:         "Wedges",
						DefaultValue: 10,
					},

					&generator.FloatParameter{
						Name:         "Wedge Spacing",
						DefaultValue: .1,
					},

					&generator.FloatParameter{
						Name:         "Wedge Radius",
						DefaultValue: .3,
					},

					&generator.FloatParameter{
						Name:         "Max Width",
						DefaultValue: .3,
					},

					&generator.FloatParameter{
						Name:         "Top Dip",
						DefaultValue: .2,
					},

					&generator.ColorParameter{
						Name:         "Light Color",
						DefaultValue: coloring.WebColor{R: 0xf4, G: 0xf5, B: 0xad, A: 255},
					},

					&generator.BoolParameter{
						Name:         "Carve",
						DefaultValue: true,
					},

					&generator.GroupParameter{
						Name: "Stem",
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Color",
								DefaultValue: coloring.WebColor{R: 0x6d, G: 0x52, B: 0x40, A: 255},
							},
							&generator.FloatParameter{
								Name:         "Base Width",
								DefaultValue: 0.07,
							},
							&generator.FloatParameter{
								Name:         "Tip Width",
								DefaultValue: 0.03,
							},
							&generator.FloatParameter{
								Name:         "Length",
								DefaultValue: 0.3,
							},
							&generator.FloatParameter{
								Name:         "Tip Offset",
								DefaultValue: 0.1,
							},
							&generator.FloatParameter{
								Name:         "Cubes Per Unit",
								DefaultValue: 100,
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"perlin.png": func(c *generator.Context) (generator.Artifact, error) {
					// dim := 128
					dim := 1024
					img := image.NewRGBA(image.Rect(0, 0, dim, dim))

					n := noise.NewTilingNoise(dim, 1/64., 5)

					for x := 0; x < dim; x++ {
						// xDim := (float64(x) / float64(dim)) * 2
						// xRot := xDim * math.Pi * 2.

						for y := 0; y < dim; y++ {
							// yDim := (float64(y) / float64(dim)) * 2
							// yRot := yDim * math.Pi * 2.

							// p := noise.Perlin3D(vector3.New(x, y, 0).ToFloat64().Scale(1./128.).Scale(4)) * 255

							// A regular doughnut
							// xDir := vector3.New(math.Cos(xRot), math.Sin(xRot), 0).
							// 	Scale(2).
							// 	Add(vector3.New(1., 0., 0.).Scale(8))
							// final := quaternion.FromTheta(yRot, vector3.Up[float64]()).
							// 	Rotate(xDir)

							// A regular sphere
							// rot1 := quaternion.FromTheta(yRot, vector3.Up[float64]())
							// rot2 := quaternion.FromTheta(xRot, vector3.Forward[float64]())
							// final := rot1.Rotate(rot2.Rotate(vector3.Right[float64]()))

							// Normal Mapping method - FAILURE
							// if xDim >= 1 {
							// 	xDim -= 1
							// }
							// if yDim >= 1 {
							// 	yDim -= 1
							// }
							// final := FromOctUV(vector2.New((xDim*2)-1, (yDim*2)-1))

							// A wiggly donut
							// xDir := vector3.New(math.Cos(xRot), math.Sin(xRot), 0).
							// 	Scale(1).
							// 	Add(vector3.New(1., 0., 0.).Scale(2))

							// len := (xDir.X() - 1) / 2
							// xDir = xDir.SetY(xDir.Y() + (1 - math.Pow((1-(len)), 4)*math.Cos(yRot)*3.6))
							// final := quaternion.FromTheta(yRot, vector3.Up[float64]()).
							// 	Rotate(xDir)

							// A dumb Doughnut
							// xDir := vector3.New(math.Cos(yRot), math.Sin(xRot), 0).
							// 	Scale(5).
							// 	Add(vector3.New(1., 0., 0.).Scale(7))
							// final := vector3.New(math.Cos(xRot), math.Sin(yRot), 0).
							// 	Add(xDir)

							// A spinny doughnut
							// rot := quaternion.FromTheta(yRot, vector3.Up[float64]())
							// xDir := rot.Rotate(vector3.New(math.Cos(xRot), math.Sin(xRot), 0).
							// 	Scale(1)).
							// 	Add(vector3.New(1., 0., 0.).Scale(1))
							// final := rot.Rotate(xDir)

							// p := noise.Perlin3D(final.Scale(.8)) * 255

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

					// normals.FromHeightmap(img)
					// return &generator.ImageArtifact{Image: img}, nil
					return &generator.ImageArtifact{Image: texturing.ToNormal(img)}, nil
				},

				"uvMap.png": func(c *generator.Context) (generator.Artifact, error) {
					img := texturing.DebugUVTexture{
						ImageResolution:      1024,
						BoardResolution:      10,
						NegativeCheckerColor: color.RGBA{0, 0, 0, 255},

						PositiveCheckerColor: color.RGBA{255, 0, 0, 255},
						XColorScale:          color.RGBA{0, 255, 0, 255},
						YColorScale:          color.RGBA{0, 0, 255, 255},
					}.Image()
					return &generator.ImageArtifact{Image: img}, nil
				},
				"pumpkin.glb": func(c *generator.Context) (generator.Artifact, error) {

					pumpkinMesh := newPumpkinMesh(
						c.Parameters.Float64("Pumpkin Cubes Per Unit"),
						c.Parameters.Float64("Max Width"),
						c.Parameters.Float64("Top Dip"),
						c.Parameters.Float64("Wedge Spacing"),
						c.Parameters.Float64("Wedge Radius"),
						c.Parameters.Int("Wedges"),
						imgData,
						c.Parameters.Bool("Carve"),
					)

					stem := pumpkinStemMesh(c.Parameters.Group("Stem"), c.Parameters.Float64("Top Dip"))

					// texturingParams := c.Parameters.Group("Texturing")

					// outerIndices := make([]int, 0)
					// innerIndices := make([]int, 0)
					// pumpkinMesh.ScanFloat1Attribute(modeling.PositionAttribute+"-winner", func(i int, v vector3.Float64) {
					// 	if v == -1 {

					// 	}
					// })

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "Pumpkin",
									Mesh: pumpkinMesh,
									Material: &gltf.PolyformMaterial{
										// PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
										// 	BaseColorTexture: &gltf.PolyformTexture{
										// 		URI: "Texturing/pumpkin.png", //"uvMap.png",
										// 		// URI: "uvMap.png", //"uvMap.png",
										// 		Sampler: &gltf.Sampler{
										// 			WrapS: gltf.SamplerWrap_REPEAT,
										// 			WrapT: gltf.SamplerWrap_REPEAT,
										// 		},
										// 	},
										// 	MetallicRoughnessTexture: &gltf.PolyformTexture{
										// 		URI: "Texturing/roughness.png",
										// 	},
										// 	// BaseColorFactor: texturingParams.Color("Base Color"),
										// 	// MetallicFactor:  1,
										// 	// RoughnessFactor: 0,
										// },
										// NormalTexture: &gltf.PolyformNormal{
										// 	PolyformTexture: gltf.PolyformTexture{
										// 		URI: "Texturing/normal.png",
										// 	},
										// },
										Extensions: []gltf.MaterialExtension{
											// gltf.PolyformMaterialsUnlit{},
										},
									},
								},
								stem,
							},
							Lights: []gltf.KHR_LightsPunctual{
								{
									Type:     gltf.KHR_LightsPunctualType_Point,
									Position: vector3.New(0., 0.5, 0.),
									Color:    c.Parameters.Color("Light Color"),
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
