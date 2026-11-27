package main

import (
	"math"
	"math/rand/v2"
	"testing"

	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestThinWallPaths follows known paths through both faces of a rough wall.
// t reports the exit direction, side, and radiance weight of each path.
func TestThinWallPaths(t *testing.T) {
	glass := scene.NewMaterial().M(0, 1).T(1, false, 1.5)
	incoming := vec3.UnitZ.Inverted()
	refracted, _ := dielectricTestReference(0, 30, 1.5, 1)
	for _, path := range []struct {
		name      string
		draws     []float64
		bounce    bounceType
		direction vec3.T
	}{
		// The second face refracts at a 30-degree facet, spreading transmitted light.
		{"transmission", []float64{0.5, 0, 0, 0.99, 0.99, 0.5, 0.25, 0.25, 0.99, 0.99}, transmissionBounce, refracted},
		// Enter the wall, reflect off its back face, and exit the entry face.
		{"internal reflection", []float64{0.5, 0, 0, 0.99, 0.99, 0.5, 0, 0, 0, 0.99, 0.5, 0, 0, 0.99, 0.99}, reflectionBounce, vec3.UnitZ},
	} {
		t.Run(path.name, func(t *testing.T) {
			draw := scriptedScatterDraws(t, path.draws...)
			sample := sampleTransmissiveDielectric(glass, 1, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, draw)
			if sample.heading == nil || sample.bounce != path.bounce || vec3.Distance(sample.heading, &path.direction) > 1e-12 || sample.weight != [3]float64{1, 1, 1} {
				t.Fatalf("got %+v, want direction %v with unit weight", sample, path.direction)
			}
		})
	}
}

// TestThinWallSmoothLimit checks two-face Fresnel reflectance and straight transmission.
// t reports the reflection frequency and angular error of transmitted rays.
func TestThinWallSmoothLimit(t *testing.T) {
	thin := scene.NewMaterial().T(1, false, 1.5)
	for _, cosine := range []float64{1, 0.5} {
		incoming := vec3.T{math.Sqrt(1 - cosine*cosine), 0, -cosine}
		fresnel := surfaceReferenceFresnel(cosine, 1.5)
		want := 2 * fresnel / (1 + fresnel)

		for _, roughness := range []float64{0, 0.001} {
			random := rand.New(rand.NewPCG(12, 46))
			const samples = 50000
			var reflected, transmitted, spread float64
			for range samples {
				sample := sampleTransmissiveDielectric(thin, roughness, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, random.Float64)
				if sample.heading == nil {
					continue
				}
				if sample.bounce == reflectionBounce {
					reflected += sample.weight[0]
				} else {
					transmitted += sample.weight[0]
					spread += sample.weight[0] * vec3.Distance(sample.heading, &incoming)
				}
			}
			tolerance := 6*math.Sqrt(want*(1-want)/samples) + 0.0001
			if !(math.Abs(reflected/samples-want) < tolerance && math.Abs(transmitted/samples-(1-want)) < tolerance && spread/transmitted < 0.001) {
				t.Fatalf("roughness %g, cosine %g: R %g, T %g, mean deflection %g, want R %g", roughness, cosine, reflected/samples, transmitted/samples, spread/transmitted, want)
			}
		}
	}
}

// TestRoughThinWallInMedium checks unchanged surrounding absorption after transmission.
// t reports any unexpected refraction, absorption, or change of medium.
func TestRoughThinWallInMedium(t *testing.T) {
	surrounding := scene.NewMaterial().T(1, true, 1.33).Absorb(color.NewColor(math.Ln2, 0, 0))
	incoming := vec3.UnitZ.Inverted()
	for _, index := range []float64{1.33, 1.5} {
		glass := scene.NewMaterial().C(color.WhiteTransparent).M(0, 1).T(1, false, index).Absorb(color.NewColorGrey(100))
		emitter := scene.NewMaterial().E(color.White, 1, true)
		world := &scene.SceneNode{Discs: []*scene.Disc{
			scene.NewDisc(&vec3.T{}, &vec3.UnitZ, 5, glass),
			scene.NewDisc(&vec3.T{0, 0, -1}, &vec3.UnitZ, 5, emitter),
		}}

		// Matching indices bypass both faces. Otherwise force normal transmission at both.
		draws := []float64{0}
		if index != surrounding.RefractionIndex {
			draws = append(draws, 0.5, 0, 0, 0.99, 0.99, 0.5, 0, 0, 0.99, 0.99)
		}
		ray := &scene.Ray{Origin: &vec3.UnitZ, Heading: &incoming}
		camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
		got := tracePathSampled(ray, camera, world, 0, []*scene.Material{surrounding}, scriptedScatterDraws(t, draws...))
		want := color.NewColor(math.Exp(-math.Ln2*(2-epsilonDistance)), 1, 1)
		assertTransmissionColor(t, got, want)
	}
}

// TestRoughThinWallAgainstSlab compares one thin hit against two explicit glass faces.
// t reports the direction-dependent lighting and white-furnace energy from each side.
func TestRoughThinWallAgainstSlab(t *testing.T) {
	const roughness = 0.8
	thin := scene.NewMaterial().C(color.WhiteTransparent).M(0, roughness).T(1, false, 1.5)
	solid := scene.NewMaterial().C(color.WhiteTransparent).M(0, roughness).T(1, true, 1.5)
	backNormal := vec3.UnitZ.Inverted()
	wall := &scene.SceneNode{Discs: []*scene.Disc{scene.NewDisc(&vec3.T{}, &vec3.UnitZ, 1e9, thin)}}
	slab := &scene.SceneNode{Discs: []*scene.Disc{
		scene.NewDisc(&vec3.T{}, &vec3.UnitZ, 1e9, solid),
		scene.NewDisc(&vec3.T{0, 0, -0.001}, &backNormal, 1e9, solid),
	}}

	// Red identifies the front hemisphere. Green and blue split the back hemisphere
	// into a central cone and its complement. Their sum measures furnace energy.
	lights := []*scene.Disc{
		scene.NewDisc(&vec3.T{0, 0, 100}, &vec3.UnitZ, 1e12, scene.NewMaterial().E(color.NewColor(1, 0, 0), 1, true)),
		scene.NewDisc(&vec3.T{0, 0, -100}, &vec3.UnitZ, 50, scene.NewMaterial().E(color.NewColor(0, 1, 0), 1, true)),
		scene.NewDisc(&vec3.T{0, 0, -101}, &vec3.UnitZ, 1e12, scene.NewMaterial().E(color.NewColor(0, 0, 1), 1, true)),
	}
	wall.Discs = append(wall.Discs, lights...)
	slab.Discs = append(slab.Discs, lights...)
	media := []*scene.Material{scene.NewMaterial().T(1, true, 1)}

	for _, side := range []float64{-1, 1} {
		ray := &scene.Ray{Origin: &vec3.T{0, 0, side}, Heading: &vec3.T{0, 0, -side}}
		const samples = 50000
		collapsed := meanTestPathRadiance(ray, wall, media, 1, samples)
		explicit := meanTestPathRadiance(ray, slab, media, 96, samples)
		t.Logf("incident side %g: thin %v, explicit faces %v", side, collapsed, explicit)
		for channel := range collapsed {
			if !(math.Abs(collapsed[channel]-explicit[channel]) < 0.012 && collapsed[channel] > 0.001) {
				t.Fatalf("thin wall %v, explicit slab %v", collapsed, explicit)
			}
		}
		if !(math.Abs(collapsed[0]+collapsed[1]+collapsed[2]-1) < 0.006) {
			t.Fatalf("thin wall lost furnace energy: %v", collapsed)
		}
	}
}

// TestTexturedThinWallRoughness checks the effective roughness at a scene intersection.
// t reports a missed lamp when the texture selects smooth or rough transmission.
func TestTexturedThinWallRoughness(t *testing.T) {
	for _, textureValue := range []float64{0, 0.5} {
		image := floatimage.NewFloatImage("wall roughness", 1, 1)
		image.Fill(color.NewColorGrey(textureValue))
		projection := scene.NewParallelImageProjection(scene.NewTexture(image, scene.TextureTypeRoughness), &vec3.T{}, vec3.UnitX, vec3.UnitY)
		projection.Initialize()
		wall := scene.NewMaterial().C(color.WhiteTransparent).M(0, 1).T(1, false, 1.5).P(projection)

		// Aim a small lamp along the analytically predicted exit direction.
		target := vec3.T{0, 0, -1}
		draws := []float64{0, 0.99}
		if textureValue > 0 {
			target, _ = dielectricTestReference(0, 30, 1.5, 1)
			slope := math.Tan(math.Pi / 6)
			width := textureValue * textureValue
			radialDraw := slope * slope / (slope*slope + width*width)
			draws = []float64{0, 0.5, 0, 0, 0.99, 0.99, 0.5, radialDraw, 0.25, 0.99, 0.99}
		}
		world := &scene.SceneNode{Discs: []*scene.Disc{
			scene.NewDisc(&vec3.T{}, &vec3.UnitZ, 10, wall),
			scene.NewDisc(&target, &vec3.UnitZ, 0.001, scene.NewMaterial().E(color.White, 1, true)),
		}}
		ray := &scene.Ray{Origin: &vec3.UnitZ, Heading: &vec3.T{0, 0, -1}}
		camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
		got := tracePathSampled(ray, camera, world, 0, []*scene.Material{scene.NewMaterial().T(1, true, 1)}, scriptedScatterDraws(t, draws...))
		assertTransmissionColor(t, got, color.White)
	}
}
