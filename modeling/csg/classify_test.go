package csg

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanningAgreesWithTheTree(t *testing.T) {
	a, err := facesOf(primitives.UVSphere(1, 16, 24))
	require.NoError(t, err)
	b, err := facesOf(primitives.UVSphere(1, 16, 24).Translate(vector3.New(0.7, 0.3, 0.2)))
	require.NoError(t, err)

	tolerance := toleranceFor(a, b)
	enough := scanBudget/len(b) + 1
	indexed := newTarget(b, tolerance, enough)
	scanned := newTarget(b, tolerance, 0)

	require.NotNil(t, indexed.tree, "sanity: the high query count should build a tree")
	require.Nil(t, scanned.tree, "sanity: no queries should skip it")

	for i, f := range a {
		assert.Equal(t, indexed.classify(f), scanned.classify(f), "face %d", i)
	}
}
