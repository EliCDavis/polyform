package history_test

import (
	"errors"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dereferenceNothing() int {
	var nothing *int
	return *nothing
}

func TestTransactKeepsTheStackOfAPanic(t *testing.T) {
	h := history.New(func() ([]byte, error) { return []byte("state"), nil }, func([]byte) error { return nil })

	err := h.Transact("probe", func() error {
		dereferenceNothing()
		return nil
	})

	var recovered *history.PanicError
	require.True(t, errors.As(err, &recovered))
	assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", err.Error())
	assert.Contains(t, recovered.Stack, "dereferenceNothing")
}
