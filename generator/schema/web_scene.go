package schema

import (
	"fmt"
	"image/color"
	"strconv"
)

type WebSceneFog struct {
	Color WebColor `json:"color"`
	Near  float32  `json:"near"`
	Far   float32  `json:"far"`
}

type WebScene struct {
	RenderWireframe bool        `json:"renderWireframe"`
	AntiAlias       bool        `json:"antiAlias"`
	XrEnabled       bool        `json:"xrEnabled"`
	Fog             WebSceneFog `json:"fog"`
	Background      WebColor    `json:"background"`
	Lighting        WebColor    `json:"lighting"`
	Ground          WebColor    `json:"ground"`
}

type WebColor struct {
	R byte
	G byte
	B byte
	A byte
}

func (c WebColor) MarshalJSON() ([]byte, error) {
	if c.A != 255 {
		return []byte(fmt.Sprintf(
			"\"#%02x%02x%02x%02x\"",
			c.R,
			c.G,
			c.B,
			c.A,
		)), nil
	}

	return []byte(fmt.Sprintf(
		"\"#%02x%02x%02x\"",
		c.R,
		c.G,
		c.B,
	)), nil
}

func (c WebColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}

func (c WebColor) RGBA8() color.RGBA {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: c.A,
	}
}

func (c *WebColor) unmarshalJson6Digit(data []byte) error {
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

	c.R = byte(r)
	c.G = byte(g)
	c.B = byte(b)
	c.A = byte(a)
	return nil
}

func (c *WebColor) unmarshalJson3Digit(data []byte) error {
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

	c.R = byte(r)
	c.G = byte(g)
	c.B = byte(b)
	c.A = byte(a)
	return nil
}

func (c *WebColor) UnmarshalJSON(data []byte) error {
	if len(data) == 6 || len(data) == 7 {
		return c.unmarshalJson3Digit(data)
	}
	return c.unmarshalJson6Digit(data)
}
