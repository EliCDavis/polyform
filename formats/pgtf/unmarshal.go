package pgtf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func jsonNameTag(v reflect.StructField) string {
	name := v.Name
	if altName, ok := v.Tag.Lookup("json"); ok {
		elements := strings.Split(altName, ",")
		name = elements[0]
	}
	return name
}

func jsonOmitTag(v reflect.StructField) bool {
	if altName, ok := v.Tag.Lookup("json"); ok {
		elements := strings.Split(altName, ",")
		if len(elements) < 2 {
			return false
		}
		if len(elements) != 2 {
			panic(fmt.Errorf("unimplemented json tag situation: %s", altName))
		}
		return elements[1] == "omitempty"
	}
	return false
}

func IsNilish(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}

	return false
}

func BuildBuffer(buf Buffer) *bytes.Reader {
	const dataUriIdentifier = "data:application/octet-stream;base64,"
	if strings.Index(buf.URI, dataUriIdentifier) == 0 {
		dataString := buf.URI[len(dataUriIdentifier):]
		data, err := base64.StdEncoding.DecodeString(dataString)
		if err != nil {
			panic(err)
		}
		return bytes.NewReader(data)
	}
	panic(fmt.Errorf("don't know how to interpret buffer: %+v", buf))
}

func rebuildBuffers(bufs []Buffer) []*bytes.Reader {
	vals := make([]*bytes.Reader, len(bufs))
	for i, v := range bufs {
		vals[i] = BuildBuffer(v)
	}
	return vals
}

func interpretBufferView(value reflect.Value, rawData map[string]any, bufferViews []BufferView, buffers []*bytes.Reader) error {

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for fieldIndex := 0; fieldIndex < valueType.NumField(); fieldIndex++ {
		viewFieldValue := value.Field(fieldIndex)
		structField := valueType.Field(fieldIndex)

		if viewFieldValue.CanInterface() {
			i := viewFieldValue.Interface()
			_, isSerializable := i.(Serializable)
			if !isSerializable {
				if isStruct(i) {
					ahh, ok := rawData[structField.Name].(map[string]any)
					if ok {
						err := interpretBufferView(viewFieldValue, ahh, bufferViews, buffers)
						if err != nil {
							return err
						}
					}
				}
				continue
			}

			bufIndexKey := "$" + structField.Name
			bufIndexRaw := rawData[bufIndexKey]

			bufIndex, isSerializable := bufIndexRaw.(float64)
			if !isSerializable {
				panic(fmt.Errorf("buffer index '%s' was not a number", bufIndexKey))
			}

			bufView := bufferViews[int(bufIndex)]

			buf := buffers[bufView.Buffer]
			_, err := buf.Seek(int64(bufView.ByteOffset), io.SeekStart)
			if err != nil {
				return err
			}

			typeEle := viewFieldValue.Type()
			for typeEle.Kind() == reflect.Pointer {
				typeEle = typeEle.Elem()
			}

			refNew := reflect.New(typeEle)
			instantiatedType := refNew.Interface()
			serializable, isSerializable := instantiatedType.(Serializable)
			if !isSerializable {
				panic("instantiated type is not serializable")
			}

			err = serializable.Deserialize(buf)
			if err != nil {
				return err
			}

			if !viewFieldValue.CanSet() {
				panic(fmt.Errorf("field '%s' was found but can not be set", structField.Name))
			}

			viewFieldValue.Set(reflect.ValueOf(serializable))
			continue
		}
	}

	return nil
}

func ParseJsonUsingBuffers[T any](buffers []Buffer, bufferViews []BufferView, data []byte) (T, error) {
	allBuffers := rebuildBuffers(buffers)

	var thingToParse T
	err := json.Unmarshal(data, &thingToParse)
	if err != nil {
		return thingToParse, err
	}

	rawData := make(map[string]any)
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		return thingToParse, err
	}

	return thingToParse, interpretBufferView(reflect.ValueOf(&thingToParse), rawData, bufferViews, allBuffers)
}

func Unmarshal[T any](data []byte) (T, error) {
	format := &Schema[T]{}
	err := json.Unmarshal(data, format)
	if err != nil {
		return format.Data, err
	}

	rawData := make(map[string]any)
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		return format.Data, err
	}

	specificData, ok := rawData["data"].(map[string]any)
	if ok {
		rawData = specificData
	}

	allBuffers := rebuildBuffers(format.Buffers)

	return format.Data, interpretBufferView(reflect.ValueOf(&format.Data), specificData, format.BufferViews, allBuffers)
}
