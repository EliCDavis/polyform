package graph

import (
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/generator/subgraph"
)

type Scope string

const RootScope Scope = "root"

func SubGraphScope(subGraphID string) Scope {
	return Scope(subgraph.RuntimeTypePath(subGraphID))
}

func (s Scope) String() string {
	return string(s)
}

func (s Scope) IsRoot() bool {
	return s == "" || s == RootScope
}

func (s Scope) SubGraphID() (string, error) {
	if s.IsRoot() {
		return "", fmt.Errorf("scope %q is not a sub-graph scope", s)
	}

	id, ok := strings.CutPrefix(string(s), subgraph.RuntimeTypePrefix)
	if !ok || id == "" {
		return "", fmt.Errorf("unknown graph scope %q", s)
	}
	return id, nil
}

func (s Scope) ResolveInstance(project *Instance) (*Graph, error) {
	if s.IsRoot() {
		return project.Graph, nil
	}

	subGraphID, err := s.SubGraphID()
	if err != nil {
		return nil, err
	}
	return project.SubGraphInstance(subGraphID)
}
