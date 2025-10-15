package main

import (
	"log"
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
)

func load(gltfPath string) error {
	doc, buffers, err := gltf.LoadFile(gltfPath, nil)
	if err != nil {
		return err
	}

	log.Printf("Discovered Buffers: %d\n", len(buffers))

	models, err := gltf.DecodeModels(doc, buffers, nil)
	if err != nil {
		return err
	}

	for _, model := range models {
		log.Printf("%s:", model.Name)
		log.Printf("\tChildren: %d", len(model.Children))
		log.Printf("\tPosition: %s", model.TRS.Position().Format("%.2f, %.2f, %.2f"))
		log.Printf("\tRotation: %s", model.TRS.Rotation().ToEulerAngles().Scale(180./math.Pi).Format("%.2f, %.2f, %.2f"))
		log.Printf("\tScale: %s", model.TRS.Scale().Format("%.2f, %.2f, %.2f"))
	}

	return gltf.SaveBinary("rewritten.glb", gltf.PolyformScene{
		Models: models,
	}, nil)
}

func main() {
	err := load("C:/Users/elida/Downloads/IridescentDishWithOlives (2).glb")
	if err != nil {
		panic(err)
	}
}
