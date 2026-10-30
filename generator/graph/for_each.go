package graph

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

// A placement set to For Each runs its definition once per item: an input
// fed a list one level deeper than it declares is iterated, anything else
// is broadcast, and every output becomes the list of the runs' results.

type placementData struct {
	Each bool `json:"each,omitempty"`
}

func (r *SubgraphInstanceNode) ForEach() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.each
}

func (r *SubgraphInstanceNode) ToJSON(*jbtf.Encoder) ([]byte, error) {
	if !r.ForEach() {
		return nil, nil
	}
	return json.Marshal(placementData{Each: true})
}

func (r *SubgraphInstanceNode) FromJSON(_ jbtf.Decoder, body []byte) error {
	var data placementData
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
	}
	r.setEach(data.Each)
	return nil
}

// Output types change with the setting, so edges that no longer fit are dropped and reported.
func (a *Graph) SetForEach(nodeID string, each bool) ([]DroppedEdge, error) {
	return a.project.write(dropConflicts, func() error {
		node, ok := a.nodesByID[nodeID]
		if !ok {
			return fmt.Errorf("no node exists with id %q", nodeID)
		}
		placement, ok := node.(*SubgraphInstanceNode)
		if !ok {
			return fmt.Errorf("node %q is not a subgraph instance", nodeID)
		}
		if placement.ForEach() == each {
			return nil
		}
		placement.setEach(each)
		a.dirty()
		return nil
	})
}

func (r *SubgraphInstanceNode) setEach(each bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.each = each
	r.copied = nil
	r.runs = nil
	r.outputs = nil
}

func (r *SubgraphInstanceNode) input(id string) nodes.OutputPort {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.inputs[id]
}

func iterates(declared string, fed nodes.OutputPort) bool {
	return fed != nil && fedType(declared, fed) != declared
}

// The type of one run's output, read from the definition: every run is fed exactly its declared types.
func (r *SubgraphInstanceNode) itemType(id, declared string) string {
	if output, ok := r.definition().outputPort(id); ok {
		return cmp.Or(output.EffectiveType(), declared)
	}
	return declared
}

// ============================================================================

// What run i's input boundary reads: element i of an iterated list, else the fed port itself.
type runInput struct {
	node     *SubgraphInstanceNode
	id       string
	portType string
	index    int
}

func (in runInput) Value() nodes.OutputPort {
	fed := in.node.input(in.id)
	if !iterates(in.portType, fed) {
		return fed
	}
	element := elementPort{list: fed, index: in.index, name: in.id, portType: in.portType}
	if builder, ok := nodes.LookupPortTypeDynamic(in.portType); ok {
		return builder.BuildDynamicOutput(element)
	}
	return element
}

// Its node and version are the list's source's, so a consumer in the run keys
// its cache on where the list comes from, as it would reading the list itself.
type elementPort struct {
	list     nodes.OutputPort
	index    int
	name     string
	portType string
}

func (p elementPort) Node() nodes.Node { return nodes.Source(p.list).Node() }
func (p elementPort) Version() int     { return nodes.Source(p.list).Version() }
func (p elementPort) Name() string     { return p.name }
func (p elementPort) Type() string     { return p.portType }

// Not the list: a proxy reading through it would hand the whole list back as one T.
func (p elementPort) CurrentSource() nodes.OutputPort { return nil }

// A list of one applies to every run.
func (p elementPort) AnyValue() any {
	list, ok := nodes.DynamicArrayValue(nil, p.list)
	if !ok {
		return nil
	}
	if list.Len() == 1 {
		return list.At(0)
	}
	return list.At(p.index)
}

// ============================================================================

// Iterated lists share a length, or are of length one to apply to every run.
func (r *SubgraphInstanceNode) runCount() (int, error) {
	definition := r.definition()
	var lengths []int
	for _, id := range slices.Sorted(maps.Keys(definition.boundaryNodes)) {
		boundary, ok := definition.inputPort(id)
		if !ok {
			continue
		}
		fed := r.input(id)
		if !iterates(boundary.BoundaryPortType(), fed) {
			continue
		}
		if list, ok := nodes.DynamicArrayValue(nil, fed); ok {
			lengths = append(lengths, list.Len())
		}
	}
	return nodes.BroadcastLength(lengths...)
}

// Runs is one current copy per item. A run that is still current keeps its caches.
func (r *SubgraphInstanceNode) Runs() ([]*Graph, error) {
	n, err := r.runCount()
	if err != nil {
		return nil, err
	}
	version := r.runtime().version
	r.mu.Lock()
	had := r.runs
	r.mu.Unlock()

	runs := make([]*copy, n)
	built := false
	for i := range runs {
		if i < len(had) && had[i].version == version {
			runs[i] = had[i]
			continue
		}
		graph, err := r.owner.copyDefinition(r, func(id, portType string) subgraph.External {
			return runInput{node: r, id: id, portType: portType, index: i}
		})
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		runs[i] = &copy{graph: graph, version: version}
		built = true
	}

	r.mu.Lock()
	r.runs = runs
	r.mu.Unlock()
	if built {
		nodes.Touch()
	}

	graphs := make([]*Graph, n)
	for i, run := range runs {
		graphs[i] = run.graph
	}
	return graphs, nil
}

