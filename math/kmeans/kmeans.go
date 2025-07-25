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
