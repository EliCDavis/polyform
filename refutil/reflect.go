package refutil

import (
	"fmt"
	"path"
	"reflect"
	"strings"
)

func GetPackagePath(v any) string {
	vType := reflect.TypeOf(v)
	if vType == nil {
		return ""
	}
	viewKind := vType.Kind()
	for viewKind == reflect.Ptr {
		vType = vType.Elem()
		viewKind = vType.Kind()
	}
	return vType.PkgPath()
}

func GetTypeWithPackage(v any) string {
	vType := reflect.TypeOf(v)
	if vType == nil {

		// e := reflect.TypeOf((&v)).Elem()

		// return e.String()
		// if e.Name() == "" {
		// 	return "nil"
		// }

		// return name

		return "nil"
	}

	viewKind := vType.Kind()
	// ptr := ""
	for viewKind == reflect.Ptr {
		vType = vType.Elem()
		viewKind = vType.Kind()
		// ptr += "*"
	}

	pkgPath := vType.PkgPath()
	if !strings.Contains(pkgPath, "/") {
		return vType.String()
	}
	return path.Dir(pkgPath) + "/" + vType.String()
}

// GetTypeName returns the name of the type of the variable provided
func GetTypeName(in any) string {
	view := reflect.TypeOf(in)
	if view == nil {
		return "nil"
	}

	// Dereference pointer
	viewKind := view.Kind()
	for viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	return view.String()
}

func GetTypeNameWithoutPackage(in any) string {
	view := reflect.TypeOf(in)
	if view == nil {
		return "nil"
	}

	// Dereference pointer
	viewKind := view.Kind()
	for viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	str := view.String()
	if !strings.Contains(str, ".") {
		return str
	}

	parts := strings.Split(str, ".")

	return parts[len(parts)-1]
}

func FuncValuesOfType[T any](in any) []string {
	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	out := make([]string, 0)

	viewType := view.Type()
	for i := 0; i < viewType.NumMethod(); i++ {
		method := viewType.Method(i)

		methodType := method.Func.Type()
		if methodType.NumOut() != 1 {
			continue
		}

		methodOutType := methodType.Out(0)
		var exampleVal *T
		if !methodOutType.Implements(reflect.TypeOf(exampleVal).Elem()) {
			continue
		}

		out = append(out, method.Name)
	}

	return out
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

			// Skip nodes that have not been set....
			// TODO: Is this really what we want to do here?
			if viewFieldValue.IsNil() {
				continue
			}

			i := viewFieldValue.Interface()
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

func SetStructField(structToSet any, field string, val any) {
	viewPointerValue := reflect.ValueOf(structToSet)

	view := viewPointerValue
	viewKind := view.Kind()

	// Dereference pointer
	for viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	if viewKind != reflect.Struct {
		panic(fmt.Errorf("value of type: '%s' has no field '%s' to set", viewKind.String(), field))
	}

	viewType := view.Type()

	for i := 0; i < viewType.NumField(); i++ {
		structField := viewType.Field(i)

		// Not our field, continue
		if structField.Name != field {
			continue
		}

		viewFieldValue := view.Field(i)

		// Can't be set
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("field '%s' was found but can not be set", field))
		}

		viewFieldValue.Set(reflect.ValueOf(val))
		return
	}

	panic(fmt.Errorf("field '%s' was not found on struct", field))
}

func GenericFieldValues(genericType string, in any) map[string]string {
	viewPointerValue := reflect.ValueOf(in)

	view := viewPointerValue
	viewKind := view.Kind()

	// Dereference pointer
	for viewKind == reflect.Ptr {
		view = view.Elem()
		viewKind = view.Kind()
	}

	if viewKind != reflect.Struct {
		panic(fmt.Errorf("value of type: '%s' can not be evaluated for generic field values", viewKind.String()))
	}

	viewType := view.Type()

	out := make(map[string]string)

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)

		typeString := viewFieldValue.Type().String()

		// It really does suck this bad at the moment while this proposal is open
		// https://stackoverflow.com/questions/73864711/get-type-parameter-from-a-generic-struct-using-reflection
		if strings.Index(typeString, genericType+"[") == 0 && typeString[len(typeString)-1:] == "]" {
			out[structField.Name] = typeString[len(genericType)+1 : len(typeString)-1]
		}
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
					break
				}
				out[structField.Name] = append(out[structField.Name], perm)
			}
			// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
		}
	}

	return out
}
