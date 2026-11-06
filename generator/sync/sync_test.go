package sync_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/stretchr/testify/assert"
)

func TestNestedSyncMap(t *testing.T) {
	syncmap := sync.NewNestedSyncMap()

	syncmap.Set("1.2.3", 4)
	syncmap.Set("1.2.4", 5)
	syncmap.Set("1.x", "somthin")

	assert.Equal(t, 4, syncmap.Get("1.2.3"))
	assert.Equal(t, 5, syncmap.Get("1.2.4"))
	assert.Equal(t, "somthin", syncmap.Get("1.x"))
}

func TestNestedSyncMapDataIsUntouchedByLaterWrites(t *testing.T) {
	syncmap := sync.NewNestedSyncMap()
	syncmap.Set("nodes.a.position", 1)

	data := syncmap.Data()
	syncmap.Set("nodes.b.position", 2)
	syncmap.Delete("nodes.a")

	assert.Equal(t, map[string]any{"nodes": map[string]any{"a": map[string]any{"position": 1}}}, data)
}

func TestNestedSyncMapGetIsUntouchedByLaterWrites(t *testing.T) {
	syncmap := sync.NewNestedSyncMap()
	syncmap.Set("nodes.a", 1)
	syncmap.Set("list", []any{map[string]any{"x": 1}})

	nodes := syncmap.Get("nodes")
	list := syncmap.Get("list")
	syncmap.Set("nodes.b", 2)
	syncmap.Get("list").([]any)[0].(map[string]any)["x"] = 2

	assert.Equal(t, map[string]any{"a": 1}, nodes)
	assert.Equal(t, []any{map[string]any{"x": 1}}, list)
}
