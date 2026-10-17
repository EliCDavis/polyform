package arrays_test

import (
	"testing"
	"time"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type slowValues struct{}

func (slowValues) Out(out *nodes.StructOutput[[]float64]) {
	time.Sleep(40 * time.Millisecond)
	out.Set([]float64{1, 2, 3})
}

func TestSpreadDoesNotBillItsInputsWorkAsItsOwn(t *testing.T) {
	spread := &nodes.Struct[arrays.SpreadNode]{}
	wire(t, spread, "Array", nodes.GetNodeOutputPort[[]float64](&nodes.Struct[slowValues]{}, "Out"))
	out := spread.Outputs()["Out"]

	_ = out.(nodes.Output[float64]).Value()
	report := out.(nodes.ObservableExecution).ExecutionReport()

	require.NotNil(t, report.SelfTime)
	assert.GreaterOrEqual(t, report.TotalTime, 40*time.Millisecond)
	assert.Less(t, *report.SelfTime, 20*time.Millisecond)
}
