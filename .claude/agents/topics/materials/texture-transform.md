# KHR_texture_transform — tile, shift and turn a texture

Repeats, offsets or rotates a texture on the surface without touching the
mesh's UVs. Use it to:
- tile a small seamless texture across a big part;
- turn a grain to run the other way;
- line a texture up with a feature.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.TextureNode` | Image, Sampler, Offset, Rotation, Scale | Out |
| `formats/gltf.TextureReferenceNode` | URI, Sampler, Offset, Rotation, Scale | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.

- `Scale` (`vector2`) is how many times the texture repeats along `u`
  and `v`: `{"x":4,"y":4}` tiles it 4 by 4.
- `Offset` (`vector2`) shifts it, in repeats.
- `Rotation` (radians) turns it counter-clockwise.

Leave all three unwired and no transform is written.

## Rules

- **The transform belongs to the texture node.** Every material slot fed
  from one `TextureNode` gets the same transform. A colour map and its
  matching normal map come from two texture nodes, so give both the
  same `Scale`, `Offset` and `Rotation`, or they drift apart.
- **Tiling needs a repeating sampler.** glTF repeats by default. A
  `SamplerNode` set to clamp stops the repeat.
- **It is not a substitute for UVs.** A mesh with no UVs or zero-area UVs
  still shows nothing useful; fix those first (see
  `texturing-and-color.md`).

## Verify

`render_preview` samples the raw UVs and ignores the transform.
`describe_manifest` lists `KHR_texture_transform` in `extensionsUsed`; the
web editor's viewer applies it.
