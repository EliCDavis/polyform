package colmap

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/sfm/colmap"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func ImageDataToPointCloud(images []colmap.Image) modeling.Mesh {

	positionData := make([]vector3.Float64, len(images))
	rotationData := make([]vector4.Float64, len(images))
	idData := make([]float64, len(images))
	cameraIdData := make([]float64, len(images))
	pointCountData := make([]float64, len(images))

	for i, p := range images {
		positionData[i] = p.Translation
		rotationData[i] = p.Rotation
		idData[i] = float64(p.Id)
		cameraIdData[i] = float64(p.CameraId)
		pointCountData[i] = float64(len(p.Points))
	}

	return modeling.NewPointCloud(
		map[string][]vector4.Float64{
			modeling.RotationAttribute: rotationData,
		},
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: positionData,
		},
		nil,
		map[string][]float64{
			"id":          idData,
			"camera id":   cameraIdData,
			"point count": pointCountData,
		},
	)
}

func LoadImageData(filename string) (modeling.Mesh, error) {
	points, err := colmap.LoadImagesBinary(filename)
	if err != nil {
		return modeling.EmptyPointcloud(), err
	}
	return ImageDataToPointCloud(points), nil
}
