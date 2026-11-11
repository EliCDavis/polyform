package coloring

import "github.com/EliCDavis/polyform/nodes"

type SRGBToLinearNode struct {
	In nodes.LiftedPort[Color] `description:"A color as picked on screen." default:"#000000"`
}

func (n SRGBToLinearNode) Description() string {
	return "Converts an sRGB color to linear space."
}

func (n SRGBToLinearNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip1(out, n.In, Color.SRGBToLinear)
}

type LinearToSRGBNode struct {
	In nodes.LiftedPort[Color] `description:"A linear-space color." default:"#000000"`
}

func (n LinearToSRGBNode) Description() string {
	return "Converts a linear-space color to sRGB. Inverse of SRGB To Linear."
}

func (n LinearToSRGBNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip1(out, n.In, Color.LinearToSRGB)
}
