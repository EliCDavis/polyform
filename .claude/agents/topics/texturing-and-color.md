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
(Width, Start, End) to set the repeat. `csg` carries through whatever
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

Where UVs do exist, the proven pipeline is `drawing/texturing.NoiseNode`
(or `SeamlessPerlinNode`) -> `texturing.ApplyGradientNode[coloring.Color]`
-> `texturing.ColorToImageNode` -> `gltf.TextureNode` ->
`MaterialNode.ColorTexture`. `ApplyGradientNode.Gradient` takes a
`coloring.Gradient[...coloring.Color]` **literal or variable** directly —
`{"keys":[{"time":0,"value":"#2a4d1e"},{"time":1,"value":"#9bbf6a"}]}` —
so a colour ramp the user should be able to tune is one variable with a
gradient-bar editor, not a `GradientColorNode` fed by `GradientKeyNode`s
(that chain still works and is what the example graphs do). Don't guess the
wiring — `Read` an existing example graph that already does exactly this
(`generator/edit/examples/terrain.json`, `doughnut.json`, or
`snowglobe.json`, which also shows the matching normal-map half via
`normals.FromHeightMapNode`) and follow its pattern.

### Normal maps

Surface relief too fine to model — grain, grooves, a brushed or cast
finish — is a normal map, and it costs one node on the chain you already
built for colour:

| PATH | inputs | outputs |
| --- | --- | --- |
| `drawing/texturing/normals.FromImageNode` | In, Scale | Heightmap, Normalmap, `Normal Map Image` |
| `drawing/texturing/normals.FromHeightMapNode` | In, Scale | `Normal Map`, `Normal Map Image` |
| `drawing/texturing/normals.NewNode` | Dimensions | `Normal Map` |
| `drawing/texturing/normals.DrawLineNode` | Start, End, Thicknesses, Texture, Subtract | `Normal Map` |
| `formats/gltf.NormalTextureNode` | Texture, Scale | Out |

`FromImageNode.In` is an `image.Image` — the same image you fed
`gltf.TextureNode` for colour. `FromHeightMapNode.In` is a
`texturing.Texture[float64]`. `NewNode` makes a blank map to draw into,
and `DrawLineNode`/`DrawLinesNode`/`DrawSphereNode`/`DrawSpheresNode`
scratch creases and bumps straight onto one (`Subtract` carves instead of
raising).

Take the same image feeding `ColorToImageNode`, run it through
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

**`render_preview` does not apply normal maps** — it is a Phong
rasterizer shading from vertex normals, so a normal map changes nothing
in the preview even when it is wired correctly and exports correctly.
Don't tune one against a preview, and don't read a flat-looking preview
as the map being broken. Check it is wired with `describe_graph`, keep
`Scale` sane, and say in the report that it is unverifiable in-preview.

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
applies the sRGB curve on output. Polyform passes color values through
unchanged, so a hex pick like `#c8322f` wired straight into a
`MaterialNode.Color` exports as linear (0.78, 0.20, 0.18) and displays as
pale salmon, not the red you picked; a mid-tone `#d3a468` comes out
near-white. `render_preview` treats colors the same way, so if the preview
looks washed out, the export will too — that is the signal, not a preview
bug.

Convert at the last step, keep the variables human:

- one color → `drawing/coloring.SRGBToLinearNode` (`In` → `Out`) between
  the color variable/literal and the `MaterialNode`/`WithColorNode`;
- a marched mesh's vertex colors → `modeling/meshops.SrgbToLinearNode`
  (`Mesh`, `Attribute: "Color"`) after `ApplyColorFieldNode` or the
  gradient subgraph, if the colors that went in were sRGB picks.

Don't do both on the same color (a colored field converted per
`WithColorNode` *and* again on the mesh) — it double-darkens. White,
black and greys near either end survive unconverted; saturated mid-tones
are where the difference is large.
