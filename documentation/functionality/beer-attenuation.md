# Absorption inside filled materials

The path tracer uses Beer–Lambert attenuation for each ray segment in a
homogeneous, non-scattering medium:

```
T_rgb(d) = exp(-sigma_a_rgb * d)
L_at_origin_rgb = T_rgb(d) * L_at_hit_rgb
```

Here `d` is world-space distance in scene units and `sigma_a` is the RGB
absorption coefficient in inverse scene units. It is nonnegative, has no upper
bound of one, and is independent of refractive index and surface transparency.
Beer attenuation scales RGB radiance along the segment. Surface tint and
emission are evaluated at the hit as described in the [material guide](metallic-roughness.md).

This is a physical law, not just a visual trick. Using three RGB coefficients
instead of wavelength-dependent spectra is a practical approximation suitable
for this renderer. Scientific results would also require calibrated spectra,
length units, material data, and the other relevant transport processes.
See [PBRT, Transmittance](https://pbr-book.org/4ed/Volume_Scattering/Transmittance).

## The previous Beer function

The old function was equivalent to `sigma_a_rgb = 0.15 * (1 - color_rgb)`.
Its exponential distance dependence was correct, but that conversion from a
surface color was an artistic choice, and no transport code called the function.
The constant `0.15` was an absorption scale in inverse scene units, not a
universal optical density. For a channel, optical depth is `sigma_a * d`;
the old final exponent `(1 - color) * k` was its negative.

The revised `Beer` function takes absorption coefficients directly. To reproduce
the old function's curve, set coefficients to `0.15 * (1 - oldColor)` explicitly.
There is no automatic conversion of existing surface colors.

## Material settings

`scene.Material` adds one optional field:

```go
Absorption *color.Color // sigma_a; inverse scene units; alpha ignored
```

Nil or zero means no volume absorption, preserving existing scenes. Set it using
`Absorb(coefficients)` or the easier `Attenuate(transmission, distance)` helper:

```go
glass := scene.NewMaterial().
    N("green tinted glass").
    T(1, true, scene.RefractionIndex_Glass).
    Attenuate(color.NewColor(0.45, 0.90, 0.55), 10)
```

Ten scene units through this material transmit 45% of red, 90% of green, and
55% of blue before surface losses. Twenty units transmit the square of each
value. `Attenuate` stores `sigma_a = -ln(transmission) / distance`. Supply linear
RGB transmission values in `(0,1]` and a finite positive reference distance.
For nearly complete absorption, use a small positive transmission value.
The helpers reject invalid inputs with a panic at scene construction time.

`Color` still controls surface albedo and tinted reflections. `Transparency`
still selects optical transmission at a boundary; it is not a per-unit-distance
transmission factor. Setting absorption alone does not make an opaque material
transparent. `SolidObject` must be true to enter/exit a filled object's medium.
Use closed surfaces with outward-facing geometric normals. A hollow glass shell
needs inner and outer boundaries (or an explicitly modeled clear interior).

| Material | Appropriate treatment |
| --- | --- |
| Vacuum | Zero absorption; refractive index 1. |
| Clear air at ordinary scene distances | Usually zero absorption is sufficient visually. Atmospheric scattering is a separate effect. |
| Clear water, glass, transparent plastic | Small or zero absorption plus the material's refractive index and surface settings. |
| Tinted glass or transparent plastic | Unequal RGB absorption coefficients. A green appearance generally means greater red/blue absorption. |
| Opaque materials | Keep optical transmission disabled; surface reflection and diffuse albedo determine appearance. |
| Milky water, wax, cloudy plastic | Absorption alone is insufficient; volume/subsurface scattering is needed. |

Lengths must be consistent: if one scene unit is one centimeter, a coefficient
quoted per meter must be multiplied by `0.01`. Enlarging an object increases its
absorption unless its coefficients or reference distance are adjusted.

## Transport and scope

The top of the existing medium stack selects absorption. Every segment is
attenuated using its incident medium, including segments ending at emissive or
opaque surfaces and segments before/after reflection or total internal reflection.
The stack changes only on accepted transmission through a solid boundary.
Nested solids replace the enclosing medium locally; their coefficients are not
added together. Thin surfaces keep the current medium.

Attenuation is applied after gathering the hit's emission and scattered radiance,
without changing alpha. The existing small ray-origin offset slightly shortens
the measured segment. This numerical approximation becomes relevant if geometry
is comparable in size to that offset or absorption is extremely strong.

Absorption is included in the render archive's material/color tables; archives
without the new field load with no absorption. Raycasting preview mode is
unchanged. Media must be properly nested or disjoint; arbitrary overlapping
solids are not resolved by the existing stack.

## World medium and cameras inside objects

`scene.Frame.WorldMedium` specifies the medium outside bounded objects. Nil
defaults to air (`RefractionIndex_Air`) with zero absorption. Both the world and
the camera's initial media reuse `scene.Material`; only `RefractionIndex` and
`Absorption` affect transport along a segment. Surface settings still belong
to the material at a boundary.

`scene.Camera.InitialMedia` lists the materials of objects containing the camera,
ordered **outermost to innermost**, excluding the world medium. For example, a
camera inside glass inside a bounded water volume uses:

```go
water := scene.NewMaterial().
    T(1, true, scene.RefractionIndex_Water).
    Attenuate(color.NewColor(0.8, 0.95, 0.98), 100)
glass := scene.NewMaterial().
    T(1, true, scene.RefractionIndex_Glass).
    Attenuate(color.NewColor(0.45, 0.90, 0.55), 10)

// Assign these same materials to the enclosing objects.
camera.InitialMedia = []*scene.Material{water, glass}
frame := scene.NewFrame("underwater", 0, camera, sceneRoot)
// frame.WorldMedium is nil: effective stack is [air, water, glass].
```

Exiting glass returns to water; exiting water returns to global air. To model
unbounded water without an outer water boundary, set:

```go
frame.WorldMedium = water
camera.InitialMedia = nil // The camera starts in water: [water].
// If a glass object contains the camera instead:
camera.InitialMedia = []*scene.Material{glass} // [water, glass].
```

`Frame.InitialMediumStack()` constructs and validates the initial stack. A missing
camera, nil camera medium entry, nonpositive/nonfinite refractive index, or
negative/nonfinite absorption coefficient returns an error. The world entry is
never popped. Rendering builds this template once per frame, then gives every
ordinary or debug camera ray a private copy of its slice storage. Materials are
shared and must remain read-only during rendering.

Both settings are serialized with the frame and camera through the existing
material/color tables, including media referenced only by these settings. Loading
preserves their order and shared material references. Older archives without
these fields default to air and an empty camera list; invalid new material
references return errors.

The renderer does not infer containment from geometry. Update `InitialMedia`
for each frame when an animated camera crosses a boundary. All aperture sample
origins must lie inside the declared enclosing volumes, and the camera must not
start exactly on a boundary. Surface normals and the declared stack must agree
with the actual nesting; only numeric validity is checked automatically.

## Mist and distant mountains

Beer attenuation also describes loss from a beam in a scattering medium, using
extinction `sigma_t = sigma_a + sigma_s`. However, applying attenuation alone
only darkens distant objects. Mist and atmospheric haze also scatter illumination
*into* the viewing ray. See [PBRT, The Equation of Transfer](https://pbr-book.org/4ed/Light_Transport_II_Volume_Rendering/The_Equation_of_Transfer).

A useful visual approximation for uniform, uniformly lit haze is:

```
T = exp(-sigma_t * distance)
L = T * L_surface + (1 - T) * L_haze
```

This makes distant mountains lose contrast and approach the haze/sky color.
It assumes a constant source term and does not reproduce volumetric shadows,
sunbeams, directional scattering, or varying fog density. Height-dependent haze
needs integrated optical depth, `T = exp(-integral(sigma_t(s) ds))`.

These haze/scattering features are not part of the absorption implementation.
For full mist transport, introduce a shared medium description with absorption,
RGB scattering coefficients, and a phase-function parameter (for example
Henyey–Greenstein `g`). Add free-flight sampling, scattering events and their
illumination, using the camera/world settings above. A simple artistic
haze mode could instead use an extinction coefficient and constant haze radiance
with the formula above. Surface `Emission` cannot supply the missing light along
the whole volume.
