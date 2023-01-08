# Polyform

Library for generating and editing 3D geometry and it's associated data.

## Processing Example

Reads in a obj and applies the cube marching algorithm over the meshes 3D SDF.

```go
package main

import (
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	loadedMesh, err := obj.Load("test-models/stanford-bunny.obj")
	check(err)

	resolution := 10.
	canvas := marching.NewMarchingCanvas(resolution)

	canvas.AddFieldParallel(marching.Mesh(
		loadedMesh.
			CenterFloat3Attribute(modeling.PositionAttribute).
			Scale(vector.Vector3Zero(), vector.Vector3(vector.NewVector3(12, 12, 12))),
		.1,
		10,
	))
	check(obj.Save("chunky-bunny.obj", canvas.MarchParallel(.3)))
}
```

Results in:

![Chunky Bunny](/examples/inflate/chunky-bunny.png)

## Helpful Procedural Generation Sub Packages

- Modeling
  - [marching](/modeling/marching/) - Multi-threaded Cube Marching algorithm and utilities.
  - [extrude](/modeling/extrude/) - Functionality for generating geometry from 2D shapes.
  - [repeat](/modeling/repeat/) - Functionality for copying geometry in common patterns.
  - [primitives](/modeling/repeat/) - Functionality pertaining to generating common geometry.
  - [triangulation](/modeling/triangulation/) - Generating meshes from a set of 2D points.
- Drawing
  - [coloring](/drawing/coloring/) - Color utilities for blending multiple colors together using weights.
  - [texturing](/drawing/texturing/) - Image processing utilities like generating Normal maps or blurring images.
- [Math](/math/README.md)
  - [noise](/math/noise/) - Utilities around noise functions for common usecases like stacking multiple samples of perlin noise from different frequencies.
  - [sample](/math/sample/) - Serves as a group of definitions for defining a mapping from one numeric value to another
  - [curves](/math/curves/) - Commonly curves used in animation like cubic bezier curves.

## Procedural Generation Examples

You can at the different projects under the [examples](/examples/) folder for different examples on how to procedurally generate meshes.

### Evergreen Trees

This was my [submission for ProcJam 2022](https://elicdavis.itch.io/evergreen-tree-generation). Pretty much uses every bit of functionality available in this repository.

[[Source Here](/examples/terrain/main.go)]

![Evergreen Tree Demo](./examples/chill/tree-demo.png)

![Evergreen Terrain Demo](./examples/chill/terrain-demo.png)

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

- [x] Cube Marching
- [x] Bezier Curves
- [ ] Constrained Delaunay Tesselation
- [ ] 3D Tesselation
- [ ] Slice By Plane
- [ ] Slice By Octree
- [ ] Poisson Reconstruction

## Resources

Resources either directly contributing to the code here or are just interesting finds while researching.

- Noise
  - [Perlin Noise](https://gpfault.net/posts/perlin-noise.txt.html)
    - [Perlin Worms](https://libnoise.sourceforge.net/examples/worms/index.html)
  - [Worley/Cellular Noise](https://thebookofshaders.com/12/)
  - [Book of Shaders on Noise](https://thebookofshaders.com/11/)
  - [Simplex Noise](https://en.wikipedia.org/wiki/Simplex_noise)
- Triangulation
  - Delaunay
    - Bowyer–Watson
      - [A short video overview](https://www.youtube.com/watch?v=4ySSsESzw2Y)
      - [General Algorithm Description](https://en.wikipedia.org/wiki/Bowyer%E2%80%93Watson_algorithm)
    - Constraint/Refinement
      - [Computing Constrained Delaunay Traingulations By Samuel Peterson](http://www.geom.uiuc.edu/~samuelp/del_project.html#implementation)
    - [3 Points To Create a Circle](https://kyndinfo.notion.site/Geometric-Drawings-2cefb8d81ced41d5af532dd7bdfdceee)
  - [Chew's Second Algorithm](https://cccg.ca/proceedings/2011/papers/paper91.pdf)
  - Polygons
    - [Wikipedia](https://en.wikipedia.org/wiki/Polygon_triangulation)
    - [Fast Polygon Triangulation Based on Seidel's Algorithm By Atul Narkhede and Dinesh Manocha](http://gamma.cs.unc.edu/SEIDEL/)
    - [Triangulating a Monotone Polygon
      ](http://homepages.math.uic.edu/~jan/mcs481/triangulating.pdf)
- Texturing
  - [Normal Map From Color Map](https://stackoverflow.com/questions/5281261/generating-a-normal-map-from-a-height-map)
- Formats
  - OBJ/MTL
    - [jburkardt MTL](https://people.sc.fsu.edu/~jburkardt/data/mtl/mtl.html)
    - [Excerpt from FILE FORMATS, Version 4.2 October 1995 MTL](http://paulbourke.net/dataformats/mtl/)
- Generative Techniques
  - [Country Flags by vividfax](https://vividfax.notion.site/Generative-Flag-Design-e663bc26f5a54ab48fad1428bc32b610)
  - [Snow by Ryan King](https://www.youtube.com/watch?v=UzJnsqIRbDw)
  - Terrain
    - [World Gen by Leather Bee](https://leatherbee.org/index.php/category/world-gen/)
    - [Procedural Hydrology: Dynamic Lake and River Simulation By: Nicholas McDonald](https://nickmcd.me/2020/04/15/procedural-hydrology/)
    - [The Canyons of Your Mind By JonathanCR](https://undiscoveredworlds.blogspot.com/2019/05/the-canyons-of-your-mind.html)
    - [Simulating hydraulic erosion By Job Talle](https://jobtalle.com/simulating_hydraulic_erosion.html)
    - [Coastal Landforms for Fantasy Mapping](https://www.youtube.com/watch?v=ztemzsxso0U)
  - Planet
    - [Planet Generation](https://archive.vn/kmVP4)
  - [Taming Randomness](https://kyndinfo.notion.site/Taming-Randomness-e4351f08ec7c43a7ad47ef2d1dfe2ed8)
- Voronoi
  - [Voronoi Edges by Inigo Quilez](https://iquilezles.org/articles/voronoilines/)
- Functions / Curves / Animation Lines
  - [Interpolation and Animation](https://kyndinfo.notion.site/Interpolation-and-Animation-44d00edd89bc41d686260d6bfd6a01d9)
  - [Cubic Bézier by Maxime](https://blog.maximeheckel.com/posts/cubic-bezier-from-math-to-motion/)
- Marching Cubes / SDFs
  - [LUT](http://paulbourke.net/geometry/polygonise/)
  - [Coding Adventure: Marching Cubes By Sebastian Lague](https://www.youtube.com/watch?v=M3iI2l0ltbE)
  - [SDFs](https://iquilezles.org/articles/distfunctions/)
- Collisions
  - [Closest point on Triangle](https://gdbooks.gitbooks.io/3dcollisions/content/Chapter4/closest_point_to_triangle.html)
