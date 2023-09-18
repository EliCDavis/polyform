package colors

import "math"

// From Three.js
// https://github.com/mrdoob/three.js/blob/e6f7c4e677cb8869502739da2640791d020d8d2f/src/math/ColorManagement.js#L6
func SRGBToLinear(c float64) float64 {
	if c < 0.04045 {
		return c * 0.0773993808
	}
	return math.Pow(c*0.9478672986+0.0521327014, 2.4)
}

// From Three.js
// https://github.com/mrdoob/three.js/blob/e6f7c4e677cb8869502739da2640791d020d8d2f/src/math/ColorManagement.js#L12
func LinearToSRGB(c float64) float64 {
	if c < 0.0031308 {
		return c * 12.92
	}
	return 1.055*(math.Pow(c, 0.41666)) - 0.055
}
