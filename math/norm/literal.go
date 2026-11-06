package norm

import "github.com/EliCDavis/polyform/nodes"

func init() {
	nodes.RegisterLiteralType(UFloat64(0))
	nodes.RegisterLiteralType(SFloat64(0))
}
