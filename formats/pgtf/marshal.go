package pgtf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
)

func isStruct(v any) bool {
	value := reflect.ValueOf(v)

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	return value.Kind() == reflect.Struct
}

func buildBufferView(buf *bytes.Buffer, currentStructure map[string]any, v any, f *format) error {
	value := reflect.ValueOf(v)

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		viewFieldValue := value.Field(i)
		structField := valueType.Field(i)

		if viewFieldValue.CanInterface() {

			i := viewFieldValue.Interface()
			perm, ok := i.(PgtfSerializable)
			if !ok {
				currentStructure[structField.Name] = i
				continue
			}

			if perm == nil {
				continue
			}

			start := buf.Len()
			err := perm.Serialize(buf)
			if err != nil {
				return err
			}

			currentStructure["$"+structField.Name] = len(f.BufferViews)

			f.BufferViews = append(f.BufferViews, bufferView{
				Buffer:     0,
				ByteOffset: start,
				ByteLength: buf.Len() - start,
			})

			continue
		}
	}

	return nil
}

func Marshal(v any) ([]byte, error) {
	structure := &format{
		Data:        v,
		Buffers:     make([]buffer, 0),
		BufferViews: make([]bufferView, 0),
	}

	buf := &bytes.Buffer{}

	if isStruct(v) {
		data := make(map[string]any)
		err := buildBufferView(buf, data, v, structure)
		if err != nil {
			return nil, err
		}
		structure.Data = data
	}

	if len(structure.BufferViews) > 0 {
		bufData := buf.Bytes()
		structure.Buffers = append(structure.Buffers, buffer{
			ByteLength: len(bufData),
			URI:        "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(bufData),
		})
	}

	return json.MarshalIndent(structure, "", "\t")
}
