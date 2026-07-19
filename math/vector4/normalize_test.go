package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestNormalizeArray(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"descriptions",
			nodetest.NewNode(vector4.NormalizeArray{}),
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
			nodetest.NewNode(vector4.NormalizeArray{}),
			nodetest.AssertOutput[[]v4.Float64]("Local", nil),
		),
		nodetest.NewTestCase(
			"Local: empty => empty",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{}),
			}),
			nodetest.AssertOutput("Local", []v4.Float64{}),
		),
		nodetest.NewTestCase(
			"Local: single vector",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
				}),
			}),
			nodetest.AssertOutput("Local", []v4.Float64{
				v4.New(0., 1., 0., 0.),
			}),
		),
		nodetest.NewTestCase(
			"Local: multiple vectors",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
					v4.New(4., 0., 0., 0.),
					v4.New(0., 0., 3., 4.),
				}),
			}),
			nodetest.AssertOutput("Local", []v4.Float64{
				v4.New(0., 1., 0., 0.),
				v4.New(1., 0., 0., 0.),
				v4.New(0., 0., 3., 4.).Normalized(),
			}),
		),
		nodetest.NewTestCase(
			"Global: nil => nil",
			nodetest.NewNode(vector4.NormalizeArray{}),
			nodetest.AssertOutput[[]v4.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: empty => nil",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{}),
			}),
			nodetest.AssertOutput[[]v4.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: all zero => error",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.Zero[float64](),
					v4.Zero[float64](),
				}),
			}),
			nodetest.AssertOutputPortValue[[]v4.Float64]{
				Port: "Global",
				Value: []v4.Float64{
					v4.Zero[float64](),
					v4.Zero[float64](),
				},
				ExecutionReport: &nodes.ExecutionReport{
					Errors: []string{"all vector data has a magnitude of 0"},
				},
			},
		),
		nodetest.NewTestCase(
			"Global: scale by longest magnitude",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
					v4.New(0., 5., 0., 0.),
					v4.New(0., 0., 3., 4.),
				}),
			}),
			nodetest.AssertOutput("Global", []v4.Float64{
				v4.New(0., 10., 0., 0.).DivByConstant(10.),
				v4.New(0., 5., 0., 0.).DivByConstant(10.),
				v4.New(0., 0., 3., 4.).DivByConstant(10.),
			}),
		),
	)
	suite.Run(t)
}
