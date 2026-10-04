package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func (i *Instance) BuildSchemaForAllNodeTypes() []schema.NodeType {
	i.mu().Lock()
	defer i.mu().Unlock()

	registeredTypes := i.typeFactory.Types()
	nodeTypes := make([]schema.NodeType, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		instance := i.typeFactory.New(registeredType)
		nodeInstance, ok := instance.(nodes.Node)
		if !ok {
			panic(fmt.Errorf("Registered type %q is not a node: %s", registeredType, instance))
		}
		if nodeInstance == nil {
			panic("New registered type is nil")
		}
		// log.Printf("%T: %+v\n", nodeInstance, nodeInstance)
		// log.Print(registeredType)
		b := BuildNodeTypeSchema(registeredType, nodeInstance)
		nodeTypes = append(nodeTypes, b)
	}
	return nodeTypes
}

func BuildNodeTypeSchema(registeredType string, node nodes.Node) schema.NodeType {
	typeSchema := schema.NodeType{
		DisplayName: "Untyped",
		Outputs:     make(map[string]schema.NodeTypeOutput),
		Inputs:      make(map[string]schema.NodeTypeInput),
	}

	outputs := node.Outputs()
	for name, o := range outputs {
		nodeType := "any"
		if typed, ok := o.(nodes.Typed); ok {
			nodeType = typed.Type()
		}

		desc := ""
		if description, ok := o.(nodes.Describable); ok {
			desc = description.Description()
		}

		dynamic := false
		if port, ok := o.(nodes.DynamicallyTypedPort); ok {
			if pattern := port.DynamicPattern(); pattern != "" {
				dynamic = true
				if nodeType == "" {
					nodeType = pattern
				}
			}
		}

		typeSchema.Outputs[name] = schema.NodeTypeOutput{
			Type:        nodeType,
			Description: desc,
			Dynamic:     dynamic,
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

		desc := ""
		if description, ok := input.(nodes.Describable); ok {
			desc = description.Description()
		}

		dynamic := false
		if pattern, ok := input.(nodes.DynamicallyTypedPort); ok {
			dynamic = true
			if nodeType == "" {
				nodeType = pattern.DynamicPattern()
			}
		}

		var accepted []string
		if options, ok := input.(nodes.TypeOptions); ok {
			accepted = options.AcceptedTypes()
			if nodeType == "" && len(accepted) > 0 {
				nodeType = accepted[0]
			}
		}

		typeSchema.Inputs[name] = schema.NodeTypeInput{
			Type:          nodeType,
			IsArray:       array,
			Description:   desc,
			Dynamic:       dynamic,
			AcceptedTypes: accepted,
		}
	}

	if param, ok := node.(Parameter); ok {
		typeSchema.Parameter = param.Schema()
	}

	if typed, ok := node.(nodes.Named); ok {
		typeSchema.DisplayName = typed.Name()
	} else if typed, ok := node.(nodes.Typed); ok {
		typeSchema.DisplayName = typed.Type()
	} else {
		resolver := refutil.TypeResolution{
			IncludePackage: false,
			IncludePointer: false,
		}
		typeSchema.DisplayName = resolver.Resolve(node)
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

	if keyworded, ok := node.(nodes.Keyworded); ok {
		typeSchema.Keywords = keyworded.Keywords()
	}

	typeSchema.Type = registeredType

	return typeSchema
}

