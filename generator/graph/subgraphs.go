package graph

import (
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

type subGraphRuntime struct {
	id          string
	name        string
	description string
	instance    *Instance
}

type BoundaryPort struct {
	Name string
	Type string
	Kind string // "input" | "output"
}

func (a *Instance) initSubGraphs() {
	if a.subGraphs == nil {
		a.subGraphs = make(map[string]*subGraphRuntime)
	}
}

func newInstance(parent *Instance) *Instance {
	return &Instance{
		typeFactory:     parent.typeFactory,
		variableFactory: parent.variableFactory,
		nodeIDs:         make(map[nodes.Node]string),
		variables:       variable.NewSystem(),
		profiles:        make(map[string]variable.Profile),
		metadata:        sync.NewNestedSyncMap(),
		namedManifests: &namedOutputManager[manifest.Manifest]{
			namedPorts: make(map[string]namedOutputEntry[manifest.Manifest]),
		},
		parent: parent,
	}
}

func (a *Instance) assertRootGraph(operationId string) error {
	if !a.IsRoot() {
		return fmt.Errorf("operation only allowed on root graph instance: %s", operationId)
	}
	return nil
}

func (a *Instance) IsRoot() bool {
	return a.parent == nil
}

func (a *Instance) Root() *Instance {
	if a.parent == nil {
		return a
	}
	return a.parent.Root()
}

func (a *Instance) IsSubGraphScope() bool {
	return a.parent != nil
}

func (a *Instance) CreateSubGraph(id, name, description string) error {
	if err := a.assertRootGraph("create subgraph"); err != nil {
		return err
	}
	a.initSubGraphs()

	if _, exists := a.subGraphs[id]; exists {
		return fmt.Errorf("sub-graph %q already exists", id)
	}

	child := newInstance(a)
	a.subGraphs[id] = &subGraphRuntime{
		id:          id,
		name:        name,
		description: description,
		instance:    child,
	}

	a.RegisterSubGraphNodeType(id)
	a.incModelVersion()
	return nil
}

func (a *Instance) DeleteSubGraph(id string) error {
	if err := a.assertRootGraph("delete subgraph"); err != nil {
		return err
	}
	a.initSubGraphs()

	if _, exists := a.subGraphs[id]; !exists {
		return fmt.Errorf("sub-graph %q does not exist", id)
	}

	if a.countSubGraphNodeInstances(id) > 0 {
		return fmt.Errorf("sub-graph %q is still referenced by %d node instance(s)", id, a.countSubGraphNodeInstances(id))
	}

	typePath := subgraph.RuntimeTypePath(id)
	a.typeFactory.Unregister(typePath)
	delete(a.subGraphs, id)
	a.incModelVersion()
	return nil
}

func (a *Instance) countSubGraphNodeInstances(subGraphID string) int {
	count := 0
	for node := range a.nodeIDs {
		runtime, ok := node.(*InstanceNode)
		if ok && runtime.subGraphID == subGraphID {
			count++
		}
	}
	return count
}

func (a *Instance) SetSubGraphInfo(id, name, description string) error {
	if err := a.assertRootGraph("update subgraph info"); err != nil {
		return err
	}
	a.initSubGraphs()

	runtime, exists := a.subGraphs[id]
	if !exists {
		return fmt.Errorf("sub-graph %q does not exist", id)
	}
	runtime.name = name
	runtime.description = description
	a.incModelVersion()
	return nil
}

func (a *Instance) SubGraphInstance(id string) (*Instance, error) {
	if err := a.assertRootGraph("fetch subgraph"); err != nil {
		return nil, err
	}
	a.initSubGraphs()

	runtime, exists := a.subGraphs[id]
	if !exists {
		return nil, fmt.Errorf("sub-graph %q does not exist", id)
	}
	return runtime.instance, nil
}

func (a *Instance) RegisterSubGraphNodeType(subGraphID string) (string, error) {
	if err := a.assertRootGraph("register subgraph node type"); err != nil {
		return "", err
	}
	typePath := subgraph.RuntimeTypePath(subGraphID)
	a.typeFactory.RegisterBuilder(typePath, func() any {
		return NewRuntimeNode(a, subGraphID)
	})
	return typePath, nil
}

func (a *Instance) CollectBoundaryPorts(subGraphID string) ([]BoundaryPort, error) {
	child, err := a.SubGraphInstance(subGraphID)
	if err != nil {
		return nil, err
	}

	ports := make([]BoundaryPort, 0)
	seen := make(map[string]struct{})

	for node := range child.nodeIDs {
		boundary, ok := subgraph.IsBoundaryNode(node)
		if !ok {
			continue
		}

		if portType := boundary.BoundaryPortType(); portType == "" {
			continue
		}

		kind := "output"
		if _, isInput := subgraph.IsInputBoundary(node); isInput {
			kind = "input"
		}

		name := boundary.BoundaryPortName()
		seenKey := kind + "/" + name
		if _, dup := seen[seenKey]; dup {
			continue
		}
		seen[seenKey] = struct{}{}

		ports = append(ports, BoundaryPort{
			Name: name,
			Type: boundary.BoundaryPortType(),
			Kind: kind,
		})
	}

	return ports, nil
}

func (a *Instance) refreshSubGraphNodeType(subGraphID string) {
	a.RegisterSubGraphNodeType(subGraphID)
	a.incModelVersion()
}

func (a *Instance) onSubGraphChildMutation(subGraphID string) {
	a.Root().refreshSubGraphNodeType(subGraphID)
}

func (a *Instance) runtimeSubGraphSchema(id string) (schema.RuntimeSubGraphDefinition, error) {
	if err := a.assertRootGraph("fetch subgraph schema"); err != nil {
		return schema.RuntimeSubGraphDefinition{}, err
	}
	a.initSubGraphs()

	runtime, exists := a.subGraphs[id]
	if !exists {
		return schema.RuntimeSubGraphDefinition{}, fmt.Errorf("sub-graph %q does not exist", id)
	}

	childSchema := runtime.instance.Schema()
	return schema.RuntimeSubGraphDefinition{
		Name:        runtime.name,
		Description: runtime.description,
		Nodes:       childSchema.Nodes,
		Notes:       childSchema.Notes,
		Variables:   childSchema.Variables,
	}, nil
}

func (a *Instance) encodeSubGraphDefinitions(encoder *jbtf.Encoder) (map[string]schema.SubGraphDefinition, error) {
	if err := a.assertRootGraph("encode subgraph definition"); err != nil {
		return nil, err
	}
	a.initSubGraphs()

	result := make(map[string]schema.SubGraphDefinition, len(a.subGraphs))
	for id, runtime := range a.subGraphs {
		child := runtime.instance
		nodeInstances := make(map[string]schema.AppNodeInstance)
		for node := range child.nodeIDs {
			nodeID := child.nodeIDs[node]
			nodeInstances[nodeID] = child.buildNodeGraphInstanceSchema(node, encoder)
		}

		variableSchema, err := child.variables.PersistedSchema(encoder)
		if err != nil {
			panic(err)
		}

		var noteMetadata map[string]any
		if notes := child.metadata.Get("notes"); notes != nil {
			if casted, ok := notes.(map[string]any); ok {
				noteMetadata = casted
			}
		}

		result[id] = schema.SubGraphDefinition{
			Name:        runtime.name,
			Description: runtime.description,
			Nodes:       nodeInstances,
			Notes:       noteMetadata,
			Variables:   variableSchema,
			Metadata:    child.metadata.Data(),
		}
	}
	return result, nil
}
