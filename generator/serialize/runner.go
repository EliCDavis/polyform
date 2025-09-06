package serialize

import (
	"slices"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type switchEntry[T any] struct {
	pkg     string
	accepts func(nodes.OutputPort) bool
	builder func(nodes.OutputPort) T
}

type TypeSwitch[T any] struct {
	types map[string]switchEntry[T]
}

func (tf TypeSwitch[T]) Types() []string {
	t := make([]string, 0, len(tf.types))
	for key := range tf.types {
		t = append(t, key)
	}
	slices.Sort(t)

	return t
}

func (tf TypeSwitch[T]) Run(op nodes.OutputPort) T {
	for _, t := range tf.types {
		if t.accepts(op) {
			return t.builder(op)
		}
	}
	panic("nothing regeistered can handle output port")
}

// func (tf TypeSwitch[T]) KeyRegistered(key string) bool {
// 	if tf.types == nil {
// 		return false
// 	}
// 	_, ok := tf.types[key]
// 	return ok
// }

// func (tf TypeSwitch[T]) TypeRegistered(v any) bool {
// 	if tf.types == nil {
// 		return false
// 	}
// 	_, ok := tf.types[refutil.GetTypeWithPackage(v)]
// 	return ok
// }

// func (tf TypeSwi3tch[T]) Run(key string) any {
// 	if tf.types != nil {
// 		if entry, ok := tf.types[key]; ok {
// 			return entry.builder()
// 		}
// 	}

// 	panic(fmt.Errorf("type factory has no type registered for key '%s'", key))
// }

func Register[T, G any](factory *TypeSwitch[T], builder func(G) T) {
	if factory.types == nil {
		factory.types = make(map[string]switchEntry[T])
	}

	factory.types[refutil.GetTypeWithPackage(new(G))] = switchEntry[T]{
		pkg: refutil.GetPackagePath(new(G)),
		accepts: func(op nodes.OutputPort) bool {
			_, ok := op.(nodes.Output[G])
			return ok
		},
		builder: func(x nodes.OutputPort) T {
			f, ok := x.(nodes.Output[G])
			if !ok {
				panic("typeswitch recieved not it's type")
			}
			return builder(f.Value())
		},
	}
}
