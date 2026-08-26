# Organic/SDF modeling

Read this when a part is an organic form (an animal, creature, plant,
character — something that should read as one continuous soft body, not
parts glued together) or needs geometry more complex than a primitive can
express directly. Deciding mechanical-vs-organic wrong at the decompose
step is the most common way an organic build goes wrong — "build a cat"
that comes out as a sphere-head-plus-cylinder-body-plus-cone-ears is the
mechanical-assembly decomposition applied to something that needed the
organic one: one subgraph (or one per continuous body region) built as
overlapping `math/sdf` primitives combined with a union node, marched into
a single mesh — not separate subgraphs stitched together with `ModelNode`
transforms.

Every node in `math/sdf` has real documentation — `get_node_types` on a
candidate before wiring it rather than assuming from the name.

## The roster (`pathPrefix: "math/sdf"`)

**Primitives**: `CubeNode`/`RoundCubeNode` (box, optionally filleted),
`SphereNode`, `CutSphereNode` (sphere with a flat cap), `TorusNode`
(donut), `RoundedConeNode` (capsule/cone between two points with an
independent radius at each end — good for tapered limbs),
`RoundedCylinderNode`, `PlaneNode` (infinite — best used as a cutting tool
via `SubtractionNode`/`IntersectionNode`, not as visible geometry on its
own), `LineNode` (single capsule between two points), `LinesNode` (a
chain of capsules through an array of points, one shared radius — for a
body that stays one constant thickness end to end: a rope, a chain link,
a cable), `VaryingRadiusLinesNode` (the same chain, but a parallel `Radii`
array gives every point its own radius, so the body can taper). **For a
tail, tentacle, horn, or anything else that should read as a smoothly
curving line getting thinner along its length, don't wire
`VaryingRadiusLinesNode` directly — use the `create_tapered_curve_subgraph`
tool instead** (see the very next section). Don't hand-build a tapering
chain out of individual `RoundedConeNode`s glued together with
`SmoothUnionNode` either — consecutive segments in a real chain already
share an exact matching endpoint and radius, so smooth-unioning them adds
unwanted extra blending at every joint (a visible lump right at each
seam, worst at a sharp curl) instead of the clean taper a plain union of
matching segments already gives for free.

## Tapering a curve: `create_tapered_curve_subgraph`, not raw `VaryingRadiusLinesNode`

`VaryingRadiusLinesNode` (and plain `LinesNode`) draws **straight**
segments between consecutive points in `Points`. A tail/tentacle/horn
needs that `Points` array **densely resampled along a curve**, not just
the 4-6 points you'd naturally place by hand to rough out its shape —
feeding those few points straight to `Points` produces a visibly
straight-segmented, faceted chain, not a smooth one (worst exactly where
the curve bends sharply, e.g. a curled tail tip). This applies regardless
of where the points come from — a handful of hardcoded literal points and
a short posable-variable array fail the exact same way.

`create_tapered_curve_subgraph` does the whole thing — resample and
taper — in one call: `create_tapered_curve_subgraph({"id": "tail"})`
creates a reusable subgraph with four boundary inputs (`Points`, your few
control points; `Base Radius`; `Tip Radius`; `Samples`, a generously
large number like 12-20 — deliberately decoupled from and bigger than
your control-point count, since it's the resample density, not the pose
resolution) and one boundary output (`Field`, a ready-to-union
`math/sdf` field). `instantiate_subgraph` it, wire the four inputs (a
literal array, or a posable-variable reference, for `Points`), and wire
`Field` straight into your body's `Union`/`SmoothUnionNode` like any
other `math/sdf` node output.

Reach for the manual pipeline instead only if you need something this
tool doesn't expose (a non-default spline `Alpha`, a `Closed` loop, a
non-linear radius profile): `math/curves.CatmullRomSplineNode` (`Points`
-> a smooth `Spline`) -> `LengthNode` (`Spline` -> arc length) ->
`math/sequence.LinearNode` (`Start: 0`, `End:` the length, `Samples:` a
generously large fixed number) -> `Distances` -> `PositionsForArrayNode`
(`Spline` + `Distances`) -> the dense point array that feeds
`VaryingRadiusLinesNode.Points`; a second `LinearNode` sharing the same
`Samples` builds the matching `Radii` for free (index *i* of one lines up
with index *i* of the other automatically, no remap needed). This is
exactly what `create_tapered_curve_subgraph` builds internally — see
`repetition-and-instancing.md` for the posable-variable framing of the
same recipe.

