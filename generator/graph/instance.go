package graph

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"

	gsync "sync"
)

type Instance struct {
	typeFactory *refutil.TypeFactory

	movelVersion uint32
	nodeIDs      map[nodes.Node]string
	metadata     *sync.NestedSyncMap
	producers    map[string]nodes.NodeOutput[artifact.Artifact]
	producerLock gsync.Mutex
}

func New(typeFactory *refutil.TypeFactory) *Instance {
	return &Instance{
		typeFactory: typeFactory,

		nodeIDs:      make(map[nodes.Node]string),
		metadata:     sync.NewNestedSyncMap(),
		producers:    make(map[string]nodes.NodeOutput[artifact.Artifact]),
		movelVersion: 0,
	}
}
func (i *Instance) ModelVersion() uint32 {
	return i.movelVersion
}

func (i *Instance) incModelVersion() {
	// TODO: Make thread safe
	i.movelVersion++
}

func (i *Instance) NodeInstanceSchema(node nodes.Node) schema.NodeInstance {
	var metadata map[string]any
	metadataPath := "nodes." + i.nodeIDs[node]

	if i.metadata.PathExists(metadataPath) {
		if data := i.metadata.Get(metadataPath); data != nil {
			metadata = data.(map[string]any)
		}
	}

	nodeInstance := schema.NodeInstance{
		Name:         "Unamed",
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]schema.NodeDependency, 0),
		Version:      node.Version(),
		Metadata:     metadata,
	}

	for _, subDependency := range node.Dependencies() {
		nodeInstance.Dependencies = append(nodeInstance.Dependencies, schema.NodeDependency{
			DependencyID:   i.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	if param, ok := node.(Parameter); ok {
		nodeInstance.Name = param.DisplayName()
		nodeInstance.Parameter = param.Schema()
	} else {
		named, ok := node.(nodes.Named)
		if ok {
			nodeInstance.Name = named.Name()
		}
	}

	return nodeInstance
}

func (i *Instance) addType(v any) {
	if !i.typeFactory.TypeRegistered(v) {
		i.typeFactory.RegisterType(v)
	}
}

func (i *Instance) buildIDsForNode(node nodes.Node) {

	// IDs for this node has already been built.
	if _, ok := i.nodeIDs[node]; ok {
		return
	}

	i.addType(node)

	for _, subDependency := range node.Dependencies() {
		i.buildIDsForNode(subDependency.Dependency())
	}

	// TODO: UGLY UGLY UGLY UGLY
	highestInded := len(i.nodeIDs)
	for {
		id := fmt.Sprintf("Node-%d", highestInded)

		taken := false
		for _, usedId := range i.nodeIDs {
			if usedId == id {
				taken = true
			}
			if taken {
				break
			}
		}
		if !taken {
			i.nodeIDs[node] = id
			break
		}
		highestInded++
	}
}

func (i *Instance) ApplyAppSchema(jsonPayload []byte) error {
	appSchema, err := jbtf.Unmarshal[schema.App](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}

	decoder, err := jbtf.NewDecoder(jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}

	i.nodeIDs = make(map[nodes.Node]string)
	i.metadata = sync.NewNestedSyncMap()
	i.metadata.OverwriteData(appSchema.Metadata)

	createdNodes := make(map[string]nodes.Node)

	// Create the Nodes
	for nodeID, instanceDetails := range appSchema.Nodes {
		if nodeID == "" {
			panic("attempting to create a node without an ID")
		}
		newNode := i.typeFactory.New(instanceDetails.Type)
		casted, ok := newNode.(nodes.Node)
		if !ok {
			panic(fmt.Errorf("graph definition contained type that instantiated a non node: %s", instanceDetails.Type))
		}
		createdNodes[nodeID] = casted
		i.nodeIDs[casted] = nodeID
	}

	// Connect the nodes we just created
	for nodeID, instanceDetails := range appSchema.Nodes {
		node := createdNodes[nodeID]
		for _, dependency := range instanceDetails.Dependencies {

			outNode := createdNodes[dependency.DependencyID]
			outPortVals := refutil.CallFuncValuesOfType(outNode, dependency.DependencyPort)
			ref := outPortVals[0].(nodes.NodeOutputReference)

			node.SetInput(dependency.Name, nodes.Output{
				NodeOutput: ref,
			})
		}
	}

	// Set the Producers
	i.producers = make(map[string]nodes.NodeOutput[artifact.Artifact])
	for fileName, producerDetails := range appSchema.Producers {
		producerNode := createdNodes[producerDetails.NodeID]
		outPortVals := refutil.CallFuncValuesOfType(producerNode, producerDetails.Port)
		ref := outPortVals[0].(nodes.NodeOutput[artifact.Artifact])
		if ref == nil {
			panic(fmt.Errorf("REF IS NIL FOR FILE %s (node id: %s) and port %s", fileName, producerDetails.NodeID, producerDetails.Port))
		}
		i.producers[fileName] = ref
	}

	// Set Parameters
	for nodeID, instanceDetails := range appSchema.Nodes {
		nodeI := createdNodes[nodeID]
		if p, ok := nodeI.(CustomGraphSerialization); ok {
			err := p.FromJSON(decoder, instanceDetails.Data)
			if err != nil {
				return err
			}
		}
	}

	i.incModelVersion()

	return nil
}

func (i *Instance) Schema() schema.GraphInstance {
	var noteMetadata map[string]any
	if notes := i.metadata.Get("notes"); notes != nil {
		casted, ok := notes.(map[string]any)
		if ok {
			noteMetadata = casted
		}
	}

	appSchema := schema.GraphInstance{
		Producers: make(map[string]schema.Producer),
		Notes:     noteMetadata,
	}

	appNodeSchema := make(map[string]schema.NodeInstance)

	for node := range i.nodeIDs {
		id, ok := i.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := appNodeSchema[id]; ok {
			panic("not sure how this happened")
		}

		appNodeSchema[id] = i.NodeInstanceSchema(node)
	}

	for key, producer := range i.producers {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := i.nodeIDs[producer.Node()]
		node := appNodeSchema[id]
		node.Name = key
		appNodeSchema[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.Port(),
		}
	}

	appSchema.Nodes = appNodeSchema

	registeredTypes := i.typeFactory.Types()
	nodeTypes := make([]schema.NodeType, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		nodeInstance, ok := i.typeFactory.New(registeredType).(nodes.Node)
		if !ok {
			panic("Registered type is somehow not a node. Not sure how we got here :/")
		}
		if nodeInstance == nil {
			panic("New registered type")
		}
		// log.Printf("%T: %+v\n", nodeInstance, nodeInstance)
		b := BuildNodeTypeSchema(nodeInstance)
		b.Type = registeredType
		nodeTypes = append(nodeTypes, b)
	}
	appSchema.Types = nodeTypes

	return appSchema
}

func (i *Instance) EncodeToAppSchema(appSchema *schema.App, encoder *jbtf.Encoder) {
	nodeInstances := make(map[string]schema.AppNodeInstance)
	for node := range i.nodeIDs {
		id, ok := i.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := nodeInstances[id]; ok {
			panic(fmt.Errorf("we've arrived to a invalid state. two nodes refer to the same ID. There's a bug somewhere"))
		}

		nodeInstances[id] = i.buildNodeGraphInstanceSchema(node, encoder)
	}

	if appSchema.Producers == nil {
		appSchema.Producers = make(map[string]schema.Producer)
	}
	for key, producer := range i.producers {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := i.nodeIDs[producer.Node()]
		node := nodeInstances[id]
		nodeInstances[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.Port(),
		}
	}
	appSchema.Nodes = nodeInstances

	// TODO: Is this unsafe? Yes.
	appSchema.Metadata = i.metadata.Data()
}

func (i *Instance) buildNodeGraphInstanceSchema(node nodes.Node, encoder *jbtf.Encoder) schema.AppNodeInstance {

	nodeInstance := schema.AppNodeInstance{
		Type:         refutil.GetTypeWithPackage(node),
		Dependencies: make([]schema.NodeDependency, 0),
	}

	for _, subDependency := range node.Dependencies() {
		nodeInstance.Dependencies = append(nodeInstance.Dependencies, schema.NodeDependency{
			DependencyID:   i.nodeIDs[subDependency.Dependency()],
			DependencyPort: subDependency.DependencyPort(),
			Name:           subDependency.Name(),
		})
	}

	sort.Slice(nodeInstance.Dependencies, func(i, j int) bool {
		return strings.ToLower(nodeInstance.Dependencies[i].Name) < strings.ToLower(nodeInstance.Dependencies[j].Name)
	})

	if param, ok := node.(CustomGraphSerialization); ok {
		data, err := param.ToJSON(encoder)
		if err != nil {
			panic(err)
		}
		nodeInstance.Data = data
	}

	return nodeInstance
}

// NODES ======================================================================

func (i *Instance) NodeId(node nodes.Node) string {
	return i.nodeIDs[node]
}

func (i *Instance) Node(nodeId string) nodes.Node {
	for n, id := range i.nodeIDs {
		if id == nodeId {
			return n
		}
	}
	panic(fmt.Errorf("no node exists with id %q", nodeId))
}

func (i *Instance) CreateNode(nodeType string) (nodes.Node, string, error) {
	if !i.typeFactory.KeyRegistered(nodeType) {
		return nil, "", fmt.Errorf("no factory registered with ID %s", nodeType)
	}

	newNode := i.typeFactory.New(nodeType)
	casted, ok := newNode.(nodes.Node)
	if !ok {
		panic(fmt.Errorf("Regiestered type did not create a node. How'd ya manage that: %s", nodeType))
	}
	i.buildIDsForNode(casted)

	return casted, i.nodeIDs[casted], nil
}

func (i *Instance) DeleteNode(nodeId string) {
	var nodeToDelete nodes.Node

	for n, id := range i.nodeIDs {
		if id == nodeId {
			nodeToDelete = n
		}
	}

	for filename, producer := range i.producers {
		if i.nodeIDs[producer.Node()] == nodeId {
			delete(i.producers, filename)
		}
	}

	delete(i.nodeIDs, nodeToDelete)
}

// PARAMETER ==================================================================

func (i *Instance) getParameters() []Parameter {
	if i.producers == nil {
		return nil
	}

	parameterSet := make(map[Parameter]struct{})
	for _, n := range i.producers {
		params := RecurseDependenciesType[Parameter](n.Node())
		for _, p := range params {
			parameterSet[p] = struct{}{}
		}
	}

	uniqueParams := make([]Parameter, 0, len(parameterSet))
	for p := range parameterSet {
		uniqueParams = append(uniqueParams, p)
	}

	return uniqueParams
}

func (i *Instance) InitializeParameters(set *flag.FlagSet) {
	for _, p := range i.getParameters() {
		p.InitializeForCLI(set)
	}
}

func (i *Instance) Parameter(nodeId string) Parameter {
	node := i.Node(nodeId)

	param, ok := node.(Parameter)
	if !ok {
		panic(fmt.Errorf("node %q is not a parameter", nodeId))
	}

	return param
}

func (i *Instance) UpdateParameter(nodeId string, data []byte) (bool, error) {
	i.producerLock.Lock()
	defer i.producerLock.Unlock()

	r, err := i.Parameter(nodeId).ApplyMessage(data)
	i.incModelVersion()
	return r, err
}

func (i *Instance) ParameterData(nodeId string) []byte {
	i.producerLock.Lock()
	defer i.producerLock.Unlock()
	return i.Parameter(nodeId).ToMessage()
}

// METADATA ===================================================================

func (i *Instance) SetMetadata(key string, value any) {
	i.metadata.Set(key, value)
}

func (i *Instance) DeleteMetadata(key string) {
	i.metadata.Delete(key)
}

// CONNECTIONS ================================================================

func (i *Instance) DeleteNodeInputConnection(nodeId, portName string) {
	i.Node(nodeId).SetInput(
		portName,
		nodes.Output{
			NodeOutput: nil,
		},
	)
	i.incModelVersion()
}

func (i *Instance) ConnectNodes(nodeOutId, outPortName, nodeInId, inPortName string) {
	inNode := i.Node(nodeInId)
	outNode := i.Node(nodeOutId)
	outPortVals := refutil.CallFuncValuesOfType(outNode, outPortName)

	ref := outPortVals[0].(nodes.NodeOutputReference)
	inNode.SetInput(
		inPortName,
		nodes.Output{
			NodeOutput: ref,
		},
	)
	i.incModelVersion()
}

// PRODUCERS ==================================================================

func (i *Instance) SetNodeAsProducer(nodeId, producerName string) {
	producerNode := i.Node(nodeId)

	if producerNode == nil {
		panic(fmt.Errorf("no node exists with id %q", nodeId))
	}

	// TODO: We need to allow users to specify which output port
	// that is the actuall artifact. can't rely on "Out"
	outPortVals := refutil.CallFuncValuesOfType(producerNode, "Out")
	ref := outPortVals[0].(nodes.NodeOutput[artifact.Artifact])
	if ref == nil {
		panic(fmt.Errorf("Couldn't find Out port on Node: %s", nodeId))
	}

	// We need to check and remove previous references...
	for filename, producer := range i.producers {

		if i.NodeId(producer.Node()) != nodeId {
			continue
		}

		// TODO: This changes once we allow multiple output
		// port artifact. Need to specify port instead of "Out"
		if producer.Port() != "Out" {
			continue
		}

		delete(i.producers, filename)
	}

	i.producers[producerName] = ref
	i.incModelVersion()
}

func (i *Instance) recursivelyRegisterNodeTypes(node nodes.Node) {
	i.addType(node)
	for _, subDependency := range node.Dependencies() {
		i.recursivelyRegisterNodeTypes(subDependency.Dependency())
	}
}

func (i *Instance) Artifact(producerName string) artifact.Artifact {
	producer, ok := i.producers[producerName]
	if !ok {
		panic(fmt.Errorf("no producer registered for: %s", producerName))
	}

	i.producerLock.Lock()
	defer i.producerLock.Unlock()

	return producer.Value()
}

func (i *Instance) AddProducer(producerName string, producer nodes.NodeOutput[artifact.Artifact]) {
	i.recursivelyRegisterNodeTypes(producer.Node())
	i.buildIDsForNode(producer.Node())
	i.producers[producerName] = producer
}

func (i *Instance) Producer(producerName string) nodes.NodeOutput[artifact.Artifact] {
	return i.producers[producerName]
}

func (i *Instance) ProducerNames() []string {
	names := make([]string, 0, len(i.producers))

	for name := range i.producers {
		names = append(names, name)
	}

	return names
}

// REFLECTION =================================================================

func RecurseDependenciesType[T any](dependent nodes.Dependent) []T {
	allDependencies := make([]T, 0)
	for _, dep := range dependent.Dependencies() {
		subDependent := dep.Dependency()
		subDependencies := RecurseDependenciesType[T](subDependent)
		allDependencies = append(allDependencies, subDependencies...)

		ofT, ok := subDependent.(T)
		if ok {
			allDependencies = append(allDependencies, ofT)
		}
	}

	return allDependencies
}

func BuildNodeTypeSchema(node nodes.Node) schema.NodeType {

	typeSchema := schema.NodeType{
		DisplayName: "Untyped",
		Outputs:     make([]schema.NodeOutput, 0),
		Inputs:      make(map[string]schema.NodeInput),
	}

	outputs := node.Outputs()
	for _, o := range outputs {
		typeSchema.Outputs = append(typeSchema.Outputs, schema.NodeOutput{
			Name: o.NodeOutput.Port(),
			Type: o.Type,
		})
	}

	inputs := node.Inputs()
	for _, o := range inputs {
		typeSchema.Inputs[o.Name] = schema.NodeInput{
			Type:    o.Type,
			IsArray: o.Array,
		}
	}

	if param, ok := node.(Parameter); ok {
		typeSchema.Parameter = param.Schema()
	}

	if typed, ok := node.(nodes.Typed); ok {
		typeSchema.DisplayName = typed.Type()
	} else {
		typeSchema.DisplayName = refutil.GetTypeName(node)
	}

	if pathed, ok := node.(nodes.Pathed); ok {
		typeSchema.Path = pathed.Path()
	} else {
		packagePath := refutil.GetPackagePath(node)
		if strings.Contains(packagePath, "/") {
			path := strings.Split(packagePath, "/")
			path = path[1:]
			if path[0] == "EliCDavis" {
				path = path[1:]
			}

			if path[0] == "polyform" {
				path = path[1:]
			}
			typeSchema.Path = strings.Join(path, "/")
		} else {
			typeSchema.Path = packagePath
		}
	}

	if described, ok := node.(nodes.Describable); ok {
		typeSchema.Info = described.Description()
	}

	return typeSchema
}
