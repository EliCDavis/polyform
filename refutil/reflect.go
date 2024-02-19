package refutil

import (
	"fmt"
	"reflect"
)

func GetName(in any) string {
	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	viewKind := view.Kind()
	if viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}
	if viewKind != reflect.Struct {
		panic(fmt.Errorf("views of type: '%s' can not evaluate name", viewKind.String()))
	}
	viewType := view.Type()
	return viewType.Name()
}

func FieldValuesOfType[T any](in any) map[string]T {

	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	viewKind := view.Kind()

	// Dereference pointer
	if viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	if viewKind != reflect.Struct {
		panic(fmt.Errorf("views of type: '%s' can not be populated", viewKind.String()))
	}

	viewType := view.Type()

	out := make(map[string]T)

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)

		viewFieldValueKind := viewFieldValue.Kind()
		if viewFieldValue.CanInterface() && viewFieldValueKind == reflect.Interface {
			i := viewFieldValue.Interface()

			// Skip nodes that have not been set....
			// TODO: Is this really what we want to do here?
			if viewFieldValue.IsNil() {
				continue
			}

			perm, ok := i.(T)
			if !ok {
				// panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
				continue
			}
			out[structField.Name] = perm
			continue
		}

		// panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s", structField.Name, viewFieldValueKind.String()))
	}

	return out
}

func FieldValuesOfTypeInArray[T any](in any) map[string][]T {

	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	viewKind := view.Kind()

	// Dereference pointer
	if viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	if viewKind != reflect.Struct {
		panic(fmt.Errorf("views of type: '%s' can not be populated", viewKind.String()))
	}

	viewType := view.Type()

	out := make(map[string][]T)

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)

		viewFieldValueKind := viewFieldValue.Kind()
		if viewFieldValueKind == reflect.Slice {

			if viewFieldValue.IsNil() {
				continue
			}

			sliceElementType := viewFieldValue.Type().Elem()
			if sliceElementType.Kind() != reflect.Interface {
				continue
			}

			var exampleVal *T
			if !sliceElementType.Implements(reflect.TypeOf(exampleVal).Elem()) {
				continue
			}

			// This workes >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			for i := 0; i < viewFieldValue.Len(); i++ {
				element := viewFieldValue.Index(i)

				if !element.CanInterface() || element.Kind() != reflect.Interface {
					break
				}

				elementInterface := element.Interface()

				perm, ok := elementInterface.(T)
				if !ok {
					// panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
					break
				}
				out[structField.Name] = append(out[structField.Name], perm)
			}
			// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		}

		// panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s", structField.Name, viewFieldValueKind.String()))
	}

	return out
}
