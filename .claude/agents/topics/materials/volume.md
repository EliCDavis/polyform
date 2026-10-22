# KHR_materials_volume — thick, refracting, tinted interiors

Turns a transmissive surface from a thin sheet into the boundary of a
solid: light bends as it enters, and coloured media darken with depth.
Use it for a transparent solid or liquid.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialVolumeExtensionNode` | `Thickness Factor`, `Attenuation Distance`, `Attenuation Color`, `Thickness Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Volume` port.

## Requirements

- **It needs transmission.** On its own it does nothing; wire
  `transmission.md` on the same material.
- **The mesh must be closed.** The surface is the boundary of the medium,
  so an open shell, a single face or a mesh with holes refracts wrongly.
- Set `Index Of Refraction` on the material (`ior.md`): it is how much
  the light bends.

## Values

- **`Thickness Factor`** is how thick the object is, **in the mesh's own
  units** before any `ModelNode` scale. A good value is about the part's
  typical thickness, e.g. a 6 cm-thick part in metres is `0.06`.
  - `0` means a thin wall again: the extension then changes nothing.
  - `Thickness Texture` (G channel, needs UVs) scales it across the
    surface, for a part whose thickness varies: thick in the middle,
    or heavy at the base. Without it one factor stands for
    the whole part.
  - Don't share a volume material between parts of very different size:
    a thickness right for one is wrong for the other. `describe_manifest`
    warns when this happens.
- **`Attenuation Color`** is the colour white light becomes after
  travelling `Attenuation Distance` through the medium. Examples: amber
  `#ffaa33`, green glass `#55aa55`, sea water `#3399aa`.
  - It is a linear factor, so pass a hex pick through
    `coloring.SRGBToLinearNode` first, the same as the base colour.
- **`Attenuation Distance`** is in world units (metres in a metre-scale
  scene). Shorter is denser.
  - Leave it unset for clear glass, which absorbs nothing.
  - Set it to about the part's size for a tint that is visible but not
    murky; a fraction of the size for a deep colour.
- Keep the base `Color` white or very pale when using attenuation. Both
  tint the light, and together they come out muddy.

## Verify

`render_preview` shows none of it. Use `describe_manifest`:
`KHR_materials_volume` in `extensionsUsed`, `thicknessFactor` and the
attenuation values on the material, and no volume-thickness warning. The
web editor's viewer renders refraction.
