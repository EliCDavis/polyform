package nodes

type Stateful interface {
	State() NodeState
}

type StateData struct {
	state NodeState
}

func (s StateData) State() NodeState {
	return s.state
}
