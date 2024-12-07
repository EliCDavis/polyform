package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestConstructor_Position(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.Position(vector3.New(0., 1., 2.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 1., 2.), position)
	assert.Equal(t, quaternion.Identity(), rotation)
	assert.Equal(t, vector3.New(1., 1., 1.), scale)

}

func TestConstructor_Rotation(t *testing.T) {

	// ARRANGE ================================================================
	rot := quaternion.FromTheta(math.Pi, vector3.Up[float64]())
	transform := trs.Rotation(rot)

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 0., 0.), position)
	assert.Equal(t, rot, rotation)
	assert.Equal(t, vector3.New(1., 1., 1.), scale)

}

func TestConstructor_Scale(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.Scale(vector3.New(0., 1., 2.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 0., 0.), position)
	assert.Equal(t, quaternion.Identity(), rotation)
	assert.Equal(t, vector3.New(0., 1., 2.), scale)

}

func TestConstructor_New(t *testing.T) {

	// ARRANGE ================================================================
	rot := quaternion.FromTheta(math.Pi, vector3.Up[float64]())
	transform := trs.New(vector3.New(1., 2., 3.), rot, vector3.New(4., 5., 6.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(1., 2., 3.), position)
	assert.Equal(t, rot, rotation)
	assert.Equal(t, vector3.New(4., 5., 6.), scale)

}
