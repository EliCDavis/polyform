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
}
