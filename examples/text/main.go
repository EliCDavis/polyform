package main

import (
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"time"
	"unicode"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func run() error {
	textToWrite := "Polyform"
	characterSpacing := 10.
	resolution := .5
	lineRadius := 2.5
	fontFile := "./data/fonts/Current-Black.ttf"

	fontByteData, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return err
	}

	parsedFont, err := truetype.Parse(fontByteData)
	if err != nil {
		return err
	}

	characterFields := make([]marching.Field, 0)

	glyph := truetype.GlyphBuf{}

	totalWidth := 0.
	for _, char := range textToWrite {

		if unicode.IsSpace(char) {
			totalWidth += 20 + characterSpacing
			continue
		}

		glyph.Load(parsedFont, 100, parsedFont.Index(char), font.HintingNone)

		startPoint := glyph.Points[0]
		letterPoints := make([]vector3.Float64, 0)
		nextEndpoint := 0
		width := 0.
		for i, p := range glyph.Points {
			width = math.Max(width, float64(p.X))

			letterPoints = append(letterPoints, vector3.New(float64(p.X), float64(p.Y), 0))

			if nextEndpoint < len(glyph.Ends) && i == glyph.Ends[nextEndpoint]-1 {
				letterPoints = append(letterPoints, vector3.New(float64(startPoint.X), float64(startPoint.Y), 0))
				characterField := marching.
					MultiSegmentLine(letterPoints, lineRadius, 2).
					Translate(vector3.New(totalWidth, 0, 0))

				characterFields = append(characterFields, characterField)
				letterPoints = make([]vector3.Float64, 0)
				if nextEndpoint < len(glyph.Ends)-1 {
					startPoint = glyph.Points[i+1]
					nextEndpoint++
				}
			}
		}

		totalWidth += width + characterSpacing
	}

	finalWords := marching.CombineFields(characterFields...)

	start := time.Now()
	mesh := finalWords.March(modeling.PositionAttribute, resolution, .0).
		// Scale(vector3.New(1, 5, 1)).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      20,
				SmoothingFactor: .1,
			},
			meshops.SmoothNormalsTransformer{},
		)
	log.Printf("time to compute: %s", time.Since(start))

	return obj.Save("tmp/text/text.obj", obj.Scene{
		Objects: []obj.Object{
			{
				Entries: []obj.Entry{
					{
						Mesh: mesh,
						Material: &obj.Material{
							Name:         "Text",
							DiffuseColor: color.RGBA{R: 90, G: 218, B: 255, A: 255},
						},
					},
				},
			},
		},
	})
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
