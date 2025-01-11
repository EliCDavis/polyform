package generator_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/stretchr/testify/assert"
)

func TestNestedSyncMap(t *testing.T) {
	syncmap := generator.NewNestedSyncMap()

	syncmap.Set("1.2.3", 4)
	syncmap.Set("1.2.4", 5)
	syncmap.Set("1.x", "somthin")

	assert.Equal(t, 4, syncmap.Get("1.2.3"))
	assert.Equal(t, 5, syncmap.Get("1.2.4"))
	assert.Equal(t, "somthin", syncmap.Get("1.x"))
}
