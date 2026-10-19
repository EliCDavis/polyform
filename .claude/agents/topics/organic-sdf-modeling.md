# Organic/SDF modeling

Read this when a part is an organic form (something that should read as
one continuous soft body, not parts glued together) or needs geometry more
complex than a primitive can express directly. Deciding mechanical-vs-organic
wrong at the decompose step is the most common way an organic build goes
wrong — a soft body that comes out as a sphere plus a cylinder plus two
cones is the mechanical-assembly decomposition applied to something that
needed the organic one. The organic one keeps the parts — each is still
a subgraph — but each outputs a `Field` of overlapping `math/sdf`
primitives, and the body unions every part's field with its own and
marches the whole thing **once**. What must be singular is the march,
not the subgraph: neither one giant `body` scope nor separately marched
parts stitched together with `ModelNode` transforms. The exception is
anything with a different finish. A march is one mesh with one material,
so those are their own meshes, seated into the body rather than blended
into it (the outline's `material:`).

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
| `math/sdf.SmoothIntersectionNode` | Fields[], Radius | **Intersection** |
| `math/sdf.SubtractionNode` | A, B | **Subtract** |
| `math/sdf.SmoothSubtractionNode` | A, B, Radius | **Subtract** |
| `math/sdf.TranslateNode` | Field, Position | **Result** |
| `math/sdf.TransformNode` | Field, Transform | **Result** |
| `math/sdf.RepeatNode` | Field, Transforms, Radius | **Result** |
| `math/sdf.MirrorNode` | Field, Union | X, Y, Z, XY, XZ, YZ, XYZ |
| `math/sdf.DisplaceNode` | Primitive, Displacement | Field |
| `math/sdf.GrowNode` | Field, Distance | **Grow** |
| `math/sdf.ShellNode` | Field, Thickness | **Shell** |
| `modeling/marching.MarchNode` | Field, Domain, Resolution, Surface | **Mesh** |
| `modeling/meshops.SmoothNormalsNode` | Mesh | Out |
| `math/sdf.RaycastNode` | Field, Origin, Direction, `Max Distance` | Point, Normal, Hit, Distance |

The output names are the trap: field nodes output `Field`, but the
combinators do not — union outputs `Union`, subtraction `Subtract`,
intersection `Intersection`, and the transforms `Result`. `MarchNode`
outputs `Mesh`, and every `meshops` node outputs `Out`. `MirrorNode` has
seven outputs, one per mirror-plane combination; pick the axis you want.

A layer that hugs another field's surface is `GrowNode` on that field,
cut to shape with planes. `ShellNode` gives a hollow wall.

**Also on this list and covered in its own section below**: `LinesNode`
(a chain of capsules through an array of points at one shared radius)
and `VaryingRadiusLinesNode` (the same chain with a parallel `Radii`
array so it can taper). `CutSphereNode` is a sphere with a flat cap.

For anything that should read as a smoothly curving line getting
thinner along its length, **don't wire
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
segments between consecutive points in `Points`. A curving tapered part
needs that `Points` array **densely resampled along a curve**, not just
the 4-6 points you'd naturally place by hand to rough out its shape —
feeding those few points straight to `Points` produces a visibly
straight-segmented, faceted chain, not a smooth one (worst exactly where
the curve bends sharply). This applies regardless
of where the points come from — a handful of hardcoded literal points and
a short posable-variable array fail the exact same way.

`create_tapered_curve_subgraph` does the whole thing — resample and
taper — in one call: `create_tapered_curve_subgraph({"id": "taper"})`
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
generously large fixed number) -> `Distance` -> `math/curves.PositionNode`
(`Spline` + `Distance`) -> the dense point array that feeds
`VaryingRadiusLinesNode.Points`; a second `LinearNode` sharing the same
`Samples` builds the matching `Radii` for free (index *i* of one lines up
with index *i* of the other automatically, no remap needed). This is
exactly what `create_tapered_curve_subgraph` builds internally — see
`repetition-and-instancing.md` for the posable-variable framing of the
same recipe.

