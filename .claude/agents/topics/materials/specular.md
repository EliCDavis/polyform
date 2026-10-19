# KHR_materials_specular — dull or tinted highlights on non-metals

Controls how strongly a non-metal reflects, and in what colour, beyond
what roughness alone can say. Use it to:
- **dull a surface** that is smooth but shouldn't shine much;
- **tint the highlight** of a few special materials, e.g. coloured
  coated plastics, some fabrics.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialSpecularExtensionNode` | Factor, Texture, `Color Factor`, `Color Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Specular` port.

- `Factor`, 0 to 1 (default 1), scales the reflection strength.
  - `1` is ordinary.
  - `0.3-0.6` gives a subdued highlight.
  - `0` removes it entirely, which rarely looks real.
- `Color Factor` (default white) tints the head-on reflection. It is
  linear, so pass hex picks through `coloring.SRGBToLinearNode`.
  - Leave it white unless the highlight itself should be coloured; a
    coloured base already shows its colour in the diffuse part.
- `Texture` (alpha channel) and `Color Texture` (sRGB) vary them across
  the surface; both need UVs.

## Only for non-metals

It acts on the non-metal share of the material only. On `Metallic Factor
1` it does nothing, because metal's reflection colour is its base
colour. Set `Metallic Factor 0` explicitly (it defaults to `1`).

## Specular, roughness or IOR?

- Roughness spreads the highlight.
- Specular `Factor` weakens it.
- `Index Of Refraction` (`ior.md`) also raises or lowers head-on
  reflectance, physically.

For "smooth but not shiny" use specular; for "shiny but blurry" use
roughness.

## Verify

`render_preview` doesn't show it. `describe_manifest` lists
`KHR_materials_specular` with `specularFactor` and `specularColorFactor`.
