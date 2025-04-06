package markdown_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdown(t *testing.T) {
	// ARRANGE ================================================================
	buf := &bytes.Buffer{}
	writer := markdown.NewWriter(buf)

	// ACT ====================================================================
	writer.Header1("Header 1")
	writer.Header2("Header 2")
	writer.Header3("Header 3")
	writer.Paragraph("A paragraph")
	writer.Bullet("Hmm....")
	writer.NewLine()

	// ASSERT =================================================================
	require.NoError(t, writer.Error())
	assert.Equal(t, `# Header 1

## Header 2

### Header 3

A paragraph

* Hmm....

`, buf.String())
}
