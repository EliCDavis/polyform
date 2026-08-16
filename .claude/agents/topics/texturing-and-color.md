# Texturing and color

Read this once geometry is assembled and you're deciding whether a large
or visually prominent surface (a car's body panels, a table's tabletop, a
wall, an animal's coat) should stay a flat `MaterialNode` color or get
real surface variation — real materials almost never are perfectly
uniform. Small or inherently-uniform parts (a bolt, a thin wire, glass, a
simple painted plastic trim piece) are usually fine flat; don't force
texture onto something that's correctly plain.

Which technique applies depends on whether the part has UVs:

## UV-textured parts

`CubeNode`/`CylinderNode`/`TorusNode`/cone/circle/quad primitives (and
mechanical assemblies built from them) generate UVs — **`UvSphereNode`/
`QuadSphereNode` do not**, despite the name (position/normal only); for a
textured sphere-like part, use vertex color instead (below) or accept a
flat/gradient `MaterialNode.Color`.

Where UVs do exist, the proven pipeline is `drawing/texturing.NoiseNode`
(or `SeamlessPerlinNode`) -> `texturing.ApplyGradientNode[coloring.Color]`
(sampling a `coloring.GradientColorNode`) -> `texturing.ColorToImageNode`
-> `gltf.TextureNode` -> `MaterialNode.ColorTexture`. Don't guess the
wiring — `Read` an existing example graph that already does exactly this
(`generator/edit/examples/terrain.json`, `doughnut.json`, or
`snowglobe.json`, which also shows the matching normal-map half via
`normals.FromHeightMapNode`) and follow its pattern.

## Vertex-colored parts (marched/SDF meshes, or anything without UVs)

`MarchNode` never generates UVs, so `gltf.MaterialNode.ColorTexture` has
nothing to map onto. Write into the mesh's `"Color"` attribute instead:

`SelectFromMeshNode`'s `Position` output -> per-vertex values
(`math/noise.Perlin3DNode` takes an array of positions, returns one float
per vertex — a real, concrete starting point) -> colors ->
`SetAttribute3DNode` with `inputs: {"Attribute": {"value": "\"Color\""}}`.

To turn per-vertex float values into per-vertex colors — a flat
dorsal/belly gradient, a noise-driven band/marking pattern — use
`drawing/coloring.InterpolateToArrayNode` (`Time` <- the per-vertex float
array, `A`/`B` <- the two colors to blend between); it lives in
`drawing/coloring`, not `drawing/texturing` (that package is 2D
image-space, not per-vertex arrays). It only blends between one pair of
colors by a factor array — there's no node that blends two full color
arrays together elementwise, so a 3+-color pattern needs multiple chained
blends, not one call.

`render_preview` reads this `"Color"` attribute directly, so a correctly
vertex-colored part renders with its real color, not flat gray — check it
the same way you check any other geometry, no browser round trip needed
for this specifically.

## The metallic-factor gotcha (applies to any material, UV or vertex)

`MetallicFactor` defaults to `1.0` (fully metallic) per the glTF spec when
left unconnected, not `0`. Lowering `RoughnessFactor` alone for a glossy
look (an eye, glass, a wet nose, ceramic, polished plastic) produces
chrome, not a shiny non-metal surface, because the metalness default was
never touched. Any material that should read as shiny *and* non-metal
needs `MetallicFactor` explicitly wired to `0` — a literal `{"value":
"0"}` is enough, it doesn't need to be a variable unless you want it
tunable.
