package texturing

import (
	"image"
	"image/color"
)

//
// [0][1][2]
// [3][4][5]
// [6][7][8]
//
//
func Convolve(src image.Image, f func(x, y int, values []color.Color)) {
	for xIndex := 0; xIndex < src.Bounds().Dx(); xIndex++ {
		xLeft := xIndex - 1
		if xIndex == 0 {
			xLeft = xIndex + 1
		}
		xMid := xIndex
		xRight := xIndex + 1
		if xIndex == src.Bounds().Dx()-1 {
			xRight = xIndex - 1
		}

		for yIndex := 0; yIndex < src.Bounds().Dy(); yIndex++ {
			yBot := yIndex - 1
			if yIndex == 0 {
				yBot = yIndex + 1
			}
			yMid := yIndex
			yTop := yIndex + 1
			if yIndex == src.Bounds().Dx()-1 {
				yTop = yIndex - 1
			}

			f(xIndex, yIndex, []color.Color{
				src.At(xLeft, yTop), src.At(xMid, yTop), src.At(xRight, yTop),
				src.At(xLeft, yMid), src.At(xMid, yMid), src.At(xRight, yMid),
				src.At(xLeft, yBot), src.At(xMid, yBot), src.At(xRight, yBot),
			})
		}
	}
}
