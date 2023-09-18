package texturing

import (
	"image"
	"image/color"
)

// [0][1][2]
// [3][4][5]
// [6][7][8]
func Convolve(src image.Image, f func(x, y int, values []color.Color)) {
	kernel := make([]color.Color, 9)
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

			kernel[0] = src.At(xLeft, yTop)
			kernel[1] = src.At(xMid, yTop)
			kernel[2] = src.At(xRight, yTop)
			kernel[3] = src.At(xLeft, yMid)
			kernel[4] = src.At(xMid, yMid)
			kernel[5] = src.At(xRight, yMid)
			kernel[6] = src.At(xLeft, yBot)
			kernel[7] = src.At(xMid, yBot)
			kernel[8] = src.At(xRight, yBot)

			f(xIndex, yIndex, kernel)
		}
	}
}

func ConvolveArray[T any](arr [][]T, f func(x, y int, kernel []T)) {
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
