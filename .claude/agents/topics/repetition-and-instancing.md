# Repetition and instancing

Read this when you're about to place more than roughly 2-3 near-identical
copies of something by a formula (evenly spaced, scattered, spiraling) —
copies in a ring, in a grid, along an edge or along a path. Stop before
hand-placing the second copy; every one of these arrangements has a
dedicated node, and hand-building N copies with individually-computed
positions is the anti-pattern: N radial copies are one `CircleNode`, not
N separate hand-placed subgraphs.

Not every multi-part object needs this — two distinct parts are just two
`ModelNode`s, no pattern node required.

## The pattern nodes: `modeling/repeat`

`search_node_types` with `pathPrefix: "modeling/repeat"`, or `get_node_types`
on any node in it — every node is documented. Each one outputs a `[]trs.TRS` transform list:

- **`CircleNode`** — evenly spaced around a circle, each transform already
  rotated to face outward. The everyday case for "N things arranged
  radially". `Times: 5` replaces 5 hand-placed copies with one node.
- **`GridNode`** — rows x columns, flat grid. It
  lies in the **XZ plane**: a grid on a vertical face needs a `trs.MultiplyNode` (`A` <- a
  −90°-about-X transform, `B` <- the grid array) to stand it up — its
  ports are lifted, so one node turns every copy — and the repeated
  piece is built in the grid's local frame, where its `Height` ends up
  along world Z.
- **`LineNode`** — evenly spaced along a straight segment.
- **`SplineNode`** — evenly spaced along a curve, tangent-aligned.
- **`FibonacciSphereNode`** / **`FibonacciSpiralNode`** — even coverage
  over a sphere's surface or a flat spiral, no visible clustering.
- **`SampleMeshSurfaceNode`** — random scatter across a mesh's surface,
  oriented to the normal (surface detail on an irregular shape).
- **`RandomPointsInSphereNode`** — random scatter through a solid sphere's
  volume.
- **`TransformationNode`** — cumulative (each copy built on top of the
  previous one) for spirals/growth sequences, not a fixed formula like the
  above.
- **`TRSNode`** — compounds two patterns together (a ring x a grid).
  Moving every transform in a list by one transform is not this: wire
  `trs.MultiplyNode` with `A` the one transform and `B` the list — its
  ports are lifted, so one node does the whole list.

## Exact type keys, do not search for these

Each `PATH` goes inside `github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Confirmed against the registry; use them directly.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.ModelNode` | Mesh, Material, Translation, Rotation, Scale, `Gpu Instances`, Children[], Name | Out |
| `formats/gltf.ManifestNode` | Models[], Animations[] | Out |
| `math/trs.NewNode` | Position, Rotation, Scale | Out |
| `math/quaternion.FromEulerAngleNode` | Angle | Out |
| `math/quaternion.RotationToNode` | From, To | Out |
| `math/sdf.RepeatNode` | Field, Transforms, Radius | Result |
| `math/sdf.MirrorNode` | Field, Union | X, Y, Z, XY, XZ, YZ, XYZ |
| `math/vector3.NewNode[float64]` | X, Y, Z | Out |

