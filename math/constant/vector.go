package constant

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Vector3[T vector.Number] struct{}

func (Vector3[T]) Name() string {
	var x T
	return fmt.Sprintf("Vector3[%s]", refutil.GetTypeNameWithoutPackage(x))
}

func (Vector3[T]) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *Vector3[T]) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Up": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Up[T](),
			PortName: "Up",
		},

		"Down": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Down[T](),
			PortName: "Down",
		},

		"Left": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Left[T](),
			PortName: "Left",
		},

		"Right": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Right[T](),
			PortName: "Right",
		},

		"Forward": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Forward[T](),
			PortName: "Forward",
		},

		"Backward": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.Backwards[T](),
			PortName: "Backward",
		},

		"One": ConstOutput[vector3.Vector[T]]{
			Ref:      p,
			Val:      vector3.One[T](),
			PortName: "One",
		},
	}
}
