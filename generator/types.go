package generator

import (
	"sync"

	"github.com/EliCDavis/polyform/refutil"
)

var types *refutil.TypeFactory = new(refutil.TypeFactory)

var typeMutex sync.Mutex

func RegisterTypes(typesToRegister *refutil.TypeFactory) {
	typeMutex.Lock()
	defer typeMutex.Unlock()
	types = types.Combine(typesToRegister)
	// for _, t := range types.Types() {
	// 	log.Printf("Registered: %s\n", t)
	// }
}

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}
	return factory.Combine(
	// parameter.Nodes(),
	// artifact.Nodes(),
	)
}
