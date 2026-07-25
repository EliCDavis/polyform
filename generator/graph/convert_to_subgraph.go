package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
)

// ConvertSelectionResult is returned after collapsing selected nodes into a
// new sub-graph definition and replacing them with a single runtime instance.
type ConvertSelectionResult struct {
	SubGraphID    string
	Name          string
	RuntimeNodeID string
	NodeType      schema.NodeType
}

// ConvertSelectionToSubGraph extracts the given nodes from scope into a new
// sub-graph definition, creates Input/Output boundaries for crossing edges,
// places a runtime instance on the parent, rewires neighbors, and deletes the
// original selection.
func (a *Instance) ConvertSelectionToSubGraph(scope Scope, nodeIDs []string, name, description string) (ConvertSelectionResult, error) {
	return convertSelectionToSubGraph(a.Root(), scope, nodeIDs, name, description)
}

type inboundCut struct {
	destNodeID   string
	destPortName string // may be indexed for arrays ("Values.0")
	srcNodeID    string
	srcPortName  string
	portType     string
}

type outboundCut struct {
	srcNodeID    string
	srcPortName  string
	destNodeID   string
	destPortName string
	portType     string
}

type nodePosition struct {
	x float64
	y float64
}

const (
	boundaryLayoutGapX  = 220.0
	boundaryLayoutGapY  = 100.0
)

func convertSelectionToSubGraph(root *Instance, scope Scope, nodeIDs []string, name, description string) (ConvertSelectionResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ConvertSelectionResult{}, fmt.Errorf("name is required")
	}
	description = strings.TrimSpace(description)

	if len(nodeIDs) == 0 {
		return ConvertSelectionResult{}, fmt.Errorf("at least one node id is required")
	}

	parent, err := scope.ResolveInstance(root)
	if err != nil {
		return ConvertSelectionResult{}, err
	}

	selection, err := validateConvertSelection(parent, nodeIDs)
	if err != nil {
		return ConvertSelectionResult{}, err
	}

	resolvedPositions := selectionPositionsFromMetadata(parent, selection)

	inbound, outbound, err := classifyConvertCuts(parent, selection)
	if err != nil {
		return ConvertSelectionResult{}, err
	}

	subGraphID, err := allocateSubGraphID(root, name)
	if err != nil {
		return ConvertSelectionResult{}, err
	}
	if err := root.CreateSubGraph(subGraphID, name, description); err != nil {
		return ConvertSelectionResult{}, err
	}

	child, err := root.SubGraphInstance(subGraphID)
	if err != nil {
		return ConvertSelectionResult{}, err
	}

	if err := copySelectionIntoSubGraph(parent, child, selection); err != nil {
		_ = root.DeleteSubGraph(subGraphID)
		return ConvertSelectionResult{}, err
	}

	inputPortNames, err := createInboundBoundaries(child, inbound, resolvedPositions)
	if err != nil {
		_ = root.DeleteSubGraph(subGraphID)
		return ConvertSelectionResult{}, err
	}

	outputPortBySource, err := createOutboundBoundaries(child, outbound, resolvedPositions)
	if err != nil {
		_ = root.DeleteSubGraph(subGraphID)
		return ConvertSelectionResult{}, err
	}

	// Ensure the registered runtime type reflects the new boundaries before
	// placing an instance on the parent.
	root.refreshSubGraphNodeType(subGraphID)

	_, runtimeNodeID, err := parent.CreateNode(subgraph.RuntimeTypePath(subGraphID))
	if err != nil {
		_ = root.DeleteSubGraph(subGraphID)
		return ConvertSelectionResult{}, err
	}

	setRuntimeNodePosition(parent, runtimeNodeID, resolvedPositions)

	for i, cut := range inbound {
		parent.ConnectNodes(cut.srcNodeID, cut.srcPortName, runtimeNodeID, inputPortNames[i])
	}

	for _, cut := range outbound {
		portName, ok := outputPortBySource[outboundSourceKey(cut.srcNodeID, cut.srcPortName)]
		if !ok {
			_ = root.DeleteSubGraph(subGraphID)
			return ConvertSelectionResult{}, fmt.Errorf("missing output boundary for %s.%s", cut.srcNodeID, cut.srcPortName)
		}
		parent.ConnectNodes(runtimeNodeID, portName, cut.destNodeID, cut.destPortName)
	}

	// Delete after rewiring so array consumers keep the new link while the
	// old selected-node link is cleared by DeleteNode.
	sortedIDs := make([]string, 0, len(selection))
	for id := range selection {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)
	for _, id := range sortedIDs {
		parent.DeleteNodeById(id)
	}

	typePath := subgraph.RuntimeTypePath(subGraphID)
	nodeType := BuildNodeTypeSchema(typePath, NewRuntimeNode(root, subGraphID))

	return ConvertSelectionResult{
		SubGraphID:    subGraphID,
		Name:          name,
		RuntimeNodeID: runtimeNodeID,
		NodeType:      nodeType,
	}, nil
}

