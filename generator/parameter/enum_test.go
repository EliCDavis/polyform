package parameter_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/generator/variable/variabletypes"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Grain int

const (
	Fine Grain = iota
	Coarse
)

var grains = parameter.DefineEnum(enum.Value(Fine, "Fine", "Small particles"), enum.Value(Coarse, "Coarse", ""))

type GrainUser struct {
	Grain nodes.Output[Grain] `description:"How fine the surface is."`
	Many  nodes.Output[[]Grain]
}

func (GrainUser) Out(out *nodes.StructOutput[string]) {
	out.Set(grains.Name(nodes.TryGetOutputValue(out, out.Node().(*nodes.Struct[GrainUser]).Data.Grain, Fine)))
}

func TestDefineEnumRegistersParameterNodeAndVariableType(t *testing.T) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[GrainUser]](factory)
	nodes.DiscoverPortTypes(factory)

	assert.Equal(t, []string{"Fine", "Coarse"}, grains.Names())

	key := "github.com/EliCDavis/polyform/generator/parameter.Enum[github.com/EliCDavis/polyform/generator/parameter_test.Grain]"
	require.True(t, registry.Types().KeyRegistered(key))
	require.True(t, registry.Types().KeyRegistered("github.com/EliCDavis/polyform/generator/enum.Members[github.com/EliCDavis/polyform/generator/parameter_test.Grain]"))
	assert.True(t, nodes.IsPortTypeKnown("github.com/EliCDavis/polyform/generator/parameter_test.Grain"))

	_, err := variabletypes.New("parameter_test.Grain")
	require.NoError(t, err)

	param := registry.Types().New(key).(*parameter.Enum[Grain])
	assert.Equal(t, "github.com/EliCDavis/polyform/generator/parameter_test.Grain", param.Outputs()["Value"].(nodes.Typed).Type())

	data, err := json.Marshal(param.Schema())
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"parameter_test.Grain"`)
	assert.Contains(t, string(data), `"currentValue":"Fine"`)

	_, err = param.ApplyMessage([]byte(`"coarse"`))
	require.NoError(t, err)
	assert.Equal(t, Coarse, param.Value())
	assert.Equal(t, `"Coarse"`, string(param.ToMessage()))

	_, err = param.ApplyMessage([]byte(`"Gritty"`))
	assert.EqualError(t, err, `"Gritty" is not a Grain; expected Fine or Coarse`)

	saved, err := param.ToJSON(nil)
	require.NoError(t, err)
	restored := registry.Types().New(key).(*parameter.Enum[Grain])
	require.NoError(t, restored.FromJSON(jbtf.Decoder{}, saved))
	assert.Equal(t, Coarse, restored.Value())
}
