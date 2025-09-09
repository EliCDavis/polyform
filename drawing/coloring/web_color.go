package coloring

import (
	"fmt"
	"image/color"
	"strconv"
)

// Like color.RGBA but we can be serialized to JSON!
type Color struct {
	R float64
	G float64
	B float64
	A float64
}

func (c Color) MarshalJSON() ([]byte, error) {
	if c.A != 1 {
		return []byte(fmt.Sprintf(
			"\"#%02x%02x%02x%02x\"",
			byte(c.R*255),
			byte(c.G*255),
			byte(c.B*255),
			byte(c.A*255),
		)), nil
	}

	return []byte(fmt.Sprintf(
		"\"#%02x%02x%02x\"",
		byte(c.R*255),
		byte(c.G*255),
		byte(c.B*255),
	)), nil
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * 255)
	r |= r << 8
	g = uint32(c.G * 255)
	g |= g << 8
	b = uint32(c.B * 255)
	b |= b << 8
	a = uint32(c.A * 255)
	a |= a << 8
	return
}

func (c Color) RGBA8() color.RGBA {
	return color.RGBA{
		R: byte((c.R) * 255),
		G: byte((c.G) * 255),
		B: byte((c.B) * 255),
		A: byte((c.A) * 255),
	}
}

func (c *Color) unmarshalJson6Digit(data []byte) error {
	hex := string(data)
	r, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse r component of color '%s': %w", hex, err)
	}
	g, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse g component of color '%s': %w", hex, err)
	}
	b, err := strconv.ParseUint(hex[6:8], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse b component of color '%s': %w", hex, err)
	}

	var a uint64 = 255
	if len(data) == 11 {
		a, err = strconv.ParseUint(hex[8:10], 16, 8)
		if err != nil {
			return fmt.Errorf("unable to parse a component of color '%s': %w", hex, err)
		}
	}

	c.R = float64(r) / 255.
	c.G = float64(g) / 255.
	c.B = float64(b) / 255.
	c.A = float64(a) / 255.
	return nil
}

func (c *Color) unmarshalJson3Digit(data []byte) error {
	hex := string(data)
	r, err := strconv.ParseUint(hex[2:3]+hex[2:3], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse r component of color '%s': %w", hex, err)
	}
	g, err := strconv.ParseUint(hex[3:4]+hex[3:4], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse g component of color '%s': %w", hex, err)
	}

	b, err := strconv.ParseUint(hex[4:5]+hex[4:5], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse b component of color '%s': %w", hex, err)
	}

	var a uint64 = 255
	if len(data) == 7 {
		a, err = strconv.ParseUint(hex[5:6]+hex[5:6], 16, 8)
		if err != nil {
			return fmt.Errorf("unable to parse a component of color '%s': %w", hex, err)
		}
	}

	c.R = float64(r) / 255.
	c.G = float64(g) / 255.
	c.B = float64(b) / 255.
	c.A = float64(a) / 255.
	return nil
}

func (c *Color) UnmarshalJSON(data []byte) error {
	if len(data) == 6 || len(data) == 7 {
		return c.unmarshalJson3Digit(data)
	}
	return c.unmarshalJson6Digit(data)
}

func (c Color) Lerp(b Color, time float64) Color {
	mt := 1 - time
	return Color{
		R: (c.R * mt) + (b.R * time),
		G: (c.G * mt) + (b.G * time),
		B: (c.B * mt) + (b.B * time),
		A: (c.A * mt) + (b.A * time),
	}
}
