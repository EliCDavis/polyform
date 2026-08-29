package variant

import (
	"encoding/json"
	"fmt"
	"math"
)

// hexColor encodes r, g, b (0-1) as a "#rrggbb" string.
func hexColor(r, g, b float64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(
		"\"#%02x%02x%02x\"",
		colorByte(r), colorByte(g), colorByte(b),
	))
}

func colorByte(v float64) byte {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return byte(v * 255)
}

// hsvToRGB converts h (degrees), s, and v (0-1) into r, g, b (0-1).
func hsvToRGB(h, s, v float64) (r, g, b float64) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var rp, gp, bp float64
	switch {
	case h < 60:
		rp, gp, bp = c, x, 0
	case h < 120:
		rp, gp, bp = x, c, 0
	case h < 180:
		rp, gp, bp = 0, c, x
	case h < 240:
		rp, gp, bp = 0, x, c
	case h < 300:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}

	return rp + m, gp + m, bp + m
}
