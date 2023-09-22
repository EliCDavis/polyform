package coloring

import (
	"fmt"
	"image/color"
	"strconv"
)

type WebColor struct {
	R byte
	G byte
	B byte
	A byte
}

func (v WebColor) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"\"#%02x%02x%02x\"",
		v.R,
		v.G,
		v.B,
	)), nil
}

func (v WebColor) RGBA() color.RGBA {
	return color.RGBA{
		R: v.R,
		G: v.G,
		B: v.B,
		A: v.A,
	}
}

func (v *WebColor) UnmarshalJSON(data []byte) error {
	hex := string(data)
	r, err := strconv.ParseInt(hex[1:3], 16, 64)
	if err != nil {
		return err
	}
	g, err := strconv.ParseInt(hex[3:5], 16, 64)
	if err != nil {
		return err
	}
	b, err := strconv.ParseInt(hex[5:7], 16, 64)
	if err != nil {
		return err
	}

	v.R = byte(r)
	v.G = byte(g)
	v.B = byte(b)
	v.A = 255
	return nil
}
