package enum_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Compression int

const (
	Fast Compression = iota
	Balanced
	Smallest
)

var compressions = enum.Define(
	enum.Value(Fast, "Fast", "Quickest encode"),
	enum.Value(Balanced, "Balanced", ""),
	enum.Value(Smallest, "Smallest", "Smallest file"),
)

type unlistedZero int

type dupName int

type dupValue int

func TestDefineRefusesAnUnlistedZero(t *testing.T) {
	assert.PanicsWithError(t, "enum unlistedZero: the zero value 0 must be listed; it is what an unwired port holds", func() {
		enum.Define(enum.Value(unlistedZero(1), "One", ""))
	})
}

func TestDefineRefusesDuplicates(t *testing.T) {
	assert.PanicsWithError(t, `enum dupName: name "a" listed twice`, func() {
		enum.Define(enum.Value(dupName(0), "A", ""), enum.Value(dupName(1), "a", ""))
	})
	assert.PanicsWithError(t, "enum dupValue: value 0 listed twice", func() {
		enum.Define(enum.Value(dupValue(0), "A", ""), enum.Value(dupValue(0), "B", ""))
	})
}

func TestDefineTwicePanics(t *testing.T) {
	assert.PanicsWithError(t, "enum Compression: defined twice", func() {
		enum.Define(enum.Value(Fast, "Fast", ""))
	})
}

func TestParseListsTheNamesAndIgnoresCase(t *testing.T) {
	c, err := compressions.Parse(" smallest ")
	require.NoError(t, err)
	assert.Equal(t, Smallest, c)

	_, err = compressions.Parse("Tiny")
	assert.EqualError(t, err, `"Tiny" is not a Compression; expected Fast, Balanced or Smallest`)
}

func TestNamesValuesAndDescriptions(t *testing.T) {
	assert.Equal(t, []string{"Fast", "Balanced", "Smallest"}, compressions.Names())
	assert.Equal(t, []Compression{Fast, Balanced, Smallest}, compressions.Values())
	assert.Equal(t, "Smallest", compressions.Name(Smallest))
	assert.Equal(t, "Smallest file", compressions.Description(Smallest))
	assert.Equal(t, "Quickest encode", compressions.Describe("fast"))
	assert.Equal(t, "", compressions.Name(Compression(9)))
}

func TestLookupByTypeNameAndByType(t *testing.T) {
	def, ok := enum.Lookup("github.com/EliCDavis/polyform/generator/enum_test.Compression")
	require.True(t, ok)
	assert.Equal(t, []string{"Fast", "Balanced", "Smallest"}, def.Names())
	assert.Equal(t, "Smallest file", def.Describe("Smallest"))

	name, ok := def.NameOf(Smallest)
	require.True(t, ok)
	assert.Equal(t, "Smallest", name)
	_, ok = def.NameOf(3.5)
	assert.False(t, ok)

	value, err := def.ValueOf("fast")
	require.NoError(t, err)
	assert.Equal(t, Fast, value)

	_, ok = enum.Lookup("int")
	assert.False(t, ok)
}
