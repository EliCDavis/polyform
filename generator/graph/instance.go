package graph

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/sync"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"

	gsync "sync"
)

type Details struct {
	Name        string
	Version     string
	Description string
	Authors     []schema.Author
}

type Instance struct {
	details     Details
	typeFactory *refutil.TypeFactory

	movelVersion   uint32
	nodeIDs        map[nodes.Node]string
	metadata       *sync.NestedSyncMap
	namedManifests *namedOutputManager[manifest.Manifest]
	variables      variable.System

	profiles map[string]variable.Profile

	// TODO: Make this a lock across the entire instance
	lock gsync.RWMutex
}

type Config struct {
	Name        string
	Version     string
	Description string
	Authors     []schema.Author
	TypeFactory *refutil.TypeFactory
}

func New(config Config) *Instance {
	return &Instance{
		details: Details{
			Name:        config.Name,
			Description: config.Description,
			Version:     config.Version,
			Authors:     config.Authors,
		},
		typeFactory: config.TypeFactory,
		variables:   variable.NewSystem(),
		nodeIDs:     make(map[nodes.Node]string),
		metadata:    sync.NewNestedSyncMap(),
		namedManifests: &namedOutputManager[manifest.Manifest]{
			namedPorts: make(map[string]namedOutputEntry[manifest.Manifest]),
		},
		movelVersion: 0,
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Details
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (a *Instance) GetName() string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.details.Name
}

func (a *Instance) GetVersion() string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.details.Version
}

func (a *Instance) GetDescription() string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.details.Description
}

func (a *Instance) GetAuthors() []schema.Author {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.details.Authors
}

func (a *Instance) SetName(name string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.details.Name = name
}

func (a *Instance) SetVersion(version string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.details.Version = version
}

func (a *Instance) SetDescription(description string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.details.Description = description
}

func (a *Instance) SetDetails(details Details) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.details = details
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// VARIABLES
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (a *Instance) NewVariable(variablePath string, variable variable.Variable) string {
	if variable == nil {
		panic(fmt.Errorf("trying to add a nil variable %q to graph", variablePath))
	}

	err := a.variables.Add(variablePath, variable)
	if err != nil {
		panic(fmt.Errorf("failed to add variable to graph: %w", err))
	}

	a.typeFactory.RegisterBuilder(variablePath, func() any {
		return variable.NodeReference()
	})

	return variablePath
}

func (a *Instance) DeleteVariable(variablePath string) {
	if !a.variables.Exists(variablePath) {
		panic(fmt.Errorf("trying to delete a variable at the path %q which doesn't contain a variable", variablePath))
	}

	varP, err := a.variables.Variable(variablePath)
	if err != nil {
		panic(err)
	}
	nodesToDelete := make([]string, 0)
	for node, nodeId := range a.nodeIDs {
		ref, ok := node.(variable.Reference)
		if ok && ref.Reference() == varP {
			nodesToDelete = append(nodesToDelete, nodeId)
		}
	}

	for _, n := range nodesToDelete {
		a.DeleteNodeById(n)
	}

	err = a.variables.Remove(variablePath)
	if err != nil {
		panic(err)
	}
	a.typeFactory.Unregister(variablePath)
}

func (a *Instance) GetVariable(variablePath string) variable.Variable {
	if !a.variables.Exists(variablePath) {
		panic(fmt.Errorf("trying to get a variable at the path %q which doesn't exist", variablePath))
	}

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		panic(err)
	}
	return variable
}

func (a *Instance) SetVariableInfo(variablePath, newPath, description string) error {
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return err
	}

	variable.Info().SetDescription(description)

	return a.variables.Move(variablePath, newPath)
}

func (a *Instance) SetVariableDescription(variablePath, description string) error {
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return err
	}

	variable.Info().SetDescription(description)

	return nil
}

func (a *Instance) UpdateVariable(variablePath string, data []byte) (bool, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return false, err
	}

	r, err := variable.ApplyMessage(data)
	a.incModelVersion()
	return r, err
}

