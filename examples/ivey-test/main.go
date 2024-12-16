package main

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fp := "C:/Users/elida/Downloads/snapshot_binary_format.bin"

	f, err := os.Open(fp)
	check(err)
	defer f.Close()

	reader := bufio.NewReader(f)
	floatBuf := make([]byte, 4)

	dimensions := 4096
	img := image.NewRGBA(image.Rect(0, 0, dimensions, dimensions))

	min := -10.
	max := 10.
	valueRange := max - min

	for x := 0; x < dimensions; x++ {
		for y := 0; y < dimensions; y++ {
			io.ReadFull(reader, floatBuf)
			f := math.Float32frombits(binary.LittleEndian.Uint32(floatBuf))
			adjusted := (float64(f) - min) / valueRange
			colorValue := byte(math.Min(1, math.Max(0, adjusted)) * 255)
			img.Set(x, y, color.RGBA{
				R: colorValue,
				G: colorValue,
				B: colorValue,
				A: 255,
			})
		}
	}

	imgFile, err := os.Create("img.png")
	check(err)
	defer imgFile.Close()
	check(png.Encode(imgFile, img))
}
