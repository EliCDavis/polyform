package texturing

import (
	"image"
	"image/color"
)

// [0][1][2]
// [3][4][5]
// [6][7][8]
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
			if yIndex == src.Bounds().Dy()-1 {
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

func ConvolveArray[T any](arr [][]T, f func(x, y int, values []T)) {
	dx := len(arr)
	dy := len(arr[0])

	kernel := make([]T, 9)

	for xIndex := 0; xIndex < dx; xIndex++ {
		xLeft := xIndex - 1
		if xIndex == 0 {
			xLeft = xIndex + 1
		}
		xMid := xIndex
		xRight := xIndex + 1
		if xIndex == dx-1 {
			xRight = xIndex - 1
		}

		for yIndex := 0; yIndex < dy; yIndex++ {
			yBot := yIndex - 1
			if yIndex == 0 {
				yBot = yIndex + 1
			}
			yMid := yIndex
			yTop := yIndex + 1
			if yIndex == dy-1 {
				yTop = yIndex - 1
			}

			kernel[0] = arr[xLeft][yTop]
			kernel[1] = arr[xMid][yTop]
			kernel[2] = arr[xRight][yTop]
			kernel[3] = arr[xLeft][yMid]
			kernel[4] = arr[xMid][yMid]
			kernel[5] = arr[xRight][yMid]
			kernel[6] = arr[xLeft][yBot]
			kernel[7] = arr[xMid][yBot]
			kernel[8] = arr[xRight][yBot]

			f(xIndex, yIndex, kernel)
		}
	}
}
