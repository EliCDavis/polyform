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

	err = os.WriteFile(gs.savePath, data, 0666)
	if err != nil {
		panic(err)
	}
	log.Printf("Graph written %s\n", gs.savePath)
}
