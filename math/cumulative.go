package math

import (
	"github.com/EliCDavis/polyform/nodes"
)

type CumulativeNode struct {
	Values nodes.LiftedPort[[]float64]
}

func (CumulativeNode) Description() string {
	return "Running totals of a list."
}

func (CumulativeNode) InclusiveDescription() string {
	return "Each total includes its own value."
}

func (CumulativeNode) ExclusiveDescription() string {
	return "Each total covers only the values before it."
}

func cumulative(values []float64, inclusive bool) []float64 {
	totals := make([]float64, len(values))
	total := 0.
	for i, v := range values {
		if inclusive {
			total += v
			totals[i] = total
			continue
		}
		totals[i] = total
		total += v
	}
	return totals
}

func (n CumulativeNode) Inclusive(out *nodes.Lifted[[]float64]) {
	nodes.Zip1(out, n.Values, func(values []float64) []float64 { return cumulative(values, true) })
}

func (n CumulativeNode) Exclusive(out *nodes.Lifted[[]float64]) {
	nodes.Zip1(out, n.Values, func(values []float64) []float64 { return cumulative(values, false) })
}
