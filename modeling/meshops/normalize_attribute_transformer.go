package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Finds the vector with the longest length and scales all vectors within the
// mesh attribute by (1 / longestLength)
type NormalizeAttribute3DTransformer struct {
	Attribute string
}

func (st NormalizeAttribute3DTransformer) attribute() string {
	return st.Attribute
}

func (st NormalizeAttribute3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(st, modeling.PositionAttribute)

	if err = requireV3Attribute(m, attribute); err != nil {
		return
	}

	return NormalizeAttribute3D(m, attribute), nil
}

func NormalizeAttribute3D(m modeling.Mesh, attribute string) modeling.Mesh {
	if err := requireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float3Attribute(attribute)
	maxLength := -math.MaxFloat64
	for i := 0; i < oldData.Len(); i++ {
		maxLength = math.Max(maxLength, oldData.At(i).Length())
	}

	scaledData := make([]vector3.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = oldData.At(i).DivByConstant(maxLength)
	}

	return m.SetFloat3Attribute(attribute, scaledData)
}

type NormalizeAttribute2DTransformer struct {
	Attribute string
}

func (st NormalizeAttribute2DTransformer) attribute() string {
	return st.Attribute
}

func (st NormalizeAttribute2DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = requireV2Attribute(m, st.Attribute); err != nil {
		return
	}

	return NormalizeAttribute2D(m, st.Attribute), nil
}

func NormalizeAttribute2D(m modeling.Mesh, attribute string) modeling.Mesh {
	if err := requireV2Attribute(m, attribute); err != nil {
		panic(err)
	}

	oldData := m.Float2Attribute(attribute)
	maxLength := -math.MaxFloat64
	for i := 0; i < oldData.Len(); i++ {
		maxLength = math.Max(maxLength, oldData.At(i).Length())
	}

	scaledData := make([]vector2.Float64, oldData.Len())
	for i := 0; i < oldData.Len(); i++ {
		scaledData[i] = oldData.At(i).DivByConstant(maxLength)
	}

	return m.SetFloat2Attribute(attribute, scaledData)
}
