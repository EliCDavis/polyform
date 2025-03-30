package constant

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

type Pi struct{}

func (Pi) Name() string {
	return "Pi"
}

func (Pi) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *Pi) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Pi": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi,
			PortName: "Pi",
		},

		"Pi / 2": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi / 2,
			PortName: "Pi / 2",
		},

		"2 Pi": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi * 2,
			PortName: "2 Pi",
		},

		"Square Root": ConstOutput[float64]{
			Ref:      p,
			Val:      math.SqrtPi,
			PortName: "Square Root",
		},
	}
}