**Combinators**: `UnionNode`/`IntersectionNode`/`SubtractionNode` (boolean
ops — `UnionNode` is a **hard** min, a visible crease where shapes meet;
each has a `Smooth…` sibling with a `Radius`, so a domed box is a box
`SmoothIntersectionNode`-ed with a big sphere, not a hard lens edge)
and **`SmoothUnionNode`**, a separate node (`Fields` + `Radius`, default
`0.1`) that blends fields into each other instead of creasing — this is
the one to reach for almost any time two organic parts meet, not
`UnionNode` plus generous overlap. `Radius` is the width of the blend region in world units,
roughly proportional to the size of the *smaller* of the two fields being
blended — too small and it barely differs from a hard union (visible seam
survives), too large and it inflates/rounds off the surface where they
meet into a blob that swallows the shapes' own form. Use `UnionNode`
deliberately where a crisp seam is actually correct (most mechanical
assemblies, or an organic feature meant to read as separate),
`SmoothUnionNode` for everything meant to read as one continuous
body. Even with a smooth union, primitives still need to **overlap
generously** (pushed well into each other, not just touching) —
the blend only activates within `Radius` of an actual overlap, it doesn't
substitute for one.

`MirrorNode` (7 output ports, one per axis/plane combination — wire only
the one you need). **With `Union` at its default `true`, its output
already is the union of the original and the reflection** — e.g. for a single part built on the `+X` side, the `X`
output port alone gives you both copies; don't also wire the un-mirrored
original into a separate `UnionNode`/`SmoothUnionNode` alongside it, that
just re-adds a copy of what's already there for free. `Union` false
outputs only the reflection, without the original.

`RepeatNode` (place copies of the field at a list of transforms and union
them — see `repetition-and-instancing.md`), `TranslateNode`/
`TransformNode` (move, or move+rotate+scale, a field before combining it
with others).

## Unioning many primitives at once

Grouping fields into sub-clusters by region before a final union is
still often worth doing for a different reason: it lets each region use
a blend `Radius` appropriate to its own scale — a large region wants a
wider blend than a small one — rather than one compromise value for the
whole body. That's an organizational choice for blend quality, not a
workaround for a bug.

A mass that should read as one continuous volume is a row of fat
primitives whose cross-sections *agree* where they overlap: three
spheres of similar radius, or a capsule, each pushed well into the
next. It is not a thin tapered core with round masses hung on its ends:
the core is narrower than the masses everywhere between them, so the
union dips there and the mass gets a waist it was never given. A
`RoundedCone` core between two spheres is exactly that construction.

A small `RoundedCone` reads by its base-radius-to-length ratio: at or
above ~0.8 it is a soft puff, at or below ~0.3 it is a spike — there is
no in-between that reads as soft.

## Field to mesh

