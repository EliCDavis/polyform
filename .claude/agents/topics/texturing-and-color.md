# Texturing and color

Read this once geometry is assembled and you're deciding whether a large
or visually prominent surface should stay a flat `MaterialNode` color or
get real surface variation — real materials almost never are perfectly
uniform. Small or inherently-uniform parts are usually fine flat; don't
force texture onto something that's correctly plain.

Which technique applies depends on whether the part has UVs:

## Material extensions: which to read

Beyond colour, metallic and roughness, a material can carry glTF
extensions. Each has its own page in `topics/materials/` with exact type
keys, values and traps. `Read` the ones the part needs before wiring
them:

| The part is... | Read |
| --- | --- |
| A see-through surface | `materials/transmission.md`, plus `materials/volume.md` if solid |
| A transparent solid that refracts or tints with depth | `materials/transmission.md`, `materials/volume.md`, `materials/ior.md` |
| A refracting solid that splits light into colours | all of the above plus `materials/dispersion.md` |
| A finish with highlights stretched along one direction | `materials/anisotropy.md` |
| A glossy coat over a base layer, or a wet surface | `materials/clearcoat.md` |
| A soft fibrous surface | `materials/sheen.md` |
| A smooth but dull non-metal | `materials/specular.md` |
| A thin-film surface whose colour shifts with viewing angle | `materials/iridescence.md` |
| Anything that emits light | `materials/emissive-strength.md` |
| Any non-metal that should look glossier or duller without changing roughness | `materials/ior.md`, `materials/specular.md` |
| Cut-outs or a faded, partly see-through surface | `Alpha Mode` on `MaterialNode`; `materials/transmission.md` explains which to use |
| Flat, unshaded colour | `materials/unlit.md` |
| A small texture tiled, shifted or turned across a big part | `materials/texture-transform.md` |

None of them shows in `render_preview`, which draws only base colour,
colour texture, metallic and roughness. Verify each with
`describe_manifest`, which lists every extension and its values per
material. The web editor's viewer renders them all.

## Exact type keys, do not search for these