A fixed rotation can be a quaternion literal (`{"x","y","z","w"}`, see
the orchestrator's value-encoding table for the 90°/180° values); a
rotation computed from a variable is a quaternion node referenced by
`nodeId`/`port`. **`FromEulerAngleNode`'s `Angle` takes one angle or an
array of them**, returning a single rotation for a `ModelNode.Rotation`
or one per element for instancing. A rotation that *points* something —
along a surface normal, at a target, down a chain segment — is
`RotationToNode` (`From`: the axis the part was built along, `To`: the
direction it should face), not an angle you worked out by hand.

**Two or three art-directed placements of one thing** — not a formula,
each one hand-posed — are the one case the pattern nodes above don't cover. That is a
`[]trs.TRS` variable (`create_variables` type `"[]trs.TRS"`), wired
straight into `GpuInstances`: the user gets one move/rotate/scale gizmo
per entry in the web UI, and the graph stays one `ModelNode` instead of
N copies.

## Placing the copies: three options, pick the cheapest that fits

Once you have a `[]trs.TRS` list, there are three ways to actually draw
the copies:

### 1. `gltf.ModelNode.GpuInstances` — the default for a final, unmodified set

Wire the transform list straight into a `ModelNode`'s `GpuInstances` port.
This draws N copies of the *same* mesh data without duplicating it
anywhere in the graph or the exported `.glb` — the cheapest option, and
the default choice unless you have a specific reason to need one real
combined mesh (see option 2).

**How the transforms actually compose, so placement math comes out
right**: each entry in `GpuInstances` is *multiplied onto* the
`ModelNode`'s own `Translation`/`Rotation`/`Scale` — confirmed from the
real glTF writer's reference implementation
(`formats/gltf/model_trackers.go`'s instance-adding logic, which composes
as `model.TRS.Multiply(instance)`), not used as an absolute world
transform on its own. Practically: if the `ModelNode`'s own
`Translation`/`Rotation`/`Scale` are left at identity, each instance's
transform *is* its final world placement, which is the simplest and
usually-correct setup — put all your per-copy variation (from `CircleNode`
etc.) into the instances themselves, and only give the `ModelNode` a
non-identity base transform if you deliberately want every instance
shifted/rotated/scaled together as a group on top of their individual
placements (e.g. orienting an entire ring of copies to a tilted face).

This is the fix for N radial copies: one `instantiate_subgraph` of
the repeated mesh, one `CircleNode` (`Times: 5`), one `ModelNode` with
`GpuInstances` wired to the `CircleNode`'s output — instead of 5x
`instantiate_subgraph` + `ModelNode` pairs.

`GpuInstances` repeats the model's own `Mesh` and **never its
`Children`** — that is how glTF instancing works. A `ModelNode` with
`GpuInstances` and children but no mesh of its own is an error. To
repeat a component made of several models — a body and two attached
pieces — give each part's `ModelNode` the same `GpuInstances` list. Each
instance is applied before that part's own transform, so build each
part's offset within the component into its mesh (a
`meshops.TransformNode` on the mesh) and leave the part's `ModelNode`
transform at identity; otherwise a rotated instance swings the part
around the wrong point.

### 2. `repeat.MeshNode` — when you need one real combined mesh

Bakes the copies into one actual mesh. Use this instead of `GpuInstances`
only when the result needs to be a single mesh downstream — further
boolean/SDF operations, or an export path that doesn't support instancing.

### 3. `math/sdf.RepeatNode` — the SDF-field equivalent

