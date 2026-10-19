# KHR_materials_ior — index of refraction

How much light bends entering a transmissive volume, and how strongly
any non-metal reflects head-on. It is a plain number on the material, not
a separate node.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialNode` | `Index Of Refraction`, ... | Out |

The `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
`Index Of Refraction` takes a float, e.g. `{"value": "1.5"}`.

## Values

glTF's default is `1.5`, which already suits glass and most plastics, so
only set it when the material differs.

| Material | IOR |
| --- | --- |
| Air-like / no bend | 1.0 |
| Ice | 1.31 |
| Water | 1.33 |
| Plastic, acrylic | 1.45-1.5 |
| Glass | 1.5-1.52 |
| Amber | 1.55 |
| Sapphire, ruby | 1.77 |
| Diamond | 2.42 |

## Where it matters

- **With transmission and volume** (`transmission.md`, `volume.md`) it
  sets how much the view through the part is bent. A high-IOR solid's
  sparkle is mostly its IOR, with `dispersion.md` adding colour.
- **On any non-metal** it also sets head-on reflectance: a higher IOR
  gives a shinier-looking dielectric. Raising it is a quick way to make
  a coated or wet surface read glossier without touching roughness.
- **On metal** (`Metallic Factor 1`) it has no effect.

## Verify

Not shown in `render_preview`. `describe_manifest` lists
`KHR_materials_ior` with the material's `ior`.
