package main

import (
	"image/color"
	"io/ioutil"
	"log"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func run() error {
	textToWrite := "Test"
	fontByteData, err := ioutil.ReadFile("./FromCartoonBlocks.ttf")

	if err != nil {
		return err
	}

	parsedFont, err := truetype.Parse(fontByteData)

	if err != nil {
		return err
	}

	characterFields := make([]marching.Field, 0)

	glyph := truetype.GlyphBuf{}

	for charIndex, char := range textToWrite {

		glyph.Load(parsedFont, 100, parsedFont.Index(char), font.HintingNone)

		startPoint := glyph.Points[0]
		letterPoints := make([]vector.Vector3, 0)
		nextEndpoint := 0
		for i, p := range glyph.Points {

			letterPoints = append(letterPoints, vector.NewVector3(float64(p.X)+(float64(charIndex)*60.), 0, float64(p.Y)))

			if nextEndpoint < len(glyph.Ends) && i == glyph.Ends[nextEndpoint]-1 {
				letterPoints = append(letterPoints, vector.NewVector3(float64(startPoint.X)+(float64(charIndex)*60.), 0, float64(startPoint.Y)))
				characterFields = append(characterFields, marching.MultiSegmentLine(letterPoints, .5, 2))
				letterPoints = make([]vector.Vector3, 0)
				if nextEndpoint < len(glyph.Ends)-1 {
					startPoint = glyph.Points[i+1]
					nextEndpoint++
				}
			}
		}

		if len(letterPoints) > 2 {
			characterFields = append(characterFields, marching.MultiSegmentLine(letterPoints, .1, 3))
		}
	}

	finalWords := marching.CombineFields(characterFields...)

	start := time.Now()
	mesh := finalWords.March(modeling.PositionAttribute, 1, .0).
		Scale(vector.Vector3Zero(), vector.NewVector3(1, 5, 1)).
		SmoothLaplacian(20, .1).
		CalculateSmoothNormals().
		SetMaterial(modeling.Material{
			Name:         "Text",
			DiffuseColor: color.RGBA{R: 90, G: 218, B: 255, A: 255},
		})
	log.Printf("time to compute: %s", time.Now().Sub(start))

	return obj.Save("tmp/text/text.obj", mesh)
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
