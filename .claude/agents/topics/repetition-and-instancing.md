# Repetition and instancing

Read this when you're about to place more than roughly 2-3 near-identical
copies of something by a formula (evenly spaced, scattered, spiraling) —
points of a star, a ring of bolts, a grid of vents, studs along an edge,
wheels on a car. Stop before hand-placing the second copy; every one of
these arrangements has a dedicated node, and hand-building N copies with
individually-computed positions is exactly the anti-pattern that produced
a 5-point star built as 5 separate hand-placed subgraphs instead of one
`CircleNode`.

Not every multi-part object needs this — a body with one head and one
tail is just two `ModelNode`s, no pattern node required.

## The pattern nodes: `modeling/repeat`

`search_node_types` with `pathPrefix: "modeling/repeat"`, or `get_node_types`
on any node in it — every node is documented, including a roster comment
on the package itself. Each one outputs a `[]trs.TRS` transform list:

- **`CircleNode`** — evenly spaced around a circle, each transform already
  rotated to face outward. The everyday case for "N things arranged
  radially" — bolts around a hole, numbers on a clock face, points of a
  star. `Times: 5` replaces 5 hand-placed copies with one node.
- **`GridNode`** — rows x columns, flat grid (vents, tile arrays).
- **`LineNode`** — evenly spaced along a straight segment (rivets on an
  edge).
- **`SplineNode`** — evenly spaced along a curve, tangent-aligned (posts
  along a winding wall).
- **`FibonacciSphereNode`** / **`FibonacciSpiralNode`** — even coverage
  over a sphere's surface or a flat spiral, no visible clustering.
- **`SampleMeshSurfaceNode`** — random scatter across a mesh's surface,
  oriented to the normal (greebling/detail on an organic or irregular
  shape).
- **`RandomPointsInSphereNode`** — random scatter through a solid sphere's
  volume.
- **`TransformationNode`** — cumulative (each copy built on top of the
  previous one) for spirals/growth sequences, not a fixed formula like the
  above.
- **`TRSNode`** — compounds two patterns together (a ring x a grid).

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
placements (e.g. orienting an entire ring of bolts to a tilted face).

This is the fix for the classic star case: one `instantiate_subgraph` of
the point mesh, one `CircleNode` (`Times: 5`), one `ModelNode` with
`GpuInstances` wired to the `CircleNode`'s output — instead of 5x
`instantiate_subgraph` + `ModelNode` pairs.

### 2. `repeat.MeshNode` — when you need one real combined mesh

Bakes the copies into one actual mesh. Use this instead of `GpuInstances`
only when the result needs to be a single mesh downstream — further
boolean/SDF operations, or an export path that doesn't support instancing.

### 3. `math/sdf.RepeatNode` — the SDF-field equivalent

Same idea, but for an SDF *field* instead of a mesh: places copies of the
field at each transform and unions them. This is how to repeat something
built with `math/sdf` (e.g. one SDF star point, repeated via a
`CircleNode`'s output) instead of unioning N hand-built copies of the
field together. `RepeatNode.Transforms` takes the exact same `[]trs.TRS`
type every `modeling/repeat` node produces, so the same `CircleNode` feeds
either an SDF repeat or a mesh repeat/instance — no glue code needed.

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
reposition points" control — a snake, tentacle, tail, rope, cable, spine.

If the body also needs to *bend* into different smooth poses rather than
stay a straight-segment chain, resample a curve through the pose points
instead of feeding them to `LinesNode` directly:
`math/curves.CatmullRomSplineNode` (`Points` <- the same pose variable —
its own doc explains why Catmull-Rom, not Bezier, is the right curve type
for this) produces a smooth `Spline` that passes through every pose point;
`LengthNode` + `math/sequence.LinearNode` (`Start: 0`, `End:` the length,
`Samples:` however many segments the body should actually have) build a
distance range; `PositionsForArrayNode` resamples the spline at those
distances into the dense point array `LinesNode.Points` actually consumes.
This decouples "how many points does the user pose" (few) from "how many
segments does the body mesh have" (many) — `get_node_types` on any node in
`math/curves` for the full pipeline, every node in it is documented.