func validateConvertSelection(inst *Instance, nodeIDs []string) (map[string]nodes.Node, error) {
	selection := make(map[string]nodes.Node, len(nodeIDs))
	for _, id := range nodeIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("node id cannot be empty")
		}
		if _, dup := selection[id]; dup {
			return nil, fmt.Errorf("duplicate node id %q", id)
		}
		if !inst.HasNodeWithId(id) {
			return nil, fmt.Errorf("no node exists with id %q", id)
		}
		node := inst.Node(id)
		if _, isBoundary := subgraph.IsBoundaryNode(node); isBoundary {
			return nil, fmt.Errorf("cannot convert boundary node %q into a sub-graph", id)
		}
		if _, isVar := node.(variable.Reference); isVar {
			return nil, fmt.Errorf("cannot convert variable reference node %q into a sub-graph", id)
		}
		selection[id] = node
	}
	return selection, nil
}

func classifyConvertCuts(inst *Instance, selection map[string]nodes.Node) ([]inboundCut, []outboundCut, error) {
	inbound := make([]inboundCut, 0)
	outbound := make([]outboundCut, 0)

	// Inbound: selected node inputs fed by nodes outside the selection.
	selectedIDs := make([]string, 0, len(selection))
	for id := range selection {
		selectedIDs = append(selectedIDs, id)
	}
	sort.Strings(selectedIDs)

	for _, destID := range selectedIDs {
		node := selection[destID]
		for inputName, input := range node.Inputs() {
			switch v := input.(type) {
			case nodes.SingleValueInputPort:
				val := v.Value()
				if val == nil {
					continue
				}
				srcID, ok := inst.nodeIDs[val.Node()]
				if !ok {
					return nil, nil, fmt.Errorf("node %q input %q references unknown node", destID, inputName)
				}
				if _, inSelection := selection[srcID]; inSelection {
					continue
				}
				portType, err := resolveOutputPortType(val)
				if err != nil {
					return nil, nil, fmt.Errorf("inbound cut %s.%s: %w", srcID, val.Name(), err)
				}
				inbound = append(inbound, inboundCut{
					destNodeID:   destID,
					destPortName: inputName,
					srcNodeID:    srcID,
					srcPortName:  val.Name(),
					portType:     portType,
				})

			case nodes.ArrayValueInputPort:
				for index, val := range v.Value() {
					if val == nil {
						continue
					}
					srcID, ok := inst.nodeIDs[val.Node()]
					if !ok {
						return nil, nil, fmt.Errorf("node %q input %q[%d] references unknown node", destID, inputName, index)
					}
					if _, inSelection := selection[srcID]; inSelection {
						continue
					}
					portType, err := resolveOutputPortType(val)
					if err != nil {
						return nil, nil, fmt.Errorf("inbound cut %s.%s: %w", srcID, val.Name(), err)
					}
					inbound = append(inbound, inboundCut{
						destNodeID:   destID,
						destPortName: fmt.Sprintf("%s.%d", inputName, index),
						srcNodeID:    srcID,
						srcPortName:  val.Name(),
						portType:     portType,
					})
				}
			}
		}
	}

	sort.Slice(inbound, func(i, j int) bool {
		if inbound[i].destNodeID != inbound[j].destNodeID {
			return inbound[i].destNodeID < inbound[j].destNodeID
		}
		return inbound[i].destPortName < inbound[j].destPortName
	})

	// Outbound: non-selected inputs fed by selected nodes.
	externalIDs := make([]string, 0)
	for _, id := range inst.nodeIDs {
		if _, inSelection := selection[id]; inSelection {
			continue
		}
		externalIDs = append(externalIDs, id)
	}
	sort.Strings(externalIDs)

	for _, destID := range externalIDs {
		node := inst.Node(destID)
		for inputName, input := range node.Inputs() {
			switch v := input.(type) {
			case nodes.SingleValueInputPort:
				val := v.Value()
				if val == nil {
					continue
				}
				srcID, ok := inst.nodeIDs[val.Node()]
				if !ok {
					continue
				}
				if _, inSelection := selection[srcID]; !inSelection {
					continue
				}
				portType, err := resolveOutputPortType(val)
				if err != nil {
					return nil, nil, fmt.Errorf("outbound cut %s.%s: %w", srcID, val.Name(), err)
				}
				outbound = append(outbound, outboundCut{
					srcNodeID:    srcID,
					srcPortName:  val.Name(),
					destNodeID:   destID,
					destPortName: inputName,
					portType:     portType,
				})

			case nodes.ArrayValueInputPort:
				for index, val := range v.Value() {
					if val == nil {
						continue
					}
					srcID, ok := inst.nodeIDs[val.Node()]
					if !ok {
						continue
					}
					if _, inSelection := selection[srcID]; !inSelection {
						continue
					}
					portType, err := resolveOutputPortType(val)
					if err != nil {
						return nil, nil, fmt.Errorf("outbound cut %s.%s: %w", srcID, val.Name(), err)
					}
					outbound = append(outbound, outboundCut{
						srcNodeID:    srcID,
						srcPortName:  val.Name(),
						destNodeID:   destID,
						destPortName: fmt.Sprintf("%s.%d", inputName, index),
						portType:     portType,
					})
				}
			}
		}
	}

	sort.Slice(outbound, func(i, j int) bool {
		if outbound[i].srcNodeID != outbound[j].srcNodeID {
			return outbound[i].srcNodeID < outbound[j].srcNodeID
		}
		if outbound[i].srcPortName != outbound[j].srcPortName {
			return outbound[i].srcPortName < outbound[j].srcPortName
		}
		if outbound[i].destNodeID != outbound[j].destNodeID {
			return outbound[i].destNodeID < outbound[j].destNodeID
		}
		return outbound[i].destPortName < outbound[j].destPortName
	})

	return inbound, outbound, nil
}

