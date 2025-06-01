package graph

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"

	gsync "sync"
)

type Instance struct {
	typeFactory *refutil.TypeFactory

	movelVersion   uint32
	nodeIDs        map[nodes.Node]string
	metadata       *sync.NestedSyncMap
	namedManifests *namedOutputManager[manifest.Manifest]
	variables      *VariableGroup

	// TODO: Make this a lock across the entire instance
	producerLock gsync.Mutex
}

func New(typeFactory *refutil.TypeFactory) *Instance {
	return &Instance{
		typeFactory: typeFactory,
		variables:   NewVariableGroup(),
		nodeIDs:     make(map[nodes.Node]string),
		metadata:    sync.NewNestedSyncMap(),
		namedManifests: &namedOutputManager[manifest.Manifest]{
			namedPorts: make(map[string]namedOutputEntry[manifest.Manifest]),
		},
		movelVersion: 0,
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// VARIABLES
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (i *Instance) NewVariable(variablePath string, variable variable.Variable) {
	if variable == nil {
		panic(fmt.Errorf("trying to add a nil variable %q to graph", variablePath))
	}

	if i.variables.HasVariable(variablePath) {
		panic(fmt.Errorf("trying to add a variable to the path %q that already has a variable", variablePath))
	}

	if i.variables.HasSubgroup(variablePath) {
		panic(fmt.Errorf("trying to add a variable to the path %q that is registered as a subgroup", variablePath))
	}

	i.variables.AddVariable(variablePath, variable)
	i.typeFactory.RegisterBuilder(variablePath, func() any {
		return variable.NodeReference()
	})
}

func (i *Instance) DeleteVariable(variablePath string) {
	if !i.variables.HasVariable(variablePath) {
		panic(fmt.Errorf("trying to delete a variable at the path %q which doesn't contain a variable", variablePath))
	}

	i.variables.RemoveVariable(variablePath)
	i.typeFactory.Unregister(variablePath)
}

func (i *Instance) GetVariable(variablePath string) variable.Variable {
	if !i.variables.HasVariable(variablePath) {
		panic(fmt.Errorf("trying to get a variable at the path %q which doesn't exist", variablePath))
	}

	return i.variables.GetVariable(variablePath)
}

func (i *Instance) UpdateVariable(variablePath string, data []byte) (bool, error) {
	i.producerLock.Lock()
	defer i.producerLock.Unlock()

	variable := i.variables.GetVariable(variablePath)
	r, err := variable.ApplyMessage(data)
	i.incModelVersion()
	return r, err
}

func (i *Instance) VariableData(variablePath string) []byte {
	i.producerLock.Lock()
	defer i.producerLock.Unlock()

	variable := i.variables.GetVariable(variablePath)
	return variable.ToMessage()
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (i *Instance) IsPortNamed(node nodes.Node, portName string) (string, bool) {
	return i.namedManifests.IsPortNamed(node, portName)
}

func (i *Instance) ModelVersion() uint32 {
	return i.movelVersion
}

func (i *Instance) NodeIds() []string {
	ids := make([]string, 0, len(i.nodeIDs))

	for _, id := range i.nodeIDs {
		ids = append(ids, id)
	}

	return ids
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

	// nodeType := ""
	// if typed, ok := node.(nodes.Typed); ok {
	// 	nodeType = typed.Type()
	// } else {
	// 	nodeType = refutil.GetTypeWithPackage(node)
	// }

	nodeInstance := schema.NodeInstance{
		Name:          "Unamed",
		Type:          refutil.GetTypeWithPackage(node),
		AssignedInput: make(map[string]schema.PortReference),
		Output:        make(map[string]schema.NodeInstanceOutputPort),
		Metadata:      metadata,
	}

	if reference, ok := node.(variable.Reference); ok {
		variable := reference.Reference()
		nodeInstance.Name = variable.Name()
		nodeInstance.Variable = variable
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

	for outputPortName, outputPort := range node.Outputs() {
		nodeInstance.Output[outputPortName] = schema.NodeInstanceOutputPort{
			Version: outputPort.Version(),
		}
	}

	for inputPortName, inputPort := range node.Inputs() {
		if single, ok := inputPort.(nodes.SingleValueInputPort); ok {
			port := single.Value()
			if port == nil {
				continue
			}

			dependencyID, ok := i.nodeIDs[port.Node()]
			if !ok {
				panic(fmt.Errorf("node %q input port %q references ouput port %q to a node we have no knowledge of", nodeInstance.Name, inputPort.Name(), port.Name()))
			}

			nodeInstance.AssignedInput[inputPortName] = schema.PortReference{
				NodeId:   dependencyID,
				PortName: port.Name(),
			}
		}

		if array, ok := inputPort.(nodes.ArrayValueInputPort); ok {
			for valIndex, port := range array.Value() {
				if port == nil {
					continue
				}

				dependencyID, ok := i.nodeIDs[port.Node()]
				if !ok {
					panic(fmt.Errorf("node %q input port %q references ouput port %q to a node we have no knowledge of", nodeInstance.Name, inputPort.Name(), port.Name()))
				}

				nodeInstance.AssignedInput[fmt.Sprintf("%s.%d", inputPortName, valIndex)] = schema.PortReference{
					NodeId:   dependencyID,
					PortName: port.Name(),
				}
			}
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
	i.variables.Traverse(func(path string, v variable.Variable) {
		i.typeFactory.Unregister(path)
	})
	i.variables = NewVariableGroup()
	i.namedManifests = &namedOutputManager[manifest.Manifest]{
		namedPorts: make(map[string]namedOutputEntry[manifest.Manifest]),
	}
}

type sortedReference struct {
	name string
	port schema.PortReference

	arrayName string
	array     int
}

func sortPortReferences(ports map[string]schema.PortReference) []sortedReference {
	sorted := make([]sortedReference, 0, len(ports))
	for name, port := range ports {
		i := -1
		arrName := ""

		split := strings.LastIndex(name, ".")
		if split != -1 {
			v, err := strconv.Atoi(name[split+1:])
			if err != nil {
				panic(err)
			}
			i = v
			arrName = name[:split]
		}

		sorted = append(sorted, sortedReference{
			name:      name,
			port:      port,
			array:     i,
			arrayName: arrName,
		})
	}

	sort.Slice(sorted, func(i int, j int) bool {
		if sorted[i].array == -1 || sorted[j].array == -1 {
			return sorted[i].name < sorted[j].name
		}

		if sorted[i].arrayName == sorted[j].arrayName {
			return sorted[i].array < sorted[j].array
		}

		return sorted[i].array < sorted[j].array
	})

	return sorted
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
	VariableGroupFromSchema(appSchema.Variables).Traverse(func(path string, v variable.Variable) {
		i.NewVariable(path, v)
	})

	createdNodes := make(map[string]nodes.Node)

	// Create the Nodes
	for nodeID, instanceDetails := range appSchema.Nodes {
		if nodeID == "" {
			panic("attempting to create a node without an ID")
		}

		if instanceDetails.Variable != nil {
			// We instantiate variables a different way
			node := i.variables.GetVariable(*instanceDetails.Variable).NodeReference()
			createdNodes[nodeID] = node
			i.nodeIDs[node] = nodeID
		} else {
			newNode := i.typeFactory.New(instanceDetails.Type)
			casted, ok := newNode.(nodes.Node)
			if !ok {
				panic(fmt.Errorf("graph definition contained type that instantiated a non node: %s", instanceDetails.Type))
			}
			createdNodes[nodeID] = casted
			i.nodeIDs[casted] = nodeID
		}

	}

	// Connect the nodes we just created
	for nodeID, instanceDetails := range appSchema.Nodes {
		node := createdNodes[nodeID]
		inputs := node.Inputs()

		sortedInput := sortPortReferences(instanceDetails.AssignedInput)

		for _, sorted := range sortedInput {

			dirtyInputName := sorted.name
			dependency := sorted.port

			inputName := dirtyInputName
			components := strings.Split(inputName, ".")
			if len(components) > 1 {
				inputName = components[0]
			}

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
				err := single.Set(output)
				if err != nil {
					panic(err)
				}
			} else if array, ok := input.(nodes.ArrayValueInputPort); ok {
				err := array.Add(output)
				if err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("not sure how to assign node %q's input %q", nodeID, inputName))
			}
		}
	}

	// Set the Producers
	for producerName, producerDetails := range appSchema.Producers {
		producerNode := createdNodes[producerDetails.NodeID]
		outputs := producerNode.Outputs()
		output, ok := outputs[producerDetails.Port]
		if !ok {
			panic(fmt.Errorf("can't assign producer: node %q contains no port %q", producerDetails.NodeID, producerDetails.Port))
		}

		casted, ok := output.(nodes.Output[manifest.Manifest])
		if !ok {
			panic(fmt.Errorf("can't assign producer: node %q port %q does not produce a manifest", producerDetails.NodeID, producerDetails.Port))
		}

		i.namedManifests.NamePort(producerName, producerDetails.Port, producerNode, casted)
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
		Variables: i.variables.Schema(),
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

	for key, producer := range i.namedManifests.namedPorts {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := i.nodeIDs[producer.node]
		node := appNodeSchema[id]
		node.Name = key
		appNodeSchema[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.port.Name(),
		}
	}

	appSchema.Nodes = appNodeSchema

	return appSchema
}

func (i *Instance) EncodeToAppSchema(appSchema *schema.App, encoder *jbtf.Encoder) {
	variableLut := make(map[variable.Variable]string)
	i.variables.ReverseLookup(variableLut, "")

	nodeInstances := make(map[string]schema.AppNodeInstance)
	for node := range i.nodeIDs {
		id, ok := i.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := nodeInstances[id]; ok {
			panic(fmt.Errorf("we've arrived to a invalid state. two nodes refer to the same ID. There's a bug somewhere"))
		}

		nodeSchema := i.buildNodeGraphInstanceSchema(node, encoder)

		if reference, ok := node.(variable.Reference); ok {
			variablePath := variableLut[reference.Reference()]
			nodeSchema.Variable = &variablePath
		}

		nodeInstances[id] = nodeSchema
	}

	if appSchema.Producers == nil {
		appSchema.Producers = make(map[string]schema.Producer)
	}

	for key, producer := range i.namedManifests.namedPorts {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := i.nodeIDs[producer.node]
		node := nodeInstances[id]
		nodeInstances[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.port.Name(),
		}
	}

	appSchema.Nodes = nodeInstances
	appSchema.Variables = i.variables.Schema()

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

func (i *Instance) HasNodeWithId(nodeId string) bool {
	for _, id := range i.nodeIDs {
		if id == nodeId {
			return true
		}
	}
	return false
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

	i.namedManifests.DeleteNode(nodeToDelete)

	delete(i.nodeIDs, nodeToDelete)
}

// PARAMETER ==================================================================

func (i *Instance) getParameters() []Parameter {
	if i.namedManifests == nil || i.namedManifests.namedPorts == nil {
		return nil
	}

	parameterSet := make(map[Parameter]struct{})
	for _, n := range i.namedManifests.namedPorts {
		params := RecurseDependenciesType[Parameter](n.node)
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

		err := array.Remove(array.Value()[portIndex])
		if err != nil {
			panic(err)
		}

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
		err := single.Set(output)
		if err != nil {
			panic(err)
		}
	} else if array, ok := input.(nodes.ArrayValueInputPort); ok {
		err := array.Add(output)
		if err != nil {
			panic(err)
		}
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

	casted, ok := output.(nodes.Output[manifest.Manifest])
	if !ok {
		panic(fmt.Errorf("node %q output %q does not produce artifacts", nodeId, nodePort))
	}

	i.namedManifests.NamePort(producerName, nodePort, producerNode, casted)
	i.incModelVersion()
}

func (i *Instance) recursivelyRegisterNodeTypes(node nodes.Node) {
	i.addType(node)

	inputReferences := flattenNodeInputReferences(node)
	for _, reference := range inputReferences {
		i.recursivelyRegisterNodeTypes(reference)
	}
}

func (i *Instance) Manifest(producerName string) manifest.Manifest {
	producer, ok := i.namedManifests.namedPorts[producerName]
	if !ok {
		panic(fmt.Errorf("no producer registered for: %s", producerName))
	}

	i.producerLock.Lock()
	defer i.producerLock.Unlock()

	return producer.port.Value()
}

func (i *Instance) AddProducer(producerName string, producer nodes.Output[manifest.Manifest]) {
	i.recursivelyRegisterNodeTypes(producer.Node())
	i.buildIDsForNode(producer.Node())
	i.namedManifests.NamePort(producerName, producer.Name(), producer.Node(), producer)
}

func (i *Instance) Producer(producerName string) nodes.Output[manifest.Manifest] {
	return i.namedManifests.namedPorts[producerName].port
}

func (i *Instance) ProducerNames() []string {
	names := make([]string, 0, len(i.namedManifests.namedPorts))

	for name := range i.namedManifests.namedPorts {
		names = append(names, name)
	}

	return names
}
