package kmeans

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

// Result holds the results of k-means clustering
type Result[T any] struct {
	Centroids  []T
	Labels     []int
	Iterations int
}

func Run4D[T vector.Number](points []vector4.Vector[T], centroidCount int, maxIterations int, tolerance float64) Result[vector4.Vector[T]] {
	return Run(points, centroidCount, maxIterations, tolerance, vector4.Space[T]{})
}

func Run3D[T vector.Number](points []vector3.Vector[T], centroidCount int, maxIterations int, tolerance float64) Result[vector3.Vector[T]] {
	return Run(points, centroidCount, maxIterations, tolerance, vector3.Space[T]{})
}

func Run2D[T vector.Number](points []vector2.Vector[T], centroidCount int, maxIterations int, tolerance float64) Result[vector2.Vector[T]] {
	return Run(points, centroidCount, maxIterations, tolerance, vector2.Space[T]{})
}

func Run1D[T vector.Number](points []T, centroidCount int, maxIterations int, tolerance float64) Result[T] {
	return Run(points, centroidCount, maxIterations, tolerance, vector1.Space[T]{})
}

// Run performs k-means clustering on a set of 3D points
func Run[T any](points []T, centroidCount int, maxIterations int, tolerance float64, space vector.Space[T]) Result[T] {
	if len(points) == 0 {
		panic(fmt.Errorf("nothing to cluster around"))
	}

	if centroidCount <= 0 {
		panic(fmt.Errorf("invalid centroid count %d", centroidCount))
	}

	if maxIterations <= 0 {
		panic(fmt.Errorf("invalid iteration count %d", maxIterations))
	}

	// Initialize centroids randomly
	centroids := initializeCentroids(points, centroidCount)
	labels := make([]int, len(points))

	for iteration := range maxIterations {
		// Assign each point to the nearest centroid
		hasChanged := assignPointsToCentroids(points, centroids, labels, space)

		// Update centroids based on assigned points
		newCentroids := updateCentroids(points, labels, centroidCount, space)

		// Check for convergence
		if !hasChanged || hasConverged(centroids, newCentroids, tolerance, space) {
			return Result[T]{
				Centroids:  newCentroids,
				Labels:     labels,
				Iterations: iteration + 1,
			}
		}

		centroids = newCentroids
	}

	return Result[T]{
		Centroids:  centroids,
		Labels:     labels,
		Iterations: maxIterations,
	}
}

func initializeCentroids[T any](points []T, k int) []T {
	centroids := make([]T, k)

	for i := range k {
		centroids[i] = points[rand.Intn(len(points))]
	}

	return centroids
}

// assignPointsToCentroids assigns each point to the nearest centroid
func assignPointsToCentroids[T any](points []T, centroids []T, labels []int, space vector.Space[T]) bool {
	hasChanged := false

	for i, point := range points {
		minDist := math.Inf(1)
		newLabel := 0

		// Find the nearest centroid
		for j, centroid := range centroids {
			dist := space.Distance(point, centroid)
			if dist < minDist {
				minDist = dist
				newLabel = j
			}
		}

		// Check if assignment changed
		if labels[i] != newLabel {
			hasChanged = true
			labels[i] = newLabel
		}
	}

	return hasChanged
}

// updateCentroids calculates new centroids based on assigned points
func updateCentroids[T any](points []T, labels []int, clusterCount int, space vector.Space[T]) []T {
	centroids := make([]T, clusterCount)
	counts := make([]int, clusterCount)

	// Sum all points assigned to each centroid
	for i, point := range points {
		label := labels[i]
		centroids[label] = space.Add(centroids[label], point)
		counts[label]++
	}

	// Calculate average (centroid) for each cluster
	for i := range clusterCount {
		if counts[i] > 0 {
			centroids[i] = space.Scale(centroids[i], 1./float64(counts[i]))
		}
	}

	return centroids
}

// hasConverged checks if centroids have converged within tolerance
func hasConverged[T any](oldCentroids, newCentroids []T, tolerance float64, space vector.Space[T]) bool {
	for i := range oldCentroids {
		if space.Distance(oldCentroids[i], newCentroids[i]) > tolerance {
			return false
		}
	}
	return true
}

// CalculateWCSS calculates Within-Cluster Sum of Squares
func CalculateWCSS(points []vector3.Float64, centroids []vector3.Float64, labels []int) float64 {
	wcss := 0.0
	for i, point := range points {
		centroid := centroids[labels[i]]
		dist := point.Distance(centroid)
		wcss += dist * dist
	}
	return wcss
}

// // Example usage
// func main() {
// 	// Create sample 3D points
// 	points := []vector3.Float64{
// 		vector3.NewFloat64(1.0, 1.0, 1.0),
// 		vector3.NewFloat64(1.5, 2.0, 1.5),
// 		vector3.NewFloat64(3.0, 4.0, 3.0),
// 		vector3.NewFloat64(5.0, 7.0, 5.0),
// 		vector3.NewFloat64(3.5, 5.0, 3.5),
// 		vector3.NewFloat64(4.5, 5.0, 4.5),
// 		vector3.NewFloat64(3.5, 4.5, 3.5),
// 		vector3.NewFloat64(10.0, 10.0, 10.0),
// 		vector3.NewFloat64(11.0, 11.0, 11.0),
// 		vector3.NewFloat64(12.0, 12.0, 12.0),
// 	}

// 	// Perform k-means clustering
// 	k := 3
// 	maxIterations := 100
// 	tolerance := 0.01

// 	result := KMeans(points, k, maxIterations, tolerance)

// 	if result != nil {
// 		fmt.Printf("K-Means completed in %d iterations\n", result.Iterations)
// 		fmt.Println("Centroids:")
// 		for i, centroid := range result.Centroids {
// 			fmt.Printf("  Cluster %d: (%.2f, %.2f, %.2f)\n", i, centroid.X(), centroid.Y(), centroid.Z())
// 		}

// 		fmt.Println("\nPoint assignments:")
// 		for i, point := range points {
// 			fmt.Printf("  Point (%.1f, %.1f, %.1f) -> Cluster %d\n",
// 				point.X(), point.Y(), point.Z(), result.Labels[i])
// 		}

// 		wcss := CalculateWCSS(points, result.Centroids, result.Labels)
// 		fmt.Printf("\nWithin-Cluster Sum of Squares: %.4f\n", wcss)
// 	} else {
// 		fmt.Println("K-Means failed")
// 	}
// }
