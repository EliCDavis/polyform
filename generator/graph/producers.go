package graph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

func (a *Instance) SetNodeAsProducer(nodeId, nodePort, producerName string) error {
	_, err := a.write(refuseConflicts, func() error {
		node, ok := a.nodesByID[nodeId]
		if !ok {
			return fmt.Errorf("no node exists with id %q", nodeId)
		}
		outputs := node.Outputs()
		key, ok := outputKey(outputs, nodePort)
		if !ok {
			return fmt.Errorf("node %q does not contain output %q", nodeId, nodePort)
		}
		producer, ok := outputs[key].(nodes.Output[manifest.Manifest])
		if !ok {
			return fmt.Errorf("node %q output %q does not produce artifacts", nodeId, nodePort)
		}
		a.namedManifests.NamePort(producerName, key, node, producer)
		return nil
	})
	return err
}

func (a *Instance) AddProducer(producerName string, producer nodes.Output[manifest.Manifest]) {
	a.write(refuseConflicts, func() error {
		a.registerWithInputs(producer.Node(), "")
		a.namedManifests.NamePort(producerName, producer.Name(), producer.Node(), producer)
		return nil
	})
}

// Reading a value has to go through here: evaluating can build a placement's copy.
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

func (a *Instance) ProducerNames() []string {
	return slices.Collect(maps.Keys(a.namedManifests.namedPorts))
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
