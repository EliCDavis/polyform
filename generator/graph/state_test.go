package graph_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryStepsDoNotCarryFileContent(t *testing.T) {
	inst := importTestInstance(t, "Files")
	fileVariable(t, inst, "Splat", bytes.Repeat([]byte("x"), 40<<20))
	height := &variable.TypeVariable[float64]{}
	require.NoError(t, errOf(inst.NewVariable("Height", height)))

	for i := range 3 {
		require.NoError(t, inst.History().Transact("height", func() error {
			height.SetValue(float64(i + 1))
			return nil
		}))
	}

	assert.Len(t, inst.History().Steps().Undo, 3)
}

func TestUndoRestoresTheFileAVariableHeld(t *testing.T) {
	inst := importTestInstance(t, "Files")
	fileVariable(t, inst, "Splat", []byte("first"))

	require.NoError(t, inst.History().Transact("replace", func() error {
		_, err := inst.UpdateVariable("Splat", []byte("second"))
		return err
	}))

	content := func() []byte {
		v, err := inst.GetVariable("Splat")
		require.NoError(t, err)
		return v.ToMessage()
	}

	_, err := inst.History().Undo()
	require.NoError(t, err)
	assert.Equal(t, []byte("first"), content())

	_, err = inst.History().Redo()
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), content())
}
