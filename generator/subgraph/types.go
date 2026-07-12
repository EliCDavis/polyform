package subgraph

const (
	InputNodeTypeKey  = "github.com/EliCDavis/polyform/generator/subgraph.InputNode"
	OutputNodeTypeKey = "github.com/EliCDavis/polyform/generator/subgraph.OutputNode"

	RuntimeTypePrefix = "subgraph/"
)

func IsRuntimeNodeType(typeName string) bool {
	return len(typeName) > len(RuntimeTypePrefix) && typeName[:len(RuntimeTypePrefix)] == RuntimeTypePrefix
}

func RuntimeTypeID(typeName string) string {
	if !IsRuntimeNodeType(typeName) {
		return ""
	}
	return typeName[len(RuntimeTypePrefix):]
}

func RuntimeTypePath(subGraphID string) string {
	return RuntimeTypePrefix + subGraphID
}
