package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Scale3DTransformer struct {
	Attribute string
	Origin    vector3.Float64
	Amount    vector3.Float64
}

func (st Scale3DTransformer) attribute() string {
	return st.Attribute
}

func (st Scale3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.PositionAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return ScaleAttribute3D(m, attribute, st.Origin, st.Amount), nil
}

func ScaleAttribute3D(m modeling.Mesh, attribute string, origin, amount vector3.Float64) modeling.Mesh {
	if err := requireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = origin.Add(oldData.At(i).Sub(origin).MultByVector(amount))
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}
