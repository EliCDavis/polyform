package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v3 "github.com/EliCDavis/vector/vector3"
)

func TestSuite(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"Normalize: nil => 0,0,0",
			nodetest.NewNode(vector3.Normalize{}),
			nodetest.AssertOutput("Normalized", v3.Zero[float64]()),
		),
		nodetest.NewTestCase(
			"Normalize: 0, 10, 0 => 0, 1, 0",
			nodetest.NewNode(vector3.Normalize{
				In: nodetest.NewPortValue(v3.New(0., 10., 0.)),
			}),
			nodetest.AssertOutput("Normalized", v3.Up[float64]()),
		),
		nodetest.NewTestCase(
			"Normalize Array: nil => nil",
			nodetest.NewNode(vector3.NormalizeArray{}),
			nodetest.AssertOutput[[]v3.Float64]("Local", nil),
		),
		nodetest.NewTestCase(
			"Normalize Array: 0, 10, 0 => 0, 1, 0",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{
					v3.New(0., 10., 0.),
				}),
			}),
			nodetest.AssertOutput("Local", []v3.Float64{
				v3.Up[float64](),
			}),
		),
	)
	suite.Run(t)
}
