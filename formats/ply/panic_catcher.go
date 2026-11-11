package ply

import "sync"

type panicCatcher struct {
	mu    sync.Mutex
	value any
}

func (c *panicCatcher) run(f func()) {
	defer func() {
		if r := recover(); r != nil {
			c.mu.Lock()
			if c.value == nil {
				c.value = r
			}
			c.mu.Unlock()
		}
	}()
	if c.failed() {
		return
	}
	f()
}

func (c *panicCatcher) failed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value != nil
}

// Call only after every goroutine that ran through run has been joined.
func (c *panicCatcher) rethrow() {
	if c.value != nil {
		panic(c.value)
	}
}
