package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestNormalize(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"nothing wired => 0,0",
			nodetest.NewNode(vector2.Normalize{}),
			nodetest.AssertOutput("Normalized", v2.Zero[float64]()),
			nodetest.AssertLiftedInput("In", vec2Type),
		),
		nodetest.NewTestCase(
			"a single vector",
			nodetest.NewNode(vector2.Normalize{
				In: nodetest.NewPortValue(v2.New(0., 10.)),
			}),
			nodetest.AssertOutput("Normalized", v2.New(0., 1.)),
		),
		nodetest.NewTestCase(
			"every vector in an array",
			nodetest.NewNode(vector2.Normalize{
				In: nodetest.NewPortValue([]v2.Float64{
					v2.New(0., 10.),
					v2.New(4., 0.),
					v2.New(3., 4.),
				}),
			}),
			nodetest.AssertOutput("Normalized", []v2.Float64{
				v2.New(0., 1.),
				v2.New(1., 0.),
				v2.New(3., 4.).Normalized(),
			}),
			nodetest.AssertOutputType("Normalized", vec2ArrayType),
		),
		nodetest.NewTestCase(
			"a zero vector inside an array stays zero rather than NaN",
			nodetest.NewNode(vector2.Normalize{
				In: nodetest.NewPortValue([]v2.Float64{v2.New(0., 10.), v2.Zero[float64]()}),
			}),
			nodetest.AssertOutput("Normalized", []v2.Float64{v2.New(0., 1.), v2.Zero[float64]()}),
		),
		nodetest.NewTestCase(
			"an empty array stays empty",
			nodetest.NewNode(vector2.Normalize{
				In: nodetest.NewPortValue([]v2.Float64{}),
			}),
			nodetest.AssertOutput("Normalized", []v2.Float64{}),
		),
	).Run(t)
}

func TestNormalizeArray(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"descriptions",
			nodetest.NewNode(vector2.NormalizeArray{}),
			nodetest.AssertNodeOutputPortDescription{
				Port:        "Global",
				Description: "Scales each vector by the inverse of the magnitude of the longest vector",
			},
		),
		nodetest.NewTestCase(
			"nil => nil",
			nodetest.NewNode(vector2.NormalizeArray{}),
			nodetest.AssertOutput[[]v2.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"empty => nil",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{}),
			}),
			nodetest.AssertOutput[[]v2.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"all zero => error",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{
					v2.Zero[float64](),
					v2.Zero[float64](),
				}),
			}),
			nodetest.AssertOutputPortValue[[]v2.Float64]{
				Port: "Global",
				Value: []v2.Float64{
					v2.Zero[float64](),
					v2.Zero[float64](),
				},
				ExecutionReport: &nodes.ExecutionReport{
					Errors: []string{"all vector data has a magnitude of 0"},
				},
			},
		),
		nodetest.NewTestCase(
			"scale by longest magnitude",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{
					v2.New(0., 10.),
					v2.New(0., 5.),
					v2.New(3., 4.),
				}),
			}),
			nodetest.AssertOutput("Global", []v2.Float64{
				v2.New(0., 10.).DivByConstant(10.),
				v2.New(0., 5.).DivByConstant(10.),
				v2.New(3., 4.).DivByConstant(10.),
			}),
		),
	).Run(t)
}
