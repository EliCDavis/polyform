package graph

import (
	"flag"
	"fmt"
	"strconv"
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
	producers    map[string]nodes.Output[artifact.Artifact]
	metadata     *sync.NestedSyncMap
	producerLock gsync.Mutex
}

func New(typeFactory *refutil.TypeFactory) *Instance {
	return &Instance{
		typeFactory: typeFactory,

		nodeIDs:      make(map[nodes.Node]string),
		metadata:     sync.NewNestedSyncMap(),
		producers:    make(map[string]nodes.Output[artifact.Artifact]),
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
		Name:          "Unamed",
		Type:          refutil.GetTypeWithPackage(node),
		AssignedInput: make(map[string]schema.PortReference),
		Output:        make(map[string]schema.NodeInstanceOutputPort),
		Metadata:      metadata,
	}

	for outputPortName, outputPort := range node.Outputs() {
		nodeInstance.Output[outputPortName] = schema.NodeInstanceOutputPort{
			Version: outputPort.Version(),
		}
	}

	for inputPortName, inputPort := range node.Inputs() {

		if single, ok := inputPort.(nodes.SingleValueInputPort); ok {
			val := single.Value()
			if val == nil {
				continue
			}

			nodeInstance.AssignedInput[inputPortName] = schema.PortReference{
				NodeId:   i.nodeIDs[val.Node()],
				PortName: val.Name(),
			}
		}

		if array, ok := inputPort.(nodes.ArrayValueInputPort); ok {
			for valIndex, val := range array.Value() {
				if val == nil {
					continue
				}

				nodeInstance.AssignedInput[fmt.Sprintf("%s.%d", inputPortName, valIndex)] = schema.PortReference{
					NodeId:   i.nodeIDs[val.Node()],
					PortName: val.Name(),
				}
			}
		}
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

	references := flattenNodeInputReferences(node)
	for _, ref := range references {
		i.buildIDsForNode(ref)
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

func (i *Instance) Reset() {
	i.nodeIDs = make(map[nodes.Node]string)
	i.metadata = sync.NewNestedSyncMap()
	i.producers = make(map[string]nodes.Output[artifact.Artifact])
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

	i.Reset()
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
		inputs := node.Inputs()
		for inputName, dependency := range instanceDetails.AssignedInput {
			input, ok := inputs[inputName]
			if !ok {
				panic(fmt.Errorf("Node %s has no input %s", nodeID, inputName))
			}

			outNode := createdNodes[dependency.NodeId]
			outNodeOutputs := outNode.Outputs()
			output, ok := outNodeOutputs[dependency.PortName]
			if !ok {
				panic(fmt.Errorf("Node %s has no output %s", dependency.NodeId, dependency.PortName))
			}

			if single, ok := input.(nodes.SingleValueInputPort); ok {
				single.Set(output)
			} else if array, ok := input.(nodes.SingleValueInputPort); ok {
				array.Set(output)
			} else {
				panic(fmt.Errorf("not sure how to assign node %q's input %q", nodeID, inputName))
			}
		}
	}

	// Set the Producers
	for fileName, producerDetails := range appSchema.Producers {
		producerNode := createdNodes[producerDetails.NodeID]
		outputs := producerNode.Outputs()
		output, ok := outputs[producerDetails.Port]
		if !ok {
			panic(fmt.Errorf("can't assign producer: node %q contains no port %q", producerDetails.NodeID, producerDetails.Port))
		}

		casted, ok := output.(nodes.Output[artifact.Artifact])
		if !ok {
			panic(fmt.Errorf("can't assign producer: node %q port %q does not produce artifacts", producerDetails.NodeID, producerDetails.Port))
		}

		i.producers[fileName] = casted
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
			Port:   producer.Name(),
		}
	}

	appSchema.Nodes = appNodeSchema

	registeredTypes := i.typeFactory.Types()
	nodeTypes := make([]schema.NodeType, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		nodeInstance, ok := i.typeFactory.New(registeredType).(nodes.Node)
		if !ok {
			panic(fmt.Errorf("Registered type %q is not a node. Not sure how we got here :/", registeredType))
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
			Port:   producer.Name(),
		}
	}
	appSchema.Nodes = nodeInstances

	// TODO: Is this unsafe? Yes.
	appSchema.Metadata = i.metadata.Data()
}

func (i *Instance) buildNodeGraphInstanceSchema(node nodes.Node, encoder *jbtf.Encoder) schema.AppNodeInstance {

	nodeInstance := schema.AppNodeInstance{
		Type:          refutil.GetTypeWithPackage(node),
		AssignedInput: make(map[string]schema.PortReference),
	}

	for inputName, input := range node.Inputs() {

		switch v := input.(type) {
		case nodes.SingleValueInputPort:
			val := v.Value()
			if val == nil {
				continue
			}

			nodeInstance.AssignedInput[inputName] = schema.PortReference{
				NodeId:   i.nodeIDs[val.Node()],
				PortName: val.Name(),
			}

		case nodes.ArrayValueInputPort:
			for index, val := range v.Value() {
				if val == nil {
					continue
				}

				nodeInstance.AssignedInput[fmt.Sprintf("%s.%d", inputName, index)] = schema.PortReference{
					NodeId:   i.nodeIDs[val.Node()],
					PortName: val.Name(),
				}
			}

		default:
			panic(fmt.Errorf("unable to recurse %v input %q", node, inputName))
		}

	}

	// sort.Slice(nodeInstance.AssignedInput, func(i, j int) bool {
	// 	return strings.ToLower(nodeInstance.AssignedInput[i].Name) < strings.ToLower(nodeInstance.AssignedInput[j].Name)
	// })

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
	node := i.Node(nodeId)

	cleanPortName := portName
	portIndex := -1
	if strings.Contains(portName, ".") {
		split := strings.Split(portName, ".")
		var err error
		portIndex, err = strconv.Atoi(split[1])
		if err != nil {
			panic(fmt.Errorf("unable to parse array index from %s: %w", portName, err))
		}
		cleanPortName = split[0]
	}

	inputs := node.Inputs()
	input, ok := inputs[cleanPortName]
	if !ok {
		panic(fmt.Errorf("node %s contains no input port %s", nodeId, cleanPortName))
	}

	if portIndex == -1 {
		input.Clear()
	} else {

		// We're dealing with the removal of a specific element in an array
		array, ok := input.(nodes.ArrayValueInputPort)

		if !ok {
			panic(fmt.Errorf("Treating node %q port %q like array, when it isn't", nodeId, portName))
		}

		array.Remove(array.Value()[portIndex])

	}

	i.incModelVersion()
}

func (i *Instance) ConnectNodes(nodeOutId, outPortName, nodeInId, inPortName string) {

	cleanedInputName := inPortName
	components := strings.Split(inPortName, ".")
	if len(components) > 1 {
		cleanedInputName = components[0]
		_, err := strconv.ParseInt(components[1], 10, 64)
		if err != nil {
			panic(fmt.Errorf("unable to parse index from %s: %w", inPortName, err))
		}
	}

	inNode := i.Node(nodeInId)
	inputs := inNode.Inputs()

	input, ok := inputs[cleanedInputName]
	if !ok {
		panic(fmt.Errorf("node %q contains no in-port %q", nodeInId, cleanedInputName))
	}

	outNode := i.Node(nodeOutId)
	outputs := outNode.Outputs()
	output, ok := outputs[outPortName]
	if !ok {
		panic(fmt.Errorf("node %q contains no out-port %q", nodeOutId, outPortName))
	}

	if single, ok := input.(nodes.SingleValueInputPort); ok {
		single.Set(output)
	} else if array, ok := input.(nodes.ArrayValueInputPort); ok {
		array.Add(output)
	} else {
		panic(fmt.Errorf("can not determine type of node %q's input %q", nodeInId, cleanedInputName))
	}

	i.incModelVersion()
}

// PRODUCERS ==================================================================

func (i *Instance) SetNodeAsProducer(nodeId, nodePort, producerName string) {
	producerNode := i.Node(nodeId)

	if producerNode == nil {
		panic(fmt.Errorf("no node exists with id %q", nodeId))
	}

	outputs := producerNode.Outputs()
	output, ok := outputs[nodePort]
	if !ok {
		panic(fmt.Errorf("node %q does not contain output %q", nodeId, nodePort))
	}

	casted, ok := output.(nodes.Output[artifact.Artifact])
	if !ok {
		panic(fmt.Errorf("node %q output %q does not produce artifacts", nodeId, nodePort))
	}

	// We need to check and remove previous references...
	for filename, producer := range i.producers {

		if i.NodeId(producer.Node()) != nodeId {
			continue
		}

		if producer.Name() != nodePort {
			continue
		}

		delete(i.producers, filename)
	}

	i.producers[producerName] = casted
	i.incModelVersion()
}

func (i *Instance) recursivelyRegisterNodeTypes(node nodes.Node) {
	i.addType(node)

	inputReferences := flattenNodeInputReferences(node)
	for _, reference := range inputReferences {
		i.recursivelyRegisterNodeTypes(reference)
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

func (i *Instance) AddProducer(producerName string, producer nodes.Output[artifact.Artifact]) {
	i.recursivelyRegisterNodeTypes(producer.Node())
	i.buildIDsForNode(producer.Node())
	i.producers[producerName] = producer
}

func (i *Instance) Producer(producerName string) nodes.Output[artifact.Artifact] {
	return i.producers[producerName]
}

func (i *Instance) ProducerNames() []string {
	names := make([]string, 0, len(i.producers))

	for name := range i.producers {
		names = append(names, name)
	}

	return names
}

func flattenNodeInputReferences(node nodes.Node) []nodes.Node {

	references := make([]nodes.Node, 0)

	for inputName, input := range node.Inputs() {

		switch v := input.(type) {
		case nodes.SingleValueInputPort:
			value := v.Value()
			if value == nil {
				continue
			}
			references = append(references, value.Node())

		case nodes.ArrayValueInputPort:
			for _, val := range v.Value() {
				if val == nil {
					continue
				}
				references = append(references, val.Node())
			}

		default:
			panic(fmt.Errorf("unable to recursive %v's input %q", node, inputName))
		}

	}

	return references
}

// REFLECTION =================================================================

func RecurseDependenciesType[T any](dependent nodes.Node) []T {
	allDependencies := make([]T, 0)

	inputReferences := flattenNodeInputReferences(dependent)

	for _, input := range inputReferences {

		subDependencies := RecurseDependenciesType[T](input)
		allDependencies = append(allDependencies, subDependencies...)

		ofT, ok := input.(T)
		if ok {
			allDependencies = append(allDependencies, ofT)
		}
	}

	return allDependencies
}

func BuildNodeTypeSchema(node nodes.Node) schema.NodeType {

	typeSchema := schema.NodeType{
		DisplayName: "Untyped",
		Outputs:     make(map[string]schema.NodeOutput),
		Inputs:      make(map[string]schema.NodeInput),
	}

	outputs := node.Outputs()
	for name, o := range outputs {
		nodeType := "any"
		if typed, ok := o.(nodes.Typed); ok {
			nodeType = typed.Type()
		}

		typeSchema.Outputs[name] = schema.NodeOutput{
			Type: nodeType,
		}
	}

	inputs := node.Inputs()
	for name, input := range inputs {

		nodeType := "any"
		if typed, ok := input.(nodes.Typed); ok {
			nodeType = typed.Type()
		}

		array := false
		if _, ok := input.(nodes.ArrayValueInputPort); ok {
			array = true
		}

		typeSchema.Inputs[name] = schema.NodeInput{
			Type:    nodeType,
			IsArray: array,
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
