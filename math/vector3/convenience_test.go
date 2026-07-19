package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes"
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
	)
	suite.Run(t)
}

func TestNormalizeArray(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"descriptions",
			nodetest.NewNode(vector3.NormalizeArray{}),
			nodetest.AssertNodeOutputPortDescription{
				Port:        "Local",
				Description: "Normalizes each component of the array",
			},
			nodetest.AssertNodeOutputPortDescription{
				Port:        "Global",
				Description: "Scales each vector by the inverse of the magnitude of the longest vector",
			},
		),
		nodetest.NewTestCase(
			"Local: nil => nil",
			nodetest.NewNode(vector3.NormalizeArray{}),
			nodetest.AssertOutput[[]v3.Float64]("Local", nil),
		),
		nodetest.NewTestCase(
			"Local: empty => empty",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{}),
			}),
			nodetest.AssertOutput("Local", []v3.Float64{}),
		),
		nodetest.NewTestCase(
			"Local: single vector",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{
					v3.New(0., 10., 0.),
				}),
			}),
			nodetest.AssertOutput("Local", []v3.Float64{
				v3.Up[float64](),
			}),
		),
		nodetest.NewTestCase(
			"Local: multiple vectors",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{
					v3.New(0., 10., 0.),
					v3.New(4., 0., 0.),
					v3.New(3., 0., 4.),
				}),
			}),
			nodetest.AssertOutput("Local", []v3.Float64{
				v3.Up[float64](),
				v3.New(1., 0., 0.),
				v3.New(3., 0., 4.).Normalized(),
			}),
		),
		nodetest.NewTestCase(
			"Global: nil => nil",
			nodetest.NewNode(vector3.NormalizeArray{}),
			nodetest.AssertOutput[[]v3.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: empty => nil",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{}),
			}),
			nodetest.AssertOutput[[]v3.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: all zero => error",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{
					v3.Zero[float64](),
					v3.Zero[float64](),
				}),
			}),
			nodetest.AssertOutputPortValue[[]v3.Float64]{
				Port: "Global",
				Value: []v3.Float64{
					v3.Zero[float64](),
					v3.Zero[float64](),
				},
				ExecutionReport: &nodes.ExecutionReport{
					Errors: []string{"all vector data has a magnitude of 0"},
				},
			},
		),
		nodetest.NewTestCase(
			"Global: scale by longest magnitude",
			nodetest.NewNode(vector3.NormalizeArray{
				In: nodetest.NewPortValue([]v3.Float64{
					v3.New(0., 10., 0.),
					v3.New(0., 5., 0.),
					v3.New(3., 0., 4.),
				}),
			}),
			nodetest.AssertOutput("Global", []v3.Float64{
				v3.New(0., 10., 0.).DivByConstant(10.),
				v3.New(0., 5., 0.).DivByConstant(10.),
				v3.New(3., 0., 4.).DivByConstant(10.),
			}),
		),
	)
	suite.Run(t)
}