**Combinators**: `UnionNode`/`IntersectionNode`/`SubtractionNode` (boolean
ops — `UnionNode` is a **hard** min, a visible crease where shapes meet)
and **`SmoothUnionNode`**, a separate node (`Fields` + `Radius`, default
`0.1`) that blends fields into each other instead of creasing — this is
the one to reach for almost any time two organic parts meet (a haunch
into a torso, a muzzle into a skull), not `UnionNode` plus generous
overlap. `Radius` is the width of the blend region in world units,
roughly proportional to the size of the *smaller* of the two fields being
blended — too small and it barely differs from a hard union (visible seam
survives), too large and it inflates/rounds off the surface where they
meet into a blob that swallows the shapes' own form. Use `UnionNode`
deliberately where a crisp seam is actually correct (most mechanical
assemblies, or an organic feature meant to read as separate — claws,
spines), `SmoothUnionNode` for everything meant to read as one continuous
body. Even with a smooth union, primitives still need to **overlap
generously** (a head pushed well into the torso, not just touching it) —
the blend only activates within `Radius` of an actual overlap, it doesn't
substitute for one.

`MirrorNode` (7 output ports, one per axis/plane combination — wire only
the one you need). **Its output already is the union of the original and
the reflection** — e.g. for a single ear built on the `+X` side, the `X`
output port alone gives you both ears; don't also wire the un-mirrored
original into a separate `UnionNode`/`SmoothUnionNode` alongside it, that
just re-adds a copy of what's already there for free. `Union` (bool,
defaults `true`) controls *how* it does that: true evaluates the field on
every mirrored side and combines the results, correct even if the field
already has real, distinct content on more than one side of the axis
(mirroring something other than a single one-sided limb/ear); false folds
a query onto one canonical side and evaluates the field once, cheaper but
only correct when the field is known to be one-sided along every mirrored
axis. Leave it at the default unless you've specifically profiled mirror
evaluation as a bottleneck.

`RepeatNode` (place copies of the field at a list of transforms and union
them — see `repetition-and-instancing.md`), `TranslateNode`/
`TransformNode` (move, or move+rotate+scale, a field before combining it
with others).

## Unioning many primitives at once

Grouping fields into anatomical sub-clusters (torso spheres, then
head+ears, then muzzle) before a final union is still often worth doing
for a different reason: it lets each region use a blend `Radius`
appropriate to its own scale — a torso wants a wider blend than an ear —
rather than one compromise value for the whole body. That's an
organizational choice for blend quality, not a workaround for a bug.

## Field to mesh

An SDF node produces a distance *field*, not a mesh. Feed the final
combined field into `modeling/marching.MarchNode` (`Resolution` high
enough that the surface doesn't look blocky — low single digits reads
chunky/voxel-y), then run *that* mesh through
`nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.SmoothNormalsImplicitWeldNode]`
(`Distance` set explicitly — its own default, `0.0001`, is too small for
real-world scale) before using it downstream. Use this node's output, not
the raw `MarchNode` output — it recomputes smooth per-vertex normals
across nearby vertices, which is what actually makes the surface read as
smooth under lighting, cheaper than raising `Resolution` further.

**Be generous with `Domain` — a too-small one silently clips the surface
with no error, not a crash.** A curled tail, a raised limb, anything that
reaches further than the torso's own bounds, gets a flat, cut-off end the
moment it exceeds the `Domain` AABB — nothing in the tool chain warns
you, it just looks wrong in the render. Marching is cheap relative to the
cost of debugging a mysteriously truncated part, so don't tightly
eyeball the domain to the geometry you *think* you built: estimate the
whole assembly's extent and pad it generously, roughly 3x on each axis,
especially on any axis a posable/variable part (a tail, a limb) can
swing into. Oversizing costs a few idle voxels; undersizing costs a
silently broken part.

