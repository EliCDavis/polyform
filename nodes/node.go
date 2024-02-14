package nodes

type NodeState int

const (
	Stale NodeState = iota
	Processed
	Error
)

type Node interface {
	Versioned
	Stateful
	Subscribable
	Dependent
}

// AHHHHHHHHHHHHH =============================================================

type nodeData struct {
	version int
	state   NodeState
	subs    []Alertable
}

func (nd *nodeData) incVersion() {
	nd.version++
}

func (s nodeData) State() NodeState {
	return s.state
}

func (v nodeData) Version() int {
	return v.version
}

func (v *nodeData) MarkStale() {
	v.state = Stale
}

func (v *nodeData) AddSubscription(a Alertable) {
	v.subs = append(v.subs, a)
}

func (v *nodeData) alertSubscribers() {
	for _, a := range v.subs {
		if a != nil {
			a.Alert(v.version, v.state)
		}
	}
}

//
// N --\
//      x -- N -- \
// N --/           \
//                  x -- N
// N --\           /
//      x -- N -- /
// N --/
//
