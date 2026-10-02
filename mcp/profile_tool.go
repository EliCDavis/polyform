package mcp

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type ProfileGraphInput struct {
	NodeId   string  `json:"nodeId,omitempty" jsonschema:"the node whose output to evaluate and time. Omit to evaluate the first named producer (what generate writes)."`
	Port     string  `json:"port,omitempty" jsonschema:"output port on nodeId; defaults to its only output"`
	Scope    string  `json:"scope,omitempty"`
	Top      int     `json:"top,omitempty" jsonschema:"how many of the slowest individual node executions to list, default 15"`
	MinShare float64 `json:"minShare,omitempty" jsonschema:"prune flame-tree scopes below this share of the total self time, default 0.01 (1%)"`
}

type ProfileTypeTotal struct {
	Type  string  `json:"type" jsonschema:"node type, package-qualified but without the nodes.Struct wrapper"`
	Count int     `json:"count" jsonschema:"executions of this type that ran (across every instance)"`
	Self  string  `json:"self" jsonschema:"summed self time"`
	Share float64 `json:"share" jsonschema:"fraction of the total self time"`
}

type ProfileNodeTotal struct {
	Path  string `json:"path" jsonschema:"instance path to the node: root node id, or outer/inner/.../nodeId through subgraph instances"`
	Type  string `json:"type"`
	Self  string `json:"self"`
	Total string `json:"total" jsonschema:"self plus time spent waiting on the node's inputs"`
}

type ProfileGraphOutput struct {
	Evaluated  string             `json:"evaluated" jsonschema:"nodeId.port that was evaluated"`
	WallTime   string             `json:"wallTime" jsonschema:"time the evaluation call took. Nodes already cached from an earlier evaluation contribute their remembered timings but no wall time, so a small wall time next to a large total means the graph was warm."`
	TotalSelf  string             `json:"totalSelf" jsonschema:"sum of every executed node's self time - the graph's real compute cost"`
	Executions int                `json:"executions" jsonschema:"node outputs that have run at least once"`
	Flame      []string           `json:"flame" jsonschema:"the flame tree, one line per scope, depth-first: indented instance path, inclusive time, share, then self time of the nodes directly in that scope. Scopes under minShare are folded into their parent."`
	ByType     []ProfileTypeTotal `json:"byType" jsonschema:"self time summed by node type, largest first - answers 'is it the CSG?'"`
	Slowest    []ProfileNodeTotal `json:"slowest" jsonschema:"the top individual node executions by self time"`
}

type profileNode struct {
	path  string
	typ   string
	self  time.Duration
	total time.Duration
}

type profileScope struct {
	path     string
	subgraph string
	self     time.Duration
	children []*profileScope
}

func (p *profileScope) inclusive() time.Duration {
	sum := p.self
	for _, c := range p.children {
		sum += c.inclusive()
	}
	return sum
}

func (s *Server) profileGraph(ctx context.Context, req *mcpsdk.CallToolRequest, in ProfileGraphInput) (*mcpsdk.CallToolResult, ProfileGraphOutput, error) {
	var out ProfileGraphOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		out, e = Profile(inst, in)
		return e
	})
	return nil, out, err
}

// Profile evaluates the requested output on inst and rolls up every
// executed node's timings; profile_graph is this over MCP.
func Profile(inst *graph.Instance, in ProfileGraphInput) (ProfileGraphOutput, error) {
	var out ProfileGraphOutput
	err := func() error {
		port, label, e := portToProfile(inst, in.NodeId, in.Port)
		if e != nil {
			return e
		}
		out.Evaluated = label

		start := time.Now()
		func() {
			defer func() {
				if r := recover(); r != nil {
					e = fmt.Errorf("evaluating %s: %v", label, r)
				}
			}()
			e = evaluatePort(port)
		}()
		if e != nil {
			return e
		}
		out.WallTime = time.Since(start).Round(time.Microsecond).String()

		root := &profileScope{}
		var executions []profileNode
		collectProfile(inst, "", root, &executions)

		type typeTotal struct {
			count int
			self  time.Duration
		}
		var total time.Duration
		byType := map[string]*typeTotal{}
		for _, n := range executions {
			total += n.self
			t, ok := byType[n.typ]
			if !ok {
				t = &typeTotal{}
				byType[n.typ] = t
			}
			t.count++
			t.self += n.self
		}
		out.TotalSelf = total.Round(time.Microsecond).String()
		out.Executions = len(executions)

		typeKeys := make([]string, 0, len(byType))
		for k := range byType {
			typeKeys = append(typeKeys, k)
		}
		sort.Slice(typeKeys, func(i, j int) bool { return byType[typeKeys[i]].self > byType[typeKeys[j]].self })
		for _, k := range typeKeys[:min(25, len(typeKeys))] {
			t := byType[k]
			share := 0.0
			if total > 0 {
				share = float64(t.self) / float64(total)
			}
			out.ByType = append(out.ByType, ProfileTypeTotal{
				Type:  k,
				Count: t.count,
				Self:  t.self.Round(time.Microsecond).String(),
				Share: share,
			})
		}

		sort.Slice(executions, func(i, j int) bool { return executions[i].self > executions[j].self })
		top := in.Top
		if top <= 0 {
			top = 15
		}
		for _, n := range executions[:min(top, len(executions))] {
			out.Slowest = append(out.Slowest, ProfileNodeTotal{
				Path:  n.path,
				Type:  n.typ,
				Self:  n.self.Round(time.Microsecond).String(),
				Total: n.total.Round(time.Microsecond).String(),
			})
		}

		minShare := in.MinShare
		if minShare <= 0 {
			minShare = 0.01
		}
		out.Flame = flameLines(root, total, minShare)
		return nil
	}()
	return out, err
}

