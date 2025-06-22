package ply

import (
	"errors"
	"fmt"
	"io"
)

type Element struct {
	Name       string     `json:"name"`
	Count      int64      `json:"count"`
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

func (e Element) DeterministicPointSize() bool {
	for _, p := range e.Properties {
		if _, ok := p.(ScalarProperty); !ok {
			return false
		}
	}
	return true
}

func (e Element) PointSize() (int, error) {
	size := 0
	for _, p := range e.Properties {
		scalar, ok := p.(ScalarProperty)

		if !ok {
			return 0, fmt.Errorf("property %q is not scalar, point size is variable", p.Name())
		}

		size += scalar.Size()
	}
	return size, nil
}

func (e Element) PropertyStart(propertyName string) (int, error) {
	size := 0
	for _, p := range e.Properties {
		if p.Name() == propertyName {
			return size, nil
		}

		scalar, ok := p.(ScalarProperty)

		if !ok {
			return -1, fmt.Errorf("property %q is not scalar, variable start for proerty %q", p.Name(), propertyName)
		}

		size += scalar.Size()
	}
	return -1, fmt.Errorf("element %q does not contain a property named %q", e.Name, propertyName)
}

func (e Element) Scan(in io.Reader, cb func(buf []byte) error) error {
	if !e.DeterministicPointSize() {
		return errors.New("unimplemented usecase where point size is variable")
	}

	size, err := e.PointSize()
	if err != nil {
		return err
	}

	buf := make([]byte, size)

	for range e.Count {
		_, err = io.ReadFull(in, buf)
		if err != nil {
			return err
		}
		err = cb(buf)
		if err != nil {
			return err
		}
	}

	return nil
}
