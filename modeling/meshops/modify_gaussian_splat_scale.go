package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type ModifyGaussianSplatScaleTransformer struct {
	Attribute string
	Scale     vector3.Float64
}

func (st ModifyGaussianSplatScaleTransformer) attribute() string {
	return st.Attribute
}

func (st ModifyGaussianSplatScaleTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.ScaleAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return ModifyGaussianSplatScale(m, attribute, st.Scale), nil
}

func ModifyGaussianSplatScale(m modeling.Mesh, attribute string, amount vector3.Float64) modeling.Mesh {
	if err := requireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(modeling.ScaleAttribute)
	scaledData := make([]vector3.Float64, oldData.Len())
	len := oldData.Len()
	for i := 0; i < len; i++ {
		scaledData[i] = oldData.At(i).Exp().MultByVector(amount).Log()
	}
	return m.SetFloat3Attribute(modeling.ScaleAttribute, scaledData)
}
