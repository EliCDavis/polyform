package rendering

import (
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

var inf float64 = 10000//math.Inf(1)

func colorFromRay(tr TemporalRay, world Hittable, background sample.Vec3ToVec3, depth int) vector3.Float64 {
	if depth < 0 {
		return vector3.Zero[float64]()
	}

	ray := tr.Ray()
	hitRecord := NewHitRecord()
	if !world.Hit(&tr, 0.001, inf, hitRecord) {
		return background(tr.direction)
	}

	scattered := geometry.NewRay(vector3.Zero[float64](), vector3.Zero[float64]())
	attenuation := vector3.Zero[float64]()
	emitted := hitRecord.Material.Emitted(hitRecord.UV, hitRecord.Point)

	if !hitRecord.Material.Scatter(ray, hitRecord, &attenuation, &scattered) {
		return emitted
	}

	return colorFromRay(
		NewTemporalRay(scattered.Origin(), scattered.Direction(), tr.time),
		world,
		background,
		depth-1,
	).MultByVector(attenuation).Add(emitted)
}

func Render(
	maxRayBounce, samplesPerPixel, imageWidth int,
	aspectRatio float64,
	hittables []Hittable,
	camera Camera,
	imgPath string,
	completion chan<- float64,
) error {
	f, err := os.Create(imgPath)
	if err != nil {
		return err
	}

	defer f.Close()

	imageHeight := int(float64(imageWidth) / aspectRatio)
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	var world HitList = hittables
	// bvh := NewBVH(hittables, camera.timeStart, camera.timeEnd)

	totalPixels := float64(imageHeight * imageWidth)

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {

			col := vector3.Zero[float64]()

			for s := 0; s < samplesPerPixel; s++ {
				u := (float64(x) + rand.Float64()) / float64(imageWidth-1)
				v := (float64(y) + rand.Float64()) / float64(imageHeight-1)
				col = col.Add(colorFromRay(camera.GetRay(u, v), &world, camera.background, maxRayBounce))
			}

			col = col.
				DivByConstant(float64(samplesPerPixel)).
				Sqrt().
				Scale(255).
				Clamp(0, 255)

			img.Set(x, imageHeight-y, &color.RGBA{
				uint8(col.X()),
				uint8(col.Y()),
				uint8(col.Z()),
				255,
			})

			if completion != nil {
				completion <- float64((y*imageWidth)+x) / totalPixels
			}
		}
	}

	err = png.Encode(f, img)

	if completion != nil {
		close(completion)
	}

	return err
}
