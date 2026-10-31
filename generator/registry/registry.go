// Package registry is the one list of node types an app can build by name.
// It sits below the rest of generator so a node package can announce
// itself without importing the app.
package registry

import (
	"sync"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

var (
	mu    sync.Mutex
	types = new(refutil.TypeFactory)
)

func Register(factory *refutil.TypeFactory) {
	mu.Lock()
	defer mu.Unlock()
	types = types.Combine(factory)
	nodes.DiscoverPortTypes(factory)
}

func Types() *refutil.TypeFactory {
	mu.Lock()
	defer mu.Unlock()
	return types
}
