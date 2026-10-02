package graph_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAFailedLoadLeavesTheOpenGraphAlone(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, existing, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)

	require.Error(t, instance.ApplyAppSchema([]byte("{ not valid json")))

	require.Len(t, instance.Schema().Nodes, 1, "the graph that was open survives")
	assert.NotNil(t, instance.Node(existing))
}

func TestAFailedLoadOfEmptyBytesLeavesTheGraphAlone(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, _, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)

	require.Error(t, instance.ApplyAppSchema(nil))
	assert.Len(t, instance.Schema().Nodes, 1)
}

func TestARollbackThatCannotLoadDoesNotEmptyTheGraph(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, _, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)

	err = instance.Transact("a step that fails", func() error {
		if _, _, e := instance.CreateNode(floatSourceType); e != nil {
			return e
		}
		return errors.New("something went wrong")
	})

	require.Error(t, err)
	assert.Len(t, instance.Schema().Nodes, 1, "rolled back to the one node that was there")
}

func TestUndoStillWorksAfterTheRestoreGuard(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	require.NoError(t, instance.Transact("add", func() error {
		_, _, e := instance.CreateNode(floatSourceType)
		return e
	}))
	require.Len(t, instance.Schema().Nodes, 1)

	label, err := instance.Undo()
	require.NoError(t, err)
	assert.Equal(t, "add", label)
	assert.Empty(t, instance.Schema().Nodes)

	_, err = instance.Redo()
	require.NoError(t, err)
	assert.Len(t, instance.Schema().Nodes, 1)
}
