package pipeline

import "github.com/EliCDavis/polyform/modeling"

type Schedule struct {
	waves [][]Command
}

func (s Schedule) Run(m modeling.Mesh) modeling.Mesh {

	finalMesh := m
	for _, wave := range s.waves {

		for _, c := range wave {
			c.operator(newView(finalMesh, c.readPermissions, c.writePermissions))
		}

	}
	return finalMesh
}
