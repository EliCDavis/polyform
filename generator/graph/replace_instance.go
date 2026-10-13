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
}

type outboundEdge struct {
	consumer nodes.Node
	input    string
	index    int
	output   string
}

// The id stays; an edge stays when both placements have its port with a fitting type.
func (a *Graph) ReplaceSubgraphInstance(nodeID, subGraphID string) (result ReplacedInstance, err error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	node, ok := a.nodesByID[nodeID]
	if !ok {
		return result, fmt.Errorf("no node exists with id %q", nodeID)
	}
	old, ok := node.(*SubgraphInstanceNode)
	if !ok {
		return result, fmt.Errorf("node %q is not a subgraph instance", nodeID)
	}
	if _, exists := a.project.subGraphs[subGraphID]; !exists {
		return result, fmt.Errorf("sub-graph %q does not exist", subGraphID)
	}
	if scope := a.SubGraphScopeID(); scope != "" && a.project.wouldContainItself(scope, subGraphID) {
		return result, refuseSelfContainment(scope, subGraphID)
	}
	if old.subGraphID == subGraphID {
		return result, fmt.Errorf("node %q is already an instance of %q", nodeID, subGraphID)
	}

	typePath := subgraph.RuntimeTypePath(subGraphID)
	fresh, ok := a.project.typeFactory.New(typePath).(*SubgraphInstanceNode)
	if !ok {
		return result, fmt.Errorf("sub-graph %q has no registered node type", subGraphID)
	}

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

	end := a.beginCompoundEdit()
	defer func() {
		if endErr := end(); endErr != nil && err == nil {
			err = endErr
		}
	}()

	a.project.namedManifests.DeleteNode(old)
	a.forget(old)
	a.register(fresh, nodeID, typePath)
	a.reads[fresh] = map[string][]source{}
	a.deferredToCompoundEdit()

	freshInputs := fresh.Inputs()
	for _, input := range slices.Sorted(maps.Keys(inbound)) {
		if _, has := freshInputs[input]; !has {
			result.LostInputs = append(result.LostInputs, fmt.Sprintf("%s: %q has no input named %q", input, subGraphID, input))
			continue
		}
		var failed error
		for _, from := range inbound[input] {
			if failed = a.connectNodes(a.nodeIDs[from.node], from.port, nodeID, input); failed != nil {
				break
			}
		}
		if failed != nil {
			result.LostInputs = append(result.LostInputs, fmt.Sprintf("%s: %v", input, failed))
			continue
		}
		result.KeptInputs = append(result.KeptInputs, input)
	}

	freshOutputs := fresh.Outputs()
	for _, o := range outbound {
		consumerID := a.nodeIDs[o.consumer]
		target := o.input
		if _, isArray := o.consumer.Inputs()[o.input].(nodes.ArrayValueInputPort); isArray {
			target = fmt.Sprintf("%s.%d", o.input, o.index)
		}
		label := fmt.Sprintf("%s -> %s.%s", o.output, consumerID, target)
		if _, has := freshOutputs[o.output]; !has {
			result.LostOutputs = append(result.LostOutputs, fmt.Sprintf("%s: %q has no output named %q", label, subGraphID, o.output))
			continue
		}
		if e := a.connectNodes(nodeID, o.output, consumerID, target); e != nil {
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

	for _, p := range producers {
		if port, ok := freshOutputs[p.port].(nodes.Output[manifest.Manifest]); ok {
			a.project.namedManifests.NamePort(p.name, p.port, fresh, port)
			continue
		}
		result.LostOutputs = append(result.LostOutputs, fmt.Sprintf("producer %q read %s, which %q does not produce a manifest on", p.name, p.port, subGraphID))
	}

	a.clearCopySource()
	a.incModelVersion()
	return result, nil
}
