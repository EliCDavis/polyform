package nodetest

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
)

type Suite struct {
	cases []TestCase
}

func (s Suite) Run(t *testing.T) {
	for _, tc := range s.cases {
		tc.Run(t)
	}
}

func NewSuite(cases ...TestCase) Suite {
	return Suite{
		cases: cases,
	}
}

type TestCase struct {
	name       string
	node       nodes.Node
	assertions []Assertion
}

func (tc TestCase) Run(t *testing.T) {
	t.Run(tc.name, func(t *testing.T) {
		for _, assertion := range tc.assertions {
			assertion.Assert(t, tc.node)
		}
	})
}

func NewTestCase(name string, node nodes.Node, assertions ...Assertion) TestCase {
	return TestCase{
		name:       name,
		node:       node,
		assertions: assertions,
	}
}
