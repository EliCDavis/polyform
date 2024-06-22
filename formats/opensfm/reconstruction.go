package opensfm

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/sfm/opensfm"
	"github.com/EliCDavis/vector/vector3"
)

func ReconstructionToPointcloud(reconstruction opensfm.ReconstructionSchema) modeling.Mesh {
	positionData := make([]vector3.Float64, len(reconstruction.Points))
	colorData := make([]vector3.Float64, len(reconstruction.Points))

	i := 0
	for _, point := range reconstruction.Points {
		positionData[i] = vector3.New(point.Coordinates[0], point.Coordinates[1], point.Coordinates[2])
		colorData[i] = vector3.New(point.Color[0]/255, point.Color[1]/255, point.Color[2]/255)
		i++
	}

	return modeling.NewPointCloud(nil, map[string][]vector3.Vector[float64]{
		modeling.PositionAttribute: positionData,
		modeling.ColorAttribute:    colorData,
	}, nil, nil, nil)
}

// Loads the feature match point data into a Pointcloud mesh
func LoadReconstructiontData(filename string) (modeling.Mesh, error) {
	reconstructions, err := opensfm.LoadReconstruction(filename)
	mesh := modeling.EmptyPointcloud()
	if err != nil {
		return mesh, err
	}

	for _, rec := range reconstructions {
		mesh = mesh.Append(ReconstructionToPointcloud(rec))
	}

	return mesh, nil
}