An SDF node produces a distance *field*, not a mesh. Feed the final
combined field into `modeling/marching.MarchNode` (`Resolution` high
enough that the surface doesn't look blocky — low single digits reads
chunky/voxel-y), then run *that* mesh through
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.SmoothNormalsNode]`
(`Mesh` → `Out`) before using it downstream. `MarchNode` writes no
normals, so its raw output shades faceted; the march already shares
vertices between neighbouring triangles, so this plain node gives smooth
normals. Don't use `SmoothNormalsImplicitWeldNode` here: it costs about a
quarter of the graph's evaluation time and smears normals across
subtraction cut edges.

**Be generous with `Domain` — a too-small one silently clips the surface
with no error, not a crash.** Any part that reaches further than the
main mass's own bounds gets a flat, cut-off end the
moment it exceeds the `Domain` AABB — nothing in the tool chain warns
you, it just looks wrong in the render. Marching is cheap relative to the
cost of debugging a mysteriously truncated part, so don't tightly
eyeball the domain to the geometry you *think* you built: estimate the
whole assembly's extent and pad it generously, roughly 3x on each axis,
especially on any axis a posable/variable part can swing into. Oversizing costs a few idle voxels; undersizing costs a
silently broken part.

A `Domain` that should track the part's size variables rather than be a
typed literal is `math/geometry/aabb.NewNode` (`Center`, `Size` — the
**full** size, unlike a literal AABB's half-extent `extents`), fed from
the same variables the geometry uses; `aabb.ExpandNode` pads it, and
`aabb.FromPointsNode` bounds a pose's point list.

**A `Resolution` that looks fine on the body can still be too coarse for
small, thin, or pointed features on the same field**: marching-cubes
voxel quantization can land the two sides
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

A measurement that a part should *keep* — where one part rests on a
domed surface, the height of the surface under a part, the spot a small
part sits on a curved surface — belongs in the graph as a `math/sdf.RaycastNode` wired from the body's
field, so it follows every upstream change. `Origin` and `Direction`
take one vector or an array, so a row of attachment points is one node.
`Point` and `Normal` are zero where `Hit` is false. Keep `raycast_field`
and `sample_field` for one-off inspection; a literal typed from their
answer does not follow the body when it moves.

A separate mesh meant to show on a marched body can be built correctly
and still render as nothing: the body's surface is in front of it.
Before a render, `sample_field` the body's final shape at the mesh's
outermost world point — negative means buried, and no render angle will
show it. Re-run that sample whenever the body field changes near an
attached mesh: the march's triangle count and fingerprint prove the
*body* is unchanged, they say nothing about an attached mesh that a
thickened region of the body has now swallowed.

## Primitives are stock, not answers

The roster above is vocabulary, not a menu to pick a part off. Asking
"which primitive *is* this part" has no good answer and gets you the closest
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
squashed sphere are the same primitive, and every part that is thin,
tapered and round-edged is the same *form*, so they are all the same
construction with different numbers.

**Then fix the proportions with a transform.** `TransformNode` with a
non-uniform `Scale` is a first-class modeling operation, not a
workaround: it is how a primitive whose topology is right but whose
proportions are wrong becomes the shape you described. Reaching for a
different primitive because the first one "isn't the right shape" is the
mistake — none of them are, and scaling is the step you skipped.

## The thinnest feature sets the budget for every softening length

Several parameters across the pipeline are all secretly the same thing —
a distance over which the model gets blurred:

| parameter | blurs |
| --- | --- |
| `RoundCubeNode.Roundness` | edges, by inflating outward |
| `SmoothUnionNode.Radius` | the joint between two fields |
| the marching voxel size (`1 / Resolution`, world units — `Domain` doesn't change it) | everything, by sampling |

**Every one of them has to be smaller than the thinnest feature it
touches**, and they are all sized by the *thinnest* part of the model, not
the average one. Above that threshold each fails in its own way and none
of them error: rounding swells a thin panel until it is a lump, a smooth
union bridges across a gap that was supposed to stay open, and a voxel coarser than the feature drops it or speckles it.

This is why thin parts are the hard case, and why "it looked fine on the
large mass but wrong on the thin part" keeps happening: the large mass
tolerates a generous blur and the thin part does not. When a model has
both, every one of these is sized by the thin part.

So when something thin comes out chunky, blobby, or oddly lit, check
these three before changing the geometry — the shape is usually right and
one of the blur lengths is too big for it.

You don't have to size these by hand: `set_outline` computes them from
the plan's leaf thicknesses and returns `minResolution` (for
`MarchNode.Resolution`) and `maxBlend` (the ceiling for the other
two), and `check_outline` flags a march or smooth union that breaks
them.

## Ask the question in the form that answers it

A render is the only way to judge *form*, and it is the worst way to
answer anything factual. Two failures cost real time on every build so
far, and neither is something a picture shows well:

- **a part that isn't there.** Nothing looks like nothing, and among
  fifteen parts that did appear, one that didn't is easy to miss.
- **a part that didn't attach.** A small part floating a fraction off its
  host looks correct from every angle that isn't exactly edge-on.

`describe_mesh` answers both outright. It reports the triangle count, the
bounding box, and **how many separate connected pieces the mesh is in** —
so a part that failed to wire doesn't move the triangle count, and a part
that failed to union is its own piece, listed with its size and centre so
you know *which* one came adrift. A body that shares one march and is
meant to be one body should report exactly 1 piece. Re-check after every
`SubtractionNode` groove or crease: a cut that runs through a valley
between two blended parts can sever a sliver on the far side that no
render angle shows.

It also settles proportion, which is the defect that survives every other
check — parts all present, all attached, all correctly coloured, and the
thing still doesn't read right. State the ratio you intended, then read
the bounding box: 3:1 when you wanted 15:1 is a number, not a judgement.

Reach for it whenever the question is factual rather than aesthetic. It
costs the same evaluation as a render but skips the rasterizing and the
image, so it is strictly cheaper than the render it replaces — and unlike
the render, it cannot be misread.

Two siblings for questions that aren't about a mesh: `evaluate_node`
reads any port's current value as JSON (a posed part's world point, a
center after rotation, a spline length) — never re-derive such
a number by hand from the parameters; and `describe_manifest` gives
per-model triangle counts and the total for an export, which is how to
compare a low-poly manifest against the full one without probe renders.
All of `sample_field`/`raycast_field`/`describe_mesh`/`evaluate_node`
take `instanceId` as a path (`outer/inner`) to reach a part nested
inside another part.

## A body with parts grown out of it is one field, not several marched separately

The most common way this goes wrong even after correctly deciding a
subject is organic: building the main mass as its own subgraph (SmoothUnion
→ March → Mesh, all internal), then building each attached part the
*same self-contained way* — its own SmoothUnion → March → Mesh — and
attaching them to the main mass only via a `ModelNode` `Translation`. Each
part is individually a valid, smooth, organic mesh. The result is still
wrong: every joint is two independently-marched surfaces forced to
overlap, which shows as a hard seam and a flat-colored patch up close
(especially if the body has a vertex-color gradient the separately-marched
part never picked up), no matter how precisely the translation is
computed. This is a structural problem, not a placement-precision one —
no amount of retuning the translation vector fixes it, because there is
no math that makes two separately-solved surfaces meet without a crease.

The fix is the interface: a part meant to grow out of another one has
**one boundary output, `Field`** (wired from the part's own internal
`SmoothUnionNode`, `type` omitted so it takes that port's) and no march of its own.
At the point where the parts come together, wrap each part's field in an
`sdf.TranslateNode`/`TransformNode` (the placement you'd otherwise have
put on a `ModelNode`), combine all of them — body plus every
attached part's field — in **one** top-level `SmoothUnionNode`, and march
**that** once, with a domain expanded to cover the whole assembled shape.
This guarantees a genuinely seamless join at every joint, structurally —
the surface is solved as one continuous field instead of assembled from
independently-marched pieces, so there's no coordinate math that can
leave a gap or a crease. If you inherit a part that only exposes `Mesh`,
add the `Field` output with `create_boundary_node` and leave the `Mesh`
port in place; don't restructure the subgraph.

Apply this to *every* organic part meant to read as grown-from-the-body,
not just whichever one happens to show a visible seam first; fixing one
part this way and leaving another as an independent mesh-plus-translation
reproduces the exact same bug on the one you skipped. Reserve
independent-mesh-plus-translation for parts that are organic in *shape*
but meant to read as separate objects, not grown from the body — anything
that visibly sits *on* the surface rather than merging into it.

If the body has a vertex-color gradient keyed on a raw axis, building it with
`create_vertex_color_gradient_subgraph` (see "Coloring a marched mesh"
below) sidesteps this entirely, since it derives the gradient's range
from the mesh's own current extent every time rather than a range fixed
when the gradient was first wired. A hand-wired gradient with a
hard-coded `In Min`/`In Max` doesn't get that for free — check it after
merging parts in, since a range that assumed only the main mass's own
bounds will clip newly-added parts to a solid color instead of shading them.

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
common two-color gradient case — one smooth fade along a single world
axis, nothing more — use `create_vertex_color_gradient_subgraph`:
one call replaces the whole select/remap/interpolate chain, and it
derives the gradient's range from the mesh's own actual extent
automatically, so it can't clip newly-added geometry (a part added after
the gradient was tuned) to a solid color the way a hand-picked range can.
See `texturing-and-color.md` for the fully manual version of that same
recipe.

If different *parts* need their own distinct color — not a fade, actual
per-region coloring (a patch/marking pattern, any case where you'd
describe the result as "this part is colored X" rather than "the color
fades from X to Y") — a single global gradient can't express that, no
matter how the axis or range is tuned. Use colored fields instead, from
`math/sdf`:

1. Build the body the same way as always — each region as one
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

### The cheaper route for markings on an existing body: a separate paint field

`ApplyColorFieldNode` reads **only the color half** of the colored field
it is given, at each vertex of the mesh it is given. Nothing requires
that field's *shape* to be the body. So an already-built plain-SDF body
(union → march → smooth normals) keeps its whole tree untouched, and the markings
come from a second, color-only field:

1. `WithColorNode(Field: <the body's final union or transform output>,
   Color: <base color>)` — the base color. Its distance is ~0 at every
   vertex, which is what makes the rest work.
2. One `WithColorNode` per marking, wrapping a cheap blob (a sphere, a
   squashed sphere, a box half-space, a `MirrorNode` of one of those) that
   *encloses the surface region* to be recolored. A vertex inside the blob
   sees a negative distance there, beats the base's ~0, and takes the
   blob's color. A marking confined to one part is a box half-space
   intersected with that part's own primitive, in the part's local frame,
   then pushed through the part's own transform. Each marking is 1-3 nodes.
3. `SmoothUnionColoredNode(Fields: [base, markings...], Radius: r)`, then
   `ApplyColorFieldNode(Mesh: <the smoothed mesh>, Field: that union)` in
   place of the gradient subgraph. `Radius` is the fade width at each
   blob's edge (0.02 on a ~0.4-unit model gave soft-but-clear edges; 0.03
   was mushy).

Two rules of that blend to know before placing blobs. **Deepest wins**:
the color comes from the field the vertex is *most inside*, so a small
blob placed inside a big one never shows — `SubtractionNode` the small
blob out of the big one first so the big one is no longer "inside" there,
or make the small one deeper than the big one is at that spot. A blob
only ever paints vertices, so a marking narrower than about two voxels
(`2 / Resolution`) can't be painted at all — a 0.002-wide line on a
300-resolution march is a separate mesh, not paint. **Only the surface
matters**: a blob may extend far inside the body with no effect, so size
it to the patch of *surface* it should cover, and check the far end: a
blob that stops short of a protruding tip leaves the tip unpainted.

Blobs live in the same space as the vertices, so a blob for a part that
is moved by a transform goes through a copy of that same transform before
its `WithColorNode`, and through any whole-body transform after it, or it
drifts off the part when the variable changes.

## Reflection and grouped rigid transforms

`sdf.TransformNode` treats a negative scale as a clean reflection (the
distance factor is the absolute scale), so `scale {-1,1,1}` on one part's
sub-union is its mirrored copy — the way to turn a mirrored pair into two
independently posable parts without duplicating the part's node tree:
`UnionNode(Transform(part, reflect), Transform(part, rotate-about-pivot))`
in place of the `MirrorNode`.

To rotate a part about a pivot `P`: `trs.PivotNode(Pivot: P, Rotation, Scale?)` — it works
out the `P - R·S·P` position itself, so a joint is two nodes
(`quaternion.FromEulerAngleNode` → `PivotNode`) and the angle stays a
live variable. Sign
convention for `FromEulerAngleNode`: a positive X angle turns +Y toward
+Z, so a part hanging down (-Y) swings toward +Z with a **negative** X
angle. A chain of joints is one such pivot per joint composed with
`trs.MultiplyNode` (A = parent, B = child) into one TRS, then that one
TRS applied to the field, its paint blobs and its attachments — not a
nested `TransformNode` per joint on each of them. The other side's
pivots are the same literals through `vector3.Multiply(B: (-1,1,1))`,
never hand-negated copies. A part that should face another part is
`trs.LookAtNode(Position, Target)` — local +Z aims at the target with +Y
kept up — not a hand-derived yaw angle that goes stale when either
position moves.

Separate meshes attached to that part (anything placed by a surface
point) follow it for free if they are `Children` of
one mesh-less `gltf.ModelNode` carrying the same `Translation`/`Rotation`
(the previewer and the glTF writer both accept a `ModelNode` with no
`Mesh` as a pure group). Build the same `PivotNode` TRS at root from the
same variable, and place any new attachment point with an
`sdf.RaycastNode` on the *pre*-transform sub-union (that part's own union,
not the whole body) so the point is in the group's local frame. Anything
outside the group that needs to know where a posed point ended up runs
the local point through
`trs.TransformPointNode(TRS: that same pivot TRS, Point)` — one node,
and it follows the variable.

A separate part placed near a blended one bridges to it inside the blend
radius: two surfaces closer than `SmoothUnionNode.Radius` merge, so two
parts 0.01 apart on a 0.03 blend read as one lump. Keep the
gap above the radius, or `sample_field` the distance before trusting the
render.

A primitive scaled to zero inside a `SmoothUnionNode` is not a no-op: it
still takes part in the blend wherever its distance is within `Radius` of
the nearest other field's, so a pose part that
must leave the rest pose byte-identical goes through its own union whose
`Radius` is also scaled by the pose (a hard min at 0), not into the
shared body union. The other exact way: collapse the pose-only primitive
*onto* an existing one at pose 0 (a capsule whose endpoints both sit on
an existing capsule's axis, with that capsule's radius) — `min` then
returns the original field
everywhere by the triangle inequality, and no gate is needed.
