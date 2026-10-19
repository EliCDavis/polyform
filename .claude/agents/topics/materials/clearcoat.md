# KHR_materials_clearcoat — a glossy layer on top

A thin, clear, shiny layer over the base material. The base keeps its own
colour and roughness underneath, and the coat adds a sharp second
highlight. Use it for:
- metallic or coloured paint under a clear finish;
- lacquered, varnished or glazed surfaces;
- wet surfaces, coated plastic.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialClearcoatExtensionNode` | `Clearcoat Factor`, `Clearcoat Roughness Factor`, `Clearcoat Texture`, `Clearcoat Roughness Texture`, `Clearcoat Normal Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Clearcoat` port.

- `Clearcoat Factor`, 0 to 1, is how strong the coat is; `1` for a full
  lacquer.
- `Clearcoat Roughness Factor`, 0 to 1, is the coat's own roughness:
  `0-0.05` mirror-like, `0.1-0.2` satin varnish.
- `Clearcoat Texture` (R channel) and `Clearcoat Roughness Texture`
  (G channel) multiply those factors across the surface, e.g. lacquer
  worn off the edges or a scuffed patch. Both need UVs.
- `Clearcoat Normal Texture` is a normal map for the coat alone, built
  the same way as the base one (`gltf.NormalTextureNode`, see
  `texturing-and-color.md`). Use it for a dimpled or brushed texture in
  the coat over a smooth base, or a smooth coat over a bumpy base
  by putting the bumps on the base `Normal Texture` only. Like any
  normal map it needs good UVs, and `render_preview` doesn't show it.

## The point is two different roughnesses

The look comes from a rough or patterned base under a smooth coat. If
the base is already smooth, the coat adds little.

| Finish | Base | Coat |
| --- | --- | --- |
| Metallic base, mirror coat | metallic `0.5-1`, roughness `0.4` | factor `1`, roughness `0.03` |
| Textured non-metal, varnish | colour texture, metallic `0`, roughness `0.6` | factor `1`, roughness `0.1` |
| Flat-coloured non-metal, satin coat | metallic `0`, roughness `0.5` | factor `0.8`, roughness `0.15` |

## Verify

`render_preview` shows only the base metallic and roughness, not the
coat. `describe_manifest` lists `KHR_materials_clearcoat` with
`clearcoatFactor` and `clearcoatRoughnessFactor` on the material, and
`clearcoatNormalTexture` when a coat normal map is wired.
