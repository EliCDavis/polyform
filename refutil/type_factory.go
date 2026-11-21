package refutil

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type typeEntry struct {
	pkg     string
	builder func() any
}

type TypeFactory struct {
	mu    sync.RWMutex
	types map[string]typeEntry
}

// type TypeFactoryEntry interface {
// 	pkg() string
// }

// type TypeEntry[T any] struct {
// 	Builder func()
// }

// func (TypeEntry[T]) pkg() string {
// 	return ""
// }

// func NewTypeFactory(entries ...TypeFactoryEntry) {

// }

func getTypeWithPackage(v any) string {
	return TypeResolution{IncludePackage: true, IncludePointer: true}.Resolve(v)
}

func (tf *TypeFactory) Types() []string {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	t := make([]string, 0, len(tf.types))
	for key := range tf.types {
		t = append(t, key)
	}
	slices.Sort(t)

	return t
}

func (tf *TypeFactory) KeyRegistered(key string) bool {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.types == nil {
		return false
	}
	_, ok := tf.types[key]
	return ok
}

func (tf *TypeFactory) Unregister(key string) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	delete(tf.types, key)
}

func (tf *TypeFactory) TypeRegistered(v any) bool {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.types == nil {
		return false
	}
	_, ok := tf.types[getTypeWithPackage(v)]
	return ok
}

func (tf *TypeFactory) New(key string) any {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.types != nil {
		if entry, ok := tf.types[key]; ok {
			return entry.builder()
		}
	}

	panic(fmt.Errorf("type factory has no type registered for key '%s'", key))
}

func (factory *TypeFactory) RegisterType(v any) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	typeEle := reflect.TypeOf(v)
	for typeEle.Kind() == reflect.Pointer {
		typeEle = typeEle.Elem()
	}

	factory.types[getTypeWithPackage(v)] = typeEntry{
		pkg: typeEle.PkgPath(),
		builder: func() any {
			return reflect.New(typeEle).Interface()
		},
	}
}

func (factory *TypeFactory) RegisterBuilder(key string, builder func() any) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}
	factory.types[key] = typeEntry{
		// pkg: typeEle.PkgPath(),
		builder: builder,
	}
}

func (factory *TypeFactory) Combine(others ...*TypeFactory) *TypeFactory {
	factory.mu.RLock()
	newFactory := make(map[string]typeEntry, len(factory.types))
	for key, val := range factory.types {
		newFactory[key] = val
	}
	factory.mu.RUnlock()

	for _, f := range others {
		if f == nil {
			continue
		}
		f.mu.RLock()
		for key, val := range f.types {
			if _, ok := newFactory[key]; ok {
				f.mu.RUnlock()
				panic(fmt.Errorf("combining type factories led to a collision: '%s'", key))
			}

			newFactory[key] = val
		}
		f.mu.RUnlock()
	}

	return &TypeFactory{
		types: newFactory,
	}
}

func RegisterType[T any](factory *TypeFactory) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	v := new(T)
	factory.types[getTypeWithPackage(*v)] = typeEntry{
		pkg: GetPackagePath(v),
		builder: func() any {
			var v T
			return &v
		},
	}
}

func RegisterTypeWithBuilder[T any](factory *TypeFactory, builder func() T) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	t := new(T)
	factory.types[getTypeWithPackage(*t)] = typeEntry{
		pkg: GetPackagePath(t),
		builder: func() any {
			v := builder()
			return &v
		},
	}
}
