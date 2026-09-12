package obj_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const triangleUsingMaterial = `mtllib lookup.mtl
v 0 0 0
v 1 0 0
v 0 1 0
usemtl brass
f 1 2 3
`

func writeObj(t *testing.T, dir string, mtl string) string {
	t.Helper()
	objPath := filepath.Join(dir, "tri.obj")
	require.NoError(t, os.WriteFile(objPath, []byte(triangleUsingMaterial), 0o644))
	if mtl != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "lookup.mtl"), []byte(mtl), 0o644))
	}
	return objPath
}

func onlyMaterial(t *testing.T, scene *obj.Scene) *obj.Material {
	t.Helper()
	require.Len(t, scene.Objects, 1)
	require.Len(t, scene.Objects[0].Entries, 1)
	return scene.Objects[0].Entries[0].Material
}

func TestLoadKeepsTheMaterialNameWhenItsLibraryIsMissing(t *testing.T) {
	scene, err := obj.Load(writeObj(t, t.TempDir(), ""))
	require.NoError(t, err)

	material := onlyMaterial(t, scene)
	require.NotNil(t, material)
	assert.Equal(t, "brass", material.Name)
	assert.Nil(t, material.DiffuseColor)
}

func TestLoadKeepsTheMaterialNameWhenItsLibraryLacksIt(t *testing.T) {
	scene, err := obj.Load(writeObj(t, t.TempDir(), "newmtl copper\nKd 0.9 0.5 0.2\n"))
	require.NoError(t, err)

	material := onlyMaterial(t, scene)
	require.NotNil(t, material)
	assert.Equal(t, "brass", material.Name)
	assert.Nil(t, material.DiffuseColor)
}

func TestLoadResolvesAMaterialItsLibraryDefines(t *testing.T) {
	scene, err := obj.Load(writeObj(t, t.TempDir(), "newmtl brass\nKd 0.9 0.5 0.2\n"))
	require.NoError(t, err)

	material := onlyMaterial(t, scene)
	require.NotNil(t, material)
	assert.Equal(t, "brass", material.Name)
	assert.NotNil(t, material.DiffuseColor)
}

func TestLoadStillFailsOnALibraryItCannotParse(t *testing.T) {
	_, err := obj.Load(writeObj(t, t.TempDir(), "newmtl brass\nKd not a color\n"))
	require.Error(t, err)
}

func TestSaveRoundTripsAMaterialThroughANewDirectory(t *testing.T) {
	scene, err := obj.Load(writeObj(t, t.TempDir(), "newmtl brass\nKd 0.9 0.5 0.2\n"))
	require.NoError(t, err)

	out := filepath.Join(t.TempDir(), "nested", "deeper", "tri.obj")
	require.NoError(t, obj.Save(out, *scene))

	written, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(written), "mtllib tri.mtl\n")

	reloaded, err := obj.Load(out)
	require.NoError(t, err)
	material := onlyMaterial(t, reloaded)
	require.NotNil(t, material)
	assert.Equal(t, "brass", material.Name)
	assert.NotNil(t, material.DiffuseColor)
}
