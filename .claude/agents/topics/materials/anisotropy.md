# KHR_materials_anisotropy — brushed and streaked highlights

Stretches the highlight along one direction, as on brushed or machined
metal. Use it for metal with a visible grain; plain polished or matte metal doesn't need it.

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.MaterialAnisotropyExtensionNode` | `Anisotropy Strength`, `Anisotropy Rotation`, `Anisotropy Texture` | Out |

Each `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
Wire `Out` into `MaterialNode`'s `Anisotropy` port.

- `Anisotropy Strength` is 0 to 1; `0.5-0.8` reads as brushed metal.
- `Anisotropy Rotation` turns the streak, in radians, away from the UV
  `u` direction. `0` streaks along `u`; `1.5708` (90°) streaks along `v`.
- `Anisotropy Texture` varies it per texel: red and green hold the
  streak direction relative to `u` and `v` (each -1 to 1, stored as 0 to
  1), blue the strength. It is linear data, not colour.

## The direction comes from the UVs, so they must be good

The viewer orients the streak by the mesh's UVs. Where they are missing
or flattened, it has no direction to use and shades the surface
blown-out white.

- **Use it only on parts whose UVs run along the brushing.**
  - A sweep (`extrude.OutlineNode`, `CircleAlongSplineNode`,
    `ScrewNode`): `u` runs along the path, so `0` brushes along the
    sweep.
  - A primitive's own UVs (cube, cylinder).
  - `meshops.CylinderProjectUVNode` on a cylindrical part: `u` runs
    around the axis, so `0` brushes around it and `1.5708` along it.
  - `meshops.BoxProjectUVNode` on a blocky part.
- **Never on a planar `ProjectUVNode` projection** that includes faces
  edge-on to it, such as a box's sides projected from above. Those faces
  get zero-area UVs.
- **Not on a mesh combined from parts with and without UVs.**
  `meshops.CombineNode` fills the missing UVs with zeros.
- Radial or concentric brushing on a flat round face needs
  `Anisotropy Texture`, with each texel's red/green pointing around the
  circle. No node draws such a direction image yet, so unless you can
  build one, fake the look with `metallicRoughness` texture detail and
  keep the scalar strength low.

## Typical values

Brushed steel: `Metallic Factor 1`, `Roughness Factor 0.3-0.45`, strength
`0.6`. Brushed aluminium: a lighter base colour, roughness `0.35`,
strength `0.7`. Without some roughness there is nothing to stretch.

## Verify

`render_preview` does not show it. `describe_manifest`:
- `KHR_materials_anisotropy` in `extensionsUsed`;
- **no warning** about zero-area UVs or missing `TEXCOORD_0` on the
  meshes using it. A warning there means the part will glow white in a
  viewer; fix the UVs or drop the extension from that part.
