package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v4 "github.com/EliCDavis/vector/vector4"
)

func TestNormalize(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0,0,0,0",
			nodetest.NewNode(vector4.Normalize{}),
			nodetest.AssertOutput("Normalized", v4.Zero[float64]()),
			nodetest.AssertLiftedInput("In", vec4Type),
		),
		nodetest.NewTestCase(
			"a single vector",
			nodetest.NewNode(vector4.Normalize{
				In: nodetest.NewPortValue(v4.New(0., 10., 0., 0.)),
			}),
			nodetest.AssertOutput("Normalized", v4.New(0., 1., 0., 0.)),
		),
		nodetest.NewTestCase(
			"every vector in an array",
			nodetest.NewNode(vector4.Normalize{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
					v4.New(4., 0., 0., 0.),
				}),
			}),
			nodetest.AssertOutput("Normalized", []v4.Float64{
				v4.New(0., 1., 0., 0.),
				v4.New(1., 0., 0., 0.),
			}),
			nodetest.AssertOutputType("Normalized", vec4ArrayType),
		),
		nodetest.NewTestCase(
			"a zero vector inside an array stays zero rather than NaN",
			nodetest.NewNode(vector4.Normalize{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
					v4.Zero[float64](),
				}),
			}),
			nodetest.AssertOutput("Normalized", []v4.Float64{
				v4.New(0., 1., 0., 0.),
				v4.Zero[float64](),
			}),
		),
	).Run(t)
}

func TestNormalizeArray(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"descriptions",
			nodetest.NewNode(vector4.NormalizeArray{}),
			nodetest.AssertNodeOutputPortDescription{
				Port:        "Global",
				Description: "Scales each vector by the inverse of the magnitude of the longest vector",
			},
		),
		nodetest.NewTestCase(
			"nil => nil",
			nodetest.NewNode(vector4.NormalizeArray{}),
			nodetest.AssertOutput[[]v4.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"all zero => error",
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
			"scale by longest magnitude",
			nodetest.NewNode(vector4.NormalizeArray{
				In: nodetest.NewPortValue([]v4.Float64{
					v4.New(0., 10., 0., 0.),
					v4.New(0., 5., 0., 0.),
				}),
			}),
			nodetest.AssertOutput("Global", []v4.Float64{
				v4.New(0., 10., 0., 0.).DivByConstant(10.),
				v4.New(0., 5., 0., 0.).DivByConstant(10.),
			}),
		),
	).Run(t)
}
