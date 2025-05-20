package graph_test

import (
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/stretchr/testify/assert"
)

func TestVariableGroup_Variable(t *testing.T) {
	// ARRANGE ================================================================
	vg := graph.NewVariableGroup()
	vp := "test"
	variable := &variable.TypeVariable[int]{}

	// ACT ====================================================================
	has1 := vg.HasVariable(vp)
	vg.AddVariable(vp, variable)
	has2 := vg.HasVariable(vp)
	back := vg.GetVariable(vp)
	vg.RemoveVariable(vp)
	has3 := vg.HasVariable(vp)

	// ASSERT =================================================================
	assert.Equal(t, variable, back)
	assert.False(t, has1)
	assert.True(t, has2)
	assert.False(t, has3)
}

func TestVariableGroup_NestedVariable(t *testing.T) {
	// ARRANGE ================================================================
	vg := graph.NewVariableGroup()
	subgroup := "subgroup"
	variableName := "test"
	variablePath := fmt.Sprintf("%s/%s", subgroup, variableName)
	variable := &variable.TypeVariable[int]{}

	// ACT/ASSERT =============================================================
	// Make sure we don't already have the variable
	assert.False(t, vg.HasVariable(variablePath))
	assert.False(t, vg.HasSubgroup(subgroup))

	// Add the variable
	vg.AddVariable(variablePath, variable)
	assert.False(t, vg.HasVariable(variableName))
	assert.True(t, vg.HasVariable(variablePath))
	assert.True(t, vg.HasSubgroup(subgroup))
	assert.Equal(t, variable, vg.GetVariable(variablePath))

	sg := vg.GetSubgroup(subgroup)
	assert.True(t, sg.HasVariable(variableName))

	// Remove the variable
	vg.RemoveVariable(variablePath)
	assert.False(t, vg.HasVariable(variablePath))
	assert.False(t, sg.HasVariable(variableName))
}

func TestVariableGroup_AddRemoveSubgroup(t *testing.T) {
	// ARRANGE ================================================================
	vg := graph.NewVariableGroup()
	vp := "subsub/sub"

	// ACT ====================================================================
	has1 := vg.HasSubgroup("subsub")
	vg.AddSubgroup(vp)
	has2 := vg.HasSubgroup(vp)
	has3 := vg.HasSubgroup("subsub")

	vg.RemoveSubgroup(vp)
	has4 := vg.HasSubgroup(vp)
	has5 := vg.HasSubgroup("subsub")

	// ASSERT =================================================================
	assert.False(t, has1)
	assert.True(t, has2)
	assert.True(t, has3)
	assert.False(t, has4)
	assert.True(t, has5)
}

func TestVariableGroup_HasSubgroup(t *testing.T) {
	// ARRANGE ================================================================
	vg := graph.NewVariableGroup()

	// ACT ====================================================================
	has1 := vg.HasSubgroup("a/b/c")

	// ASSERT =================================================================
	assert.False(t, has1)
}

func TestVariableGroup_HasVariable(t *testing.T) {
	// ARRANGE ================================================================
	vg := graph.NewVariableGroup()

	// ACT ====================================================================
	has1 := vg.HasVariable("a/b/c")

	// ASSERT =================================================================
	assert.False(t, has1)
}
