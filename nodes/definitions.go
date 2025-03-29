package nodes

type Named interface {
	Name() string
}

type Typed interface {
	Type() string
}

type Pathed interface {
	Path() string
}

type Describable interface {
	Description() string
}
