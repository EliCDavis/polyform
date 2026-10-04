package graph_test

import (
	"errors"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nodeCount(t *testing.T, instance *graph.Instance) int {
	t.Helper()
	return len(instance.Schema().Nodes)
}

func TestTransactRecordsAnUndoableStep(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	require.False(t, instance.History().CanUndo())

	require.NoError(t, instance.History().Transact("add a source", func() error {
		_, _, err := instance.CreateNode(floatSourceType)
		return err
	}))

	assert.Equal(t, 1, nodeCount(t, instance))
	require.True(t, instance.History().CanUndo())
	assert.Equal(t, []string{"add a source"}, instance.History().Steps().Undo)
}

func TestUndoAndRedoWalkTheStack(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	require.NoError(t, instance.History().Transact("first", func() error {
		_, _, err := instance.CreateNode(floatSourceType)
		return err
	}))
	require.NoError(t, instance.History().Transact("second", func() error {
		_, _, err := instance.CreateNode(floatSourceType)
		return err
	}))
	require.Equal(t, 2, nodeCount(t, instance))

	label, err := instance.History().Undo()
	require.NoError(t, err)
	assert.Equal(t, "second", label)
	assert.Equal(t, 1, nodeCount(t, instance))

	label, err = instance.History().Undo()
	require.NoError(t, err)
	assert.Equal(t, "first", label)
	assert.Equal(t, 0, nodeCount(t, instance))

	require.False(t, instance.History().CanUndo())
	_, err = instance.History().Undo()
	assert.Error(t, err)

	label, err = instance.History().Redo()
	require.NoError(t, err)
	assert.Equal(t, "first", label)
	assert.Equal(t, 1, nodeCount(t, instance))
}

func TestActingAfterAnUndoAbandonsTheRedoBranch(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	require.NoError(t, instance.History().Transact("first", func() error {
		_, _, err := instance.CreateNode(floatSourceType)
		return err
	}))
	_, err := instance.History().Undo()
	require.NoError(t, err)
	require.True(t, instance.History().CanRedo())

	require.NoError(t, instance.History().Transact("different", func() error {
		_, _, err := instance.CreateNode(floatArraySourceType)
		return err
	}))

	assert.False(t, instance.History().CanRedo(), "the undone branch is gone")
	assert.Equal(t, []string{"different"}, instance.History().Steps().Undo)
}

func TestATransactionThatFailsPartwayLeavesNothingBehind(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, existing, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)

	failure := errors.New("the second half went wrong")
	err = instance.History().Transact("two changes", func() error {
		if _, _, e := instance.CreateNode(floatSourceType); e != nil {
			return e
		}
		return failure
	})

	require.ErrorIs(t, err, failure)
	assert.Equal(t, 1, nodeCount(t, instance), "the first change was rolled back too")
	assert.NotNil(t, instance.Node(existing), "and what was already there survives")
	assert.False(t, instance.History().CanUndo(), "a failed step is not an undoable one")
}

func TestATransactionThatPanicsIsRolledBack(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	err := instance.History().Transact("connect something impossible", func() error {
		_, consumer, e := instance.CreateNode(floatConsumerType)
		if e != nil {
			return e
		}
		panic(instance.ConnectNodes(consumer, "Out", consumer, "In"))
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "itself")
	assert.Equal(t, 0, nodeCount(t, instance))
	assert.False(t, instance.History().CanUndo())
}

func TestUndoRestoresConnectionsNotJustNodes(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	_, source, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(floatConsumerType)
	require.NoError(t, err)

	require.NoError(t, instance.History().Transact("wire them", func() error {
		require.NoError(t, instance.ConnectNodes(source, "Out", consumer, "In"))
		return nil
	}))
	require.NotNil(t, instance.Schema().Nodes[consumer].AssignedInput["In"])

	_, err = instance.History().Undo()
	require.NoError(t, err)
	assert.Empty(t, instance.Schema().Nodes[consumer].AssignedInput,
		"the connection went away with the step that made it")
}

func TestHistoryIsBoundedByDepth(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	for range 80 {
		require.NoError(t, instance.History().Transact("step", func() error {
			_, _, err := instance.CreateNode(floatSourceType)
			return err
		}))
	}

	assert.LessOrEqual(t, len(instance.History().Steps().Undo), 64)
	assert.Greater(t, len(instance.History().Steps().Undo), 0)
}
