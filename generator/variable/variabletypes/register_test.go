package variabletypes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Weave int

const (
	Plain Weave = iota
	Twill
)

var weaves = enum.Define(enum.Value(Plain, "Plain", ""), enum.Value(Twill, "Twill", ""))

func TestRegisterEnumAddsARowAndRefusesASecond(t *testing.T) {
	variabletypes.RegisterEnum(weaves)

	v, err := variabletypes.New("variabletypes_test.Weave")
	require.NoError(t, err)
	assert.IsType(t, &variable.EnumTypeVariable[Weave]{}, v)
	assert.Contains(t, variabletypes.Keys(), "variabletypes_test.Weave")

	assert.PanicsWithError(t, `variable type "variabletypes_test.Weave" registered twice`, func() {
		variabletypes.RegisterEnum(weaves)
	})
}
