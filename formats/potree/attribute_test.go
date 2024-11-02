package potree_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/stretchr/testify/assert"
)

func TestAttributeSize(t *testing.T) {
	tests := map[string]struct {
		input  potree.AttributeType
		result int
	}{
		"int8": {
			input:  potree.Int8AttributeType,
			result: 1,
		},
		"uint8": {
			input:  potree.UInt8AttributeType,
			result: 1,
		},
		"int16": {
			input:  potree.Int16AttributeType,
			result: 2,
		},
		"uint16": {
			input:  potree.UInt16AttributeType,
			result: 2,
		},
		"int32": {
			input:  potree.Int32AttributeType,
			result: 4,
		},
		"uint32": {
			input:  potree.UInt32AttributeType,
			result: 4,
		},
		"int64": {
			input:  potree.Int64AttributeType,
			result: 8,
		},
		"uint64": {
			input:  potree.UInt64AttributeType,
			result: 8,
		},
		"double": {
			input:  potree.DoubleAttributeType,
			result: 8,
		},
		"float": {
			input:  potree.FloatAttributeType,
			result: 4,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.result, test.input.Size())
		})
	}

	assert.PanicsWithError(t, "unimplemented byte size for attribute type: undefined", func() {
		potree.UndefinedAttributeType.Size()
	})
}
