package parameter_test

import (
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormParametersClampWhenRead(t *testing.T) {
	u := parameter.UFloat64{}
	_, err := u.ApplyMessage([]byte(`2`))
	require.NoError(t, err)
	assert.Equal(t, 1., u.Value().Value())

	s := parameter.SFloat64{}
	_, err = s.ApplyMessage([]byte(`-3`))
	require.NoError(t, err)
	assert.Equal(t, -1., s.Value().Value())
}

func TestNormParameterSurvivesASave(t *testing.T) {
	saved := parameter.UFloat64{Name: "Shine", CurrentValue: 0.25}
	data, err := saved.ToJSON(nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Shine","description":"","currentValue":0.25}`, string(data))

	loaded := parameter.UFloat64{}
	require.NoError(t, loaded.FromJSON(jbtf.Decoder{}, data))
	assert.Equal(t, "Shine", loaded.Name)
	assert.Equal(t, 0.25, loaded.Value().Value())
}

func TestNormParameterSchemaNamesItsType(t *testing.T) {
	assert.Equal(t, "norm.UFloat64", (&parameter.UFloat64{}).Schema().ValueType())
	assert.Equal(t, "norm.SFloat64", (&parameter.SFloat64{}).Schema().ValueType())
}
