package nodes

type TransformerNode[Tin any, Tout any] struct {
	NodeData

	value Tout
	err   error

	in        Tin
	transform func(in Tin) (Tout, error)
}

func (tn TransformerNode[Tin, Tout]) Dependencies() []Dependency {
	return FieldValuesOfType[Dependency](tn.in)
}

func (tn TransformerNode[Tin, Tout]) Data() Tout {
	return tn.value
}

func (tn *TransformerNode[Tin, Tout]) Process() {
	tn.value, tn.err = tn.transform(tn.in)
	tn.version++
	tn.state = Processed
}

func Transformer[Tin any, Tout any](in Tin, trasnformer func(in Tin) (Tout, error)) *TransformerNode[Tin, Tout] {
	return &TransformerNode[Tin, Tout]{
		NodeData: NodeData{
			version: 0,
			state:   Stale,
			subs:    make([]Alertable, 0),
		},
		in:        in,
		transform: trasnformer,
	}
}
