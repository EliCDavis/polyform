package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
)

type CustomTransformer struct {
	Func func(m modeling.Mesh) (results modeling.Mesh, err error)
}

func (ct CustomTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	return ct.Func(m)
}
