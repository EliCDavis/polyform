package ply

import (
	"fmt"
	"io"
)

type Element struct {
	name       string
	count      int
	properties []Property
}

func (e Element) Write(out io.Writer) error {
	fmt.Fprintf(out, "element %s %d\n", e.name, e.count)
	for _, prop := range e.properties {
		err := prop.Write(out)
		if err != nil {
			return err
		}
	}
	return nil
}
