# KHR_materials_dispersion — rainbow fire in refracting solids

Splits light passing through a refracting volume into colours: the
coloured flashes at a faceted edge.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialDispersionExtensionNode` | Dispersion | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Dispersion` port.

## Requirements

It only acts on light passing through a volume, so the material needs
all of:
- `transmission.md` (Factor near 1);
- `volume.md` with a `Thickness Factor` above 0;
- an `Index Of Refraction` (`ior.md`).

Without them it is written to the file and does nothing.

## Values

`Dispersion` is 20 divided by the material's Abbe number.

| Material | Dispersion |
| --- | --- |
| None (default) | 0 |
| Crown glass | 0.33 |
| Diamond, water | 0.36 |
| Dense flint glass | 0.5 |
| Polycarbonate | 0.63 |
| Rutile (strong fire) | 2.0 |

Values above 1 are valid and give a stylised, exaggerated effect.

## Verify

Not shown in `render_preview`. `describe_manifest` lists
`KHR_materials_dispersion` with the value, alongside
`KHR_materials_transmission`, `KHR_materials_volume` and
`KHR_materials_ior` on the same material. If any of those three is
missing, the dispersion is inert.
