package variable_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Finish int

const (
	Matte Finish = iota
	Gloss
	Satin
)

var finishes = enum.Define(
	enum.Value(Matte, "Matte", ""),
	enum.Value(Gloss, "Gloss", ""),
	enum.Value(Satin, "Satin", ""),
)

func TestEnumVariableRoundTripsByName(t *testing.T) {
	v := variable.NewEnumTypeVariable(finishes)
	v.SetValue(Satin)

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"variable_test.Finish","value":"Satin"}`, string(data))

	applied, err := v.ApplyMessage([]byte(`"gloss"`))
	require.NoError(t, err)
	assert.True(t, applied)
	assert.Equal(t, Gloss, v.GetValue())
	assert.Equal(t, `"Gloss"`, string(v.ToMessage()))
}

func TestEnumVariableSchemaNamesTheValue(t *testing.T) {
	sys := variable.NewSystem()
	require.NoError(t, sys.Add("Paint/Finish", variable.NewEnumTypeVariable(finishes)))

	runtime, err := sys.RuntimeSchema()
	require.NoError(t, err)
	finish := runtime.SubGroups["Paint"].Variables["Finish"]
	assert.Equal(t, "variable_test.Finish", finish.Type)
	assert.Equal(t, "Matte", finish.Value)

	prop := sys.SwaggerDefinition().Properties["Paint/Finish"]
	assert.Equal(t, swagger.StringPropertyType, prop.Type)
	assert.Equal(t, []any{"Matte", "Gloss", "Satin"}, prop.Enum)
}

func TestProfileWithAnUnknownEnumValueIsRefusedWithThePath(t *testing.T) {
	sys := variable.NewSystem()
	require.NoError(t, sys.Add("Paint/Finish", variable.NewEnumTypeVariable(finishes)))

	err := sys.ApplyProfile(variable.Profile{"Paint/Finish": json.RawMessage(`"Shiny"`)})
	assert.EqualError(t, err, `Paint/Finish: "Shiny" is not a Finish; expected Matte, Gloss or Satin`)

	require.NoError(t, sys.ApplyProfile(variable.Profile{"Paint/Finish": json.RawMessage(`"Satin"`)}))
	v, err := sys.Variable("Paint/Finish")
	require.NoError(t, err)
	assert.Equal(t, Satin, v.(*variable.EnumTypeVariable[Finish]).GetValue())
}
