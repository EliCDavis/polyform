package markdown_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeToDoc(writer markdown.Writer) {
	writer.Header1("Header 1")
	writer.Header2("Header 2")
	writer.Header3("Header 3")
	writer.Paragraph("A paragraph")
	writer.StartBulletList()

	writer.StartBullet()
	writer.Text("Hmm....")
	writer.EndBullet()
	writer.StartBullet()
	writer.Text("Another!")
	writer.EndBullet()

	writer.EndBulletList()
	writer.NewLine()
}

func TestMarkdown(t *testing.T) {
	// ARRANGE ================================================================
	buf := &bytes.Buffer{}
	writer := markdown.NewWriter(buf)

	// ACT ====================================================================
	writeToDoc(writer)

	// ASSERT =================================================================
	require.NoError(t, writer.Error())
	assert.Equal(t, `# Header 1

## Header 2

### Header 3

A paragraph

* Hmm....
* Another!


`, buf.String())
}

func TestHTML(t *testing.T) {
	// ARRANGE ================================================================
	buf := &bytes.Buffer{}
	writer := markdown.NewHtmlWriter(buf)

	// ACT ====================================================================
	writeToDoc(writer)

	// ASSERT =================================================================
	require.NoError(t, writer.Error())
	assert.Equal(t, `<h1>Header 1</h1>
<h2>Header 2</h2>
<h3>Header 3</h3>
<p>A paragraph</p>
<ul><li>Hmm....</li>
<li>Another!</li>
</ul>
<br>`, buf.String())
}
