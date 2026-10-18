package graph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
)

type ReplacedInstance struct {
	KeptInputs  []string
	KeptOutputs []string
	LostInputs  []string
	LostOutputs []string
	Dropped     []DroppedEdge
}

type outboundEdge struct {
	consumer nodes.Node
	input    string
	index    int
	output   string
}

// The id stays; an edge stays when both placements have a port of that
// name with a fitting type.
func (a *Graph) ReplaceSubgraphInstance(nodeID, subGraphID string) (result ReplacedInstance, err error) {
	result.Dropped, err = a.project.write(dropConflicts, func() error {
		return a.replaceSubgraphInstance(nodeID, subGraphID, &result)
	})
	return result, err
}

func (a *Graph) replaceSubgraphInstance(nodeID, subGraphID string, result *ReplacedInstance) error {
	node, ok := a.nodesByID[nodeID]
	if !ok {
		return fmt.Errorf("no node exists with id %q", nodeID)
	}
	old, ok := node.(*SubgraphInstanceNode)
	if !ok {
		return fmt.Errorf("node %q is not a subgraph instance", nodeID)
	}
	if _, exists := a.project.subGraphs[subGraphID]; !exists {
		return fmt.Errorf("sub-graph %q does not exist", subGraphID)
	}
	if a.scope != "" && a.project.wouldContainItself(a.scope, subGraphID) {
		return refuseSelfContainment(a.scope, subGraphID)
	}
	if old.subGraphID == subGraphID {
		return fmt.Errorf("node %q is already an instance of %q", nodeID, subGraphID)
	}

	typePath := subgraph.RuntimeTypePath(subGraphID)
	fresh, ok := a.project.typeFactory.New(typePath).(*SubgraphInstanceNode)
	if !ok {
		return fmt.Errorf("sub-graph %q has no registered node type", subGraphID)
	}

	oldInputs, oldOutputs := old.Inputs(), old.Outputs()
	inbound := a.reads[old]
	var outbound []outboundEdge
	for _, consumerID := range slices.Sorted(maps.Keys(a.nodesByID)) {
		consumer := a.nodesByID[consumerID]
		if consumer == old {
			continue
		}
		for _, input := range slices.Sorted(maps.Keys(a.reads[consumer])) {
			for i, from := range a.reads[consumer][input] {
				if from.node == old {
					outbound = append(outbound, outboundEdge{consumer: consumer, input: input, index: i, output: from.port})
				}
			}
		}
	}
	type producer struct{ name, port string }
	var producers []producer
	for name, entry := range a.project.namedManifests.namedPorts {
		if entry.node == old {
			producers = append(producers, producer{name, entry.portName})
		}
	}

	a.project.namedManifests.DeleteNode(old)
	a.forget(old)
	a.register(fresh, nodeID, typePath)
	a.reads[fresh] = map[string][]source{}
	a.dirty()

	for _, input := range slices.Sorted(maps.Keys(inbound)) {
		name := label(oldInputs[input])
		var failed error
		for _, from := range inbound[input] {
			if failed = a.connectNodes(a.nodeIDs[from.node], from.port, nodeID, name); failed != nil {
				break
			}
		}
		if failed != nil {
			result.LostInputs = append(result.LostInputs, fmt.Sprintf("%s: %v", name, failed))
			continue
		}
		result.KeptInputs = append(result.KeptInputs, name)
	}

	for _, o := range outbound {
		consumerID := a.nodeIDs[o.consumer]
		target := o.input
		if _, isArray := o.consumer.Inputs()[o.input].(nodes.ArrayValueInputPort); isArray {
			target = fmt.Sprintf("%s.%d", o.input, o.index)
		}
		name := label(oldOutputs[o.output])
		label := fmt.Sprintf("%s -> %s.%s", name, consumerID, target)
		if e := a.connectNodes(nodeID, name, consumerID, target); e != nil {
			result.LostOutputs = append(result.LostOutputs, fmt.Sprintf("%s: %v", label, e))
			continue
		}
		result.KeptOutputs = append(result.KeptOutputs, label)
	}
	for consumer, inputs := range a.reads {
		for input, sources := range inputs {
			kept := slices.DeleteFunc(slices.Clone(sources), func(from source) bool { return from.node == old })
			if len(kept) != len(sources) {
				a.setSources(consumer, input, kept)
			}
		}
	}

	freshOutputs := fresh.Outputs()
	for _, p := range producers {
		key, _ := outputKey(freshOutputs, label(oldOutputs[p.port]))
		if port, ok := freshOutputs[key].(nodes.Output[manifest.Manifest]); ok {
			a.project.namedManifests.NamePort(p.name, key, fresh, port)
			continue
		}
		result.LostOutputs = append(result.LostOutputs, fmt.Sprintf("producer %q read %s, which %q does not produce a manifest on", p.name, label(oldOutputs[p.port]), subGraphID))
	}
	return nil
}
