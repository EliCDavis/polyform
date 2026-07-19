package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	v2 "github.com/EliCDavis/vector/vector2"
)

func TestNormalizeArray(t *testing.T) {
	suite := nodetest.NewSuite(
		nodetest.NewTestCase(
			"descriptions",
			nodetest.NewNode(vector2.NormalizeArray{}),
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
			nodetest.NewNode(vector2.NormalizeArray{}),
			nodetest.AssertOutput[[]v2.Float64]("Local", nil),
		),
		nodetest.NewTestCase(
			"Local: empty => empty",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{}),
			}),
			nodetest.AssertOutput("Local", []v2.Float64{}),
		),
		nodetest.NewTestCase(
			"Local: single vector",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{
					v2.New(0., 10.),
				}),
			}),
			nodetest.AssertOutput("Local", []v2.Float64{
				v2.New(0., 1.),
			}),
		),
		nodetest.NewTestCase(
			"Local: multiple vectors",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{
					v2.New(0., 10.),
					v2.New(4., 0.),
					v2.New(3., 4.),
				}),
			}),
			nodetest.AssertOutput("Local", []v2.Float64{
				v2.New(0., 1.),
				v2.New(1., 0.),
				v2.New(3., 4.).Normalized(),
			}),
		),
		nodetest.NewTestCase(
			"Global: nil => nil",
			nodetest.NewNode(vector2.NormalizeArray{}),
			nodetest.AssertOutput[[]v2.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: empty => nil",
			nodetest.NewNode(vector2.NormalizeArray{
				In: nodetest.NewPortValue([]v2.Float64{}),
			}),
			nodetest.AssertOutput[[]v2.Float64]("Global", nil),
		),
		nodetest.NewTestCase(
			"Global: all zero => error",
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
			"Global: scale by longest magnitude",
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
	)
	suite.Run(t)
}