func (a *Instance) VariableData(variablePath string) ([]byte, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return nil, err
	}
	return variable.ToMessage(), nil
}

func (a *Instance) SwaggerDefinition() swagger.Definition {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.variables.SwaggerDefinition()
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Profiles
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (a *Instance) SaveProfile(profileName string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.profiles[profileName] = a.variables.GetProfile()
}

func (a *Instance) LoadProfile(profileName string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if profile, ok := a.profiles[profileName]; ok {
		return a.variables.ApplyProfile(profile)
	}

	return fmt.Errorf("no profile exists with name %q", profileName)
}

func (a *Instance) RenameProfile(profile, newName string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if _, ok := a.profiles[profile]; !ok {
		return fmt.Errorf("profile %q does not exist", profile)
	}

	if _, ok := a.profiles[newName]; ok {
		return fmt.Errorf("profile %q already exists", profile)
	}

	a.profiles[newName] = a.profiles[profile]

	delete(a.profiles, profile)

	return nil
}

func (a *Instance) DeleteProfile(profileName string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if _, ok := a.profiles[profileName]; !ok {
		return fmt.Errorf("graph contains no profile named %q", profileName)
	}

	delete(a.profiles, profileName)
	return nil
}

func (a *Instance) Profiles() []string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	profiles := make([]string, 0, len(a.profiles))

	for profile := range a.profiles {
		profiles = append(profiles, profile)
	}

	sort.Strings(profiles)

	return profiles
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (a *Instance) IsPortNamed(node nodes.Node, portName string) (string, bool) {
	return a.namedManifests.IsPortNamed(node, portName)
}

func (a *Instance) ModelVersion() uint32 {
	return a.movelVersion
}

func (a *Instance) NodeIds() []string {
	ids := make([]string, 0, len(a.nodeIDs))

	for _, id := range a.nodeIDs {
		ids = append(ids, id)
	}

	return ids
}

func (a *Instance) incModelVersion() {
	// TODO: Make thread safe
	a.movelVersion++
}

func (a *Instance) NodeInstanceSchema(node nodes.Node) schema.NodeInstance {
	var metadata map[string]any
	metadataPath := "nodes." + a.nodeIDs[node]

	if a.metadata.PathExists(metadataPath) {
		if data := a.metadata.Get(metadataPath); data != nil {
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
		nodeInstance.Name = variable.Info().Name()
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

			dependencyID, ok := a.nodeIDs[port.Node()]
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

				dependencyID, ok := a.nodeIDs[port.Node()]
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

// func (i *Instance) addType(v any) {
// 	if !i.typeFactory.TypeRegistered(v) {
// 		i.typeFactory.RegisterType(v)
// 	}
// }

func (a *Instance) buildIDsForNode(node nodes.Node) {

	// IDs for this node has already been built.
	if _, ok := a.nodeIDs[node]; ok {
		return
	}

	// Try to remove this
	// i.addType(node)

	references := flattenNodeInputReferences(node)
	for _, ref := range references {
		a.buildIDsForNode(ref)
	}

	// TODO: UGLY UGLY UGLY UGLY
	highestInded := len(a.nodeIDs)
	for {
		id := fmt.Sprintf("Node-%d", highestInded)

		taken := false
		for _, usedId := range a.nodeIDs {
			if usedId == id {
				taken = true
			}
			if taken {
				break
			}
		}
		if !taken {
			a.nodeIDs[node] = id
			break
		}
		highestInded++
	}
}

func (a *Instance) Reset() {
	a.details = Details{
		Name:        "New Graph",
		Description: "",
		Version:     "v0.0.0",
		Authors:     []schema.Author{},
	}
	a.nodeIDs = make(map[nodes.Node]string)
	a.metadata = sync.NewNestedSyncMap()
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		a.typeFactory.Unregister(path)
	})
	a.variables = variable.NewSystem()
	a.namedManifests = &namedOutputManager[manifest.Manifest]{
		namedPorts: make(map[string]namedOutputEntry[manifest.Manifest]),
	}
	a.profiles = make(map[string]variable.Profile)
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

func (a *Instance) ApplyAppSchema(jsonPayload []byte) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Reset()

	appSchema, err := jbtf.Unmarshal[schema.App](jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to parse graph as a jbtf: %w", err)
	}

	a.details.Name = appSchema.Name
	a.details.Authors = appSchema.Authors
	a.details.Version = appSchema.Version
	a.details.Description = appSchema.Description

	decoder, err := jbtf.NewDecoder(jsonPayload)
	if err != nil {
		return fmt.Errorf("unable to build a jbtf decoder: %w", err)
	}

	a.metadata.OverwriteData(appSchema.Metadata)
	appSchema.Variables.Traverse(func(path string, v schema.PersistedVariable) bool {
		var varabl variable.Variable
		varabl, err = variable.DeserializePersistantVariableJSON(v.Data, decoder)
		if err == nil {
			a.NewVariable(path, varabl)

			info, err := a.variables.Info(path)
			if err != nil {
				panic(fmt.Errorf("failed to add variable to graph: %w", err))
			}
			info.SetDescription(v.Description)
		}

		return err == nil
	})
	if err != nil {
		return err
	}

	for profile, data := range appSchema.Profiles {
		a.profiles[profile] = data.Data
	}

	createdNodes := make(map[string]nodes.Node)

	// Create the Nodes
	for nodeID, instanceDetails := range appSchema.Nodes {
		if nodeID == "" {
			panic("attempting to create a node without an ID")
		}

		if instanceDetails.Variable != nil {
			// We instantiate variables a different way
			variable, err := a.variables.Variable(*instanceDetails.Variable)
			if err != nil {
				return err
			}
			node := variable.NodeReference()
			createdNodes[nodeID] = node
			a.nodeIDs[node] = nodeID
		} else {
			newNode := a.typeFactory.New(instanceDetails.Type)
			casted, ok := newNode.(nodes.Node)
			if !ok {
				panic(fmt.Errorf("graph definition contained type that instantiated a non node: %s", instanceDetails.Type))
			}
			createdNodes[nodeID] = casted
			a.nodeIDs[casted] = nodeID
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

		a.namedManifests.NamePort(producerName, producerDetails.Port, producerNode, casted)
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

	a.incModelVersion()

	return nil
}

func (a *Instance) Schema() schema.GraphInstance {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var noteMetadata map[string]any
	if notes := a.metadata.Get("notes"); notes != nil {
		casted, ok := notes.(map[string]any)
		if ok {
			noteMetadata = casted
		}
	}

	variableSchema, err := a.variables.RuntimeSchema()
	if err != nil {
		panic(err)
	}

	appSchema := schema.GraphInstance{
		Producers: make(map[string]schema.Producer),
		Notes:     noteMetadata,
		Variables: variableSchema,
		Profiles:  make([]string, 0, len(a.profiles)),
	}

	for profile := range a.profiles {
		appSchema.Profiles = append(appSchema.Profiles, profile)
	}
	sort.Strings(appSchema.Profiles)

	appNodeSchema := make(map[string]schema.NodeInstance)

	for node := range a.nodeIDs {
		id, ok := a.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := appNodeSchema[id]; ok {
			panic("not sure how this happened")
		}

		appNodeSchema[id] = a.NodeInstanceSchema(node)
	}

	for key, producer := range a.namedManifests.namedPorts {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := a.nodeIDs[producer.node]
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

func (a *Instance) EncodeToAppSchema() ([]byte, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	encoder := &jbtf.Encoder{}

	appSchema := schema.App{
		Name:        a.details.Name,
		Version:     a.details.Version,
		Description: a.details.Description,
		Authors:     a.details.Authors,
	}

	variableLut := make(map[variable.Variable]string)
	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		variableLut[v] = path
	})

	nodeInstances := make(map[string]schema.AppNodeInstance)
	for node := range a.nodeIDs {
		id, ok := a.nodeIDs[node]
		if !ok {
			panic(fmt.Errorf("node %v has not had an ID generated for it", node))
		}

		if _, ok := nodeInstances[id]; ok {
			panic(fmt.Errorf("we've arrived to a invalid state. two nodes refer to the same ID. There's a bug somewhere"))
		}

		nodeSchema := a.buildNodeGraphInstanceSchema(node, encoder)

		if reference, ok := node.(variable.Reference); ok {
			variablePath := variableLut[reference.Reference()]
			nodeSchema.Variable = &variablePath
		}

		nodeInstances[id] = nodeSchema
	}

	if appSchema.Profiles == nil {
		appSchema.Profiles = make(map[string]schema.AppProfile)
	}
	for name, data := range a.profiles {
		appSchema.Profiles[name] = schema.AppProfile{
			Data: data,
		}
	}

	if appSchema.Producers == nil {
		appSchema.Producers = make(map[string]schema.Producer)
	}

	for key, producer := range a.namedManifests.namedPorts {
		// a.buildSchemaForNode(producer.Node(), appNodeSchema)
		id := a.nodeIDs[producer.node]
		node := nodeInstances[id]
		nodeInstances[id] = node

		appSchema.Producers[key] = schema.Producer{
			NodeID: id,
			Port:   producer.port.Name(),
		}
	}

	variableSchema, err := a.variables.PersistedSchema(encoder)
	if err != nil {
		panic(err)
	}

	appSchema.Nodes = nodeInstances
	appSchema.Variables = variableSchema

	// TODO: Is this unsafe? Yes.
	appSchema.Metadata = a.metadata.Data()

	return encoder.ToPgtf(appSchema)
}

func (a *Instance) buildNodeGraphInstanceSchema(node nodes.Node, encoder *jbtf.Encoder) schema.AppNodeInstance {

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
				NodeId:   a.nodeIDs[val.Node()],
				PortName: val.Name(),
			}

		case nodes.ArrayValueInputPort:
			for index, val := range v.Value() {
				if val == nil {
					continue
				}

				nodeInstance.AssignedInput[fmt.Sprintf("%s.%d", inputName, index)] = schema.PortReference{
					NodeId:   a.nodeIDs[val.Node()],
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

func (a *Instance) NodeId(node nodes.Node) string {
	return a.nodeIDs[node]
}

func (a *Instance) HasNodeWithId(nodeId string) bool {
	for _, id := range a.nodeIDs {
		if id == nodeId {
			return true
		}
	}
	return false
}

func (a *Instance) Node(nodeId string) nodes.Node {
	for n, id := range a.nodeIDs {
		if id == nodeId {
			return n
		}
	}
	panic(fmt.Errorf("no node exists with id %q", nodeId))
}

func (a *Instance) CreateNode(nodeType string) (nodes.Node, string, error) {
	if !a.typeFactory.KeyRegistered(nodeType) {
		return nil, "", fmt.Errorf("no factory registered with ID %s", nodeType)
	}

	newNode := a.typeFactory.New(nodeType)
	casted, ok := newNode.(nodes.Node)
	if !ok {
		panic(fmt.Errorf("Regiestered type did not create a node. How'd ya manage that: %s", nodeType))
	}
	a.buildIDsForNode(casted)

	return casted, a.nodeIDs[casted], nil
}

func (a *Instance) DeleteNodeById(nodeId string) {
	var nodeToDelete nodes.Node

	for n, id := range a.nodeIDs {
		if id == nodeId {
			nodeToDelete = n
		}
	}

	if nodeToDelete == nil {
		panic(fmt.Errorf("can't delete, no node registered with ID %s", nodeId))
	}

	a.DeleteNode(nodeToDelete)
}

func (a *Instance) DeleteNode(nodeToDelete nodes.Node) {
	a.namedManifests.DeleteNode(nodeToDelete)
	delete(a.nodeIDs, nodeToDelete)

	// Delete all nodes connecting to this
	for node := range a.nodeIDs {
		for inputName, input := range node.Inputs() {

			switch v := input.(type) {
			case nodes.SingleValueInputPort:
				value := v.Value()
				if value == nil {
					continue
				}
				if value.Node() == nodeToDelete {
					v.Clear()
				}

			case nodes.ArrayValueInputPort:
				for _, val := range v.Value() {
					if val == nil {
						continue
					}

					if val.Node() == nodeToDelete {
						err := v.Remove(val)
						if err != nil {
							panic(err)
						}
					}
				}

			default:
				panic(fmt.Errorf("unable to interpret %v's input %q", node, inputName))
			}

		}
	}
}

// PARAMETER ==================================================================

func (a *Instance) getParameters() []Parameter {
	if a.namedManifests == nil || a.namedManifests.namedPorts == nil {
		return nil
	}

	parameterSet := make(map[Parameter]struct{})
	for _, n := range a.namedManifests.namedPorts {
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

func (a *Instance) InitializeFromCLI(set *flag.FlagSet) {
	// 	for _, p := range a.getParameters() {
	// 		p.InitializeForCLI(set)
	// 	}
}

func (a *Instance) Parameter(nodeId string) Parameter {
	node := a.Node(nodeId)

	param, ok := node.(Parameter)
	if !ok {
		panic(fmt.Errorf("node %q is not a parameter", nodeId))
	}

	return param
}

func (a *Instance) UpdateParameter(nodeId string, data []byte) (bool, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	r, err := a.Parameter(nodeId).ApplyMessage(data)
	a.incModelVersion()
	return r, err
}

func (a *Instance) ParameterData(nodeId string) []byte {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.Parameter(nodeId).ToMessage()
}

// METADATA ===================================================================

func (a *Instance) SetMetadata(key string, value any) {
	a.metadata.Set(key, value)
}

func (a *Instance) DeleteMetadata(key string) {
	a.metadata.Delete(key)
}

// CONNECTIONS ================================================================

func (a *Instance) DeleteNodeInputConnection(nodeId, portName string) {
	node := a.Node(nodeId)

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

	a.incModelVersion()
}

func (a *Instance) ConnectNodes(nodeOutId, outPortName, nodeInId, inPortName string) {

	cleanedInputName := inPortName
	components := strings.Split(inPortName, ".")
	if len(components) > 1 {
		cleanedInputName = components[0]
		_, err := strconv.ParseInt(components[1], 10, 64)
		if err != nil {
			panic(fmt.Errorf("unable to parse index from %s: %w", inPortName, err))
		}
	}

	inNode := a.Node(nodeInId)
	inputs := inNode.Inputs()

	input, ok := inputs[cleanedInputName]
	if !ok {
		panic(fmt.Errorf("node %q contains no in-port %q", nodeInId, cleanedInputName))
	}

	outNode := a.Node(nodeOutId)
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

	a.incModelVersion()
}

// PRODUCERS ==================================================================

func (a *Instance) SetNodeAsProducer(nodeId, nodePort, producerName string) {
	producerNode := a.Node(nodeId)

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

	a.namedManifests.NamePort(producerName, nodePort, producerNode, casted)
	a.incModelVersion()
}

// func (i *Instance) recursivelyRegisterNodeTypes(node nodes.Node) {
// 	i.addType(node)

// 	inputReferences := flattenNodeInputReferences(node)
// 	for _, reference := range inputReferences {
// 		i.recursivelyRegisterNodeTypes(reference)
// 	}
// }

func (a *Instance) Manifest(producerName string) manifest.Manifest {
	producer, ok := a.namedManifests.namedPorts[producerName]
	if !ok {
		panic(fmt.Errorf("no producer registered for: %s", producerName))
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	return producer.port.Value()
}

func (a *Instance) AddProducer(producerName string, producer nodes.Output[manifest.Manifest]) {
	// i.recursivelyRegisterNodeTypes(producer.Node())
	a.buildIDsForNode(producer.Node())
	a.namedManifests.NamePort(producerName, producer.Name(), producer.Node(), producer)
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
