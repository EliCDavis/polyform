package trees_test

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

func bunny(t *testing.T) modeling.Mesh {
	t.Helper()
	loaded, err := ply.Load("../test-models/stanford-bunny.ply")
	if err != nil {
		t.Skipf("no bunny to read: %v", err)
	}
	return meshops.Weld(*loaded, modeling.PositionAttribute, 1e-6)
}

// The scene shape the whole idea is aimed at: a detailed model standing on a
// floor whose two triangles each span everything.
func onAGroundPlane(m modeling.Mesh) modeling.Mesh {
	bounds := m.BoundingBox(modeling.PositionAttribute)
	reach := bounds.Size().MaxComponent() * 5
	floor := bounds.Center().Y() - bounds.Size().Y()/2

	positions := []vector3.Float64{
		vector3.New(-reach, floor, -reach),
		vector3.New(reach, floor, -reach),
		vector3.New(reach, floor, reach),
		vector3.New(-reach, floor, reach),
	}
	plane := modeling.NewTriangleMesh([]int{0, 1, 2, 0, 2, 3}).
		SetFloat3Data(map[string][]vector3.Float64{modeling.PositionAttribute: positions})

	return m.Append(plane)
}

func meshElements(m modeling.Mesh) []trees.Element {
	elements := make([]trees.Element, m.PrimitiveCount())
	m.ScanPrimitives(func(i int, p modeling.Primitive) {
		elements[i] = p.Scope(modeling.PositionAttribute)
	})
	return elements
}

func raysAt(bounds geometry.AABB, count int) []geometry.Ray {
	random := rand.New(rand.NewSource(42))
	reach := bounds.Size().MaxComponent()
	rays := make([]geometry.Ray, count)
	for i := range rays {
		origin := bounds.Center().Add(vector3.New(
			random.NormFloat64(), random.NormFloat64(), random.NormFloat64()).
			Normalized().Scale(reach * 2))
		aim := bounds.Center().Add(vector3.New(
			random.NormFloat64(), random.NormFloat64(), random.NormFloat64()).
			Scale(reach * 0.25))
		rays[i] = geometry.NewRay(origin, aim.Sub(origin).Normalized())
	}
	return rays
}

// RETENTION=1 MESH=bunny TOLERANCE=0.25 go test ./trees/ -count=1 -run TestRetentionOnRealMeshes -v
func TestRetentionOnRealMeshes(t *testing.T) {
	if os.Getenv("RETENTION") == "" {
		t.Skip("set RETENTION=1 to run the sweep")
	}

	name := os.Getenv("MESH")
	var mesh modeling.Mesh
	switch {
	case name == "bunny":
		mesh = bunny(t)
	case name == "sphere":
		mesh = primitives.UVSphere(1, 200, 200)
	case strings.HasSuffix(name, ".obj"):
		loaded, err := obj.LoadMesh(name)
		if err != nil {
			t.Fatalf("loading %s: %v", name, err)
		}
		mesh = meshops.Weld(*loaded, modeling.PositionAttribute, 1e-6)
		name = strings.TrimSuffix(filepath.Base(name), ".obj")
	default:
		t.Fatalf("set MESH to bunny, sphere or a path to an .obj")
	}
	if os.Getenv("FLOOR") != "" {
		mesh = onAGroundPlane(mesh)
		name += "+floor"
	}

	label := os.Getenv("TOLERANCE")
	tolerance := trees.NoRetention
	if label != "off" {
		parsed, err := strconv.ParseFloat(label, 64)
		if err != nil {
			t.Fatalf("set TOLERANCE to a number or to off: %v", err)
		}
		tolerance = parsed
	}

	elements := meshElements(mesh)
	depth := trees.OctreeDepthFromCount(len(elements))
	bounds := mesh.BoundingBox(modeling.PositionAttribute)
	rays := raysAt(bounds, 50_000)

	reach := bounds.Size().MaxComponent() * 8
	build := time.Duration(math.MaxInt64)
	var tree *trees.OctTree
	for i := 0; i < 5; i++ {
		start := time.Now()
		tree = trees.NewOctreeWithRetention(elements, depth, tolerance)
		if taken := time.Since(start); taken < build {
			build = taken
		}
	}

	stats := trees.Stats(tree)
	tests := 0
	for _, ray := range rays {
		tests += trees.RayTests(tree, ray, 0, reach)
	}

	hits, checksum := 0, 0
	rayTime := time.Duration(math.MaxInt64)
	for i := 0; i < 5; i++ {
		found, sum := 0, 0
		start := time.Now()
		for _, ray := range rays {
			for _, index := range tree.ElementsIntersectingRay(ray, 0, reach) {
				found++
				sum += index
			}
		}
		if taken := time.Since(start); taken < rayTime {
			rayTime = taken
		}
		hits, checksum = found, sum
	}

	fmt.Printf("MESH %s %s %d %.1f %d %d %d %.0f %.0f %.2f %d\n",
		name, label, len(elements), float64(build.Microseconds())/1000,
		stats.Nodes, stats.MaxLeaf, stats.Inner,
		float64(tests)/float64(len(rays)),
		float64(rayTime.Nanoseconds())/float64(len(rays)), float64(hits)/float64(len(rays)), checksum)
}