Each `PATH` below goes inside `github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`
(the `parameter.Value` rows are the exception — they are already the
whole key, written out in full, and take no wrapper). Confirmed against the registry; **use them directly rather
than spending a `search_node_types`/`get_node_types` pair.**

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialNode` | Color, `Color Texture`, `Metallic Factor`, `Roughness Factor`, `Emissive Factor`, `Emissive Strength`, `Normal Texture`, `Occlusion Texture`, `Alpha Mode`, `Alpha Cutoff`, Unlit, Name, ... | Out |
| `formats/gltf.OcclusionTextureNode` | Texture, Strength | Out |
| `math/noise.Perlin3DNode` | Amplitude, Frequency, Shift, Time | Out |
| `modeling.SetAttribute3DNode` | Mesh, Attribute, Data | Out |
| `modeling.FillAttribute3DNode` | Mesh, Attribute, Value | Out |
| `drawing/coloring.MultiplyNode` | A, B | Out |
| `drawing/coloring.InterpolateNode` | A, B, Time, Clamp | Out |
| `math.MultiplyNode[float64]` | Values | Float, Int |
| `math.AddNode[float64]` | Values | Float, Int |
| `math.SubtractNode[float64]` | A, B | Float, Int |
| `math.MinNode[float64]` | In | `Float 64`, Int |
| `math.MaxNode[float64]` | In | `Float 64`, Int |
| `drawing/coloring.ToVectorNode` | In | `Vector 3`, `Vector 4` |
| `math.RemapNode[float64]` | Value, `In Min`, `In Max`, `Out Min`, `Out Max` | Out |
| `github.com/EliCDavis/polyform/generator/parameter.Value[float64]` | *(none)* | Value |
| `github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/drawing/coloring.Color]` | *(none)* | Value |

`OcclusionTextureNode` wraps an ambient-occlusion image (red channel, 0
fully occluded, 1 open) for `MaterialNode`'s `Occlusion Texture`, the
way `NormalTextureNode` wraps a normal map; it darkens crevices under
ambient light and needs UVs. `Alpha Mode` takes `"OPAQUE"`, `"MASK"` or
`"BLEND"`, and `Alpha Cutoff` only applies with `MASK`.

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
mechanical assemblies built from them) and `QuadSphereNode` generate UVs
— **`UvSphereNode` does not**, despite the name (position only, plus
normals when `Weld` is true), and neither do `extrude.AreaNode` or `extrude.CircleNode`.
`extrude.OutlineNode`, `CircleAlongSplineNode` and `ScrewNode` lay the
sweep out as a strip — the outline's perimeter runs across it, the path
along it, each cap taking the whole strip as its own unwrap — spanning 0
to 1 both ways unless their `UVs` port takes a `primitives.StripUVsNode`
(Width, Start, End) to set the repeat. On a `ScrewNode` whose radius
varies, that stretches a texture big at the rim and pinched near the
axis; set its `Tile Size` (world size of one repeat) instead, and the
texture keeps one size everywhere, with a whole number of repeats around
each ring so the seam meets. `csg` carries through whatever
UVs its operands already had, so make them before the cuts, not after.
`meshops.CombineNode` fills a missing attribute with zeros, so combining
a mesh without UVs into one with them gives it all-zero UVs, not none.

Where a mesh genuinely has none — a marched surface, an `AreaNode` slab,
or a boolean whose operands never had any — project UVs from positions.
Apply the projection **last**, after every cut and combine. Pick by
shape:

| PATH | inputs | outputs |
| --- | --- | --- |
| `modeling/meshops.BoxProjectUVNode` | Mesh, Scale, Offset | Out |
| `modeling/meshops.CylinderProjectUVNode` | Mesh, Axis, Center, Around, Height | Out |
| `modeling/meshops.ProjectUVNode` | Mesh, Normal, Scale, Offset, Rotation | Out |

- **`BoxProjectUVNode`** projects each face from the axis it faces most,
  like a box around the part. Use it for anything blocky or slab-like,
  edges included. No face is left edge-on.
- **`CylinderProjectUVNode`** wraps UVs around `Axis` (default +Y),
  through `Center` (default the mesh's bounds centre), like a band
  wrapped around it. `u` runs around, `Around` times per turn, and `v`
  runs along the axis, one repeat per `Height` metres. Faces pointing
  along the axis (the caps) get a flat projection instead. Use it for
  cylindrical or turned parts, and for a brushed finish that should
  streak around.
- **`ProjectUVNode`** is one flat projection along `Normal` (default
  +Z: u along X, v along Y). It suits only a thin plate seen from one
  side.

`Scale` on the box and plane projections is the world size of one
texture repeat per axis, so an unequal scale like `{"x":0.25,"y":2}`
stretches the texture into grain running up the part.

A flat projection gives faces edge-on to it zero-area UVs. A colour
texture just smears there, but a normal map or anisotropy has no
direction on them, and a viewer shades them blown-out white.
`describe_manifest` warns when a normal-mapped or anisotropic mesh has
zero-area UV triangles; the fix is the box or cylinder projection. Use
the strip when the part *is* a sweep. For a sphere-like part, vertex
color (below) still reads better.

`render_preview` with `uvChecker: true` paints every part that has UVs
with a checker, one 8x8 grid per UV tile, red growing with U and green
with V. Check it before texturing: stretched or sheared cells, a size
change across the part, a seam or a pinched pole all show directly.

Where UVs do exist, the proven pipeline is `drawing/texturing.NoiseNode`
(or `SeamlessPerlinNode`) -> `texturing.ApplyGradientNode[coloring.Color]`
-> `gltf.TextureNode` (a color texture casts to the `Image` it takes, no
conversion node) -> `MaterialNode.ColorTexture`. `ApplyGradientNode.Gradient` takes a
`coloring.Gradient[...coloring.Color]` **literal or variable** directly —
`{"keys":[{"time":0,"value":"#2a4d1e"},{"time":1,"value":"#9bbf6a"}]}` —
so a colour ramp the user should be able to tune is one variable with a
gradient-bar editor, not a `GradientColorNode` fed by `GradientKeyNode`s
(that chain still works and is what the example graphs do). Don't guess the
wiring — `Read` an existing example graph that already does exactly this
(`generator/edit/examples/terrain.json`, `doughnut.json`, or
`snowglobe.json`, which also shows the matching normal-map half via
`normals.FromHeightMapNode`) and follow its pattern.
`generator/edit/examples/texturing.json` wires every node in this topic
into twenty-four small materials, one subgraph each; read the subgraph
matching what you need:
- **Node walkthroughs:** patterns, UV warps, texture math, normals,
  occlusion, packed metallic-roughness.
- **Classic materials:** wood, marble, brushed metal (anisotropy),
  weathered paint (masks built from masks: paint, chips, rust), water,
  emissive lava, brick, scales, layered terrain (masks, puddles, height
  shaping), leather, and a riveted plate (normals drawn from a position
  list).
- **Material extensions:** frosted glass (transmission, volume), car
  paint (clearcoat), knit wool (sheen), gemstone (dispersion), glazed ceramic (specular).

Each one generates a texture once and wires it into several maps
(color, roughness, normal, an extension's texture). Do the same: the
noise behind a color is usually the right roughness or height too.

### Procedural patterns

A texture that isn't noise — cloth, weave, tiles, stripes, plaid — is
built from textures end to end, so every step previews:

1. A `UVNode` (Dimensions) makes a UV texture: each pixel holds its own
   0-1 coordinate. Bend it with the warp nodes if the pattern needs it.
2. Pattern nodes read that UV texture and turn it into `Texture[float64]`
   values. Their `UV` input is required; unwired, they report an error.
3. The texture math nodes shape and combine them; `ApplyGradientNode`,
   `FromHeightMapNode` and the occlusion node take the result.

Every texture in one chain must be the same size. The `UVNode`'s
Dimensions set it for everything downstream; one `UVNode` can feed
several patterns.

| PATH | inputs | outputs |
| --- | --- | --- |
| `drawing/texturing.UVNode` | Dimensions | UV |
| `drawing/texturing/pattern.TransformUVNode` | In, Scale, Rotation, Offset | UV |
| `drawing/texturing/pattern.ZigzagUVNode` | In, Count, Amplitude | UV |
| `drawing/texturing/pattern.DisplaceUVNode` | In, Amount, Shift | UV |
| `drawing/texturing/pattern.CellNode` | UV, Columns, Rows, `Row Offset`, Seed | Column, Row, `Local UV`, Random |
| `drawing/texturing/pattern.DistanceNode` | UV, Center | Value |
| `drawing/texturing/pattern.SplitUVNode` | UV | U, V |
| `drawing/texturing/pattern.StripesNode` | UV, Count, Width, Softness, Angle | Value |
| `drawing/texturing/pattern.CheckerNode` | UV, Columns, Rows | Value |
| `drawing/texturing/pattern.GinghamNode` | UV, Columns, Rows, Width, Softness | Value |
| `drawing/texturing/pattern.WeaveNode` | UV, Columns, Rows, Over, Under, Herringbone, Gap, Seed | Height, Weft, Warp, Random |

- **`TransformUVNode`** scales (tiles), rotates about the centre and
  slides the pattern. **`ZigzagUVNode`** shifts V up and down along U,
  turning straight bands into chevrons: a plait is a weave or stripes
  sampled through it. **`DisplaceUVNode`** moves each UV by `Shift` times
  the matching pixel of `Amount`; with noise as `Amount`, stripes become
  wood rings, marble veins or ripples.
- **`CellNode`** splits UV into a grid: which cell, the UV inside it,
  and a seeded random value per cell, for per-tile or per-plank
  variation. Feed `Local UV` to another pattern to repeat it per cell.
  `Row Offset` 0.5 slides every other row half a cell: running-bond
  brick, scales, shingles. Never hand-roll a hash from `sin`.
- **`DistanceNode`** is each pixel's distance from `Center` in UV units.
  Fed `Local UV`, it gives one disc per cell for dots, studs or scales.
- **`SplitUVNode`** gives U and V as 0-1 ramps. Fed `Local UV`, each
  ramp restarts per cell; scaled by a cell's `Random` and added to a
  height, it tilts every cell its own way (uneven tiles, warped planks).
- **`GinghamNode`** gives 0, 0.5 or 1: no band, one band, or two crossing.
  A three-key gradient colours it.
- **`WeaveNode`** gives strands passing over and under each other:
  - `Over` and `Under` set the pattern: 1 and 1 is a plain weave, 2 and 2
    a twill.
  - `Herringbone` reverses the diagonal every that many columns.
  - `Gap` opens space between strands.
  - `Height` feeds `FromHeightMapNode` and the occlusion node below.
  - `Weft` and `Warp` are 1 where that direction's strand shows, for
    colouring the two directions differently.
  - `Random` is one value per strand; run it through a gradient for
    strand-to-strand colour variation.

Whole column and row counts tile across the UV seam.

Texture math, every input a texture of the same size:

| PATH | inputs | outputs |
| --- | --- | --- |
| `drawing/texturing.AddFloat1Node` | Textures | Result |
| `drawing/texturing.SubtractFloat1Node` | A, B | Result |
| `drawing/texturing.MultiplyFloat1Node` | Textures | Result |
| `drawing/texturing.MinFloat1Node` / `MaxFloat1Node` | Textures | Result |
| `drawing/texturing.ScaleFloat1UniformNode` | Texture, Scale | Result |
| `drawing/texturing.OneMinusNode` | Texture | Result |
| `drawing/texturing.RemapFloat1Node` | Texture, `In Min`, `In Max`, `Out Min`, `Out Max`, Clamp | Result |
| `drawing/texturing.ClampFloat1Node` | Texture, Min, Max | Result |
| `drawing/texturing.PowerFloat1Node` | Texture, Exponent | Result |
| `drawing/texturing.SmoothstepFloat1Node` | Texture, Low, High | Result |
| `drawing/texturing.MixFloat1Node` | A, B, Factor, Amount | Result |
| `drawing/texturing.MixColorNode` | A, B, Factor, Amount | Result |

`Mix` blends by `Factor` per pixel when it is wired, by `Amount`
everywhere otherwise; `MixColorNode` with a pattern as `Factor` lays one
colour texture over another. `Smoothstep` with `Low` equal to `High` is
a hard threshold.

### Normal maps

Surface relief too fine to model — grain, grooves, a brushed or cast
finish — is a normal map, and it costs one node on the chain you already
built for colour:

| PATH | inputs | outputs |
| --- | --- | --- |
| `drawing/texturing/normals.FromImageNode` | In, Scale | Heightmap, Normalmap, `Normal Map Image` |
| `drawing/texturing/normals.FromHeightMapNode` | In, Scale | `Normal Map`, `Normal Map Image` |
| `drawing/texturing/normals.OcclusionFromHeightMapNode` | In, Radius, Strength | Occlusion, `Occlusion Image` |
| `drawing/texturing/normals.NewNode` | Dimensions | `Normal Map` |
| `drawing/texturing/normals.DrawLineNode` | Start, End, Thicknesses, Texture, Subtract | `Normal Map` |
| `formats/gltf.NormalTextureNode` | Texture, Scale | Out |

`FromImageNode.In` is an `image.Image` — the same image or color texture
you fed `gltf.TextureNode` for colour. `FromHeightMapNode.In` is a
`texturing.Texture[float64]`. `NewNode` makes a blank map to draw into,
and `DrawLineNode`/`DrawLinesNode`/`DrawSphereNode`/`DrawSpheresNode`
scratch creases and bumps straight onto one (`Subtract` carves instead of
raising).

Take the same texture feeding the color `gltf.TextureNode`, run it through
`normals.FromImageNode` with a `Scale` (relief depth; start near 1 and
judge it), take **`Normal Map Image`**, and wire that into a second
`gltf.TextureNode`. **That does not go straight into the material**:
`MaterialNode`'s `Normal Texture` takes a `gltf.PolyformNormal`, not a
`PolyformTexture`, so the texture passes through
`gltf.NormalTextureNode` (`Texture`, `Scale` → `Out`) first — its own
`Scale` is the strength the map applies at, separate from the relief
baked into the image. The colour texture and the normal map then line up
because they share UVs.

glTF does not require a TANGENT attribute: a viewer computes the tangent
basis from the normals and UVs when the mesh has none, so there is
nothing extra to generate.

The height map that made the normal map also makes occlusion:
`OcclusionFromHeightMapNode` darkens the low points that higher
surroundings shade. Its `Occlusion Image` goes through a
`gltf.TextureNode` into `OcclusionTextureNode` and the material's
`Occlusion Texture`. Alternatively, multiply its `Occlusion` into the
colour so the gaps read dark everywhere.

`render_preview` applies the normal map, the occlusion map and emission
on parts with UVs. Relief reads best under a raking light (`light`
azimuth 60-80, low ambient).

## Vertex-colored parts (marched/SDF meshes, or anything without UVs)

`MarchNode` never generates UVs, so `gltf.MaterialNode.ColorTexture` has
nothing to map onto. Write into the mesh's `"Color"` attribute instead.

**One flat color per part is one fill**:
`modeling.FillAttribute3DNode` (`Mesh`, `Attribute: "Color"`, `Value`)
writes the same vector to every vertex, creating the attribute if it
isn't there — no `SelectFromMeshNode`, no per-vertex array to build.
Feed `Value` from `drawing/coloring.ToVectorNode`'s `Vector 3` so the
color can stay a `Color` variable upstream. Use it whenever a part is
one color and only the *assembly* is multi-colored: each part gets its own fill, and `MaterialNode.Color` stays
white. `1D`/`2D`/`4D` variants take a float / `vector2` / `vector4` for
any other attribute.

**Color that varies across a part** is the per-vertex pipeline:
`SelectFromMeshNode`'s `Position` output -> per-vertex values
(`math/noise.Perlin3DNode` takes an array of positions, returns one float
per vertex — a real, concrete starting point) -> colors ->
`SetAttribute3DNode` with `inputs: {"Attribute": {"value": "\"Color\""}}`.

**For a simple two-color gradient along one world axis**, skip the manual wiring entirely —
`create_vertex_color_gradient_subgraph` builds the whole
select/remap/interpolate/write chain in one call and derives the
gradient's range from the mesh's own actual extent, not a guessed one.

For anything more elaborate (a noise-driven band pattern, a
gradient that isn't a straight world-axis fade), wire it manually: to
turn per-vertex float values into per-vertex colors, use
`drawing/coloring.InterpolateNode` (`Time` <- the per-vertex float
array, `A`/`B` <- the two colors to blend between); it lives in
`drawing/coloring`, not `drawing/texturing` (that package is 2D
image-space, not per-vertex arrays). Any of its three inputs takes a
single value or a per-vertex array, so blending two full *color arrays*
under a mask is the same node with arrays wired into `A` and `B`. `Time`
outside 0..1 extrapolates past `A`/`B` into saturated colors unless
`Clamp` is true; `RemapNode.Clamp` clamps to its *output* range, not 0..1.

**Combining two per-vertex signals is elementwise array math.**
`math.MultiplyNode[float64]` gives `A[i]*B[i]` — a Y-gradient times a
radial mask is one node. `AddNode`, `SubtractNode`, `MinNode`,
`MaxNode` do the same for `+ - min max`. There is no separate array
variant of any of these: one node covers single values, an array
against a single value, and two arrays. Two arrays of different lengths
are refused with an error on the node, which is almost always two
arrays sampled from different meshes.

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
  (fine grain, mottling, a subtle fade) and put the actual hue on
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
look produces chrome, not a shiny non-metal surface, because the
metalness default was never touched. Any material that should read as
shiny *and* non-metal needs `MetallicFactor` explicitly wired to `0` — a
literal `{"value": "0"}` is enough, it doesn't need to be a variable
unless you want it tunable.

The reverse trap for things that *are* metal: `MetallicFactor 1.0`
reflects only its environment, and a viewer with no environment map —
most quick viewers — renders it near-black. `render_preview` lights metal
with a stand-in sky, so it shows the metal's color and gloss but not that
darkness. A small metal detail reads as metal at metallic `0.4-0.6`,
roughness `0.3-0.4`, with the base color a step lighter than the intended
color; reserve `1.0` for a scene that will be lit by an HDRI.

## Hex colors are sRGB; glTF wants linear (applies to every color you set)

glTF reads `baseColorFactor` and the vertex `Color` attribute as **linear**
multipliers, and every viewer (the web editor, model-viewer, Blender)
applies the sRGB curve on output. An sRGB hex read as linear displays
washed out: `#c8322f` becomes pale salmon, a mid-tone `#d3a468`
near-white. `render_preview` treats colors the same way, so if the preview
looks washed out, the export will too — that is the signal, not a preview
bug.

- `MaterialNode`'s `Color` and `Emissive Factor` take the hex as picked
  and convert it themselves. Wire the variable straight in; a
  `SRGBToLinearNode` in front of them converts twice and darkens.
- Color textures are sRGB by spec; leave them alone.
- Everything else is linear and not converted for you. Convert at the
  last step, keep the variables human:
  - one color (a `WithColorNode` color, an extension node's color
    factor) → `drawing/coloring.SRGBToLinearNode` (`In` → `Out`);
  - a marched mesh's vertex colors → `modeling/meshops.SrgbToLinearNode`
    (`Mesh`, `Attribute: "Color"`) after `ApplyColorFieldNode` or the
    gradient subgraph, if the colors that went in were sRGB picks.

Don't do both on the same color (a colored field converted per
`WithColorNode` *and* again on the mesh) — it double-darkens. White,
black and greys near either end survive unconverted; saturated mid-tones
are where the difference is large.
