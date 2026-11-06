package cast_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/cast"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCastConvertsOncePerSourceVersion(t *testing.T) {
	calls := 0
	table := &cast.Table{}
	cast.Register(table, func(v float64) int {
		calls++
		return int(v)
	})

	source := nodes.NewValue(2.)
	out, ok := table.Output(source.Outputs()["Value"], "int")
	require.True(t, ok)
	read := out.(nodes.Output[int])

	assert.Equal(t, 2, read.Value())
	assert.Equal(t, 2, read.Value())
	assert.Equal(t, 1, calls)

	source.Set(5)
	assert.Equal(t, 5, read.Value())
	assert.Equal(t, 5, read.Value())
	assert.Equal(t, 2, calls)
}
