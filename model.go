package mesh

import (
	"errors"
	"io"

	"github.com/EliCDavis/vector"
)

// Model is built with a collection of polygons
type Model struct {
	faces []Polygon
}

// NewModel builds a new model
func NewModel(faces []Polygon) (Model, error) {
	if faces == nil {
		return Model{}, errors.New("Can not have nil faces")
	}
	return Model{faces}, nil
}

// GetFaces returns all faces of the model
func (m Model) GetFaces() []Polygon {
	return m.faces
}

// Merge combines the faces of both the models into a new model
func (m Model) Merge(other Model) Model {
	return Model{append(m.faces, other.faces...)}
}

func (m Model) Translate(movement vector.Vector3) Model {
	newFaces := make([]Polygon, len(m.faces))
	for f := range m.faces {
		newFaces[f] = m.faces[f].Translate(movement)
	}
	return Model{newFaces}
}

func (m Model) Rotate(amount vector.Vector3, pivot vector.Vector3) Model {
	newFaces := make([]Polygon, len(m.faces))
	for f := range m.faces {
		newFaces[f] = m.faces[f].Rotate(amount, pivot)
	}
	return Model{newFaces}
}

// Save Writes a model to obj format
func (m Model) Save(w io.Writer) error {

	w.Write([]byte("mtllib master.mtl\n"))
	w.Write([]byte("usemtl wood\n"))

	offset := 1
	var err error
	for _, face := range m.faces {
		offset, err = face.Save(w, offset)
		if err != nil {
			return err
		}
	}

	return nil
}
