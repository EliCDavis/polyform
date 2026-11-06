package history_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EliCDavis/polyform/generator/graph/history"
	"github.com/stretchr/testify/assert"
)

func TestStepsTakenTogetherRunOneAtATime(t *testing.T) {
	h := history.New(func() ([]byte, error) { return nil, nil }, func([]byte) error { return nil })

	var running, most atomic.Int32
	var steps sync.WaitGroup
	for range 8 {
		steps.Add(1)
		go func() {
			defer steps.Done()
			_ = h.Transact("step", func() error {
				most.Store(max(most.Load(), running.Add(1)))
				time.Sleep(time.Millisecond)
				running.Add(-1)
				return nil
			})
		}()
	}
	steps.Wait()

	assert.Equal(t, int32(1), most.Load())
	assert.Len(t, h.Steps().Undo, 8)
}

func TestFailedStepKeepsAStepRequestedDuringIt(t *testing.T) {
	var mu sync.Mutex
	state := "start"
	set := func(next string) {
		mu.Lock()
		defer mu.Unlock()
		state = next
	}
	h := history.New(
		func() ([]byte, error) {
			mu.Lock()
			defer mu.Unlock()
			return []byte(state), nil
		},
		func(saved []byte) error {
			set(string(saved))
			return nil
		},
	)

	started, finished := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(finished)
		<-started
		_ = h.Transact("works", func() error {
			set("changed")
			return nil
		})
	}()

	err := h.Transact("fails", func() error {
		close(started)
		time.Sleep(20 * time.Millisecond)
		return errors.New("no")
	})
	<-finished

	assert.Error(t, err)
	assert.Equal(t, "changed", state)
}
