package generator

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

type TestDocNode = nodes.Struct[TestDocNodeData]

type TestDocNodeData struct {
	A nodes.Output[int]
	B nodes.Output[int] `description:"B has a description"`
}

func (TestDocNodeData) Out(out *nodes.StructOutput[float64]) {
	out.Set(1.)
}

func TestDocumentation_SingleMarkdown(t *testing.T) {
	// ARRANGE ================================================================
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[TestDocNode](factory)

	doc := DocumentationWriter{
		Title:       "Test",
		Description: "A Description",
		Version:     "Yee",
		NodeTypes:   factory,
	}

	// ACT ====================================================================
	buf := bytes.Buffer{}
	err := doc.WriteSingleMarkdown(&buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `# Test

*Version: Yee*

A Description

## Table Of Contents

*1 nodes across 1 packages*

* [generator](#0)
    * [Test Doc](#0-0)

## <a id="0">generator</a>

### <a id="0-0">Test Doc</a>

Inputs:

* **A**: int
* **B**: int - B has a description

Outputs:

* **Out**: float64

`, buf.String())

}
