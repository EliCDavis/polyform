package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func randomBallsScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
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
					sphere_material = materials.NewLambertian(textures.NewSolidColorTexture(albedo))
					dir := vector3.RandNormal().Scale(rand.Float64())
					world = append(
						world,
						rendering.NewAnimatedSphere(
							0.2,
							sphere_material,
							func(t float64) vector3.Float64 {
								return center.Add(dir.Scale(t))
							},
						))
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

	material2 := materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.4, 0.2, 0.1)))
	world = append(world, rendering.NewSphere(vector3.New(-4., 1., 0.), 1.0, material2))

	material3 := materials.NewFuzzyMetal(vector3.New(0.7, 0.6, 0.5), 0.0)
	world = append(world, rendering.NewSphere(vector3.New(4., 1., 0.), 1.0, material3))

	return world
}

func SimpleLightScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))
	world = append(world, rendering.NewSphere(vector3.New(0., 2., 0.), 2, ground_material))

	diffuseLight := materials.NewDiffuseLightWithColor(vector3.New(4., 4., 4.))
	world = append(world, rendering.NewXYRectangle(vector2.New(3., 1.), vector2.New(4., 3.), -2., diffuseLight))
	return world
}

func main() {
	// origin := vector3.New(13., 2., 3.)
	origin := vector3.New(26., 3., 6.)
	lookat := vector3.New(0., 2., 0.)
	aperatre := 0.1

	aspectRatio := 3. / 2.

	camera := rendering.NewCamera(20., aspectRatio, aperatre, 10., origin, lookat, vector3.Up[float64](), 0, 1)

	t1 := time.Now()

	completion := make(chan float64, 1)

	go func() {
		err := rendering.Render(50, 400, 800, aspectRatio, SimpleLightScene(), camera, "example2.png", completion)
		if err != nil {
			log.Print(err)
			panic(err)
		}
	}()

	lastProgress := -1.
	for progress := range completion {
		if progress-lastProgress > .001 {
			lastProgress = progress
			log.Printf("Image Progress: %.2f%%\n", progress*100.)
		}
	}

	fmt.Println(time.Since(t1))
}
