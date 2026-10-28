# Animation

Read this when a build asks for motion in the export — parts that spin,
orbit, swing or loop — or when a part is likely to be animated later. Read it **before** building the
moving parts, not after: the main cost of animating a finished model is
splitting merged meshes back into separately movable pieces.

Animation in polyform is glTF keyframe animation of whole `ModelNode`s.
A channel moves one model node's rotation, translation or scale over time.
Nothing deforms; nothing inside a combined mesh moves on its own.

## Exact type keys, do not search for these

Each `PATH` goes inside `github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.

| PATH | inputs | outputs | notes |
| --- | --- | --- | --- |
| `math/sequence.LinearNode` | Start, End, Samples | Out | `[]float64`, both ends inclusive |
| `math/quaternion.FromThetaNode` | Theta, Direction | Out | Theta in radians, Direction is the axis. Lifted: an angle array gives a quaternion array |
| `modeling/animation.UniformFramesNode[github.com/EliCDavis/polyform/math/quaternion.Quaternion]` | Data, Duration | Out | Duration in seconds; Out is the keyframes |
| `modeling/animation.UniformFramesNode[github.com/EliCDavis/vector/vector3.Vector[float64]]` | Data, Duration | Out | Duration in seconds |
| `formats/gltf.RotationAnimationChannelNode` | Target, Interpolation, Frames | Out | |
| `formats/gltf.TranslationAnimationChannelNode` | Target, Interpolation, Frames | Out | |
| `formats/gltf.ScaleAnimationChannelNode` | Target, Interpolation, Frames | Out | |
| `formats/gltf.AnimationNode` | Name, Channels[] | Out | |
| `formats/gltf.ManifestNode` | Models[], Animations[] | Out | |

`Interpolation` defaults to `LINEAR` (slerp for rotations) when unwired.
There is no literal for it; leave it unwired.

## The spin recipe

A part turning `turns` times about an axis over `seconds`, starting from
angle `a0`:

```
LinearNode(Start: a0, End: a0 + 2π·turns, Samples: ceil(4·|turns|) + 1, minimum 5)
  → FromThetaNode(Theta: ↑, Direction: axis)
  → UniformFramesNode[Quaternion](Data: ↑, Duration: seconds)
  → RotationAnimationChannelNode(Target: the model, Frames: ↑)
  → AnimationNode.Channels
```

Build it once as a helper subgraph (`spin_channel`: Target, Turns, Start
Angle, Seconds, Axis → Channel) and instantiate it per moving part. The
`Samples` formula is not optional, see rule 4. A negative `turns` spins
the other way.

## Rules

**1. Everything that moves is its own `ModelNode`, built on its pivot.**
A channel rotates a model about that model's own origin. A spinning part
must be a model whose mesh is centred on its rotation axis and whose
`Translation` places it. Parts that turn together on one axis share one
model, with children for different materials. Anything folded into a
`CombineNode` mesh with its neighbours cannot move separately. If a pivot
isn't at a part's natural origin, wrap the part in a parent `ModelNode`
whose `Translation` is the pivot, with the part as a child offset from
it. A part orbiting a point is a model at that point with the part
offset inside it.

**2. A channel replaces the property; it does not add to it.** A rotation
channel overwrites the model's static `Rotation` for the whole clip. If
the model already has a rotation (a starting angle or phase), either make that angle the first keyframe (`Start: a0`), or
put the static rotation on a child and animate a parent with identity
rotation. Translation and scale channels are absolute positions and
scales, not offsets. Frame 0 should equal the static pose. Check it with a
`render_preview` before and after: they must match.

**3. `Target` must be the exact model output that is in the manifest
tree.** The writer finds the channel's glTF node by model pointer. A
target model that is not wired into `Models` (or a descendant's
`Children`) silently becomes **glTF node 0**, so the wrong part animates
and no error is raised. Wire the channel's `Target` from the same output
that feeds the tree, not from an upstream node or a copy.

**4. Keep every key step at 90° or less.** `LINEAR` rotation slerps the
short way, so keys 180° apart are ambiguous and anything past 180° spins
backwards. `ceil(4·|turns|) + 1` samples keeps each step at or under 90°;
rounding instead of `ceil` can push a step past it.

**5. Seamless loops need whole symmetry periods.** Over one clip, every
moving part must return to a pose that looks identical:
- a part with no rotational symmetry: a whole number of turns;
- a part with N-fold symmetry: a whole number of 1/N turns, and **every
  feature counts**. A part with 40 copies of one feature and 5 of
  another only repeats every 1/5 turn.
- a driven part turns at a fixed ratio of its driver's speed, and parts
  on one axis share an angle. Pick the driver's turns so every
  downstream part lands on its own period. If no small driver count
  works, change the feature counts, and say so in the report.
- use stylized whole-number ratios (e.g. turns 8, 5, 4, 3, 2, 1),
  not real-world periods; real-world ratios never close.

**6. One `AnimationNode` for the whole model.** The web editor plays only
`animations[0]`. Two channels in one animation may not target the same
node and property.

**7. Optional parts need a fallback model, not an empty branch.** A
`SelectNode` with an unwired branch passes nothing. `AnimationNode.Channels`
doesn't drop that; it receives a zero channel with no target and no
sampler, which crashes export or animates node 0. For a part that can be
turned off (`Part Count`, `Attached Part Index`), fall back to an empty
placeholder `ModelNode` (no mesh), and target the channel at the same
select output that feeds the tree, so the channel always has a real node.
`AnimationNode.Channels` takes one channel per wire; an array of
channels goes in through `arrays.SpreadNode`.

**8. Timing is a control.** Expose the clip length and the turn counts as
variables (`Clip Seconds`, `Turns` as an int/float array, `Drive
Turns`), and derive everything else from them, including driven speeds
from feature counts. A hand-typed speed is a snapshot that breaks the loop
the moment a feature count changes.

## Verifying an animation

None of the usual tools shows motion:

- `render_preview` renders the rest pose only. Use it to confirm the
  static pose didn't change, nothing more.
- `describe_manifest` on the manifest node lists every animation with
  its clip length and each channel's target model path, node index,
  property (`translation`/`rotation`/`scale`), key count and last key
  time, exactly as the file will carry them. A channel whose target is
  not under `Models` comes back as a `warnings` entry naming the node the
  file aims it at instead. This is the check for targets; never `Read`
  the `.glb` for it.
- `evaluate_node` on an `AnimationNode` or a channel serialises every
  keyframe **and the whole target model recursively**: about 20k tokens
  for 12 channels. Don't. Evaluate the `LinearNode` inside one
  `spin_channel` instance (by instance path, with `maxItems`) to check key
  spacing; `describe_manifest` already gives the key count.

Report each channel in the hand-back: target name, axis, turns per clip,
key count. Also list every part whose loop is not seamless, and why.
