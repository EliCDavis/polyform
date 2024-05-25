package swagger_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/stretchr/testify/assert"
)

func TestDefinitionRef(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"simple": {
			input: "simple",
			want:  "#/definitions/simple",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, swagger.DefinitionRefPath(tc.input))
		})
	}
}

func TestDefinitionRef_PanicOnEmpty(t *testing.T) {
	assert.PanicsWithError(t, "can not build definition reference from an empty string", func() {
		swagger.DefinitionRefPath("")
	})
}
