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
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

// perm = range(256)
// random.shuffle(perm)
// perm += perm
// dirs = [(math.cos(a * 2.0 * math.pi / 256),
//          math.sin(a * 2.0 * math.pi / 256))
//          for a in range(256)]

type TilingNoise struct {
	dirs []vector2.Float64
	perm []int
}

func (tn *TilingNoise) init() {
	size := 256

	tn.perm = make([]int, size)
	for i := 0; i < size; i++ {
		tn.perm[i] = i
	}
	rand.Shuffle(len(tn.perm), func(i, j int) { tn.perm[i], tn.perm[j] = tn.perm[j], tn.perm[i] })
	tn.perm = append(tn.perm, tn.perm...)

	tn.dirs = make([]vector2.Float64, size)
	for i := 0; i < size; i++ {
		a := float64(i)
		tn.dirs[i] = vector2.New(
			math.Cos((a*2.*math.Pi)/float64(size)),
			math.Sin((a*2.*math.Pi)/float64(size)),
		)
	}
}

// def noise(x, y, per):
//     def surflet(gridX, gridY):
//         distX, distY = abs(x-gridX), abs(y-gridY)
//         polyX = 1 - 6*distX**5 + 15*distX**4 - 10*distX**3
//         polyY = 1 - 6*distY**5 + 15*distY**4 - 10*distY**3
//         hashed = perm[perm[int(gridX)%per] + int(gridY)%per]
//         grad = (x-gridX)*dirs[hashed][0] + (y-gridY)*dirs[hashed][1]
//         return polyX * polyY * grad
//     intX, intY = int(x), int(y)
//     return (surflet(intX+0, intY+0) + surflet(intX+1, intY+0) +
//             surflet(intX+0, intY+1) + surflet(intX+1, intY+1))

// https://gamedev.stackexchange.com/questions/23625/how-do-you-generate-tileable-perlin-noise
func (tn *TilingNoise) surflet(v vector2.Float64, g vector2.Int, per int) float64 {
	dist := v.Sub(g.ToFloat64()).Abs()
	polyX := 1 - (6 * math.Pow(dist.X(), 5)) + (15 * math.Pow(dist.X(), 4)) - (10 * math.Pow(dist.X(), 3))
	polyY := 1 - (6 * math.Pow(dist.Y(), 5)) + (15 * math.Pow(dist.Y(), 4)) - (10 * math.Pow(dist.Y(), 3))

	hashed := tn.perm[tn.perm[g.X()%per]+(g.Y()%per)]

	hashedDir := tn.dirs[hashed]
	grad := ((v.X() - float64(g.X())) * hashedDir.X()) + ((v.Y() - float64(g.Y())) * hashedDir.Y())
	return polyX * polyY * grad
}

func (tn *TilingNoise) Noise(v vector2.Float64, per int) float64 {
	i := v.FloorToInt()
	return tn.surflet(v, i, per) +
		tn.surflet(v, i.Add(vector2.Right[int]()), per) +
		tn.surflet(v, i.Add(vector2.Up[int]()), per) +
		tn.surflet(v, i.Add(vector2.One[int]()), per)
}

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

	n := &TilingNoise{}
	n.init()

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := 0.
			freq := 1. / 64.
			for o := 0; o < 5; o++ {
				op2 := math.Pow(2, float64(o))
				n := n.Noise(
					vector2.New(
						(float64(x)*freq)*op2,
						(float64(y)*freq)*op2,
					),
					int(float64(dim)*freq)*int(op2),
				)
				val += math.Pow(0.5, float64(o)) * n
			}
			// p := n.Noise(vector2.New(xDim*10, yDim*10), 100)
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

