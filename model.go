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

	if len(faces) == 0 {
		return Model{}, errors.New("Can't have a model with 0 faces")
	}

	var center vector.Vector3
	for _, f := range faces {
		center = center.Add(f.center)
	}

	return Model{faces}, nil
}

func (m Model) GetCenterOfBoundingBox() vector.Vector3 {
	totalX := 0.0
	totalY := 0.0
	totalZ := 0.0
	counted := 0.0

	for _, poly := range m.faces {
		for _, vertice := range poly.GetVertices() {
			totalX += vertice.X()
			totalY += vertice.Y()
			totalZ += vertice.Z()
			counted += 1.0
		}
	}

	return vector.NewVector3(totalX, totalY, totalZ).DivByConstant(counted)
}

// GetFaces returns all faces of the model
func (m Model) GetFaces() []Polygon {
	return m.faces
}

// Merge combines the faces of both the models into a new model
func (m Model) Merge(other Model) Model {
	model, _ := NewModel(append(m.faces, other.faces...))
	return model
}

func (m Model) Translate(movement vector.Vector3) Model {
	newFaces := make([]Polygon, len(m.faces))
	for f := range m.faces {
		newFaces[f] = m.faces[f].Translate(movement)
	}
	model, _ := NewModel(newFaces)
	return model
}

// Adjusts each vertices position relative to the origin
func (m Model) Scale(amount vector.Vector3, pivot vector.Vector3) Model {
	newFaces := make([]Polygon, len(m.faces))
	for f := range m.faces {
		newFaces[f] = m.faces[f].Scale(amount, pivot)
	}
	model, _ := NewModel(newFaces)
	return model
}

func (m Model) Rotate(amount vector.Vector3, pivot vector.Vector3) Model {
	newFaces := make([]Polygon, len(m.faces))
	for f := range m.faces {
		newFaces[f] = m.faces[f].Rotate(amount, pivot)
	}
	model, _ := NewModel(newFaces)
	return model
}

// Save Writes a model to obj format
func (m Model) Save(w io.Writer) error {

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
