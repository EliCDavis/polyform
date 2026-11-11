# Randomness: keys, seeds and distributions

Read this before wiring anything "random": a varied size per copy, a
jittered placement, one of several kinds per slot, a scatter, a seeded
noise. Everything random in polyform lives in `math/chance` and follows
one contract; a graph that fights the contract (sub-seeds by arithmetic,
`Math` chains to reshape a 0-1 value, K parallel chains for K kinds) is
the anti-pattern.

## Type keys

Every key below is `github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/<path>]`
unless written out. `V3` stands for `github.com/EliCDavis/vector/vector3.Vector[float64]`,
`V2` and `V4` likewise.

| Node | `<path>` | Inputs → output |
|---|---|---|
| Random (float) | `math/chance.RandomNode[float64]` | `Key`, `Seed`, `Distribution` → `Out` float64 |
| Random (vector) | `math/chance.RandomNode[V2]` / `[V3]` / `[V4]` | same, with a vector distribution → vector |
| Random Bool | `math/chance.RandomBoolNode` | `Key`, `Seed`, `Probability` → `Out` bool |
| Random Index | `math/chance.RandomIndexNode` | `Key`, `Seed`, `Weights` ([]float64) → `Out` int |
| Sample | `math/chance.SampleNode` | `Distribution`, `U` → `Out` float64 |
| Seed | `math/chance.SeedNode` | `Seed`, `Label` (string) → `Out` int |
| Constant | `math/chance.ConstantNode` | `Value` |
| Uniform | `math/chance.UniformNode` | `Min`, `Max` |
| Normal | `math/chance.NormalNode` | `Mean`, `Deviation`, `Clamp` (deviations, 0 = open tails) |
| Log Normal | `math/chance.LogNormalNode` | `Median`, `Spread` |
| Exponential | `math/chance.ExponentialNode` | `Median` |
| Triangular | `math/chance.TriangularNode` | `Min`, `Mode`, `Max` |
| Power | `math/chance.PowerNode` | `Min`, `Max`, `Exponent` |
| Discrete | `math/chance.DiscreteNode` | `Values`, `Weights` ([]float64 each) |
| Ramp | `math/chance.RampNode` | `Gradient` (`coloring.Gradient[float64]`) |
| Compose 2D/3D/4D | `math/chance.Compose2DNode` / `Compose3DNode` / `Compose4DNode` | `X`, `Y`, `Z`, `W`: one scalar distribution each |
| Box | `math/chance.BoxNode` | `Min`, `Max` (V3) |
| Inside Sphere | `math/chance.InsideSphereNode` | `Radius` |
| Direction | `math/chance.DirectionNode` | `Axis` (V3), `ConeAngle` (degrees) |
| Integers | `math/sequence.IntegersNode` | `Count`, `Start` → `Out` []int |
| Linear | `math/sequence.LinearNode` | `Start`, `End`, `Samples` → `Out` []float64 |
| Indices | `nodes/arrays.IndicesNode` | `Array` → `Out` []int |
| Group | `nodes/arrays.GroupNode` | `Array`, `Groups` ([]int), `Count` → `Out` list of lists |
| Gather | `nodes/arrays.GatherNode` | `Array`, `Indices` → `Out` |
| Filter | `nodes/arrays.FilterNode` | `Array`, `Keep` ([]bool) → `Kept`, `Dropped` |
| Shuffle / Sample / Random Pick | `nodes/arrays.ShuffleNode` / `SampleNode` / `RandomPickNode` | `Array`, (`Count`,) `Seed` |
| Select (per element) | `nodes/logic.SelectNode` | `Condition` (bool or []bool), `A`, `B` |
| Sample Mesh Surface | `modeling/repeat.SampleMeshSurfaceNode` | `Mesh`, `Attribute`, `Samples`, `Seed` → `[]trs.TRS` facing the normal |
| Random Points In Sphere | `modeling/repeat.RandomPointsInSphereNode` | `Radius`, `Points`, `Seed` → `TRS`, `Vector3` |
| Perlin 1D/2D/3D | `math/noise.Perlin1DNode` / `Perlin2DNode` / `Perlin3DNode` | `Time`, `Amplitude`, `Frequency`, `Seed` |
| Perlin field | `math/noise.Perlin3DFieldNode` | `Amplitude`, `Frequency`, `Seed` |
| Noise texture | `drawing/texturing.NoiseNode` | ..., `Seed` (int) |

