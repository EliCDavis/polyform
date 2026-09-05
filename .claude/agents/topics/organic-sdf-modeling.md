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

## The roster — exact type keys, do not search for these

Every key below is the `PATH` inside the wrapper
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]` — so `math/sdf.SphereNode`
means the type key
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]`.
Port names are exact. **These are confirmed against the registry: use them
directly, do not spend a `search_node_types`/`get_node_types` pair
rediscovering them.** Search only for something genuinely not on this list.

| PATH | inputs | outputs |
| --- | --- | --- |
| `math/sdf.SphereNode` | Position, Radius | Field |
| `math/sdf.CubeNode` | Position, Size | Field |
| `math/sdf.RoundCubeNode` | Position, Size, Roundness | Field |
| `math/sdf.RoundedConeNode` | A, B, `Radius 1`, `Radius 2` | Field |
| `math/sdf.RoundedCylinderNode` | Position, Radius, `Body Height`, `Rounding Radius` | Field |
| `math/sdf.TorusNode` | Position, `Ring Radius`, `Tube Radius` | Field |
| `math/sdf.LineNode` | Start, End, Radius | Field |
| `math/sdf.PlaneNode` | Position, Normal, Height | Field |
| `math/sdf.UnionNode` | Fields[] | **Union** |
| `math/sdf.SmoothUnionNode` | Fields[], Radius | **Union** |
| `math/sdf.IntersectionNode` | Fields[] | **Intersection** |
| `math/sdf.SubtractionNode` | A, B | **Subtract** |
| `math/sdf.SmoothSubtractionNode` | A, B, Radius | **Subtract** |
| `math/sdf.TranslateNode` | Field, Position | **Result** |
| `math/sdf.TransformNode` | Field, Transform | **Result** |
| `math/sdf.RepeatNode` | Field, Transforms, Radius | **Result** |
| `math/sdf.MirrorNode` | Field, Union | X, Y, Z, XY, XZ, YZ, XYZ |
| `math/sdf.DisplaceNode` | Primitive, Displacement | Field |
| `modeling/marching.MarchNode` | Field, Domain, Resolution, Surface | **Mesh** |
| `modeling/meshops.SmoothNormalsImplicitWeldNode` | Mesh, Distance | Out |

The output names are the trap: field nodes output `Field`, but the
combinators do not — union outputs `Union`, subtraction `Subtract`,
intersection `Intersection`, and the transforms `Result`. `MarchNode`
outputs `Mesh`, and every `meshops` node outputs `Out`. `MirrorNode` has
seven outputs, one per mirror-plane combination; pick the axis you want.

**Also on this list and covered in its own section below**: `LinesNode`
(a chain of capsules through an array of points at one shared radius — a
rope, cable, or chain link) and `VaryingRadiusLinesNode` (the same chain
with a parallel `Radii` array so it can taper). `CutSphereNode` is a
sphere with a flat cap.

