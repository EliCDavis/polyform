package coloring

import "github.com/EliCDavis/polyform/nodes"

func init() {
	nodes.RegisterLiteralType(Color{})
	nodes.RegisterLiteralType(NewGradientColor(
		GradientKey[Color]{Time: 0, Value: Black()},
		GradientKey[Color]{Time: 1, Value: White()},
	))
}