**A `Resolution` that looks fine on the body can still be too coarse for
small, thin, or pointed features on the same field** — ears, claws,
spikes, fingers: marching-cubes voxel quantization can land the two sides
of a genuinely symmetric, mirrored field on slightly different voxel
boundaries, reading as asymmetric in the render. **Don't assume a
mirrored/symmetric part looking asymmetric means the SDF or the mirror is
wrong** — check `Resolution` relative to the *smallest* feature in the
field first, not just relative to the body's overall size.

## A body with limbs/tail/ears is one field, not several marched separately

The most common way this goes wrong even after correctly deciding a
subject is organic: building the torso as its own subgraph (SmoothUnion
→ March → Mesh, all internal), then building each leg and the tail the
*same self-contained way* — its own SmoothUnion → March → Mesh — and
attaching them to the torso only via a `ModelNode` `Translation`. Each
part is individually a valid, smooth, organic mesh. The result is still
wrong: every joint is two independently-marched surfaces forced to
overlap, which shows as a hard seam and a flat-colored patch up close
(especially if the body has a vertex-color gradient the separately-marched
limb never picked up), no matter how precisely the translation is
computed. This is a structural problem, not a placement-precision one —
no amount of retuning the translation vector fixes it, because there is
no math that makes two separately-solved surfaces meet without a crease.

The fix: give every part meant to visually grow out of another one **a
second boundary output carrying its pre-march field** (`create_boundary_node`,
kind `output`, type `math/sample.Vec3ToFloat`, wired from the subgraph's
own internal `SmoothUnionNode` — not its `March`/`Mesh` output). At the
point where the parts come together, wrap each part's field in an
`sdf.TranslateNode` (the same translation you'd otherwise have put on a
`ModelNode`), combine all of them — body plus every limb/tail/ear field —
in **one** top-level `SmoothUnionNode`, and march **that** once, with a
domain expanded to cover the whole assembled shape. This guarantees a
genuinely seamless join at every joint, structurally — the surface is
solved as one continuous field instead of assembled from independently-
marched pieces, so there's no coordinate math that can leave a gap or a
crease. Each part subgraph's `Mesh` output becomes unused once this is
wired — that's fine, leave the boundary port in place rather than
restructuring the subgraph.

Apply this to *every* organic part meant to read as grown-from-the-body —
legs, tail, ears, horns — not just whichever one happens to show a visible
seam first; fixing one limb this way and leaving another as an
independent mesh-plus-translation reproduces the exact same bug on the
one you skipped. Reserve independent-mesh-plus-translation for parts
that are organic in *shape* but meant to read as separate objects, not
grown from the body — a collar, a saddle, anything that visibly sits *on*
the surface rather than merging into it.

If the body has a vertex-color gradient keyed on a raw axis (e.g. Y for a
cream-belly-to-coat-color fade), building it with
`create_vertex_color_gradient_subgraph` (see "Coloring a marched mesh"
below) sidesteps this entirely, since it derives the gradient's range
from the mesh's own current extent every time rather than a range fixed
when the gradient was first wired. A hand-wired gradient with a
hard-coded `In Min`/`In Max` doesn't get that for free — check it after
merging limbs in, since a range that assumed only the torso's own bounds
will clip newly-added legs to a solid color instead of shading them.

## Debugging a field without rendering

