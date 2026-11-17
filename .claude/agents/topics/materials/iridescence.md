# KHR_materials_iridescence — thin-film rainbow sheen

Colours that shift with viewing angle, from a microscopically thin film
on the surface.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialIridescenceExtensionNode` | Factor, Texture, `Index Of Refraction`, `Thickness Minimum`, `Thickness Maximum`, `Thickness Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Iridescence` port.

- **`Factor`**, 0 to 1 (default 0, which is no effect): how strong the
  film is.
  - A film that dominates: `1`.
  - A faint film: `0.3-0.6`.
- **`Index Of Refraction`** is the film's own IOR, not the material's;
  glTF's default is `1.3`. Soap film `1.33`, oil `1.4-1.5`.
- **`Thickness Minimum` / `Thickness Maximum`**, in nanometres (glTF
  defaults 100 and 400), set which colours appear.
  - Without `Thickness Texture` the film is the maximum thickness
    everywhere, so the colours come only from viewing angle.
  - With a texture (G channel, needs UVs) the thickness varies across the
    surface between the two, which gives colour swirls. A noise
    texture works well.
  - Roughly: `200-400` gives vivid colours; above `800` the
    colours wash out.
- `Texture` (R channel) varies the strength across the surface.

## With the rest of the material

- It works on metal and non-metal bases.
- A thin transparent film is transmission (`transmission.md`, thin wall,
  no volume) plus iridescence `1` on a near-white base.
- Keep the base colour fairly dark or neutral when the film should
  dominate; on a bright base the effect is subtle.

## Verify

`render_preview` doesn't show it. `describe_manifest` lists
`KHR_materials_iridescence` with `iridescenceFactor` above 0 and the
thickness range.
