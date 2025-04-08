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

	writer.Header2WithId("Header 2 With ID", "id-2")
	writer.Header3WithId("Header 3 With ID", "id-3")

	writer.Paragraph("A paragraph")
	{
		writer.StartBulletList()

		{
			writer.StartBullet()
			writer.Text("Hmm....")
			writer.EndBullet()
		}

		{
			writer.StartBullet()
			writer.Link("ima link!!!", "id-2")
			writer.EndBullet()
		}

		{
			writer.StartBulletList()
			writer.StartBullet()
			writer.Text("Sub bullet!")
			writer.EndBullet()
			writer.EndBulletList()
		}

		writer.EndBulletList()
	}
	writer.NewLine()

	writer.StartBold()
	writer.Text("Im Bold!!")
	writer.EndBold()
	writer.Text(" and ")

	writer.StartItalics()
	writer.Text("I'm italic!")
	writer.EndItalics()
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

## <a id="id-2">Header 2 With ID</a>

### <a id="id-3">Header 3 With ID</a>

A paragraph

* Hmm....
* [ima link!!!](#id-2)
    * Sub bullet!


**Im Bold!!** and *I'm italic!*`, buf.String())
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
<h2 id="id-2">Header 2 With ID</h2>
<h3 id="id-3">Header 3 With ID</h3>
<p>A paragraph</p>
<ul><li>Hmm....</li>
<li><a href="#id-2">ima link!!!</a>
</li>
<ul><li>Sub bullet!</li>
</ul>
</ul>
<br><b>Im Bold!!</b>
 and <i>I'm italic!</i>
`, buf.String())
}
