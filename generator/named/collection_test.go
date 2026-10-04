package named_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/named"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamesAreSorted(t *testing.T) {
	c := named.New[int]("thing")
	c.Set("zebra", 1)
	c.Set("apple", 2)
	assert.Equal(t, []string{"apple", "zebra"}, c.Names())
}

func TestGettingWhatWasNeverSavedIsAnError(t *testing.T) {
	c := named.New[int]("thing")
	_, err := c.Get("missing")
	assert.ErrorContains(t, err, `no thing exists with name "missing"`)
}

func TestRenameMovesTheItem(t *testing.T) {
	c := named.New[int]("thing")
	c.Set("old", 7)
	require.NoError(t, c.Rename("old", "new"))

	got, err := c.Get("new")
	require.NoError(t, err)
	assert.Equal(t, 7, got)
	assert.Equal(t, []string{"new"}, c.Names())
}

func TestRenameRefusesAMissingItemOrATakenName(t *testing.T) {
	c := named.New[int]("thing")
	c.Set("a", 1)
	c.Set("b", 2)
	assert.Error(t, c.Rename("missing", "c"))
	assert.ErrorContains(t, c.Rename("a", "b"), `"b" already exists`)
}

func TestDeleteRemovesOnce(t *testing.T) {
	c := named.New[int]("thing")
	c.Set("temp", 1)
	require.NoError(t, c.Delete("temp"))
	assert.Empty(t, c.Names())
	assert.Error(t, c.Delete("temp"))
}

func TestAllIsACopy(t *testing.T) {
	c := named.New[int]("thing")
	c.Set("a", 1)
	c.All()["b"] = 2
	assert.Equal(t, []string{"a"}, c.Names())
}
