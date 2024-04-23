package pgtf_test

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/EliCDavis/polyform/formats/pgtf"
	"github.com/stretchr/testify/assert"
)

type basicStruct struct {
	A int
	B bool
	C string
}

type pgtfSerializableStruct struct {
	data int32
}

func (pss *pgtfSerializableStruct) Deserialize(in io.Reader) (err error) {
	data := make([]byte, 4)
	_, err = io.ReadFull(in, data)
	i := binary.LittleEndian.Uint32(data)
	pss.data = int32(i)
	return
}

func (pss pgtfSerializableStruct) Serialize(out io.Writer) (err error) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(pss.data))
	_, err = out.Write(bytes)
	return
}

type embeddedBinaryStruct struct {
	Basic        basicStruct
	Serializable *pgtfSerializableStruct
}

type multipleEmbeddedBinaryStruct struct {
	Basic         basicStruct
	SerializableA *pgtfSerializableStruct
	SerializableB *pgtfSerializableStruct
}

// TESTING ====================================================================

type testCase interface {
	Run(t *testing.T)
}

type typedTestCase[T any] struct {
	input T
	want  string
}

func (tc typedTestCase[T]) Run(t *testing.T) {
	out, err := pgtf.Marshal(tc.input)
	assert.NoError(t, err)
	assert.Equal(t, tc.want, string(out))

	v, err := pgtf.Unmarshal[T](out)
	assert.NoError(t, err)
	assert.Equal(t, tc.input, v)
}

func TestMarshal(t *testing.T) {

	tests := map[string]testCase{
		"single bool: True": typedTestCase[bool]{
			input: true,
			want: `{
	"data": true
}`,
		},
		"single bool: False": typedTestCase[bool]{
			input: false,
			want: `{
	"data": false
}`,
		},
		"single int: 123": typedTestCase[int]{
			input: 123,
			want: `{
	"data": 123
}`,
		},
		"single string: bababa": typedTestCase[string]{
			input: "bababa",
			want: `{
	"data": "bababa"
}`,
		},
		"basic struct": typedTestCase[basicStruct]{
			input: basicStruct{A: 123, B: true, C: "yee haw"},
			want: `{
	"data": {
		"A": 123,
		"B": true,
		"C": "yee haw"
	}
}`,
		},
		"embedded binary": typedTestCase[embeddedBinaryStruct]{
			input: embeddedBinaryStruct{
				Basic:        basicStruct{A: 123, B: true, C: "yee haw"},
				Serializable: &pgtfSerializableStruct{data: 12345},
			},
			want: `{
	"buffers": [
		{
			"byteLength": 4,
			"uri": "data:application/octet-stream;base64,OTAAAA=="
		}
	],
	"bufferViews": [
		{
			"buffer": 0,
			"byteLength": 4
		}
	],
	"data": {
		"$Serializable": 0,
		"Basic": {
			"A": 123,
			"B": true,
			"C": "yee haw"
		}
	}
}`,
		},
		"multiple embedded binary": typedTestCase[multipleEmbeddedBinaryStruct]{
			input: multipleEmbeddedBinaryStruct{
				Basic:         basicStruct{A: 123, B: true, C: "yee haw"},
				SerializableA: &pgtfSerializableStruct{data: 12345},
				SerializableB: &pgtfSerializableStruct{data: 67890},
			},
			want: `{
	"buffers": [
		{
			"byteLength": 8,
			"uri": "data:application/octet-stream;base64,OTAAADIJAQA="
		}
	],
	"bufferViews": [
		{
			"buffer": 0,
			"byteLength": 4
		},
		{
			"buffer": 0,
			"byteOffset": 4,
			"byteLength": 4
		}
	],
	"data": {
		"$SerializableA": 0,
		"$SerializableB": 1,
		"Basic": {
			"A": 123,
			"B": true,
			"C": "yee haw"
		}
	}
}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.Run(t)
		})
	}
}
