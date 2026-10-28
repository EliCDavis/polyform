# KHR_materials_unlit — flat colour, no lighting

The surface shows its colour exactly as set, with no shading,
highlights or shadows. Use it for:
- stylised parts that should read as flat colour;
- an image on a surface that should not be shaded;
- debug markers and gizmos;
- a baked-lighting look where the colour already contains the light.

It is a bool port on the material, not a separate node.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialNode` | Unlit, Color, `Color Texture`, ... | Out |

The `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
`Unlit` takes `{"value": "true"}`.

## What still applies

Only `Color`, `Color Texture` and vertex colour are shown, plus alpha
through `Alpha Mode`. Metallic, roughness, normal maps, occlusion,
emissive and every other extension are ignored by an unlit viewer.
Viewers without unlit support fall back to ordinary shading of the same
colour.

Unlit is not glowing. For a surface that should light up and bloom, use
`emissive-strength.md` instead.

## Verify

`render_preview` shades it like any other part. `describe_manifest` lists
`KHR_materials_unlit` on the material.
