package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Rotate3DTransformer struct {
	Attribute string
	Amount    modeling.Quaternion
}

func (st Rotate3DTransformer) attribute() string {
	return st.Attribute
}

func (st Rotate3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.PositionAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return RotateAttribute3D(m, attribute, st.Amount), nil
}

func RotateAttribute3D(m modeling.Mesh, attribute string, q modeling.Quaternion) modeling.Mesh {
	if err := requireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = q.Rotate(oldData.At(i))
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}
