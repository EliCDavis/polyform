package schema

type NodeDependency struct {
	DependencyID   string `json:"dependencyID"`
	DependencyPort string `json:"dependencyPort"`
	Name           string `json:"name"`
}
