package sequence

import "github.com/EliCDavis/polyform/nodes"

type IntegersNode struct {
	Count nodes.Output[int] `description:"How many values to produce."`
	Start nodes.Output[int] `description:"The first value."`
}

func (IntegersNode) Description() string {
	return "Count whole numbers, counting up from Start."
}

func (IntegersNode) Keywords() []string {
	return []string{"enumerate", "iota"}
}

func (n IntegersNode) Out(out *nodes.StructOutput[[]int]) {
	count := max(nodes.GetOutputValue(out, n.Count), 0)
	start := nodes.GetOutputValue(out, n.Start)
	arr := make([]int, count)
	for i := range arr {
		arr[i] = start + i
	}
	out.Set(arr)
}
