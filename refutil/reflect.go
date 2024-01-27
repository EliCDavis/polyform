package refutil

import (
	"fmt"
	"reflect"
)

func FieldValuesOfType[T any](in any) []T {

	deps := make([]T, 0)

	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	viewKind := view.Kind()
	if viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	if viewKind != reflect.Struct {
		panic(fmt.Errorf("views of type: '%s' can not be populated", viewKind.String()))
	}

	viewType := view.Type()

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)
		viewFieldValueKind := viewFieldValue.Kind()
		if viewFieldValue.CanInterface() && viewFieldValueKind == reflect.Interface {
			i := viewFieldValue.Interface()
			perm, ok := i.(T)
			if !ok {
				panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
			}
			deps = append(deps, perm)

			continue
		}

		// panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s", structField.Name, viewFieldValueKind.String()))
	}

	return deps
}
