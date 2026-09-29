package generator

import (
	"sync"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

var types *refutil.TypeFactory = new(refutil.TypeFactory)

var typeMutex sync.Mutex

func RegisterTypes(typesToRegister *refutil.TypeFactory) {
	typeMutex.Lock()
	defer typeMutex.Unlock()
	types = types.Combine(typesToRegister)
	nodes.DiscoverPortTypes(typesToRegister)
	// for _, t := range types.Types() {
	// 	log.Printf("Registered: %s\n", t)
	// }
}
