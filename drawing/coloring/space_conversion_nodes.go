package coloring

import "github.com/EliCDavis/polyform/nodes"

type SRGBToLinearNode struct {
	In nodes.LiftedPort[Color] `description:"A color as picked on screen."`
}

func (n SRGBToLinearNode) Description() string {
	return "Converts an sRGB color to linear space."
}

func (n SRGBToLinearNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip1(out, nodes.LiftedOr(n.In, Color{A: 1}), func(in Color) Color {
		return Color{
			R: SRGBToLinear(in.R),
			G: SRGBToLinear(in.G),
			B: SRGBToLinear(in.B),
			A: in.A,
		}
	})
}

type LinearToSRGBNode struct {
	In nodes.LiftedPort[Color] `description:"A linear-space color."`
}

func (n LinearToSRGBNode) Description() string {
	return "Converts a linear-space color to sRGB. Inverse of SRGB To Linear."
}

func (n LinearToSRGBNode) Out(out *nodes.Lifted[Color]) {
	nodes.Zip1(out, nodes.LiftedOr(n.In, Color{A: 1}), func(in Color) Color {
		return Color{
			R: LinearToSRGB(in.R),
			G: LinearToSRGB(in.G),
			B: LinearToSRGB(in.B),
			A: in.A,
		}
	})
}
