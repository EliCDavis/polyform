package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sort"
	"time"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	Width   = 1920
	Height  = 1080
	OutName = "render.png"
)

type Camera struct {
	FocalLength float64 // Focal length in pixels
	Width       int     // Image width in pixels
	Height      int     // Image height in pixels
}

// ProjectPoint projects a 3D point (x, y, z) to 2D image coordinates (u, v)
func (c Camera) ProjectPoint(point vector3.Float64) (vector2.Float64, error) {
	if point.Z() <= 0 {
		return vector2.Zero[float64](), fmt.Errorf("point is behind the camera or on the image plane (z = %v)", point.Z())
	}

	// Pinhole camera projection
	u := c.FocalLength*point.X()/point.Z() + float64(c.Width)/2
	v := c.FocalLength*point.Y()/point.Z() + float64(c.Height)/2

	return vector2.New(u, v), nil
}

// ComputeVisibility calculates visibility for each point using solid angle method
func ComputeVisibility(depthBuffer texturing.Texture[float64], positionBuffer texturing.Texture[vector3.Float64], colorBuf texturing.Texture[color.Color], visibilityThreshold float64) (texturing.Texture[float64], texturing.Texture[color.Color]) {
	// Number of angular sectors to divide space into
	numSectors := 8
	outDepth := depthBuffer.Copy()
	outColor := colorBuf.Copy()

	// Search radius (in pixels)
	searchRadius := 3

	// Process each pixel in the depth buffer
	for y := searchRadius; y < depthBuffer.Height()-searchRadius; y++ {
		for x := searchRadius; x < depthBuffer.Width()-searchRadius; x++ {
			depth := depthBuffer.Get(x, y)
			// Skip background pixels (no point)
			if depth <= 0 {
				continue
			}
			p0 := positionBuffer.Get(x, y)

			// Vector from point to camera
			rayToCamera := p0.Scale(-1).Normalized()

			// Calculate total solid angle
			totalSolidAngle := 0.0

			// For each sector
			sectorAngle := 2 * math.Pi / float64(numSectors)

			for s := range numSectors {
				// Define sector direction in camera plane
				// This is a simplified approximation for demonstration
				sectorStart := float64(s) * sectorAngle
				sectorEnd := float64(s+1) * sectorAngle

				// Find the largest angle in this sector
				largestAngle := -1.0 // cosine of angle, -1 means 180 degrees (completely open)

				// Check all neighbors within search radius
				for ny := y - searchRadius; ny <= y+searchRadius; ny++ {
					for nx := x - searchRadius; nx <= x+searchRadius; nx++ {
						if nx == x && ny == y {
							continue // Skip the center pixel
						}

						neighborDepth := depthBuffer.Get(nx, ny)

						if neighborDepth <= 0 {
							continue // Skip background pixels
						}

						// Get 3D position of neighbor
						pi := positionBuffer.Get(nx, ny)

						// Vector from current point to neighbor
						vectorToNeighbor := pi.Sub(p0).Normalized()

						// Calculate dot product
						dotProduct := rayToCamera.Dot(vectorToNeighbor)

						// Check if it's in this sector
						neighborAngle := math.Atan2(
							float64(y-ny),
							float64(x-nx),
						)
						if neighborAngle < 0 {
							neighborAngle += 2 * math.Pi
						}

						if neighborAngle >= sectorStart && neighborAngle < sectorEnd {
							// Update the horizon angle if this point blocks more visibility
							if dotProduct > largestAngle {
								largestAngle = dotProduct
							}
						}
					}
				}

				// Calculate solid angle for this sector
				// If we didn't find any horizon points, the sector is completely open
				sectorSolidAngle := 0.0

				if largestAngle < 0 {
					// Each sector represents 1/numSectors of the full 4π steradians
					sectorSolidAngle = 4 * math.Pi / float64(numSectors)
				} else {
					// For a sector with a horizon point
					// Calculate the solid angle of the visible cone within this sector
					horizonAngle := math.Acos(largestAngle)

					// Solid angle of a cone is 2π(1-cos(θ))
					coneSolidAngle := 2 * math.Pi * (1 - math.Cos(horizonAngle))

					// Scale by the sector's angular proportion
					sectorSolidAngle = coneSolidAngle * (sectorAngle / (2 * math.Pi))
				}

				totalSolidAngle += sectorSolidAngle
			}

			// Mark point as visible if solid angle is above threshold
			// point.WorldPoint.IsVisible = totalSolidAngle >= visibilityThreshold
			if totalSolidAngle < visibilityThreshold {
				outDepth.Set(x, y, 0)
				outColor.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	return outDepth, outColor
}

func clamp(val, minV, maxV float64) float64 {
	return math.Min(math.Max(val, minV), maxV)
}

func ApplyAnisotropicFilling(inDepth texturing.Texture[float64], inColor texturing.Texture[color.Color], iterations int, depthContribution float64) (texturing.Texture[float64], texturing.Texture[color.Color]) {
	outDepth := inDepth
	outColor := inColor

	for range iterations {
		currentDepth := outDepth.Copy()
		currentColor := outColor.Copy()
		for inY := 3; inY < inDepth.Height()-3; inY++ {
			for inX := 3; inX < inDepth.Width()-3; inX++ {

				// Don't update valid pixels
				if inDepth.Get(inX, inY) > 0 {
					continue
				}

				pos := vector2.New(inX, inY)
				curDepth := outDepth.Get(inX, inY)

				// empty pixels / non-valid depth
				if curDepth <= 0 {
					type dc struct {
						Depth float64
						Color color.Color
					}
					vals := make([]dc, 0)

					outDepth.SearchNeighborhood(pos, 1, func(x, y int, v float64) {
						if v <= 0 {
							return
						}
						vals = append(vals, dc{
							Depth: v,
							Color: outColor.Get(x, y),
						})
					})

					if len(vals) == 0 {
						continue
					}
					sort.Slice(vals, func(i, j int) bool {
						return vals[i].Depth < vals[j].Depth
					})
					currentDepth.Set(inX, inY, vals[len(vals)/2].Depth)
					currentColor.Set(inX, inY, vals[len(vals)/2].Color)
				} else {
					sum := 0.
					rSum := 0.
					gSum := 0.
					bSum := 0.
					totalWeight := 0.

					outDepth.SearchNeighborhood(pos, 1, func(x, y int, v float64) {
						if v <= 0 {
							return
						}

						if x == inX && y == inY {
							return
						}

						r, g, b, _ := outColor.Get(x, y).RGBA()

						radius := pos.ToFloat64().Sub(vector2.New(x, y).ToFloat64()).Length()
						radiualWeight := 1. - ((radius) / 2.)

						depthWeight := (1. - min(1., math.Abs(v-curDepth)/depthContribution))

						weight := depthWeight * radiualWeight

						sum += weight * v
						rSum += weight * (float64(r>>8) / 255.)
						gSum += weight * (float64(g>>8) / 255.)
						bSum += weight * (float64(b>>8) / 255.)

						totalWeight += weight
					})

					if totalWeight > 0 {
						currentDepth.Set(inX, inY, sum/totalWeight)
						currentColor.Set(inX, inY, color.RGBA{
							R: uint8(math.Round(clamp(rSum/totalWeight, 0, 1.) * 255.)),
							G: uint8(math.Round(clamp(gSum/totalWeight, 0, 1.) * 255.)),
							B: uint8(math.Round(clamp(bSum/totalWeight, 0, 1.) * 255.)),
							A: 255,
						})
					}
				}
			}
		}
		outDepth = currentDepth
		outColor = currentColor
	}

	return outDepth, outColor
}

func SaveOutDepth(name string, tex texturing.Texture[float64]) error {
	maxDepth := 0.
	minDepth := math.MaxFloat64

	tex.Scan(func(x, y int, depth float64) {
		if depth == 0 {
			return
		}
		maxDepth = max(maxDepth, depth)
		minDepth = min(minDepth, depth)
	})

	depthRange := maxDepth - minDepth
	return tex.SaveImage(name, func(c float64) color.Color {
		if c == 0 {
			return color.RGBA{R: 255, B: 255, G: 255, A: 255}
		}
		v := byte(((c - minDepth) / depthRange) * 255)
		return color.RGBA{R: v, B: v, G: v, A: 255}
	})
}

func colorPassthrough(c color.Color) color.Color { return c }

func loadMesh() (*modeling.Mesh, error) {
	return ply.Load("./test-models/stanford-bunny.ply")

	doc, buffs, err := gltf.ExperimentalLoad("C:/Users/elida/Downloads/example.gltf")
	if err != nil {
		return nil, err
	}

	models, err := gltf.ExperimentalDecodeModels(doc, buffs, "C:/Users/elida/Downloads/")
	if err != nil {
		return nil, err
	}

	finalMesh := modeling.EmptyMesh(models[0].Mesh.Topology())
	for _, v := range models {
		mesh := *v.Mesh

		if v.TRS != nil {
			mesh = mesh.ApplyTRS(*v.TRS)
		}

		finalMesh = finalMesh.Append(mesh)
	}
	return &finalMesh, nil
}

func run(mesh modeling.Mesh, camera Camera, i int) {
	start := time.Now()
	colorTex := texturing.NewTexture[color.Color](Width, Height)
	colorTex.Fill(color.White)

	depthTexture := texturing.NewTexture[float64](Width, Height)
	positionTexture := texturing.NewTexture[vector3.Float64](Width, Height)

	points := mesh.Float3Attribute(modeling.PositionAttribute)
	colors := iter.Array(make([]vector4.Float64, points.Len()))
	if mesh.HasFloat4Attribute(modeling.ColorAttribute) {
		colors = mesh.Float4Attribute(modeling.ColorAttribute)
	}

	for p := range points.Len() {
		point := points.At(p)

		// Prune points behind camera
		if point.Z() <= 0 {
			continue
		}

		cord, err := camera.ProjectPoint(point)
		noErr(err)

		if cord.ContainsNaN() || cord.X() < 0 || cord.Y() < 0 || cord.X() > Width || cord.Y() > Height {
			continue
		}

		cordI := cord.ToInt()
		x := cordI.X()
		y := cordI.Y()

		depth := point.Length()

		// Don't render a point that's behind another point
		if cur := depthTexture.Get(x, y); cur > 0 && cur < depth {
			continue
		}

		depthTexture.Set(x, y, depth)
		positionTexture.Set(x, y, point)

		pointColor := colors.At(p)
		colorTex.Set(x, y, color.RGBA{
			R: byte(pointColor.X() * 255),
			G: byte(pointColor.Y() * 255),
			B: byte(pointColor.Z() * 255),
			A: 255,
		})
	}

	noErr(colorTex.SaveImage(fmt.Sprintf("%d-%s", i, OutName), colorPassthrough))
	noErr(SaveOutDepth("depth.png", depthTexture))

	visibilityDepth, visbilityColor := ComputeVisibility(depthTexture, positionTexture, colorTex, 11)

	noErr(SaveOutDepth("vis-depth.png", visibilityDepth))
	noErr(visbilityColor.SaveImage("vis-"+OutName, colorPassthrough))

	depthContribution := 2.01 // The greater the value, the larger the contribution
	anisotropicFillingDepth, anisotropicFillingColor := ApplyAnisotropicFilling(visibilityDepth, visbilityColor, 6, depthContribution)

	noErr(SaveOutDepth("ani-depth.png", anisotropicFillingDepth))
	noErr(anisotropicFillingColor.SaveImage(fmt.Sprintf("ani-vis-%d.png", i), colorPassthrough))

	log.Printf("Computed in %s", time.Since(start))
}

func main() {
	loadedMesh, err := loadMesh()
	noErr(err)
	centeredMesh := meshops.CenterFloat3Attribute(*loadedMesh, modeling.PositionAttribute)

	// mesh := loadedMesh.
	// 	Scale(vector3.New(1, -1, -1.).Scale(.01)).
	// 	Rotate(quaternion.FromEulerAngle(vector3.New(0., 0., 0.))).
	// 	Translate(vector3.New(0, -.13, 0.25))

	mesh := centeredMesh.
		Scale(vector3.New(1, -1, -1.).Scale(1)).
		Rotate(quaternion.FromEulerAngle(vector3.New(0., 0., 0.))).
		Translate(vector3.New(0, 0, 0.25))

	camera := Camera{
		FocalLength: 800,    // in pixels
		Width:       Width,  // image width
		Height:      Height, // image height
	}

	for i := range 1 {
		run(mesh.Translate(vector3.New(0., 0., -0.0025*float64(i))), camera, i)
	}
}
