package csg

import (
	"fmt"
	"sync"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Solid struct {
	mesh  modeling.Mesh
	faces []face

	once      sync.Once
	cornerIDs [][3]int
	weld      *welder
	tolerance float64
	err       error
}

func NewSolid(m modeling.Mesh) (*Solid, error) {
	faces, err := facesOf(m)
	if err != nil {
		return nil, err
	}
	s := &Solid{mesh: m, faces: faces}
	if err := s.prepare(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Solid) Mesh() modeling.Mesh { return s.mesh }

func (s *Solid) Union(other *Solid) (*Solid, error) {
	return s.csg(other, union)
}

func (s *Solid) Intersect(other *Solid) (*Solid, error) {
	return s.csg(other, intersection)
}

func (s *Solid) Subtract(other *Solid) (*Solid, error) {
	return s.csg(other, difference)
}

func (s *Solid) csg(other *Solid, op operation) (*Solid, error) {
	if err := s.prepare(); err != nil {
		return nil, fmt.Errorf("first %w", err)
	}
	if err := other.prepare(); err != nil {
		return nil, fmt.Errorf("second %w", err)
	}
	mesh, faces, err := combine(s, other, op)
	if err != nil {
		return nil, err
	}
	return &Solid{mesh: mesh, faces: faces}, nil
}

func (s *Solid) prepare() error {
	s.once.Do(func() {
		s.tolerance = toleranceFor(s.faces, nil)
		s.weld = newWelder(s.tolerance)
		s.cornerIDs = weldFaces(s.weld, s.faces)
		s.err = validate(s.mesh, s.faces, s.cornerIDs)
	})
	return s.err
}

type idSpace struct {
	known *welder
	added *welder
	base  int
}

func newIDSpace(known *welder, tolerance float64) *idSpace {
	return &idSpace{known: known, added: newWelder(tolerance), base: len(known.Points())}
}

func (s *idSpace) index(point vector3.Float64) int {
	if i, known := s.known.Find(point); known {
		return i
	}
	return s.base + s.added.Index(point)
}
