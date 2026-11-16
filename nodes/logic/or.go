package logic

import "github.com/EliCDavis/polyform/nodes"

type OrNode struct {
	Values []nodes.LiftedPort[bool] `description:"The conditions, any one of which is enough."`
}

func (OrNode) Description() string {
	return "True when any input is true."
}

func (n OrNode) Out(out *nodes.Lifted[bool]) {
	nodes.ZipAll(out, n.Values, func(values []bool) bool {
		for _, v := range values {
			if v {
				return true
			}
		}
		return false
	})
}
