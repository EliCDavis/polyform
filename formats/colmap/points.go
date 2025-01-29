package colmap

import (
	"io"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/sfm/colmap"
	"github.com/EliCDavis/vector/vector3"
)

func PointDataToPointCloud(points []colmap.Point3D) modeling.Mesh {
	positionData := make([]vector3.Float64, len(points))
	colorData := make([]vector3.Float64, len(points))
	errorData := make([]float64, len(points))
	idData := make([]float64, len(points))
	lenTrackData := make([]float64, len(points))

	for i, p := range points {
		positionData[i] = p.Position
		colorData[i] = vector3.FromColor(p.Color)
		errorData[i] = p.Error
		idData[i] = float64(p.ID)
		lenTrackData[i] = float64(len(p.Tracks))
	}

	return modeling.NewPointCloud(
		nil,
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: positionData,
			modeling.ColorAttribute:    colorData,
		},
		nil,
		map[string][]float64{
			"error":       errorData,
			"id":          idData,
			"track count": lenTrackData,
		},
		nil,
	)
}

func ReadSparsePointData(in io.Reader) (modeling.Mesh, error) {
	points, err := colmap.ReadPoints3DBinary(in)
	if err != nil {
		return modeling.EmptyPointcloud(), err
	}
	return PointDataToPointCloud(points), nil
}

// Loads the feature match point data into a Pointcloud mesh
func LoadSparsePointData(filename string) (modeling.Mesh, error) {
	points, err := colmap.LoadPoints3DBinary(filename)
	if err != nil {
		return modeling.EmptyPointcloud(), err
	}
	return PointDataToPointCloud(points), nil
}
