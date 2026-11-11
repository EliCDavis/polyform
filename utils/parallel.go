package utils

import (
	"runtime"
	"sync"
)

func ParallelRange(n int, f func(start, end int)) {
	workers := min(runtime.NumCPU(), n)
	if workers <= 1 {
		f(0, n)
		return
	}
	var wg sync.WaitGroup
	for i := range workers {
		start, end := n*i/workers, n*(i+1)/workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			f(start, end)
		}()
	}
	wg.Wait()
}