When something about a union looks wrong and you need to isolate *why* —
which primitive is responsible, whether a suspected point is really
inside the merged field — reach for `sample_field` before reaching for
`render_preview` in a loop or disconnecting nodes to test them one at a
time. It evaluates any `math/sdf` node's `Field` output (a primitive, or
the combined result of a `Union`/`SmoothUnionNode`/etc.) at explicit
world-space points and returns the raw signed distance instantly — no
marching, no rasterizing, no image to read. Negative means inside,
positive means outside, zero is exactly on the surface. Compare a single
field's value against the whole union's value at the same point to check
whether that field is really the one determining the surface there; check
a point that should be in open space (e.g. the midpoint between two
clusters that shouldn't touch) for an unexpectedly negative value instead
of rendering and squinting at the image. Faster and more precise than a
render for anything where the question is really "what is the field's
value here," not "what does this look like."

`render_preview` also takes an `exclude` list of node ids — parts left
out of that one render only, without touching the graph. Use it instead
of disconnecting a suspected part, rendering, and reconnecting it: pass
the node id, compare against a render with nothing excluded, and the
graph is never at risk of being left in a broken state if something
interrupts the investigation partway through.

## Coloring a marched mesh

`MarchNode` never generates UVs, so `gltf.MaterialNode.ColorTexture` has
nothing to map onto — a marched mesh needs vertex color instead. For the
common two-color gradient case (a cream belly fading into a coat color, a
dark spine fading into a light belly) — one smooth fade along a single
world axis, nothing more — use `create_vertex_color_gradient_subgraph`:
one call replaces the whole select/remap/interpolate chain, and it
derives the gradient's range from the mesh's own actual extent
automatically, so it can't clip newly-added geometry (a leg added after
the gradient was tuned) to a solid color the way a hand-picked range can.
See `texturing-and-color.md` for the fully manual version of that same
recipe.

If different *parts* need their own distinct color — not a fade, actual
per-region coloring (white socks on a cat's paws, a dark mask around the
eyes, a patch/marking pattern, any case where you'd describe the result
as "this part is colored X" rather than "the color fades from X to Y") —
a single global gradient can't express that, no matter how the axis or
range is tuned. Use colored fields instead, from `math/sdf`:

1. Build the body the same way as always — each anatomical region as one
   or more overlapping primitives — but wrap **every leaf primitive that
   needs its own distinct color** in `WithColorNode` (`Field` <- the
   primitive's own `Field` output, `Color` <- that region's color)
   *before* unioning it in. A primitive that should just blend into its
   neighbor's color doesn't need its own wrap — share the neighbor's
   `WithColorNode` output, or leave it out of the colored side of the
   tree entirely if it's plain body color.
2. Union the `ColoredField`s with `SmoothUnionColoredNode`/
   `UnionColoredNode` instead of `SmoothUnionNode`/`UnionNode` — same
   `Fields`(`/Radius`) shape as their plain-field counterparts, but they
   blend (or don't) color the same way they blend shape. Use
   `UnionColoredNode` deliberately for a crisp-edged marking (a hard seam
   between two colors, no fade); `SmoothUnionColoredNode` blends color
   using the *identical* blend weight it uses for the shape, so a soft
   marking transitions over exactly the same region the geometry seam
   does — the same choice `UnionNode` vs `SmoothUnionNode` already is for
   geometry, just carried over to color. Nest these the same way you'd
   nest ordinary `Union`/`SmoothUnionNode` calls for per-region blend
   radii (see "Unioning many primitives at once" above) — either
   combined-color node's own output is just another `ColoredField`, so it
   composes into a bigger union the same way, and the two kinds can mix
   (a `UnionColoredNode` feeding into a `SmoothUnionColoredNode`, or vice
   versa).
3. Once every part is combined into one final `ColoredField`, split it in
   two: `ColoredFieldDistanceNode` extracts the plain distance half — wire
   that into `MarchNode` exactly as before, nothing about marching
   changes. Separately, wire the *same* final `ColoredField` (not the
   distance-only extraction) into `modeling/marching.ApplyColorFieldNode`
   alongside `MarchNode`'s `Mesh` output — it samples the color half at
   each real vertex position and writes `"Color"` directly. This replaces
   `create_vertex_color_gradient_subgraph`/the manual select-remap-
   interpolate chain entirely for this mesh; don't also run a gradient
   pass on top; the color is already baked in per region.

`get_node_types` on `WithColorNode`/`UnionColoredNode`/
`SmoothUnionColoredNode`/`ColoredFieldDistanceNode` (`math/sdf`) and
`ApplyColorFieldNode` (`modeling/marching`) for exact ports — don't guess
the wiring.