// ============================================================================

type eachOutputPort struct {
	node     *SubgraphInstanceNode
	id       string
	label    string
	portType string
}

func (p eachOutputPort) Node() nodes.Node                       { return p.node }
func (p eachOutputPort) Name() string                           { return p.id }
func (p eachOutputPort) Label() string                          { return p.label }
func (p eachOutputPort) CurrentSource() nodes.OutputPort        { return nil }
func (p eachOutputPort) Type() string                           { return subgraph.ArrayOf(p.node.itemType(p.id, p.portType)) }
func (p eachOutputPort) InstanceType() string                   { return p.Type() }
func (p eachOutputPort) Version() int                           { return p.node.outputVersion(p.id) }
func (p eachOutputPort) AnyValue() any                          { return p.node.outputValue(p.id) }
func (p eachOutputPort) ExecutionReport() nodes.ExecutionReport { return p.node.outputReport(p.id) }

// The version is memoized against a key, as a Struct output's is: it bumps
// once per change to what the output is made of, and never reads a value.
type eachOutput struct {
	key     string
	version int
	err     error
}

// What the output is made of: the fed inputs, the definition, and the runs' outputs.
func (r *SubgraphInstanceNode) outputKey(id string) string {
	version := r.runtime().version
	r.mu.Lock()
	inputs := maps.Clone(r.inputs)
	runs := slices.Clone(r.runs)
	r.mu.Unlock()

	var key strings.Builder
	for _, in := range slices.Sorted(maps.Keys(inputs)) {
		key.WriteString(in)
		key.WriteByte('=')
		key.WriteString(nodes.Identity(inputs[in]))
		key.WriteByte(';')
	}
	key.WriteByte('v')
	key.WriteString(strconv.FormatUint(version, 10))
	for _, run := range runs {
		key.WriteByte('|')
		if output, ok := run.graph.outputPort(id); ok && output.ConnectedSource() != nil {
			key.WriteString(nodes.Identity(output.ConnectedSource()))
		}
	}
	return key.String()
}

func (r *SubgraphInstanceNode) outputVersion(id string) int {
	key := r.outputKey(id)
	r.mu.Lock()
	defer r.mu.Unlock()
	memo, ok := r.outputs[id]
	switch {
	case !ok:
		return -1
	case memo.key == key:
		return memo.version
	}
	return memo.version + 1
}

func (r *SubgraphInstanceNode) remember(id, key string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.outputs == nil {
		r.outputs = map[string]*eachOutput{}
	}
	memo, ok := r.outputs[id]
	switch {
	case !ok:
		r.outputs[id] = &eachOutput{key: key, err: err}
	case memo.key != key:
		memo.key, memo.version, memo.err = key, memo.version+1, err
	default:
		memo.err = err
	}
}

func (r *SubgraphInstanceNode) outputValue(id string) any {
	runs, err := r.Runs()
	values := make([]any, len(runs))
	var sample nodes.OutputPort
	for i, run := range runs {
		output, ok := run.outputPort(id)
		if !ok || output.ConnectedSource() == nil {
			continue
		}
		sample = output.ConnectedSource()
		values[i], _ = nodes.DynamicAnyValue(nil, sample)
	}

	var value any
	if sample != nil {
		value, _ = nodes.DynamicArrayOf(sample, values)
	}
	r.remember(id, r.outputKey(id), err)
	return value
}

func (r *SubgraphInstanceNode) outputReport(id string) nodes.ExecutionReport {
	r.mu.Lock()
	runs := slices.Clone(r.runs)
	memo := r.outputs[id]
	r.mu.Unlock()

	var report nodes.ExecutionReport
	if memo != nil && memo.err != nil {
		report.Errors = append(report.Errors, memo.err.Error())
	}
	for i, run := range runs {
		output, ok := run.graph.outputPort(id)
		if !ok {
			continue
		}
		observable, ok := output.ConnectedSource().(nodes.ObservableExecution)
		if !ok {
			continue
		}
		item := observable.ExecutionReport()
		for _, e := range item.Errors {
			report.Errors = append(report.Errors, fmt.Sprintf("item %d: %s", i, e))
		}
		report.Steps = append(report.Steps, nodes.StepTiming{Label: "item " + strconv.Itoa(i), Duration: item.TotalTime})
		report.TotalTime += item.TotalTime
	}
	return report
}
