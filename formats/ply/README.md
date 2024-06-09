# PLY Format

PLY format implementation for interfacing with polyform meshes.

## Saving and Loading

```golang
package example

import (
    "github.com/EliCDavis/polyform/formats/ply"
    "github.com/EliCDavis/vector/vector3"
)


func ExampleReadWrite() {
    mesh, _ := ply.Load("model.ply")
    scaledMesh := mesh.Scale(vector3.New(2., 2., 2.))
    ply.Save(out, scaledMesh, ply.ASCII)
}
```