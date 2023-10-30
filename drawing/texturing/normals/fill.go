package normals

import (
	"image"
	"image/color"
	"image/draw"
)

func Fill(img draw.Image) {
	blue := color.RGBA{128, 128, 255, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.Point{}, draw.Src)
}
