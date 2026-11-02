package trs

import "github.com/EliCDavis/polyform/nodes"

func init() {
	nodes.RegisterLiteralType(Identity())
}
