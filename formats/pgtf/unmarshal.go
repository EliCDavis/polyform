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

func BuildBuffer(buf buffer) *bytes.Reader {
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

func rebuildBuffers(bufs []buffer) []*bytes.Reader {
	vals := make([]*bytes.Reader, len(bufs))
	for i, v := range bufs {
		vals[i] = BuildBuffer(v)
	}
	return vals
}

func interpretBufferView[T any](v *T, f *typedFormat[T], rawData map[string]any) error {
	value := reflect.ValueOf(v)

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	allBuffers := rebuildBuffers(f.Buffers)

	valueType := value.Type()
	for fieldIndex := 0; fieldIndex < valueType.NumField(); fieldIndex++ {
		viewFieldValue := value.Field(fieldIndex)
		structField := valueType.Field(fieldIndex)

		if viewFieldValue.CanInterface() {
			i := viewFieldValue.Interface()
			_, ok := i.(PgtfSerializable)
			if !ok {
				continue
			}

			bufIndexKey := "$" + structField.Name
			bufIndexRaw := rawData[bufIndexKey]

			bufIndex, ok := bufIndexRaw.(float64)
			if !ok {
				panic("fuck")
			}

			bufView := f.BufferViews[int(bufIndex)]

			buf := allBuffers[bufView.Buffer]
			buf.Seek(int64(bufView.ByteOffset), io.SeekStart)

			typeEle := viewFieldValue.Type()
			for typeEle.Kind() == reflect.Pointer {
				typeEle = typeEle.Elem()
			}

			refNew := reflect.New(typeEle)
			instantiatedType := refNew.Interface()
			serializable, ok := instantiatedType.(PgtfSerializable)
			if !ok {
				panic("fuck")
			}

			err := serializable.Deserialize(buf)
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

func Unmarshal[T any](data []byte) (T, error) {
	format := &typedFormat[T]{}
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

	return format.Data, interpretBufferView(&format.Data, format, specificData)
}
