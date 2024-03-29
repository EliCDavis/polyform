package nodes

type ReferencesNode interface {
	Node() Node
}

// Producer ===================================================================

type Output struct {
	Name string
	Type string
}

type Input struct {
	Name string
	Type string
}

// type Producer interface {
// 	Outputs() []Output
// }

// Dependent ==================================================================

type Dependent interface {
	Dependencies() []NodeDependency
}

// Node Dependency ============================================================

type NodeDependency interface {
	Named
	Dependency() Node
}

// STATE ======================================================================

type Stateful interface {
	State() NodeState
}

type StateData struct {
	state NodeState
}

func (s StateData) State() NodeState {
	return s.state
}

// Subscription ===============================================================

type Subscribable interface {
	AddSubscription(a Alertable)
}

type Alertable interface {
	Alert(version int, state NodeState)
}

// Named ======================================================================

type Named interface {
	Name() string
}
