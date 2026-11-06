package logic

import "github.com/EliCDavis/polyform/nodes"

type AndNode struct {
	Values []nodes.LiftedPort[bool] `description:"The conditions that must all hold."`
}

func (AndNode) Description() string {
	return "True when every input is true."
}

func (n AndNode) Out(out *nodes.Lifted[bool]) {
	nodes.ZipAll(out, n.Values, func(values []bool) bool {
		for _, v := range values {
			if !v {
				return false
			}
		}
		return true
	})
}
