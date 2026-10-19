# KHR_materials_sheen — soft fabric glow

The soft, bright rim fibres catch at grazing angles. Use it for cloth
that should read as fabric rather than painted plastic.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialSheenExtensionNode` | `Color Factor`, `Color Texture`, `Roughness Factor`, `Roughness Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Sheen` port.

- **`Color Factor` is the sheen's colour, and it defaults to black,
  which is no sheen at all.** Always set it. It is linear, so pass a hex
  pick through `coloring.SRGBToLinearNode`, as for the base colour.
  - Velvet: a lighter, more saturated version of the fabric colour.
  - Satin: near-white.
- **`Roughness Factor`**, 0 to 1 (default 0):
  - `0.2-0.4`: tight, shiny sheen (satin, silk);
  - `0.5-0.8`: broad, soft glow (velvet, felt, plush).
- `Color Texture` (sRGB) and `Roughness Texture` (alpha channel) vary
  them across the surface; both need UVs.

## The base under it

Fabric is a matte non-metal. Set `Metallic Factor 0` explicitly (it
defaults to `1`) and `Roughness Factor` high, `0.7-1`. A smooth base with
sheen reads as plastic with a halo.

## Verify

`render_preview` doesn't show sheen. `describe_manifest` lists
`KHR_materials_sheen` with `sheenColorFactor` not black and the
`sheenRoughnessFactor`.
