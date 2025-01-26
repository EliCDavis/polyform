package generator

import (
	"log"
	"os"
	"sync"
)

type GraphSaver struct {
	app          *App
	autsaveMutex sync.Mutex
	savePath     string
}

func (gs *GraphSaver) Save() {
	if gs == nil {
		return
	}

	gs.autsaveMutex.Lock()
	defer gs.autsaveMutex.Unlock()
	err := os.WriteFile(gs.savePath, gs.app.Graph(), 0666)
	if err != nil {
		panic(err)
	}
	log.Printf("Graph written %s\n", gs.savePath)
}
