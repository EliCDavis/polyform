package nodes

import "reflect"

// DynamicAnyValue reads a dynamic input without knowing its type, which is
// what a node generic over every type has to do.
func DynamicAnyValue(port OutputPort) (any, bool) {
	value, ok := dynamicValueMethod(port)
	if !ok {
		return nil, false
	}
	return value.Interface(), true
}

func dynamicValueMethod(port OutputPort) (reflect.Value, bool) {
	if port == nil {
		return reflect.Value{}, false
	}
	method := reflect.ValueOf(port).MethodByName("Value")
	if !method.IsValid() {
		return reflect.Value{}, false
	}
	t := method.Type()
	if t.NumIn() != 0 || t.NumOut() != 1 {
		return reflect.Value{}, false
	}
	return method.Call(nil)[0], true
}

// Declared, not held: an Output[image.Image] carrying a *image.RGBA still
// has to build []image.Image.
func dynamicElemType(port OutputPort) (reflect.Type, bool) {
	return outputValueReturnType(port)
}

// DynamicArray is an array-carrying dynamic input, read elementwise without
// naming its type. Results built from it keep the original element type.
type DynamicArray struct {
	slice reflect.Value
}

// DynamicArrayValue reads a port declared DynamicPort[[]V].
func DynamicArrayValue(port OutputPort) (DynamicArray, bool) {
	value, ok := dynamicValueMethod(port)
	if !ok {
		return DynamicArray{}, false
	}
	if value.Kind() != reflect.Slice {
		return DynamicArray{}, false
	}
	return DynamicArray{slice: value}, true
}

func (a DynamicArray) Len() int {
	if !a.slice.IsValid() {
		return 0
	}
	return a.slice.Len()
}

func (a DynamicArray) At(index int) any {
	if !a.slice.IsValid() || index < 0 || index >= a.slice.Len() {
		return nil
	}
	return a.slice.Index(index).Interface()
}

func (a DynamicArray) Values() []any {
	out := make([]any, a.Len())
	for i := range out {
		out[i] = a.At(i)
	}
	return out
}

// Build makes a new array of the same element type from values.
func (a DynamicArray) Build(values []any) any {
	if !a.slice.IsValid() {
		return nil
	}
	return buildSlice(a.slice.Type().Elem(), values)
}

// Range is Build for a contiguous run, clamped to what is actually there.
func (a DynamicArray) Range(from, to int) any {
	if !a.slice.IsValid() {
		return nil
	}
	length := a.slice.Len()
	from = max(from, 0)
	to = min(to, length)
	if from >= to {
		return reflect.MakeSlice(a.slice.Type(), 0, 0).Interface()
	}
	return a.slice.Slice(from, to).Interface()
}

// T comes from the sample port, not the values, so an array of an interface
// type does not collapse to whatever concrete type happened to be first.
func DynamicArrayOf(sample OutputPort, values []any) (any, bool) {
	elem, ok := dynamicElemType(sample)
	if !ok {
		return nil, false
	}
	return buildSlice(elem, values), true
}

func buildSlice(elem reflect.Type, values []any) any {
	slice := reflect.MakeSlice(reflect.SliceOf(elem), 0, len(values))
	for _, value := range values {
		if value == nil {
			slice = reflect.Append(slice, reflect.Zero(elem))
			continue
		}
		rv := reflect.ValueOf(value)
		if !rv.Type().AssignableTo(elem) {
			continue
		}
		slice = reflect.Append(slice, rv)
	}
	return slice.Interface()
}