func normalImage() image.Image {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := &TilingNoise{}
	n.init()

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := 0.
			freq := 1. / 64.
			for o := 0; o < 5; o++ {
				op2 := math.Pow(2, float64(o))
				n := n.Noise(
					vector2.New(
						(float64(x)*freq)*op2,
						(float64(y)*freq)*op2,
					),
					int(float64(dim)*freq)*int(op2),
				)
				val += math.Pow(0.5, float64(o)) * n
			}
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
		rot := modeling.UnitQuaternionFromTheta(angleInc*float64(i), vector3.Up[float64]())

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
	log.Printf("time to add field: %s", time.Since(addFieldStart))

	marchStart := time.Now()
	log.Println("starting march...")
	mesh := canvas.MarchParallel(0)
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

func pumpkinStem(maxWidth, minWidth, length, tipOffset float64) marching.Field {
	return marching.VarryingThicknessLine([]sdf.LinePoint{
		{Point: vector3.New(0., 0., 0.), Radius: maxWidth},
		{Point: vector3.New(0., length*.8, 0.), Radius: minWidth},
		{Point: vector3.New(tipOffset, length, 0.), Radius: minWidth},
	}, 1)
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

func loadImageFromPath(imageName string) (image.Image, error) {
	logoFile, err := os.Open(imageName)
	if err != nil {
		return nil, err
	}
	defer logoFile.Close()

	img, _, err := image.Decode(logoFile)

	return img, err
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
		Version:     "0.0.1",
		Description: "Making a pumpkin for Haloween",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
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
								Name:         "Line Color",
								DefaultValue: coloring.WebColor{R: 0, G: 0x81, B: 0x1f, A: 255},
							},
							&generator.IntParameter{
								Name:         "Lines",
								DefaultValue: 8,
							},
						},
					},
					Producers: map[string]generator.Producer{
						"pumpkin.png": func(c *generator.Context) (generator.Artifact, error) {
							const texDimension = 1024

							ctx := gg.NewContext(texDimension, texDimension)
							ctx.SetColor(c.Parameters.Color("Base Color"))

							ctx.DrawRectangle(0, 0, texDimension, texDimension)
							ctx.Fill()

							// lines := c.Parameters.Int("Lines")

							// ctx.SetColor(c.Parameters.Color("Line Color"))
							// ctx.SetLineWidth(2)
							// spacing := texDimension / (lines)
							// for i := 0; i < lines; i++ {
							// 	xDim := float64((spacing / 2) + (spacing * i))
							// 	ctx.DrawLine(xDim, 0, xDim, texDimension)
							// 	ctx.Stroke()
							// }

							return generator.ImageArtifact{
								Image: ctx.Image(),
							}, nil
						},
					},
				},
			},
			Parameters: &generator.GroupParameter{
				Name: "Pumpkin",
				Parameters: []generator.Parameter{
					&generator.FloatParameter{
						Name:         "Cubes Per Unit",
						DefaultValue: 40,
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
						},
					},
				},
			},
			Producers: map[string]generator.Producer{
				"perlin.png": func(c *generator.Context) (generator.Artifact, error) {
					// dim := 128
					dim := 1024
					img := image.NewRGBA(image.Rect(0, 0, dim, dim))

					n := &TilingNoise{}
					n.init()

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
							// final := modeling.UnitQuaternionFromTheta(yRot, vector3.Up[float64]()).
							// 	Rotate(xDir)

							// A regular sphere
							// rot1 := modeling.UnitQuaternionFromTheta(yRot, vector3.Up[float64]())
							// rot2 := modeling.UnitQuaternionFromTheta(xRot, vector3.Forward[float64]())
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
							// final := modeling.UnitQuaternionFromTheta(yRot, vector3.Up[float64]()).
							// 	Rotate(xDir)

							// A dumb Doughnut
							// xDir := vector3.New(math.Cos(yRot), math.Sin(xRot), 0).
							// 	Scale(5).
							// 	Add(vector3.New(1., 0., 0.).Scale(7))
							// final := vector3.New(math.Cos(xRot), math.Sin(yRot), 0).
							// 	Add(xDir)

							// A spinny doughnut
							// rot := modeling.UnitQuaternionFromTheta(yRot, vector3.Up[float64]())
							// xDir := rot.Rotate(vector3.New(math.Cos(xRot), math.Sin(xRot), 0).
							// 	Scale(1)).
							// 	Add(vector3.New(1., 0., 0.).Scale(1))
							// final := rot.Rotate(xDir)

							// p := noise.Perlin3D(final.Scale(.8)) * 255

							val := 0.
							freq := 1. / 64.
							for o := 0; o < 5; o++ {
								op2 := math.Pow(2, float64(o))
								n := n.Noise(
									vector2.New(
										(float64(x)*freq)*op2,
										(float64(y)*freq)*op2,
									),
									int(float64(dim)*freq)*int(op2),
								)
								val += math.Pow(0.5, float64(o)) * n
							}
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
				"normal.png": func(c *generator.Context) (generator.Artifact, error) {
					return &generator.ImageArtifact{Image: normalImage()}, nil
				},
				"roughness.png": func(c *generator.Context) (generator.Artifact, error) {
					return &generator.ImageArtifact{Image: metalRoughness()}, nil
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
						c.Parameters.Float64("Cubes Per Unit"),
						c.Parameters.Float64("Max Width"),
						c.Parameters.Float64("Top Dip"),
						c.Parameters.Float64("Wedge Spacing"),
						c.Parameters.Float64("Wedge Radius"),
						c.Parameters.Int("Wedges"),
						imgData,
						c.Parameters.Bool("Carve"),
					)

					stemParams := c.Parameters.Group("Stem")
					stemCanvas := marching.NewMarchingCanvas(c.Parameters.Float64("Cubes Per Unit"))
					stemCanvas.AddFieldParallel(pumpkinStem(
						stemParams.Float64("Base Width"),
						stemParams.Float64("Tip Width"),
						stemParams.Float64("Length"),
						stemParams.Float64("Tip Offset"),
					))
					stem := stemCanvas.
						MarchParallel(0).
						Transform(
							meshops.LaplacianSmoothTransformer{
								Iterations:      20,
								SmoothingFactor: 0.1,
							},
							meshops.TranslateAttribute3DTransformer{
								Amount: vector3.New(0., 1-c.Parameters.Float64("Top Dip"), 0.),
							},
						)

					// texturingParams := c.Parameters.Group("Texturing")

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "Pumpkin",
									Mesh: pumpkinMesh,
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
												URI: "roughness.png",
											},
											// BaseColorFactor: texturingParams.Color("Base Color"),
											// MetallicFactor:  1,
											// RoughnessFactor: 0,
										},
										NormalTexture: &gltf.PolyformNormal{
											PolyformTexture: gltf.PolyformTexture{
												URI: "normal.png",
											},
										},
										Extensions: []gltf.MaterialExtension{
											// gltf.PolyformMaterialsUnlit{},
										},
									},
								},
								{
									Name: "Stem",
									Mesh: stem,
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorFactor: stemParams.Color("Color"),
										},
									},
								},
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