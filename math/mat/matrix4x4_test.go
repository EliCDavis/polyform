package mat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/stretchr/testify/assert"
)

func TestInverse(t *testing.T) {
	inv := mat.Matrix4x4{
		1, 4, 2, 3,
		6, 10, 15, 2.4,
		7, 3.2, 2, 3.4,
		5, 4.2, 4.3, 3,
	}.Inverse()

	assert.InDelta(t, -0.03802871556072906, inv.X00, 0.000001)
	assert.InDelta(t, 0.10962359332557213, inv.X01, 0.000001)
	assert.InDelta(t, 0.4714784633294516, inv.X02, 0.000001)
	assert.InDelta(t, -0.5840124175397738, inv.X03, 0.000001)

	assert.InDelta(t, 0.7697904540162962, inv.X10, 0.000001)
	assert.InDelta(t, 0.5296856810244461, inv.X11, 0.000001)
	assert.InDelta(t, 1.5148428405122205, inv.X12, 0.000001)
	assert.InDelta(t, -2.910360884749703, inv.X13, 0.000001)

	assert.InDelta(t, -0.4355840124175382, inv.X20, 0.000001)
	assert.InDelta(t, -0.23670935195964263, inv.X21, 0.000001)
	assert.InDelta(t, -0.9516880093131526, inv.X22, 0.000001)
	assert.InDelta(t, 1.7035312378734926, inv.X23, 0.000001)

	assert.InDelta(t, -0.38998835855645947, inv.X30, 0.000001)
	assert.InDelta(t, -0.5849825378346901, inv.X31, 0.000001)
	assert.InDelta(t, -1.5424912689173425, inv.X32, 0.000001)
	assert.InDelta(t, 2.9394644935971996, inv.X33, 0.000001)
}

func TestMultiply(t *testing.T) {
	a := mat.Matrix4x4{
		1.1, 2.1, 3.1, 4.1,
		5.1, 6.1, 7.1, 8.1,
		9.1, 10.1, 11.1, 12.1,
		13.1, 14.1, 15.1, 16.1,
	}
	b := mat.Matrix4x4{
		1.2, 2.2, 3.2, 4.2,
		5.2, 6.2, 7.2, 8.2,
		9.2, 10.2, 11.2, 12.2,
		13.2, 14.2, 15.2, 16.2,
	}

	c := a.Multiply(b)

	assert.InDelta(t, 94.88, c.X00, 0.000001)
	assert.InDelta(t, 105.28, c.X01, 0.000001)
	assert.InDelta(t, 115.68, c.X02, 0.000001)
	assert.InDelta(t, 126.08, c.X03, 0.000001)

	assert.InDelta(t, 210.08, c.X10, 0.000001)
	assert.InDelta(t, 236.48, c.X11, 0.000001)
	assert.InDelta(t, 262.88, c.X12, 0.000001)
	assert.InDelta(t, 289.28, c.X13, 0.000001)

	assert.InDelta(t, 325.28, c.X20, 0.000001)
	assert.InDelta(t, 367.68, c.X21, 0.000001)
	assert.InDelta(t, 410.08, c.X22, 0.000001)
	assert.InDelta(t, 452.48, c.X23, 0.000001)

	assert.InDelta(t, 440.48, c.X30, 0.000001)
	assert.InDelta(t, 498.88, c.X31, 0.000001)
	assert.InDelta(t, 557.28, c.X32, 0.000001)
	assert.InDelta(t, 615.68, c.X33, 0.000001)
}

func TestIdentity(t *testing.T) {
	inv := mat.Identity()

	assert.InDelta(t, 1, inv.X00, 0.000001)
	assert.InDelta(t, 0, inv.X01, 0.000001)
	assert.InDelta(t, 0, inv.X02, 0.000001)
	assert.InDelta(t, 0, inv.X03, 0.000001)

	assert.InDelta(t, 0, inv.X10, 0.000001)
	assert.InDelta(t, 1, inv.X11, 0.000001)
	assert.InDelta(t, 0, inv.X12, 0.000001)
	assert.InDelta(t, 0, inv.X13, 0.000001)

	assert.InDelta(t, 0, inv.X20, 0.000001)
	assert.InDelta(t, 0, inv.X21, 0.000001)
	assert.InDelta(t, 1, inv.X22, 0.000001)
	assert.InDelta(t, 0, inv.X23, 0.000001)

	assert.InDelta(t, 0, inv.X30, 0.000001)
	assert.InDelta(t, 0, inv.X31, 0.000001)
	assert.InDelta(t, 0, inv.X32, 0.000001)
	assert.InDelta(t, 1, inv.X33, 0.000001)
}

func TestAdd(t *testing.T) {
	a := mat.Matrix4x4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	b := mat.Matrix4x4{
		-1, -2, -3, -4,
		-5, -6, -7, -8,
		-9, -10, -11, -12,
		-13, -14, -15, -16,
	}

	c := a.Add(b)

	assert.InDelta(t, 0, c.X00, 0.000001)
	assert.InDelta(t, 0, c.X01, 0.000001)
	assert.InDelta(t, 0, c.X02, 0.000001)
	assert.InDelta(t, 0, c.X03, 0.000001)

	assert.InDelta(t, 0, c.X10, 0.000001)
	assert.InDelta(t, 0, c.X11, 0.000001)
	assert.InDelta(t, 0, c.X12, 0.000001)
	assert.InDelta(t, 0, c.X13, 0.000001)

	assert.InDelta(t, 0, c.X20, 0.000001)
	assert.InDelta(t, 0, c.X21, 0.000001)
	assert.InDelta(t, 0, c.X22, 0.000001)
	assert.InDelta(t, 0, c.X23, 0.000001)

	assert.InDelta(t, 0, c.X30, 0.000001)
	assert.InDelta(t, 0, c.X31, 0.000001)
	assert.InDelta(t, 0, c.X32, 0.000001)
	assert.InDelta(t, 0, c.X33, 0.000001)
}
