package generator

import (
	"sync"

	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/refutil"
)

var types *refutil.TypeFactory = new(refutil.TypeFactory)

var typeMutex sync.Mutex

func RegisterTypes(typesToRegister *refutil.TypeFactory) {
	typeMutex.Lock()
	defer typeMutex.Unlock()
	types = types.Combine(typesToRegister)
	subgraph.DiscoverPortTypes(typesToRegister)
	// for _, t := range types.Types() {
	// 	log.Printf("Registered: %s\n", t)
	// }
}

// Types returns the global registry of all node types that have registered
// themselves via RegisterTypes (typically through package init() functions).
// External tools that build a graph.Instance outside of the polyform CLI
// (e.g. cmd/polyform-mcp) use this to obtain a fully populated
// *refutil.TypeFactory after blank-importing the desired node packages.
func Types() *refutil.TypeFactory {
	typeMutex.Lock()
	defer typeMutex.Unlock()
	return types
}
