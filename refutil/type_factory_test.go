package refutil_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

func TestTypeFactory_RegisterType(t *testing.T) {
	factory := &refutil.TypeFactory{}
	assert.False(t, factory.KeyRegistered("int"))
	assert.False(t, factory.TypeRegistered(7))

	assert.Len(t, factory.Types(), 0)

	refutil.RegisterType[int](factory)
	built := factory.New("int")
	cast, ok := built.(*int)
	assert.NotNil(t, cast)
	assert.Equal(t, 0, *cast)
	assert.True(t, ok)

	var reader io.Reader = &bytes.Buffer{}
	factory.RegisterType(reader)
	built2 := factory.New("bytes.Buffer")
	cast2, ok := built2.(*bytes.Buffer)
	assert.NotNil(t, cast2)
	assert.True(t, ok)

	types := factory.Types()
	assert.Len(t, types, 2)
	assert.Equal(t, "bytes.Buffer", types[0])
	assert.Equal(t, "int", types[1])

	assert.True(t, factory.KeyRegistered("int"))
	assert.True(t, factory.TypeRegistered(7))
	assert.False(t, factory.KeyRegistered("string"))
	assert.False(t, factory.TypeRegistered("string"))
}

func TestTypeFactory_Combine(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	other := &refutil.TypeFactory{}

	refutil.RegisterType[float64](factory)
	refutil.RegisterType[int](other)

	// ACT ====================================================================
	combined := factory.Combine(other)

	// ASSERT =================================================================
	types := combined.Types()
	assert.Len(t, types, 2)
	assert.Equal(t, "float64", types[0])
	assert.Equal(t, "int", types[1])
}

func TestTypeFactory_CombinePanicsOnCollision(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	other := &refutil.TypeFactory{}

	refutil.RegisterType[float64](factory)
	refutil.RegisterType[float64](other)

	// ACT ====================================================================
	assert.PanicsWithError(t, "combining type factories led to a collision: 'float64'", func() {
		factory.Combine(other)
	})
}

func TestTypeFactory_RegisterTypeWithBuilder(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	refutil.RegisterTypeWithBuilder(factory, func() int {
		return 7
	})

	// ACT ====================================================================
	built := factory.New("int")
	cast, ok := built.(*int)

	// ASSERT =================================================================
	assert.NotNil(t, cast)
	assert.Equal(t, 7, *cast)
	assert.True(t, ok)
}

func TestTypeFactory_BuildType(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	refutil.RegisterTypeWithBuilder(factory, func() int {
		return 7
	})

	// ACT ====================================================================
	built := refutil.BuildType[int](factory)

	// ASSERT =================================================================
	assert.Equal(t, 7, *built)
}
