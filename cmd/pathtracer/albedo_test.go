package main

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

func TestSphereAlbedo(t *testing.T) {
	for _, tc := range []struct {
		name    string
		texture *color.Color
		want    *color.Color
	}{
		{"untextured", nil, color.NewColor(0.8, 0.28, 0.14)},
		{"half-strength tint", color.NewColorRGBA(0.5, 0, 1, 0.5), color.NewColor(0.6, 0.14, 0.14)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Matching the surrounding IOR isolates diffuse reflection under unit white light.
			material := scene.NewMaterial().C(color.NewColor(0.8, 0.28, 0.14)).T(0, false, 1)
			if tc.texture != nil {
				addTestProjection(material, scene.TextureTypeAlbedo, tc.texture)
			}
			sphere := scene.NewSphere(&vec3.T{0, 0, -1}, 1, material)
			light := scene.NewSphere(&vec3.T{}, 100, scene.NewMaterial().E(color.White, 1, true))
			scn := scene.NewSceneNode().S(sphere, light)
			assertTransmissionColor(t, bounceTestTrace(scn), tc.want)

			ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0, 0, -1}}
			camera := &scene.Camera{RenderType: scene.Raycasting}
			assertTransmissionColor(t, tracePath(ray, camera, scn, 0, nil), tc.want)

			material.Color.A = 0.5
			material.E(color.NewColorRGBA(2, 1, 0.5, 0.25), 1, true)
			wantEmission := tc.want.Copy().ChannelMultiply(material.Emission).Multiply(0.5 * 0.25)
			assertTransmissionColor(t, bounceTestTrace(scn), wantEmission)
		})
	}
}

func addTestProjection(material *scene.Material, kind scene.TextureType, texel *color.Color) *scene.ImageProjection {
	image := floatimage.NewFloatImage(kind.String(), 1, 1)
	image.Fill(texel)

	projection := scene.NewParallelImageProjection(scene.NewTexture(image, kind), &vec3.T{}, vec3.UnitX, vec3.UnitY)
	projection.Initialize()
	material.P(projection)
	return projection
}
