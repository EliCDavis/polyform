package refutil

import (
	"fmt"
	"reflect"
	"slices"
)

type typeEntry struct {
	pkg     string
	builder func() any
}

type TypeFactory struct {
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

func (tf TypeFactory) Types() []string {
	t := make([]string, 0, len(tf.types))
	for key := range tf.types {
		t = append(t, key)
	}
	slices.Sort(t)

	return t
}

func (tf TypeFactory) KeyRegistered(key string) bool {
	if tf.types == nil {
		return false
	}
	_, ok := tf.types[key]
	return ok
}

func (tf TypeFactory) TypeRegistered(v any) bool {
	if tf.types == nil {
		return false
	}
	_, ok := tf.types[GetTypeWithPackage(v)]
	return ok
}

func (tf TypeFactory) New(key string) any {
	if tf.types != nil {
		if entry, ok := tf.types[key]; ok {
			return entry.builder()
		}
	}

	panic(fmt.Errorf("type factory has no type registered for key '%s'", key))
}

func (factory *TypeFactory) RegisterType(v any) {
	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	typeEle := reflect.TypeOf(v)
	for typeEle.Kind() == reflect.Pointer {
		typeEle = typeEle.Elem()
	}

	factory.types[GetTypeWithPackage(v)] = typeEntry{
		pkg: typeEle.PkgPath(),
		builder: func() any {
			return reflect.New(typeEle).Interface()
		},
	}
}

func (factory *TypeFactory) RegisterBuilder(key string, builder func() any) {
	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}
	factory.types[key] = typeEntry{
		// pkg: typeEle.PkgPath(),
		builder: builder,
	}
}

func (factory TypeFactory) Combine(others ...*TypeFactory) *TypeFactory {
	newFactory := make(map[string]typeEntry)

	for key, val := range factory.types {
		newFactory[key] = val
	}

	for _, f := range others {
		for key, val := range f.types {
			if _, ok := newFactory[key]; ok {
				panic(fmt.Errorf("combining type factories led to a collision: '%s'", key))
			}

			newFactory[key] = val
		}
	}

	return &TypeFactory{
		types: newFactory,
	}
}

func RegisterType[T any](factory *TypeFactory) {
	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	factory.types[GetTypeWithPackage(new(T))] = typeEntry{
		pkg: GetPackagePath(new(T)),
		builder: func() any {
			var v T
			return &v
		},
	}
}

func RegisterTypeWithBuilder[T any](factory *TypeFactory, builder func() T) {
	if factory.types == nil {
		factory.types = make(map[string]typeEntry)
	}

	factory.types[GetTypeWithPackage(new(T))] = typeEntry{
		pkg: GetPackagePath(new(T)),
		builder: func() any {
			v := builder()
			return &v
		},
	}
}

func BuildType[T any](factory *TypeFactory) *T {
	typeName := GetTypeWithPackage(new(T))
	built := factory.New(typeName)
	cast, ok := built.(*T)

	if !ok {
		panic(fmt.Errorf("unable to construct type %s, instead constructed %v", typeName, built))
	}

	return cast
}