func portToProfile(inst *graph.Instance, nodeID, portName string) (nodes.OutputPort, string, error) {
	if nodeID == "" {
		names := inst.ProducerNames()
		if len(names) == 0 {
			return nil, "", fmt.Errorf("no nodeId given and the graph has no producer; set_producer first or pass nodeId")
		}
		sort.Strings(names)
		producer := inst.Producer(names[0])
		return producer, fmt.Sprintf("producer %s", names[0]), nil
	}
	node := inst.Node(nodeID)
	if node == nil {
		return nil, "", fmt.Errorf("no node exists with id %q", nodeID)
	}
	outputs := node.Outputs()
	if portName == "" {
		if len(outputs) != 1 {
			return nil, "", fmt.Errorf("node %q has %d outputs (%s); pass port", nodeID, len(outputs), strings.Join(sortedOutputNames(outputs), ", "))
		}
		for name := range outputs {
			portName = name
		}
	}
	port, ok := outputs[resolveOutputPortName(node, portName)]
	if !ok {
		return nil, "", fmt.Errorf("node %q has no output port %q; it has %s", nodeID, portName, strings.Join(sortedOutputNames(outputs), ", "))
	}
	return port, nodeID + "." + portName, nil
}

func evaluatePort(port nodes.OutputPort) error {
	method := reflect.ValueOf(port).MethodByName("Value")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return fmt.Errorf("port %q does not expose a value", port.Name())
	}
	method.Call(nil)
	return nil
}

// collectProfile walks inst and every subgraph instance under it,
// recording each executed output's timings and building the scope tree.
func collectProfile(inst *graph.Instance, prefix string, scope *profileScope, executions *[]profileNode) {
	schema := inst.Schema()
	for _, id := range inst.NodeIds() {
		node := inst.Node(id)
		path := prefix + id
		typ := schema.Nodes[id].Type

		if sub, ok := node.(*graph.SubgraphInstanceNode); ok {
			child := &profileScope{path: path, subgraph: sub.SubGraphID()}
			scope.children = append(scope.children, child)
			collectProfile(sub.LiveGraph(), path+"/", child, executions)
			continue
		}

		var self, total time.Duration
		ran := false
		for _, port := range node.Outputs() {
			observable, ok := port.(nodes.ObservableExecution)
			if !ok {
				continue
			}
			// SelfTime is only set once the output has actually computed;
			// TotalTime can legitimately read 0 for a trivial node.
			report := observable.ExecutionReport()
			if report.SelfTime == nil {
				continue
			}
			ran = true
			total += report.TotalTime
			self += *report.SelfTime
		}
		if !ran {
			continue
		}
		scope.self += self
		*executions = append(*executions, profileNode{path: path, typ: shortTypeKey(typ), self: self, total: total})
	}
}

func flameLines(root *profileScope, total time.Duration, minShare float64) []string {
	var lines []string
	var walk func(scope *profileScope, depth int)
	walk = func(scope *profileScope, depth int) {
		inclusive := scope.inclusive()
		share := 0.0
		if total > 0 {
			share = float64(inclusive) / float64(total)
		}
		label := "(root)"
		if scope.path != "" {
			label = scope.path + " [" + scope.subgraph + "]"
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s  %.1f%%  self %s",
			strings.Repeat("  ", depth), label, inclusive.Round(time.Microsecond), share*100, scope.self.Round(time.Microsecond)))

		children := append([]*profileScope{}, scope.children...)
		sort.Slice(children, func(i, j int) bool { return children[i].inclusive() > children[j].inclusive() })
		var folded time.Duration
		foldedCount := 0
		for _, c := range children {
			ci := c.inclusive()
			if total > 0 && float64(ci)/float64(total) < minShare {
				folded += ci
				foldedCount++
				continue
			}
			walk(c, depth+1)
		}
		if foldedCount > 0 {
			lines = append(lines, fmt.Sprintf("%s(%d smaller instances)  %s", strings.Repeat("  ", depth+1), foldedCount, folded.Round(time.Microsecond)))
		}
	}
	walk(root, 0)
	return lines
}

func shortTypeKey(typeKey string) string {
	const wrapper = "github.com/EliCDavis/polyform/nodes.Struct["
	if strings.HasPrefix(typeKey, wrapper) && strings.HasSuffix(typeKey, "]") {
		typeKey = typeKey[len(wrapper) : len(typeKey)-1]
	}
	return strings.TrimPrefix(typeKey, "github.com/EliCDavis/polyform/")
}

func (s *Server) registerProfileGraphTool() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "profile_graph",
		Description: "Evaluate a node output (default: the producer) and report where the compute time went: a flame tree of subgraph instances with inclusive and self times, self time summed by node type, and the slowest individual node executions with their instance paths. Use it when a render or generate is slow before guessing - a 2000-node graph whose time is 90% in six csg nodes is a different fix from one spread across everything. Timings come from each node's last execution, so change something (or update a variable and set it back) before profiling to force a cold run.",
	}, s.profileGraph)
}
