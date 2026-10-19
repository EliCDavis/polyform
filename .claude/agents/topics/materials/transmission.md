# KHR_materials_transmission — see-through surfaces

Use it for anything light passes through.

Transmission or alpha? `MaterialNode`'s `Alpha Mode` is the other way to
see through a surface, and they are not interchangeable:
- **Transmission** is a physical material: it still reflects, bends
  light with a volume, and frosts with roughness. Use it for clear
  solids and liquids.
- **`Alpha Mode` `MASK`** cuts holes along a texture's or colour's
  alpha. Alpha comes from `Color`'s alpha and `Color
  Texture`'s alpha channel, and `Alpha Cutoff` sets the threshold.
- **`Alpha Mode` `BLEND`** fades the whole surface, reflections
  included. It sorts poorly
  in viewers, so don't use it for glass.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialTransmissionExtensionNode` | Factor, Texture | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Transmission` port.

- `Factor` is the share of light that passes through, 0 to 1. Clear
  glass and water are `1`; frosted or milky plastic `0.5-0.9`.
- `Texture` (R channel, multiplied by `Factor`) varies it across the
  surface: a clear region in an opaque surface, worn frosting.

## How it combines with the rest of the material

- **It only applies to non-metal.** Set `Metallic Factor` to `0`
  explicitly: it defaults to `1`, and the metallic share of a surface is
  opaque, so an unset metallic gives a mirror, not glass.
- **`Color` tints the light coming through.** Clear glass is white or
  near-white; a green-tinted glass is a pale green `Color`. For colour that
  deepens with thickness, use `volume.md` instead.
- **`Roughness Factor` frosts it.** `0` is clear, `0.2-0.4` reads as
  frosted glass.
- **It still reflects.** A transmissive surface keeps its specular
  highlights and Fresnel edge. It is not an invisible surface, and that
  is what makes it read as glass.

## Thin wall or solid

On its own, transmission treats the surface as an infinitely thin
sheet: light passes straight through without bending. That is right for
a thin transparent panel or film. For a solid object that should
refract what is behind it, add `volume.md` with a thickness. Set `ior.md` for the bend.

## Verify

`render_preview` does not render transmission: the part shows as an
opaque surface. Check the export with `describe_manifest`:
`KHR_materials_transmission` must be in `extensionsUsed`, with the
material's `transmissionFactor`. The web editor's viewer renders it.
Note that a viewer shows other transmissive objects behind it only
partly or not at all, so glass behind glass looks less clear than in
life.
