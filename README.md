# Mesh

Library for editing and generating meshes.

```
go get github.com/EliCDavis/mesh
```

## Processing Example

Reads in a obj file and welds vertices, applies laplacian smoothing, and calculates smoothed normals.

```go
package main

import (
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/obj"
)

func main() {
	inFile, err := os.Open("dirty.obj")
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	loadedMesh, err := obj.ToMesh(inFile)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("smooth.obj")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	smoothedMesh := loadedMesh.
		WeldByVertices(4).
		SmoothLaplacian(5, 0.5).
		CalculateSmoothNormals()

	obj.WriteMesh(&smoothedMesh, "", outFile)
}

```

## Helpful Procedural Generation Sub Packages

- [extrude](/extrude/) - Functionality for generating geometry from 2D shapes.
- [repeat](/repeat/) - Functionality for copying geometry in common patterns.
- [primitives](/repeat/) - Functionality pertaining to generating common geometry.
- [noise](/noise/) - Utilities around noise functions for common usecases like stacking multiple samples of perlin noise from different frequencies.
- [coloring](/coloring/) - Color utilities for blending multiple colors together using weights.

## Procedural Generation Examples

You can at the different projects under the [cmd](/cmd/) folder for different examples on how to procedurally generate meshes.

### Terrain

This shows off how to use Delaunay triangulation, perlin noise, and the coloring utilities in this repository.

[[Source Here](/examples/terrain/main.go)]

![terrain](/examples/terrain/terrain.png)

### UFO

Shows off how to use the repeat, primitives, and extrude utilities in this repository.

[[Source Here](/examples/ufo/main.go)]


![ufo](/examples/ufo/ufo.png)

### Candle

Shows off how to use the primitives and extrude utilities in this repository.

[[Source Here](/examples/candle/main.go)]


![candle](/examples/candle/candle.png)

## Todo List

Things I want to implement eventually...

- [ ] Cube Marching
- [ ] Bezier Curves
- [ ] Constrained Delaunay Tesselation
- [ ] 3D Tesselation
- [ ] Slice By Plane
- [ ] Slice By Octree
- [ ] Poisson Reconstruction

## Resources Used

* [Perlin Noise](https://gpfault.net/posts/perlin-noise.txt.html)