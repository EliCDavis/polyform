
## Benchmarking

```bash
go test ./trees/ -run XXX -bench BenchmarkOctree -benchmem -count 10 > new.txt
benchstat old.txt new.txt
```

Install `benchstat` using  `go install golang.org/x/perf/cmd/benchstat@latest`.

The in-repo benchmarks use synthetic shells so they run anywhere. To benchmark against real meshes, point the `MODELS` env at a directory of `.obj` files:

```go
package trees_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

func BenchmarkModels(b *testing.B) {
	paths, _ := filepath.Glob(filepath.Join(os.Getenv("MODELS"), "*.obj"))
	for _, path := range paths {
		mesh, err := obj.LoadMesh(path)
		if err != nil {
			b.Fatal(err)
		}

		elements := make([]trees.Element, mesh.PrimitiveCount())
		mesh.ScanPrimitives(func(i int, p modeling.Primitive) {
			elements[i] = p.Scope(modeling.PositionAttribute)
		})

		bounds := mesh.BoundingBox(modeling.PositionAttribute)
		ray := geometry.NewRay(
			bounds.Center().Add(vector3.New(0., 0., bounds.Size().Z()*2)),
			vector3.New(0., 0., -1.),
		)

		name := filepath.Base(path)
		b.Run(name+"/construct", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				trees.NewOctree(elements)
			}
		})

		tree := trees.NewOctree(elements)
		b.Run(name+"/ray", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tree.ElementsIntersectingRay(ray, 0, bounds.Size().Z()*4)
			}
		})
		b.Run(name+"/closest", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tree.ClosestPoint(bounds.Center())
			}
		})
	}
}
```

```bash
MODELS=/path/to/models go test ./trees/ -run XXX -bench BenchmarkModels -benchmem -count 10
```

This repo tested against [common-3d-test-models](https://github.com/alecjacobson/common-3d-test-models).

