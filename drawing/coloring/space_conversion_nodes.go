package coloring

import "github.com/EliCDavis/polyform/nodes"

type SRGBToLinearNode struct {
	In nodes.Output[Color] `description:"A color as picked on screen."`
}

func (n SRGBToLinearNode) Description() string {
	return "Converts an sRGB color to linear space."
}

func (n SRGBToLinearNode) Out(out *nodes.StructOutput[Color]) {
	if n.In == nil {
		out.Set(Color{A: 1})
		return
	}
	in := nodes.GetOutputValue(out, n.In)
	out.Set(Color{
		R: SRGBToLinear(in.R),
		G: SRGBToLinear(in.G),
		B: SRGBToLinear(in.B),
		A: in.A,
	})
}

type LinearToSRGBNode struct {
	In nodes.Output[Color] `description:"A linear-space color."`
}

func (n LinearToSRGBNode) Description() string {
	return "Converts a linear-space color to sRGB. Inverse of SRGB To Linear."
}

func (n LinearToSRGBNode) Out(out *nodes.StructOutput[Color]) {
	if n.In == nil {
		out.Set(Color{A: 1})
		return
	}
	in := nodes.GetOutputValue(out, n.In)
	out.Set(Color{
		R: LinearToSRGB(in.R),
		G: LinearToSRGB(in.G),
		B: LinearToSRGB(in.B),
		A: in.A,
	})
}