Distribution ports carry a `math/chance.Distribution[float64]` value (or
`Distribution[V3]` for vectors); a boundary input of that type passes a
distribution into a subgraph like any other value.

## The contract: a value is a hash of (Seed, Key)

Nothing in `math/chance` has state. `RandomNode` hashes its `Seed` with
each `Key` and shapes the result through its `Distribution`. That gives
three guarantees to build on:

- **A key's value never changes when keys are added or removed around
  it.** Raising a count adds elements without moving the existing ones;
  deleting one leaves the rest alone. Keys are the element's index
  (`IndicesNode` of the list being varied, or `IntegersNode` when there
  is no list yet) or any stable id list.
- **The same seed and key always give the same value**, in any node, in
  any run. Two `RandomNode`s fed the same `Seed` and `Key` draw the same
  stream: that is how values are made to move together (below).
- **Different labels give unrelated values.** `SeedNode` derives a seed
  per `Label` from one master seed. Feed it a list of labels for several
  at once. Never build sub-seeds by arithmetic (`Seed + 1`, `Seed * 7`):
  adjacent integers are unrelated hashes already, but a label says what
  the stream is for and survives a renumbering.

The variable surface that follows: **one master seed variable** at the
root, every randomized part reading its own label off it through
`SeedNode`, so one integer gives a whole new arrangement and no two parts
share a stream unless they are meant to.

`Distribution` on `RandomNode` is required. The simplest draw, a value in
[0, 1), is `UniformNode` with its defaults.

## Distributions: the shape of the values

A distribution is a value built by its own node and wired in, so one
`RandomNode` serves every shape and the shape can be a boundary input.
Pick by the shape you want, not by the name:

| Want | Use |
|---|---|
| any value in a range, all equally likely | `Uniform(Min, Max)` |
| clustered around a typical value, symmetric | `Normal(Mean, Deviation, Clamp)`; `Clamp: 3` keeps every value within 3 deviations with nothing piling up at the edges |
| sizes and scales: always positive, most near a typical value, a few much larger | `LogNormal(Median, Spread)`; `Spread 0.25` is a gentle variation, `0.5` a wide one |
| mostly small, a thinning tail of large | `Exponential(Median)` |
| a most-likely value with hard limits either side | `Triangular(Min, Mode, Max)` |
| crowded toward one end of a range | `Power(Min, Max, Exponent)`; above 1 crowds toward `Min`, below 1 toward `Max` |
| one of a fixed set of values, weighted | `Discrete(Values, Weights)`; a weight of 0 never comes up |
| a custom shape | `Ramp(Gradient)`: the gradient's value at time t is what the quantile t gives, so a flat stretch is common and a steep one is rare |
| always the same value | `Constant(Value)`: the way to turn variation off at one input without rewiring |

Every scalar distribution is an inverse CDF, also called a quantile
function: the value at u is what a random draw lands below u of the
time, so u = 0.5 is the median and u = 0.9 the value only one draw in
ten exceeds. `SampleNode` applies that function to a `U` you supply. Use
it to shape a value that is not random at all, such as height along a
trunk or normalized index, the way a distribution would, or to draw a
distribution's curve by feeding it `LinearNode` quantiles. Keep supplied
quantiles inside (0, 1): the open-tailed `Exponential` and `LogNormal`
go to infinity at exactly 1, so use bin midpoints (`Start 0.025`,
`End 0.975` for 20) rather than 0 to 1.

### Vectors

A vector `RandomNode` takes a vector distribution:

- `Compose3DNode` (`X`, `Y`, `Z`, one scalar distribution each; an
  unwired axis is uniform 0 to 1). Each axis draws its own value from the
  key's stream, so the components are independent even when the same
  scalar distribution feeds all three. `Compose2D` and `Compose4D` do the
  same for `V2` and `V4`.
- `BoxNode` (`Min`, `Max`): uniform through a box. Offsets and jitter.
- `InsideSphereNode` (`Radius`): uniform through a solid ball. Isotropic
  jitter, scatter inside a volume.
- `DirectionNode` (`Axis`, `ConeAngle`): a unit vector within the cone.
  `0` always gives the axis, `180` any direction. Tilts, growth
  directions, normals for a scatter.

