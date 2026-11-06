package graph_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodingSurvivesLayoutWrites(t *testing.T) {
	inst := testInstanceWithSubGraphTypes(t)
	position := func(i int) (string, map[string]any) {
		return fmt.Sprintf("nodes.Node-%d.position", i), map[string]any{"x": float64(i), "y": 0.}
	}
	for i := range 2000 {
		inst.SetMetadata(position(i))
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 2000; i < 200000; i++ {
			inst.SetMetadata(position(i % 4000))
		}
	}()

	for running := true; running; {
		select {
		case <-done:
			running = false
		default:
			_, err := inst.EncodeToAppSchema()
			require.NoError(t, err)
		}
	}
}
