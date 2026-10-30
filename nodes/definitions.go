package nodes

import "github.com/EliCDavis/polyform/refutil"

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

type Advanced interface {
	Advanced() bool
}

func advancedTag(data any, field string) bool {
	return refutil.GetStructTag(data, field, "advanced") == "true"
}

type Keyworded interface {
	Keywords() []string
}
