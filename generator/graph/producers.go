package graph

import (
	"fmt"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Instance) SetNodeAsProducer(nodeId, nodePort, producerName string) error {
	a.mu().Lock()
	defer a.mu().Unlock()

	node, ok := a.nodesByID[nodeId]
	if !ok {
		return fmt.Errorf("no node exists with id %q", nodeId)
	}
	output, ok := node.Outputs()[nodePort]
	if !ok {
		return fmt.Errorf("node %q does not contain output %q", nodeId, nodePort)
	}
	producer, ok := output.(nodes.Output[manifest.Manifest])
	if !ok {
		return fmt.Errorf("node %q output %q does not produce artifacts", nodeId, nodePort)
	}

	a.namedManifests.NamePort(producerName, nodePort, node, producer)
	a.incModelVersion()
	return nil
}

func (a *Instance) AddProducer(producerName string, producer nodes.Output[manifest.Manifest]) {
	a.mu().Lock()
	defer a.mu().Unlock()

	a.registerWithInputs(producer.Node(), "")
	a.namedManifests.NamePort(producerName, producer.Name(), producer.Node(), producer)
}

// Evaluate runs f with the graph held still. Reading a node's value while
// the graph could be edited has to go through it: evaluating can build the
// copy a subgraph placement evaluates through.
func (a *Instance) Evaluate(f func()) {
	a.mu().Lock()
	defer a.mu().Unlock()
	f()
}

func (a *Instance) Manifest(producerName string) (result manifest.Manifest, err error) {
	a.Evaluate(func() {
		producer, ok := a.namedManifests.namedPorts[producerName]
		if !ok {
			err = fmt.Errorf("no producer registered for: %s", producerName)
			return
		}
		result = producer.port.Value()
	})
	return result, err
}

func (a *Instance) Producer(producerName string) nodes.Output[manifest.Manifest] {
	return a.namedManifests.namedPorts[producerName].port
}

func (a *Instance) ProducerNames() []string {
	names := make([]string, 0, len(a.namedManifests.namedPorts))
	for name := range a.namedManifests.namedPorts {
		names = append(names, name)
	}
	return names
}

func (a *Instance) IsPortNamed(node nodes.Node, portName string) (string, bool) {
	return a.namedManifests.IsPortNamed(node, portName)
}

func (a *Instance) producerSchema() map[string]schema.Producer {
	producers := make(map[string]schema.Producer, len(a.namedManifests.namedPorts))
	for name, producer := range a.namedManifests.namedPorts {
		producers[name] = schema.Producer{
			NodeID: a.nodeIDs[producer.node],
			Port:   producer.port.Name(),
		}
	}
	return producers
}
