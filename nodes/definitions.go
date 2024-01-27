package nodes

type Dependency interface {
	Versioned
	Stateful
	Subscribable
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

// Dependent ==================================================================

type Dependent interface {
	Dependencies() []Dependency
}
