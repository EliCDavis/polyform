package graph_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type savedBuffers struct {
	Buffers []struct {
		ByteLength int `json:"byteLength"`
	} `json:"buffers"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
	} `json:"bufferViews"`
	Data struct {
		Profiles map[string]struct {
			Data map[string]json.RawMessage `json:"data"`
		} `json:"profiles"`
	} `json:"data"`
}

func parseSaved(t testing.TB, payload []byte) savedBuffers {
	t.Helper()
	var saved savedBuffers
	require.NoError(t, json.Unmarshal(payload, &saved))
	return saved
}

func fileVariable(t testing.TB, inst *graph.Instance, path string, content []byte) *variable.FileVariable {
	t.Helper()
	v := &variable.FileVariable{}
	v.SetValue(content)
	require.NoError(t, errOf(inst.NewVariable(path, v)))
	return v
}

func TestProfileHoldingAFileSharesTheVariablesBuffer(t *testing.T) {
	inst := importTestInstance(t, "Files")
	content := bytes.Repeat([]byte("ply"), 100)
	fileVariable(t, inst, "Splat", content)
	inst.SaveProfile("default")

	saved := parseSaved(t, encode(t, inst))
	require.Len(t, saved.Buffers, 1)
	assert.Equal(t, len(content), saved.Buffers[0].ByteLength)
	require.Len(t, saved.BufferViews, 1)
	assert.Equal(t, len(content), saved.BufferViews[0].ByteLength)
	assert.JSONEq(t, `{"$Value": 0}`, string(saved.Data.Profiles["default"].Data["Splat"]))
}

func TestTwoProfilesHoldingTheSameFileWriteItOnce(t *testing.T) {
	inst := importTestInstance(t, "Files")
	content := bytes.Repeat([]byte("ply"), 100)
	fileVariable(t, inst, "Splat", content)
	inst.SaveProfile("one")
	inst.SaveProfile("two")

	saved := parseSaved(t, encode(t, inst))
	require.Len(t, saved.BufferViews, 1)
	assert.Equal(t, len(content), saved.Buffers[0].ByteLength)
}

func TestProfileHoldingADifferentFileWritesASecondBuffer(t *testing.T) {
	inst := importTestInstance(t, "Files")
	first := bytes.Repeat([]byte("one"), 100)
	second := bytes.Repeat([]byte("two"), 50)
	v := fileVariable(t, inst, "Splat", first)
	inst.SaveProfile("first")
	v.SetValue(second)
	inst.SaveProfile("second")

	saved := parseSaved(t, encode(t, inst))
	require.Len(t, saved.BufferViews, 2)
	assert.Equal(t, len(first)+len(second), saved.Buffers[0].ByteLength)
	assert.NotEqual(t, saved.Data.Profiles["first"].Data["Splat"], saved.Data.Profiles["second"].Data["Splat"])
}

func TestLoadingAndApplyingAProfileRestoresTheFileWithoutCopying(t *testing.T) {
	inst := importTestInstance(t, "Files")
	content := bytes.Repeat([]byte("ply"), 100)
	fileVariable(t, inst, "Splat", content)
	inst.SaveProfile("default")

	reloaded := importTestInstance(t, "Reloaded")
	require.NoError(t, reloaded.ApplyAppSchema(encode(t, inst)))
	loaded, err := reloaded.GetVariable("Splat")
	require.NoError(t, err)
	file := loaded.(*variable.FileVariable)
	assert.Equal(t, content, file.GetValue())
	shared := &file.GetValue()[0]

	file.SetValue([]byte("something else"))
	require.NoError(t, reloaded.LoadProfile("default"))
	assert.Equal(t, content, file.GetValue())
	assert.Same(t, shared, &file.GetValue()[0], "applying the profile reuses the loaded bytes")

	saved := parseSaved(t, encode(t, reloaded))
	require.Len(t, saved.BufferViews, 1)
}

func TestDeletingTheOnlyProfileHoldingAFileReleasesIt(t *testing.T) {
	inst := importTestInstance(t, "Files")
	v := fileVariable(t, inst, "Splat", []byte("old content"))
	inst.SaveProfile("old")
	old, err := inst.VariableProfile("Splat")
	require.NoError(t, err)

	v.SetValue([]byte("new content"))
	require.NoError(t, inst.LoadProfile("old"))
	assert.Equal(t, []byte("old content"), v.GetValue())

	v.SetValue([]byte("new content"))
	require.NoError(t, inst.Profiles().Delete("old"))
	err = inst.ApplyProfile(variable.Profile{"Splat": old})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not loaded")
}

func TestEncodingIsStableAcrossSavesWithSeveralFilesAndProfiles(t *testing.T) {
	inst := importTestInstance(t, "Files")
	a := fileVariable(t, inst, "A", bytes.Repeat([]byte("a"), 64))
	fileVariable(t, inst, "B", bytes.Repeat([]byte("b"), 32))
	inst.SaveProfile("one")
	a.SetValue(bytes.Repeat([]byte("c"), 16))
	inst.SaveProfile("two")

	first := encode(t, inst)
	for range 5 {
		assert.Equal(t, first, encode(t, inst))
	}
}

func TestFileProfileEntryAcceptsInlineBytes(t *testing.T) {
	inst := importTestInstance(t, "Files")
	v := fileVariable(t, inst, "Splat", []byte("start"))
	inline, err := json.Marshal([]byte("from the command line"))
	require.NoError(t, err)

	require.NoError(t, inst.ApplyProfile(variable.Profile{"Splat": inline}))
	assert.Equal(t, []byte("from the command line"), v.GetValue())
}