Same idea, but for an SDF *field* instead of a mesh: places copies of the
field at each transform and unions them. This is how to repeat something
built with `math/sdf` (e.g. one SDF piece, repeated via a
`CircleNode`'s output) instead of unioning N hand-built copies of the
field together. `RepeatNode.Transforms` takes the exact same `[]trs.TRS`
type every `modeling/repeat` node produces, so the same `CircleNode` feeds
either an SDF repeat or a mesh repeat/instance — no glue code needed.

## Scatter doesn't know about your features — mask it

Every pattern node here is deterministic and blind: `FibonacciSphereNode`,
`GridNode`, `RandomPointsInSphereNode` and the rest place copies from a
count and a shape, with no idea which parts of the surface you care
about. So scatter lands on features, across openings, over the one
region that was supposed to stay clean — and because the placement is
deterministic, it looks fine at one count and wrong at another, which
makes it read as bad luck rather than a missing step.

If a region has to stay clear, say so explicitly rather than tuning the
count until it happens to miss:

- **SDF**: subtract a keep-out volume from the scattered field — a sphere
  or box covering the protected feature, `SubtractionNode`'d out of the
  scatter before it is unioned into the body. The keep-out shape is
  geometry like any other, so it can be sized off the feature it
  protects.
- **Meshes/instances**: filter the transform array before drawing it.
  `math/trs.FilterPositionNode` splits a `[]TRS` by an axis-aligned box
  (`MinX`/`MinY`/`MinZ`/`MaxX`/`MaxY`/`MaxZ`) into the transforms inside
  it and those outside — take the outside set and the protected box stays
  empty.

Name the keep-out volume after what it protects, and check it in the same
render where you check the scatter: a mask that is slightly too small
reads exactly like no mask at all until you look at the boundary.

## The same anti-pattern applies to shape control points, not just copies

Wiring a configurable-pose body to separate top-level variables per joint
(`Point 1`...`Point 4`, each into its own body segment) bakes the point
*count* into the graph's structure — adding a 5th point of articulation
means editing the graph, not just changing a value, defeating the entire
purpose of a "configurable pose."

Instead: one `create_variable` of type `"[]vector3.vector[float64]"`
(e.g. path `"Body Pose"`) wired straight to `math/sdf.LinesNode.Points`,
not one variable per joint. This variable type gets a real point-list
editor in the web UI (an Add/Delete list plus a draggable 3D gizmo per
point), so it's the seed variable for any user-facing "add/remove/
reposition points" control.

If the body also needs to *bend* into different smooth poses rather than
stay a straight-segment chain — and especially if it should also *taper*
(thick at the base, thin at the tip) — don't feed
the pose variable to `LinesNode`/`VaryingRadiusLinesNode` directly.
Resample it through a curve first, via `create_tapered_curve_subgraph`:
place it (`instantiate_subgraph`, or a `create_nodes` entry with
`subgraphId`), wire `Points` to the pose variable (not a
literal — that's the whole point of a posable body; and it is one
`variable`/`value` carrying the list, never `elements`, which is only for
many-connection ports like `Meshes[]`), `Base Radius`/
`Tip Radius` to your desired thickness (equal values give a constant-
thickness bend, not a taper, so this tool covers both cases), and
`Samples` to however many segments the body mesh should actually have —
deliberately decoupled from and usually much bigger than how many pose
points the user drags around. Wire its `Field` output straight into
`LinesNode`'s replacement position, i.e. directly into your body's
`Union`/`SmoothUnionNode` — the tool's output is already a complete
`math/sdf` field, not a point array, so `LinesNode`/
`VaryingRadiusLinesNode` aren't needed at all once you're using it. This
decouples "how many points does the user pose" (few) from "how many
segments does the body mesh have" (many), automatically.

The tool builds this internally, for reference (or to hand-wire if you
need something it doesn't expose — a non-default spline `Alpha`, a
`Closed` loop): `math/curves.CatmullRomSplineNode` (`Points` <- the pose
variable — its own doc explains why Catmull-Rom, not Bezier, is right
here) -> `Spline`; `LengthNode` (`Spline`) -> arc length;
`math/sequence.LinearNode` (`Start: 0`, `End:` the length, `Samples:` the
segment count) -> `Distance`; `math/curves.PositionNode` (`Spline` +
`Distance`) -> the dense point array; a second `LinearNode` sharing the
same `Samples` builds the matching `Radii` for a taper, index-for-index,
no remap needed. Don't build a taper by hand-chaining individual
`RoundedConeNode`s and `SmoothUnionNode`-ing them together instead — see
`organic-sdf-modeling.md` for why that produces a lumpy seam at every
joint rather than a clean taper.

A few positions that each come from their own math (four corners
computed from two spacings, a mirrored pair) collect into one array
with `arrays.FromElementsNode` (`Elements[]`, one element per node),
and `math/trs.NewNode` turns that into the `[]trs.TRS` a
`GpuInstances` or `RepeatNode` wants — its `Position`/`Rotation`/`Scale`
are lifted, so feeding an array of positions gives back an array of
transforms from one node. No need to reach for a grid node whose
spacing you then have to back-solve. Going the other way,
`arrays.IndexNode` (`Array`, `Index`; negative counts from the end)
reads one element back out, and `arrays.LastNode` the final one. A
point that lies on a curve (a sphere or a colour marking at its tip)
is better read off the spline itself: `PositionNode` (`Spline`,
`Distance` <- `LengthNode`) is the end point, and follows the pose
variable instead of duplicating its last entry as a literal. Something at *every* pose point (rings along a
chain) is `sdf.RepeatNode` over `trs.NewNode` with only `Position` <-
the points connected; rotation and scale default to identity.

## An array into a list input: `arrays.SpreadNode`

A list input (`meshops.CombineNode.Meshes`, a manifest's `Models`, a
`ModelNode`'s `Children`, the csg, sdf and sketch unions' inputs,
`sketch.FillNode`'s `Outlines` and `Holes`, `math.AddNode`'s `Values`) takes one item per wire. When the items arrive as one array —
the output of a lifted node fed a list, or an `arrays` node —
`arrays.SpreadNode` (`Array` -> `Out`) hands each element over as if it
had its own wire. Wire its `Out` into the list input like any other
element; it can sit next to ordinary wires on the same input.

- Twelve text meshes from one array -> Spread -> `CombineNode.Meshes`
  gives one combined mesh; -> a manifest's `Models` exports twelve models.
  The array comes from a lifted node: `TextNode` fed a `[]string` (cut
  to a variable count with `arrays.SliceNode`), placed by
  `meshops.TransformNode` with a `[]trs.TRS`.
- `[1, 2, 3]` -> Spread -> `math.AddNode.Values` sums to 6. A spread of
  arrays (`[][]float64`) adds the arrays element by element instead.
- The `arrays` nodes' list inputs take it too: `FromElements`, `Append`'s
  `Elements`, `Concat`'s `Arrays` (a spread of arrays), and
  `logic.SwitchNode`'s `Cases`, where each element counts as one case.
- Only list inputs accept it. A single-value input, a lifted single input
  and subgraph boundaries all refuse it, because each would read just one
  element.
- An empty array contributes nothing, so a combine of only an empty
  spread is an empty mesh and a sum is 0. Min and Max of nothing report
  "no values to compare".