func resolveOutputPortType(port nodes.OutputPort) (string, error) {
	if port == nil {
		return "", fmt.Errorf("nil output port")
	}
	if typed, ok := port.(nodes.Typed); ok {
		t := strings.TrimSpace(typed.Type())
		if t != "" {
			// Prefer typed name; CreateBoundaryNode rejects unknown types.
			return t, nil
		}
	}
	return "", fmt.Errorf("unable to resolve port type for output %q", port.Name())
}

func allocateSubGraphID(root *Instance, name string) (string, error) {
	if err := root.assertRootGraph("allocate subgraph id"); err != nil {
		return "", err
	}
	root.initSubGraphs()

	base := strings.TrimSpace(name)
	base = strings.Join(strings.Fields(base), "_")
	if base == "" {
		base = "Subgraph"
	}

	if _, exists := root.subGraphs[base]; !exists {
		return base, nil
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if _, exists := root.subGraphs[candidate]; !exists {
			return candidate, nil
		}
	}
}

func copySelectionIntoSubGraph(parent, child *Instance, selection map[string]nodes.Node) error {
	encoder := &jbtf.Encoder{}
	nodeDefs := make(map[string]persistence.Node, len(selection))

	ids := make([]string, 0, len(selection))
	for id := range selection {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		node := selection[id]
		encoded := parent.buildNodeGraphInstanceSchema(node, encoder)
		if key, ok := parent.nodeTypeKeys[node]; ok && key != "" {
			encoded.Type = key
		}
		// Keep only wires whose source is also in the selection.
		filtered := make(map[string]schema.PortReference)
		for portName, ref := range encoded.AssignedInput {
			if _, ok := selection[ref.NodeId]; ok {
				filtered[portName] = ref
			}
		}
		encoded.AssignedInput = filtered
		nodeDefs[id] = encoded
	}

	payload, err := encoder.ToPgtf(persistence.App{
		Nodes: nodeDefs,
	})
	if err != nil {
		return fmt.Errorf("encode selection for sub-graph copy: %w", err)
	}

	decoder, err := jbtf.NewDecoder(payload)
	if err != nil {
		return fmt.Errorf("decode selection for sub-graph copy: %w", err)
	}

	appSchema, err := jbtf.Unmarshal[persistence.App](payload)
	if err != nil {
		return fmt.Errorf("unmarshal selection for sub-graph copy: %w", err)
	}

	createdNodes := make(map[string]nodes.Node, len(appSchema.Nodes))
	for nodeID, details := range appSchema.Nodes {
		node, err := child.instantiateAppNode(nodeID, details)
		if err != nil {
			return err
		}
		if node != nil {
			createdNodes[nodeID] = node
		}
	}

	if err := applyPersistedNodeData(appSchema.Nodes, createdNodes, decoder); err != nil {
		return err
	}
	if err := child.connectAppNodes(appSchema.Nodes, createdNodes); err != nil {
		return err
	}

	// Copy per-node metadata from the parent scope.
	for _, id := range ids {
		metaPath := "nodes." + id
		if data := metadataMapAt(parent, metaPath); data != nil {
			child.metadata.Set(metaPath, cloneMetadataMap(data))
		}
	}

	return nil
}

func cloneMetadataMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if nested, ok := v.(map[string]any); ok {
			out[k] = cloneMetadataMap(nested)
			continue
		}
		out[k] = v
	}
	return out
}

func createInboundBoundaries(child *Instance, inbound []inboundCut, positions map[string]nodePosition) ([]string, error) {
	minX, minY, _, _, hasBounds := selectionBounds(positions)
	portNames := make([]string, len(inbound))
	for i, cut := range inbound {
		portName := fmt.Sprintf("Input %d", i+1)
		_, boundaryID, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, cut.portType)
		if err != nil {
			return nil, fmt.Errorf("create input boundary %q: %w", portName, err)
		}
		if err := child.SetBoundaryNodeInfo(boundaryID, portName); err != nil {
			return nil, err
		}
		child.ConnectNodes(boundaryID, subgraph.ValuePortName, cut.destNodeID, cut.destPortName)
		if hasBounds {
			setNodePositionMetadata(child, boundaryID, nodePosition{
				x: minX - boundaryLayoutGapX,
				y: minY + float64(i)*boundaryLayoutGapY,
			})
		}
		portNames[i] = portName
	}
	return portNames, nil
}

func createOutboundBoundaries(child *Instance, outbound []outboundCut, positions map[string]nodePosition) (map[string]string, error) {
	// One Output N per unique (srcNode, srcPort); fan-out shares that boundary.
	type uniqueSource struct {
		srcNodeID   string
		srcPortName string
		portType    string
	}

	seen := make(map[string]struct{})
	unique := make([]uniqueSource, 0)
	for _, cut := range outbound {
		key := outboundSourceKey(cut.srcNodeID, cut.srcPortName)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, uniqueSource{
			srcNodeID:   cut.srcNodeID,
			srcPortName: cut.srcPortName,
			portType:    cut.portType,
		})
	}

	sort.Slice(unique, func(i, j int) bool {
		if unique[i].srcNodeID != unique[j].srcNodeID {
			return unique[i].srcNodeID < unique[j].srcNodeID
		}
		return unique[i].srcPortName < unique[j].srcPortName
	})

	_, minY, maxX, _, hasBounds := selectionBounds(positions)
	result := make(map[string]string, len(unique))
	for i, src := range unique {
		portName := fmt.Sprintf("Output %d", i+1)
		_, boundaryID, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, src.portType)
		if err != nil {
			return nil, fmt.Errorf("create output boundary %q: %w", portName, err)
		}
		if err := child.SetBoundaryNodeInfo(boundaryID, portName); err != nil {
			return nil, err
		}
		child.ConnectNodes(src.srcNodeID, src.srcPortName, boundaryID, subgraph.ValuePortName)
		if hasBounds {
			setNodePositionMetadata(child, boundaryID, nodePosition{
				x: maxX + boundaryLayoutGapX,
				y: minY + float64(i)*boundaryLayoutGapY,
			})
		}
		result[outboundSourceKey(src.srcNodeID, src.srcPortName)] = portName
	}
	return result, nil
}

