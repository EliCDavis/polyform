package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetypeRefusesAValueOutsideANormRange(t *testing.T) {
	inst := importTestInstance(t, "retype")
	floatVariable(t, inst, "Wear", 2.5)

	err := inst.RetypeVariable("Wear", "norm.UFloat64", nil)

	assert.ErrorContains(t, err, "2.5")
	held, lookupErr := inst.GetVariable("Wear")
	require.NoError(t, lookupErr)
	assert.IsType(t, &variable.TypeVariable[float64]{}, held)
}

func TestRetypeRefusesAProfileValueOutsideANormRange(t *testing.T) {
	inst := importTestInstance(t, "retype")
	floatVariable(t, inst, "Lean", -3)
	inst.SaveProfile("Toppled")
	_, err := inst.UpdateVariable("Lean", []byte("0.25"))
	require.NoError(t, err)

	err = inst.RetypeVariable("Lean", "norm.SFloat64", nil)

	assert.ErrorContains(t, err, "Toppled")
}

func TestRetypeTakesAValueInsideANormRange(t *testing.T) {
	inst := importTestInstance(t, "retype")
	floatVariable(t, inst, "Wear", 0.35)

	require.NoError(t, inst.RetypeVariable("Wear", "norm.UFloat64", nil))
	require.NoError(t, inst.RetypeVariable("Wear", "float64", []byte("7")))
}
