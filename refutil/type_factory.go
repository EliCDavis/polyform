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

func (tf TypeFactory) Types() []string {
	t := make([]string, 0, len(tf.types))
	for key := range tf.types {
		t = append(t, key)
	}
	slices.Sort(t)

	return t
}

func (tf TypeFactory) New(key string) any {
	if tf.types != nil {
		if entry, ok := tf.types[key]; ok {
			return entry.builder()
		}
	}

	panic(fmt.Errorf("type factory has no type registered for key %s", key))
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
