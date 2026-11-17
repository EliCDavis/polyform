# Lights: KHR_lights_punctual

| PATH | inputs | outputs |
| --- | --- | --- |
| `formats/gltf.LightNode` | Type, Color, Intensity, Range, `Inner Cone Angle`, `Outer Cone Angle` | Out |
| `formats/gltf.ModelNode` | ..., Light | Out |

The `PATH` goes inside
`github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/PATH]`.
A light is placed by wiring it into a `ModelNode`'s `Light` port; it sits
at that model's translation. Spot and directional lights shine along the
model's local -Z, so rotate the model to aim them. With `Gpu Instances`,
every instance gets its own copy of the light.

## Intensity: real units

`Intensity` is what the glTF spec says: candela for point and spot,
lux for directional. Use real-world values. Other tools read them as
physical units, so a made-up scale only works in one viewer.

The web viewer shows a model's lights through an exposure of EV100 7.7,
a normal interior exposure. Its own sun and ambient are equivalent to
about 550 and 250 lux at that exposure. Size each light from the
illuminance it should put on the surface it is for:

    Intensity (cd) = E (lux) × d²

`d` is the distance in metres from the light to that surface.

- `E` around `200-500` lux for a light that clearly lights its surface,
  `50-150` for an accent or fill. Above about `1000` reads as blown out
  at the default exposure.
- Lights whose pools overlap add up. With `N` lights on one surface,
  divide `E` by `N`.
- A directional light standing in for daylight is `10000-100000` lux.
  That only reads at a daylight exposure (about EV100 15), where interior
  lights all but vanish. Do not mix a daylight sun with interior lights.
- Set `Range` to two or three times `d` so the light stops before it
  floods the rest of the scene. `0` means unlimited.
- Expose `E` as the variable, not raw candela, when `d` is itself a
  variable, and compute `E × d²` in the graph.

## Light and emission

The visible bulb or tube is an emissive material (`materials/emissive-strength.md`);
the light is a separate `LightNode`. Keep the light's `Color` and the
emissive colour the same so the glow matches what it casts.

## Verify

`render_preview` draws no punctual lights, only emissive surfaces, so a
render will not show a blow-out. Check every light's
`Intensity / d²` by hand against the `E` range above before reporting.
`describe_manifest` lists each model's `light` with its type, colour,
intensity, range and cone angles.
