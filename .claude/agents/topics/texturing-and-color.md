# Texturing and color

Read this once geometry is assembled and you're deciding whether a large
or visually prominent surface (a car's body panels, a table's tabletop, a
wall, an animal's coat) should stay a flat `MaterialNode` color or get
real surface variation — real materials almost never are perfectly
uniform. Small or inherently-uniform parts (a bolt, a thin wire, glass, a
simple painted plastic trim piece) are usually fine flat; don't force
texture onto something that's correctly plain.

Which technique applies depends on whether the part has UVs:

## Exact type keys, do not search for these

Each `PATH` below goes inside `github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`
(the `parameter.Value` rows are the exception — they are already the
whole key, written out in full, and take no wrapper). Confirmed against the registry; **use them directly rather
than spending a `search_node_types`/`get_node_types` pair.**

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialNode` | Color, `Color Texture`, `Metallic Factor`, `Roughness Factor`, `Emissive Factor`, `Emissive Strength`, `Normal Texture`, Name, ... | Out |
| `math/noise.Perlin3DNode` | Amplitude, Frequency, Shift, Time | Out |
| `modeling.SetAttribute3DNode` | Mesh, Attribute, Data | Out |
| `drawing/coloring.MultiplyNode` | A, B | Out |
| `drawing/coloring.InterpolateToArrayNode` | A, B, Time | Out |
| `drawing/coloring.ToVectorNode` | In | `Vector 3`, `Vector 4` |
| `math.RemapNode[float64]` | Value, `In Min`, `In Max`, `Out Min`, `Out Max` | Out |
| `math.RemapToArrayNode[float64]` | Value, `In Min`, `In Max`, `Out Min`, `Out Max` | Out |
| `github.com/EliCDavis/polyform/generator/parameter.Value[float64]` | *(none)* | Value |
| `github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/drawing/coloring.Color]` | *(none)* | Value |

`ToVectorNode` is the color-to-`vector3` bridge the vertex-color recipe
needs; take its `Vector 3` output. `SetAttribute3DNode` lives in the bare
`modeling` package, not `meshops`. A `coloring.color` value is a hex
string (`"#cc3333"`, or `"#cc3333ff"` with alpha), never an `{r,g,b,a}`
object.

`render_preview` multiplies a mesh's `"Color"` attribute by the
material's `Color`, the same way glTF combines `COLOR_0` with
`baseColorFactor`. So either bake the full color into the vertex
attribute and leave `MaterialNode.Color` white, or keep the vertex
attribute a near-white variation and let the material carry the hue —
both read correctly, and both agree with a real viewer. What you must not
do is set both to a strong color and expect one to win; they compound.


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

**For a simple two-color gradient along one world axis** (a
dorsal/belly fade, most organic coats), skip the manual wiring entirely —
`create_vertex_color_gradient_subgraph` builds the whole
select/remap/interpolate/write chain in one call and derives the
gradient's range from the mesh's own actual extent, not a guessed one.

For anything more elaborate (a noise-driven band/marking pattern, a
gradient that isn't a straight world-axis fade), wire it manually: to
turn per-vertex float values into per-vertex colors, use
`drawing/coloring.InterpolateToArrayNode` (`Time` <- the per-vertex float
array, `A`/`B` <- the two colors to blend between); it lives in
`drawing/coloring`, not `drawing/texturing` (that package is 2D
image-space, not per-vertex arrays). It only blends between one pair of
colors by a factor array — there's no node that blends two full color
arrays together elementwise, so a 3+-color pattern needs multiple chained
blends, not one call.

`render_preview` reads this `"Color"` attribute directly, so a correctly
vertex-colored part renders with its real color, not flat gray. If you
have `render_preview` (the orchestrator does; a spawned
`polyform-part-builder` doesn't — use `sample_field` instead, see
`organic-sdf-modeling.md`'s debugging section), check it the same way you
check any other geometry.

**Vertex color and `MaterialNode.Color` multiply**, the way glTF combines
`COLOR_0` with `baseColorFactor` — neither one wins outright. That makes
both idioms legitimate, and you should pick deliberately:

- **Tint at the material.** Write near-white per-vertex *variation*
  (leather grain, mottling, a subtle fade) and put the actual hue on
  `MaterialNode.Color`, wired to a color variable. The variable then
  really controls the part's color, which is usually what a top-level
  control is for.
- **Bake into the vertices.** Compute the finished color per vertex and
  leave `MaterialNode.Color` white. Use this when the pattern needs two
  unrelated hues that no single tint can produce.

What you must not do is both — a variable-driven material color *and* the
same color multiplied into the vertex data — which applies it twice and
renders far darker than either alone. If a part comes out muddier than
the color you chose, that is the first thing to check.

A `ColorTexture`, where a part has one, replaces both. Note that
`render_preview` does not multiply a texture by `MaterialNode.Color` even
though a real glTF viewer does, so keep a textured part's material color
white and the two agree.

## The metallic-factor gotcha (applies to any material, UV or vertex)

`MetallicFactor` defaults to `1.0` (fully metallic) per the glTF spec when
left unconnected, not `0`. Lowering `RoughnessFactor` alone for a glossy
look (an eye, glass, a wet nose, ceramic, polished plastic) produces
chrome, not a shiny non-metal surface, because the metalness default was
never touched. Any material that should read as shiny *and* non-metal
needs `MetallicFactor` explicitly wired to `0` — a literal `{"value":
"0"}` is enough, it doesn't need to be a variable unless you want it
tunable.