func outboundSourceKey(nodeID, portName string) string {
	return nodeID + "\x00" + portName
}

func selectionPositionsFromMetadata(inst *Instance, selection map[string]nodes.Node) map[string]nodePosition {
	out := make(map[string]nodePosition, len(selection))
	for id := range selection {
		if pos, ok := metadataPosition(inst, id); ok {
			out[id] = pos
		}
	}
	return out
}

func metadataMapAt(inst *Instance, metaPath string) map[string]any {
	if !inst.metadata.PathExists(metaPath) {
		return nil
	}
	data := inst.metadata.Get(metaPath)
	if data == nil {
		return nil
	}
	meta, ok := data.(map[string]any)
	if !ok || meta == nil {
		return nil
	}
	return meta
}

func metadataPosition(inst *Instance, nodeID string) (nodePosition, bool) {
	meta := metadataMapAt(inst, "nodes."+nodeID)
	if meta == nil {
		return nodePosition{}, false
	}
	posRaw, ok := meta["position"]
	if !ok {
		return nodePosition{}, false
	}
	pos, ok := posRaw.(map[string]any)
	if !ok {
		return nodePosition{}, false
	}
	x, xOk := asFloat64(pos["x"])
	y, yOk := asFloat64(pos["y"])
	if !xOk || !yOk {
		return nodePosition{}, false
	}
	return nodePosition{x: x, y: y}, true
}

func setNodePositionMetadata(inst *Instance, nodeID string, pos nodePosition) {
	inst.metadata.Set("nodes."+nodeID+".position", map[string]any{
		"x": pos.x,
		"y": pos.y,
	})
}

func selectionBounds(positions map[string]nodePosition) (minX, minY, maxX, maxY float64, ok bool) {
	first := true
	for _, pos := range positions {
		if first {
			minX, maxX = pos.x, pos.x
			minY, maxY = pos.y, pos.y
			first = false
			continue
		}
		if pos.x < minX {
			minX = pos.x
		}
		if pos.x > maxX {
			maxX = pos.x
		}
		if pos.y < minY {
			minY = pos.y
		}
		if pos.y > maxY {
			maxY = pos.y
		}
	}
	return minX, minY, maxX, maxY, !first
}

func setRuntimeNodePosition(inst *Instance, runtimeNodeID string, positions map[string]nodePosition) {
	if len(positions) == 0 {
		return
	}
	var sumX, sumY float64
	for _, pos := range positions {
		sumX += pos.x
		sumY += pos.y
	}
	n := float64(len(positions))
	setNodePositionMetadata(inst, runtimeNodeID, nodePosition{
		x: sumX / n,
		y: sumY / n,
	})
}

func asFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
