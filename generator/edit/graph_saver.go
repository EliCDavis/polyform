package edit

import (
	"log"
	"os"
	"sync"

	"github.com/EliCDavis/polyform/generator/graph"
)

type GraphSaver struct {
	graph        *graph.Instance
	autsaveMutex sync.Mutex
	savePath     string
}

func (gs *GraphSaver) Save() {
	if gs == nil {
		return
	}

	gs.autsaveMutex.Lock()
	defer gs.autsaveMutex.Unlock()

	data, err := gs.graph.EncodeToAppSchema()
	if err != nil {
		panic(err)
	}

	// Written beside and renamed over: overwriting a large file in place waits
	// on the virus scan of its previous version, and a crash mid-write keeps the old graph.
	partial := gs.savePath + ".partial"
	if err = os.WriteFile(partial, data, 0666); err != nil {
		panic(err)
	}
	if err = os.Rename(partial, gs.savePath); err != nil {
		panic(err)
	}
	log.Printf("Graph written %s\n", gs.savePath)
}
