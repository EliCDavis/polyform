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
chain of capsules through an array of points, one shared radius — the
tool for a snake/tentacle/tail/rope/chain body; see
`repetition-and-instancing.md` for the point-array variable + curve-bend
recipe).

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
the one you need), `RepeatNode` (place copies of the field at a list of
transforms and union them — see `repetition-and-instancing.md`),
`TranslateNode`/`TransformNode` (move, or move+rotate+scale, a field
before combining it with others).

## Unioning many primitives at once

Grouping fields into anatomical sub-clusters (torso spheres, then
head+ears, then muzzle) before a final union is still often worth doing
for a different reason: it lets each region use a blend `Radius`
appropriate to its own scale — a torso wants a wider blend than an ear —
rather than one compromise value for the whole body. That's an
organizational choice for blend quality, not a workaround for a bug.

## Field to mesh

An SDF node produces a distance *field*, not a mesh. Feed the final
combined field into `modeling/marching.MarchNode` (`Domain`: an `AABB`
covering the whole shape with margin; `Resolution` high enough that the
surface doesn't look blocky — low single digits reads chunky/voxel-y),
then run *that* mesh through
`nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.SmoothNormalsImplicitWeldNode]`
(`Distance` set explicitly — its own default, `0.0001`, is too small for
real-world scale) before using it downstream. Use this node's output, not
the raw `MarchNode` output — it recomputes smooth per-vertex normals
across nearby vertices, which is what actually makes the surface read as
smooth under lighting, cheaper than raising `Resolution` further.

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
cream-belly-to-coat-color fade), check it after merging limbs in: a
gradient clamped to a Y range that assumed only the torso's own bounds
will clip newly-added legs to a solid color instead of shading them. Use
distance from the reference plane/line instead of the raw coordinate if
parts now extend well outside the original range, so the color fades
correctly in both directions instead of clamping.

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
nothing to map onto — a marched mesh needs vertex color instead. See
`texturing-and-color.md` for the recipe.
