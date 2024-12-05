package main

import (
	"fmt"
	"image/color"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
)

func plumbob() modeling.Mesh {
	return primitives.
		UVSphere(1, 2, 8).
		Transform(
			meshops.ScaleAttribute3DTransformer{
				Amount: vector3.New(1., 2., 1.),
			},
			meshops.UnweldTransformer{},
			meshops.FlatNormalsTransformer{},
			meshops.ScaleAttribute3DTransformer{
				Amount: vector3.Fill(10.0),
			},
		)
}

func main() {
	plumbobMesh := plumbob()

	backColor := color.RGBA{80, 80, 80, 0}

	backMetallicF := 0.8
	backMoughnessF := 1.0

	meshMat := &gltf.PolyformMaterial{
		Name: "plumbobBlack",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			MetallicFactor:  &backMetallicF,
			RoughnessFactor: &backMoughnessF,
			BaseColorFactor: backColor,
		},
	}

	closeV := vector3.New[float64](0, 0, 20)
	farV := vector3.New[float64](0, 0, -20)
	rightV := vector3.New[float64](20, 0, 0)
	leftV := vector3.New[float64](-20, 0, -0)

	scaleUniform15 := vector3.New[float64](1.5, 1.5, 1.5)
	scaleDistort := vector3.New[float64](0.5, 2.5, 0.5)
	scaleUniform05 := vector3.New[float64](0.5, 0.5, 0.5)
	rotQuat := quaternion.New(vector3.New[float64](1, 0, 0), -0.5)

	closeMesh := plumbobMesh.Translate(closeV).Scale(scaleUniform15)
	farMesh := plumbobMesh.Translate(farV).Scale(scaleDistort)
	rightMesh := plumbobMesh.Translate(rightV)
	leftMesh := plumbobMesh.Translate(leftV).Scale(scaleUniform05).Rotate(rotQuat)

	modelNaive := []gltf.PolyformModel{
		{Name: "scaled", Mesh: &closeMesh, Material: meshMat},
		{Name: "distorted", Mesh: &farMesh, Material: meshMat},
		{Name: "simple", Mesh: &rightMesh, Material: meshMat},
		{Name: "rotated", Mesh: &leftMesh, Material: meshMat},
	}

	sceneNaive := gltf.PolyformScene{Models: modelNaive}

	if err := gltf.SaveText("./plumbob_naive.gltf", sceneNaive); err != nil {
		panic(fmt.Errorf("failed to save GLTF: %w", err))
	}

	sceneOptimised := gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{Name: "scaled", Mesh: &plumbobMesh, Material: meshMat, Translation: &closeV, Scale: &scaleUniform15},
			{Name: "distorted", Mesh: &plumbobMesh, Material: meshMat, Translation: &farV, Scale: &scaleDistort},
			{Name: "simple", Mesh: &plumbobMesh, Material: meshMat, Translation: &rightV},
			{Name: "rotated", Mesh: &plumbobMesh, Material: meshMat, Translation: &leftV, Scale: &scaleUniform05, Quaternion: &rotQuat},
		},
	}

	if err := gltf.SaveText("./plumbob_optimised.gltf", sceneOptimised); err != nil {
		panic(fmt.Errorf("failed to save GLTF: %w", err))
	}
}