`DirectionNode` gives a direction, not a rotation: turn it into one with
`quaternion.RotationToNode` from the part's own up axis, or compose Euler
angles through `Compose3DNode` into `quaternion.FromEulerAngleNode`.

## Jitter is composition, not a node

To jitter a list of transforms: `IndicesNode` of the list for keys, a
vector `RandomNode` with `BoxNode` or `InsideSphereNode` for the offset,
another with `Compose3DNode` of angle distributions into
`quaternion.FromEulerAngleNode` for the turn, a float `RandomNode` with
`LogNormalNode` for the scale, all into `trs.NewNode`, then the lifted
`trs.MultiplyNode` with `A` the original transforms and `B` the jitter
list. Each of the three draws reads its own `SeedNode` label. Nothing in
that chain is specific to jitter, which is why there is no jitter node.

## Values that move together

Two `RandomNode`s with the same `Seed` and `Key` draw the same stream, so
"the tall ones are also the fat ones" is one key list into two
`RandomNode`s with different distributions: the correlation is in the
shared wires, not in a node. To decorrelate one of them, give it a
`SeedNode` label. A vector distribution's components are independent by
construction, so a correlated vector is three correlated scalars through
`vector3.NewNode`, not a `Compose3D`.

## One of several kinds per slot

- `RandomIndexNode` (`Key`, `Seed`, `Weights`) gives one kind index per
  key, in proportion to the weights. "Empty" is a kind with its own
  weight, and the weights are the natural variables (`Weights/Sphere`,
  `Weights/Empty`).
- `GroupNode` (`Array`, `Groups`, `Count`) sorts the slots' transforms
  into `Count` lists, list i holding the slots of kind i, empty lists
  kept.
- Run one subgraph For Each over the groups, its boundary taking one
  kind's transforms (a list) plus whatever differs per kind, rather than
  K parallel chains. See `topics/repetition-and-instancing.md` for For
  Each itself.
- `GatherNode` with the same indices picks a per-slot value from any
  list (a color, a size) instead of a whole model.
- `DiscreteNode` into a float `RandomNode` is the same choice as a
  value rather than an index, for a per-slot size or tint drawn from a
  fixed set.

## Keep-outs and masks

A per-element `logic.SelectNode` takes a `[]bool` `Condition`: `A` and
`B` are lists of the same length, or one element used for every flag.
Writing `-1` over a kind index where a mask is true drops that slot in
`GroupNode`; `FilterNode` (`Array`, `Keep`) splits any list by the same
flags. The flags come from `RandomBoolNode` (a random fraction of slots),
from `math.CompareNode` over a position list (a geometric keep-out), or
from a noise sample thresholded (clearings).

## Scattering

- `SampleMeshSurfaceNode` scatters by triangle area, each transform
  facing the surface normal, from one seed.
- `RandomPointsInSphereNode` and `InsideSphereNode` are the same
  distribution; the node is a convenience for a count and a seed, the
  distribution composes with keys and other draws.
- A scatter plus a density mask is a scatter, then a `FilterNode` whose
  flags compare a noise sample against a threshold variable.

## Noise seeds

`Perlin1D/2D/3DNode`, `Perlin3DFieldNode` and texturing's `NoiseNode` all
take an integer `Seed` and move to a different stretch of noise per seed.
Read the seed off the master through `SeedNode` like any other stream; a
noise seed and a placement seed with the same label are unrelated
because the nodes consume them differently, but giving them distinct
labels keeps that true if either changes.

## Traps

- A For Each run whose per-item port is itself a list (a group's
  transforms) is fed a list of lists. The graph handles it; the editor
  draws the wire only as of the For Each port-type fix, so a missing
  wire on such a port in an older editor is a drawing gap, not a broken
  graph.
- `RandomNode` with nothing wired into `Distribution` errors rather than
  drawing. Wire `UniformNode` for a raw 0 to 1 value.
- A list of the same scalar distribution on every `Compose3D` axis is
  fine; a single `RandomNode[float64]` fanned into `vector3.NewNode` for
  X, Y and Z is not: that is one value copied three times.
- `Normal` with `Clamp 0` and `Exponential` reach arbitrarily far on rare
  keys; clamp or choose `Triangular` when a hard limit matters for
  geometry (a part must not pierce another).
- Seeds are integers. A float wired into a `Seed` port is cast and
  truncated, so `0.3` and `0.7` are the same seed.
