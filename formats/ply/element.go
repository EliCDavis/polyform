package ply

import (
	"fmt"
	"io"
)

type Element struct {
	Name       string     `json:"name"`
	Count      int        `json:"count"`
	Properties []Property `json:"properties"`
}

func (e Element) Write(out io.Writer) error {
	fmt.Fprintf(out, "element %s %d\n", e.Name, e.Count)
	for _, prop := range e.Properties {
		err := prop.Write(out)
		if err != nil {
			return err
		}
	}
	return nil
}
