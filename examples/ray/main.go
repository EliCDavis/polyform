package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/vector/vector3"
)

func random_scene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	ground_material := materials.NewLambertian(vector3.New(0.5, 0.5, 0.5))
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			choose_mat := rand.Float64()
			center := vector3.New(
				float64(a)+(0.9*rand.Float64()),
				0.2,
				float64(b)+(0.9*rand.Float64()),
			)

			if center.Sub(vector3.New(4., 0.2, 0.)).Length() > 0.9 {
				var sphere_material rendering.Material = nil

				if choose_mat < 0.8 {
					// diffuse
					albedo := vector3.Rand().MultByVector(vector3.Rand())
					sphere_material = materials.NewLambertian(albedo)
					world = append(world, rendering.NewSphere(center, 0.2, sphere_material))
				} else if choose_mat < 0.95 {
					// metal
					albedo := vector3.RandRange(0.4, 1.)
					fuzz := rand.Float64() * 0.5
					sphere_material = materials.NewFuzzyMetal(albedo, fuzz)
					world = append(world, rendering.NewSphere(center, 0.2, sphere_material))

				} else {
					// glass
					sphere_material := materials.NewDielectric(1.5)
					world = append(world, rendering.NewSphere(center, 0.2, sphere_material))
				}
			}
		}
	}

	material1 := materials.NewDielectric(1.5)
	world = append(world, rendering.NewSphere(vector3.New(0., 1., 0.), 1.0, material1))

	material2 := materials.NewLambertian(vector3.New(0.4, 0.2, 0.1))
	world = append(world, rendering.NewSphere(vector3.New(-4., 1., 0.), 1.0, material2))

	material3 := materials.NewFuzzyMetal(vector3.New(0.7, 0.6, 0.5), 0.0)
	world = append(world, rendering.NewSphere(vector3.New(4., 1., 0.), 1.0, material3))

	return world
}

func main() {
	origin := vector3.New(13., 2., 3.)
	lookat := vector3.New(0., 0., 0.)
	aperatre := 0.1

	aspectRatio := 3. / 2.

	// camera := rendering.NewCamera(20., aspectRatio, 2.0, lookat.Sub(origin).Length(), origin, lookat, vector3.Up[float64]())

	// world := []rendering.Hittable{
	// 	rendering.NewSphere(vector3.New(0., -100.5, -1.), 100, materials.NewLambertian(vector3.New(0., 0.8, 0.10))),

	// 	// NewSphere(vector3.New(0., 0., -1.), 0.5, NewDielectric(1.5)),
	// 	rendering.NewSphere(vector3.New(0., 0., -1.), 0.5, materials.NewLambertian(vector3.New(0.8, 0.3, 0.3))),
	// 	rendering.NewSphere(vector3.New(1., 0., -1.), 0.5, materials.NewFuzzyMetal(vector3.New(0.8, 0.6, 0.2), .1)),
	// 	rendering.NewSphere(vector3.New(-1., 0., -1.), 0.5, materials.NewDielectric(1.5)),
	// 	rendering.NewSphere(vector3.New(-1., 0., -1.), -0.4, materials.NewDielectric(1.5)),
	// }

	camera := rendering.NewCamera(20., aspectRatio, aperatre, 10., origin, lookat, vector3.Up[float64](), 0, 1)

	t1 := time.Now()

	completion := make(chan float64, 1)

	go func() {
		err := rendering.Render(50, 50, 200, aspectRatio, random_scene(), camera, "example2.png", completion)
		if err != nil {
			log.Print(err)
			panic(err)
		}
	}()

	lastProgress := -1.
	for progress := range completion {
		if progress-lastProgress > .01 {
			lastProgress = progress
			log.Printf("Image Progress: %.2f\n", progress*100.)
		}
	}

	fmt.Println(time.Since(t1))
}
