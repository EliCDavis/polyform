package texturing

import "image/color"

func SimpleEdgeTest(kernel []color.Color) bool {
	for _, val := range kernel {
		if kernel[4] != val {
			return true
		}
	}
	return false
}
