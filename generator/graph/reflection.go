package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

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

func BuildSchemaForAllNodeTypes(typeFactory *refutil.TypeFactory) []schema.NodeType {
	registeredTypes := typeFactory.Types()
	nodeTypes := make([]schema.NodeType, 0, len(registeredTypes))
	for _, registeredType := range registeredTypes {
		instance := typeFactory.New(registeredType)
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
		Outputs:     make(map[string]schema.NodeOutput),
		Inputs:      make(map[string]schema.NodeInput),
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

		typeSchema.Outputs[name] = schema.NodeOutput{
			Type:        nodeType,
			Description: desc,
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

		typeSchema.Inputs[name] = schema.NodeInput{
			Type:        nodeType,
			IsArray:     array,
			Description: desc,
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

	typeSchema.Type = registeredType

	return typeSchema
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
