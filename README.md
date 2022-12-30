# Polyform

Library for generating and editing 3D geometry and it's associated data.

❌ Doing one thing really well.

✔️ Doing everything mediocre.

## Processing Example

Reads in a obj file and welds vertices, applies laplacian smoothing, and calculates smoothed normals.

```go
package main

import (
  "os"

  "github.com/EliCDavis/polyform/formats/obj"
)

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func main() {
  loadedMesh, err := obj.Load("dirty.obj")
  check(err)

  smoothedMesh := loadedMesh.
    WeldByVertices(4).
    SmoothLaplacian(5, 0.5).
    CalculateSmoothNormals()

  check(obj.Save("smoothed.obj", smoothedMesh))
}

```

## Helpful Procedural Generation Sub Packages

- Modeling
  - [extrude](/modeling/extrude/) - Functionality for generating geometry from 2D shapes.
  - [repeat](/modeling/repeat/) - Functionality for copying geometry in common patterns.
  - [primitives](/modeling/repeat/) - Functionality pertaining to generating common geometry.
  - [triangulation](/modeling/triangulation/) - Generating meshes from a set of 2D points.
- Drawing
  - [coloring](/drawing/coloring/) - Color utilities for blending multiple colors together using weights.
  - [texturing](/drawing/texturing/) - Image processing utilities like generating Normal maps or blurring images.
- Math
  - [noise](/math/noise/) - Utilities around noise functions for common usecases like stacking multiple samples of perlin noise from different frequencies.

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

- [ ] Cube Marching
- [ ] Bezier Curves
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
- Voronoi
  - [Voronoi Edges by Inigo Quilez](https://iquilezles.org/articles/voronoilines/)
