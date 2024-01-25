package nodes

type NodeState int

const (
	Stale NodeState = iota
	Processed
)

type Node[T any] interface {
	Versioned
	Stateful
	Subscribable
	Data() T
	Dependencies() []Dependency
}

// AHHHHHHHHHHHHH =============================================================

type NodeData struct {
	version int
	state   NodeState
	subs    []Alertable
}

func (s NodeData) State() NodeState {
	return s.state
}

func (v NodeData) Version() int {
	return v.version
}

func (v *NodeData) MarkStale() {
	v.state = Stale
}

func (v *NodeData) AddSubscription(a Alertable) {
	v.subs = append(v.subs, a)
}

func (v *NodeData) alertSubscribers() {
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

type Dependent interface {
	Versioned
	Stateful
	Staleable
	Dependencies() []Dependency
	Process()
}

type ProcessingBlockSubscription struct {
	Parent               *ProcessingBlock
	Dependency           Dependency
	lastVersionProcessed int
}

func (gds *ProcessingBlockSubscription) Alert(version int, state NodeState) {
	if gds.lastVersionProcessed < version || state == Stale {
		gds.Parent.dependent.MarkStale()
	}
}

type ProcessingBlock struct {
	dependent     Dependent
	subscriptions []*ProcessingBlockSubscription
}

func (pb ProcessingBlock) IsStale() bool {
	return pb.dependent.State() == Stale
}

func (pb ProcessingBlock) ReadyToProcess() bool {
	for _, sub := range pb.subscriptions {
		if sub.Dependency.State() == Stale {
			return false
		}
	}

	return true
}

func (pb ProcessingBlock) Process() {
	pb.dependent.Process()
	for _, sub := range pb.subscriptions {
		sub.lastVersionProcessed = sub.Dependency.Version()
	}
}

type ProcessManager struct {
	blocks map[Dependent]*ProcessingBlock
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		blocks: make(map[Dependent]*ProcessingBlock),
	}
}

func (g *ProcessManager) AddProcessNode(d Dependent) {

	if _, ok := g.blocks[d]; ok {
		return
	}

	for _, dependency := range d.Dependencies() {
		if alsoDependent, ok := dependency.(Dependent); ok {
			g.AddProcessNode(alsoDependent)
		}
	}

	pb := &ProcessingBlock{
		dependent:     d,
		subscriptions: make([]*ProcessingBlockSubscription, 0),
	}

	for _, dependency := range d.Dependencies() {
		subscription := &ProcessingBlockSubscription{
			lastVersionProcessed: -1,
			Parent:               pb,
			Dependency:           dependency,
		}
		dependency.AddSubscription(subscription)
		pb.subscriptions = append(pb.subscriptions, subscription)
	}

	g.blocks[d] = pb
}

func (g ProcessManager) Process() {
	for {
		doneProcessing := true
		staleLock := true

		for _, d := range g.blocks {
			if !d.IsStale() {
				continue
			}

			doneProcessing = false
			if d.ReadyToProcess() {
				staleLock = false
				d.Process()
			}
		}

		if doneProcessing {
			return
		}

		if staleLock {
			panic("stale lock")
		}
	}
}
