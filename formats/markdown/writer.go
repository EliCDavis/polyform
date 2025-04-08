package markdown

import (
	"io"

	"github.com/EliCDavis/polyform/formats/txt"
)

type Writer interface {
	Header1(text string) (int, error)
	Header2(text string) (int, error)
	Header3(text string) (int, error)
	Header2WithId(text string, id string) (int, error)
	Header3WithId(text string, id string) (int, error)
	Link(text, link string) (int, error)
	Paragraph(text string) (int, error)
	NewLine() (int, error)

	StartBulletList() (int, error)
	EndBulletList() (int, error)

	StartBullet() (int, error)
	EndBullet() (int, error)

	StartBold() (int, error)
	EndBold() (int, error)

	StartItalics() (int, error)
	EndItalics() (int, error)

	Text(text string) (int, error)
	Error() error
}

func NewWriter(out io.Writer) Writer {
	return &markdownWriter{
		writer: txt.NewWriter(out),
	}
}

func NewHtmlWriter(out io.Writer) Writer {
	return &htmlWriter{
		writer: txt.NewWriter(out),
	}
}
