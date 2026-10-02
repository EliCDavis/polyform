package variable_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyMessageRefusesNull(t *testing.T) {
	v := &variable.TypeVariable[int]{}
	applied, err := v.ApplyMessage([]byte("30"))
	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, 30, v.GetValue())

	applied, err = v.ApplyMessage([]byte("null"))

	require.Error(t, err)
	assert.False(t, applied)
	assert.Equal(t, 30, v.GetValue(), "the value it already had survives")
}

func TestApplyMessageRefusesAFloatForAnInt(t *testing.T) {
	v := &variable.TypeVariable[int]{}
	require.NoError(t, func() error { _, err := v.ApplyMessage([]byte("7")); return err }())

	_, err := v.ApplyMessage([]byte("7.5"))
	require.Error(t, err)
	assert.Equal(t, 7, v.GetValue())
}
