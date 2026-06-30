package graph

import (
	"fmt"

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

	prefix := subgraph.RuntimeTypePrefix
	scope := string(s)
	if len(scope) <= len(prefix) || scope[:len(prefix)] != prefix {
		return "", fmt.Errorf("unknown graph scope %q", s)
	}
	return scope[len(prefix):], nil
}

func (s Scope) ResolveInstance(graph *Instance) (*Instance, error) {
	if s.IsRoot() {
		return graph.Root(), nil
	}

	subGraphID, err := s.SubGraphID()
	if err != nil {
		return nil, err
	}
	return graph.Root().SubGraphInstance(subGraphID)
}
