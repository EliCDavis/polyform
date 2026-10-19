# Emissive and KHR_materials_emissive_strength — things that glow

A surface that gives off its own light. Core glTF caps emissive at
the colour's own brightness. `Emissive Strength` multiplies it past 1 so
a viewer with bloom makes it glow.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialNode` | `Emissive Factor`, `Emissive Strength`, `Emissive Texture`, ... | Out |

The `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
There is no separate extension node: these are plain `MaterialNode`
ports.
- `Emissive Factor` is the glow colour, e.g. `{"value": "\"#ff2a1a\""}`.
- `Emissive Strength` is a float multiplier, e.g. `{"value": "6"}`.
- `Emissive Texture` is an optional pattern (sRGB), multiplied by the
  factor; it needs UVs.

## Rules

- **`Emissive Factor` must be set.** It defaults to black, and black
  times any strength is nothing. `Emissive Strength` alone writes the
  extension and glows not at all.
- `Emissive Factor` is linear, so pass a hex pick through
  `coloring.SRGBToLinearNode`, like the base colour.
- Strength:
  - `1`: lit but not glowing;
  - `2-4`: a visible glow;
  - `5-10`: a bright light source.
- Emission is not a light. It does not light the parts around it, and it
  doesn't cast or remove shadows; only bloom in the viewer spreads it.
  The writer can export `KHR_lights_punctual` lights, but no node places
  one yet.
- Glow only the part that emits: a small emissive core inside a clear or
  diffuse body, not the whole body. Making the whole body emissive turns
  it into a glowing slab that reads as some other part.
- To make it blink, animate a separate emissive model's scale (see
  `animation.md`): polyform doesn't support `KHR_animation_pointer`, so
  strength itself can't be keyed.

## Verify

`render_preview` ignores emission; the part shows its base colour.
`describe_manifest` lists the material's `emissive` colour and
`KHR_materials_emissive_strength` with `emissiveStrength`. If `emissive`
is missing or `#000000`, nothing glows whatever the strength.
