package meshops

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type RotateAttribute3DTransformer struct {
	Attribute string
	Amount    quaternion.Quaternion
}

func (rat RotateAttribute3DTransformer) attribute() string {
	return rat.Attribute
}

func (rat RotateAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(rat, modeling.PositionAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return RotateAttribute3D(m, attribute, rat.Amount), nil
}

func RotateAttribute3D(m modeling.Mesh, attribute string, q quaternion.Quaternion) modeling.Mesh {
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