For a tail, tentacle, horn, or anything else that should read as a
smoothly curving line getting thinner along its length, **don't wire
`VaryingRadiusLinesNode` directly — use `create_tapered_curve_subgraph`**
(next section). Don't hand-build a tapering chain out of individual
`RoundedConeNode`s glued together with `SmoothUnionNode` either:
consecutive segments in a real chain already share an exact matching
endpoint and radius, so smooth-unioning them adds unwanted extra blending
at every joint — a visible lump at each seam, worst at a sharp curl —
instead of the clean taper a plain union of matching segments gives for
free.

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
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.SmoothNormalsImplicitWeldNode]`
(`Distance` set explicitly — its own default, `0.0001`, is too small for
real-world scale) before using it downstream. Use this node's output, not
the raw `MarchNode` output — it recomputes smooth per-vertex normals
across nearby vertices, which is what actually makes the surface read as
smooth under lighting, cheaper than raising `Resolution` further.

`Distance` is one of the softening lengths (see the budget section
below): at or above a feature's thickness the weld averages that
feature's front and back normals together and it shades like a rounded
lump — geometry right, lighting wrong, which reads as a modeling failure
and sends you back to the SDF looking for a problem that isn't there.
Everything sharing one march shares one weld, so the value is set by the
thinnest thing in the mesh, not the bulk of it.

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

## Finding the surface of a field

A field is queryable, which means you never have to estimate anything
about it. Two tools, and the difference matters:

- `sample_field` answers **"what is here"** — inside, outside, how far
  from the surface. Use it to check a suspicion about a region.
- `raycast_field` answers **"where is the surface, and which way does it
  face"** — fire a ray, get back the point and the outward normal. Use it
  whenever a number describes a *relationship* to this body: where
  something attaches, how wide a part actually came out, where a cavity
  wall sits.

Both take many queries per call, so a whole row of attachment points, or
the extents of a part along all three axes, is one call rather than one
per number.

March **outward from inside** to find a cavity wall, **inward from
outside** to find the skin. This is the general case of the positioning
rule in the orchestrator instructions — a blended body has no formula to
solve, so it gets measured instead.

## Primitives are stock, not answers

The roster above is vocabulary, not a menu to pick a part off. Asking
"which primitive *is* a fin" has no good answer and gets you the closest
box; the question is "what form is this, and how do I get there from raw
stock".

**Describe the form before you reach for a node.** Three things, in
ordinary words, none of which mention a primitive:

- its **proportions** — the ratios, not the sizes. "About fifteen times
  longer than it is thick" is a fact you can check later; "smallish" isn't.
- its **cross-section** — round, flat, square, hollow.
- **how that cross-section changes** along its length — constant,
  tapering, swelling in the middle, splitting.

That description is what you match a primitive to, and you match on
**topology, not name**: something that tapers between two ends is a
rounded cone whatever it depicts; something with a hole is a torus or
tube; something with genuinely hard corners is a box. A sphere and a
squashed sphere are the same primitive, and a fin, a leaf, a fluke and a
gill cover are all the same *form* — thin, tapered, round-edged — so they
are all the same construction with different numbers.

**Then fix the proportions with a transform.** `TransformNode` with a
non-uniform `Scale` is a first-class modeling operation, not a
workaround: it is how a primitive whose topology is right but whose
proportions are wrong becomes the shape you described. Reaching for a
different primitive because the first one "isn't fin-shaped" is the
mistake — none of them are, and scaling is the step you skipped.

(Non-uniform `Scale` on a field was broken until recently: it returned the
field's own local distances, which marched as speckle. If you have some
recollection that squashing an SDF doesn't work, that was why, and it is
fixed.)

## The thinnest feature sets the budget for every softening length

Several parameters across the pipeline are all secretly the same thing —
a distance over which the model gets blurred:

| parameter | blurs |
| --- | --- |
| `RoundCubeNode.Roundness` | edges, by inflating outward |
| `SmoothUnionNode.Radius` | the joint between two fields |
| `SmoothNormalsImplicitWeldNode.Distance` | normals across nearby surfaces |
| the marching voxel size (`Domain` extent / `Resolution`) | everything, by sampling |

**Every one of them has to be smaller than the thinnest feature it
touches**, and they are all sized by the *thinnest* part of the model, not
the average one. Above that threshold each fails in its own way and none
of them error: rounding swells a thin panel until it is a lump, a smooth
union bridges across a gap that was supposed to stay open, welding
averages a thin blade's two faces together so it shades like a tube, and
a voxel coarser than the feature drops it or speckles it.

This is why thin parts are the hard case, and why "it looked fine on the
body but wrong on the fins" keeps happening: the body tolerates a
generous blur and the fins do not. When a model has both, every one of
these is sized by the fins.

So when something thin comes out chunky, blobby, or oddly lit, check
these four before changing the geometry — the shape is usually right and
one of the blur lengths is too big for it.

## Ask the question in the form that answers it

A render is the only way to judge *form*, and it is the worst way to
answer anything factual. Two failures cost real time on every build so
far, and neither is something a picture shows well:

- **a part that isn't there.** Nothing looks like nothing, and among
  fifteen parts that did appear, one that didn't is easy to miss.
- **a part that didn't attach.** A tooth floating a hair off a jaw looks
  correct from every angle that isn't exactly edge-on.

`describe_mesh` answers both outright. It reports the triangle count, the
bounding box, and **how many separate connected pieces the mesh is in** —
so a part that failed to wire doesn't move the triangle count, and a part
that failed to union is its own piece, listed with its size and centre so
you know *which* one came adrift. A body that shares one march and is
meant to be one body should report exactly 1 piece.

It also settles proportion, which is the defect that survives every other
check — parts all present, all attached, all correctly coloured, and the
thing still doesn't read right. State the ratio you intended, then read
the bounding box: 3:1 when you wanted 15:1 is a number, not a judgement.

Reach for it whenever the question is factual rather than aesthetic. It
costs the same evaluation as a render but skips the rasterizing and the
image, so it is strictly cheaper than the render it replaces — and unlike
the render, it cannot be misread.

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
time. Unlike the render-based techniques below, `sample_field` is
available no matter which agent is reading this (the orchestrator and a
spawned `polyform-part-builder` both have it) — the part-builder has no
`render_preview` at all, so this is its primary way to verify a field
numerically rather than guessing. It evaluates any `math/sdf` node's
`Field` output (a primitive, or the combined result of a
`Union`/`SmoothUnionNode`/etc.) at explicit world-space points and returns
the raw signed distance instantly — no marching, no rasterizing, no image
to read. Negative means inside, positive means outside, zero is exactly on
the surface. Compare a single field's value against the whole union's
value at the same point to check whether that field is really the one
determining the surface there; check a point that should be in open space
(e.g. the midpoint between two clusters that shouldn't touch) for an
unexpectedly negative value instead of rendering and squinting at the
image. Faster and more precise than a render for anything where the
question is really "what is the field's value here," not "what does this
look like."

If you have `render_preview` (the orchestrator does), it also takes an
`exclude` list of node ids — parts left out of that one render only,
without touching the graph. Use it instead of disconnecting a suspected
part, rendering, and reconnecting it: pass the node id, compare against a
render with nothing excluded, and the graph is never at risk of being left
in a broken state if something interrupts the investigation partway
through.

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
