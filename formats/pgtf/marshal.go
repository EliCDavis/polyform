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

func buildBufferView(buf *bytes.Buffer, currentStructure, jsonInterprettedData map[string]any, v any, views []BufferView, bufferIndex int) ([]BufferView, error) {
	value := reflect.ValueOf(v)

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return views, nil
	}

	var err error

	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		viewFieldValue := value.Field(i)
		structField := valueType.Field(i)

		if viewFieldValue.CanInterface() {

			i := viewFieldValue.Interface()
			perm, isSerializable := i.(Serializable)
			if !isSerializable {
				jsonTagName := jsonNameTag(structField)

				// Golang's own JSON package didn't serialize it. The only case
				// I can think of this occuring is that the 'omitempty' is
				// present and we had the 0 value present.
				nestedInterprettedData, ok := jsonInterprettedData[jsonTagName]
				if !ok {
					continue
				}

				nestedInterprettedDataAsStruct, nestedJsonIsStruct := nestedInterprettedData.(map[string]any)

				if isStruct(i) && nestedJsonIsStruct {
					m := make(map[string]any)
					currentStructure[jsonTagName] = m
					views, err = buildBufferView(buf, m, nestedInterprettedDataAsStruct, i, views, bufferIndex)
					if err != nil {
						return views, err
					}
				} else {
					currentStructure[jsonTagName] = i
				}
				continue
			}

			if perm == nil {
				continue
			}

			start := buf.Len()
			err := perm.Serialize(buf)
			if err != nil {
				return views, err
			}

			currentStructure["$"+structField.Name] = len(views)

			views = append(views, BufferView{
				Buffer:     bufferIndex,
				ByteOffset: start,
				ByteLength: buf.Len() - start,
			})

			continue
		}
	}

	return views, nil
}

func toJsonMap(v any) (map[string]any, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	jsonInterprettedData := make(map[string]any)
	err = json.Unmarshal(jsonData, &jsonInterprettedData)
	if err != nil {
		return nil, err
	}

	return jsonInterprettedData, nil
}

func Marshal(v any) ([]byte, error) {
	structure := &schema{
		Data:        v,
		Buffers:     make([]Buffer, 0),
		BufferViews: make([]BufferView, 0),
	}

	buf := &bytes.Buffer{}

	if isStruct(v) {
		jsonInterprettedData, err := toJsonMap(v)
		if err != nil {
			return nil, err
		}

		builtDataView := make(map[string]any)
		views, err := buildBufferView(buf, builtDataView, jsonInterprettedData, v, nil, 0)
		if err != nil {
			return nil, err
		}
		structure.Data = builtDataView
		structure.BufferViews = views
	}

	if len(structure.BufferViews) > 0 {
		bufData := buf.Bytes()
		structure.Buffers = append(structure.Buffers, Buffer{
			ByteLength: len(bufData),
			URI:        "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(bufData),
		})
	}

	return json.MarshalIndent(structure, "", "\t")
}
