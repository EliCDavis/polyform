package schema_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/stretchr/testify/assert"
)

func TestParameter(t *testing.T) {
	base := schema.ParameterBase{
		Name: "name",
		Type: "type",
	}

	assert.Equal(t, "name", base.DisplayName())
	assert.Equal(t, "type", base.ValueType())
}
