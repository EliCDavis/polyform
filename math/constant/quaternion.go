package constant

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
)

type Quaternion struct{}

func (Quaternion) Name() string {
	return "Quaternion"
}

func (Quaternion) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *Quaternion) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Identity": nodes.ConstOutput[quaternion.Quaternion]{
			Ref:      p,
			Val:      quaternion.Identity(),
			PortName: "Identity",
		},
	}
}
