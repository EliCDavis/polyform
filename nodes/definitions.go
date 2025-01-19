package nodes

type NodeOutputReference interface {
	Node() Node
	Port() string
}

// Producer ===================================================================

type Output struct {
	Type       string
	NodeOutput NodeOutputReference
}

type Input struct {
	Name  string
	Type  string
	Array bool
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
	DependencyPort() string
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

// Typed ======================================================================

type Typed interface {
	Type() string
}

// Pathed =====================================================================

type Pathed interface {
	Path() string
}

// Describable ================================================================

type Describable interface {
	Description() string
}
