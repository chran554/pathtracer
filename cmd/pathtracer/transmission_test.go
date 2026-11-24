package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestRoughDielectricBoundary checks immediate escape against analytic Snell and Fresnel values.
func TestRoughDielectricBoundary(t *testing.T) {
	incoming := vec3.UnitZ.Inverted()
	// At normal incidence a GGX disk sample u has slope alpha*sqrt(u/(1-u)).
	// Azimuth 1/4 turns that slope toward +X in the sampler's shading frame.
	const roughness = 0.7
	for _, tc := range []struct {
		name                  string
		incident, transmitted float64
		facetDegrees          float64
	}{
		{"enter", 1, 1.5, 30},
		{"exit", 1.5, 1, 30},
		{"local total internal reflection", 1.5, 1, 43},
	} {
		t.Run(tc.name, func(t *testing.T) {
			angle := tc.facetDegrees * math.Pi / 180
			slope := math.Tan(angle)
			alpha := roughness * roughness
			u := slope * slope / (alpha*alpha + slope*slope)
			refracted, fresnel := dielectricTestReference(0, tc.facetDegrees, tc.incident, tc.transmitted)

			// Probe both sides of the analytic microfacet Fresnel threshold.
			for _, roulette := range []float64{fresnel - 1e-6, min(fresnel+1e-6, math.Nextafter(1, 0))} {
				random := glassTestDraws(t, 0.5, u, 0.25, roulette, math.Nextafter(1, 0))
				glass := scene.NewMaterial().M(0, roughness).T(1, true, 1.5)
				sample := sampleTransmissiveDielectric(glass, roughness, &vec3.UnitZ, &vec3.UnitZ, &incoming, tc.incident, tc.transmitted, random)
				want, bounce := refracted, transmissionBounce
				scale := math.Pow(tc.incident/tc.transmitted, 2)
				if roulette < fresnel {
					want = vec3.T{math.Sin(2 * angle), 0, math.Cos(2 * angle)}
					bounce, scale = reflectionBounce, 1
				}
				if sample.heading == nil || sample.bounce != bounce || vec3.Distance(sample.heading, &want) > 1e-12 {
					t.Fatalf("draw %g: got bounce %d, direction %v; want %d, %v", roulette, sample.bounce, sample.heading, bounce, want)
				}

				// Escape was sampled explicitly. Only the net radiance conversion remains.
				for _, weight := range sample.weight {
					if !(math.Abs(weight-scale) < 1e-12) {
						t.Fatalf("draw %g: weight %g, want %g", roulette, weight, scale)
					}
				}
			}
		})
	}
}

// TestTraceRoughGlassSphere verifies independent entry and exit samples along a known ray.
func TestTraceRoughGlassSphere(t *testing.T) {
	// Send a central ray through a flat entry microfacet. At the opposite pole,
	// choose a 30-degree facet on exit. Only the predicted exit ray hits the lamp.
	glass := scene.NewMaterial().M(0, 1).T(1, true, 1.5)
	exit, _ := dielectricTestReference(0, 30, 1.5, 1)
	target := exit.Added(&vec3.T{0, 0, -1 - epsilonDistance})
	scn := &scene.SceneNode{
		Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 1, glass)},
		Discs:   []*scene.Disc{scene.NewDisc(&target, &vec3.UnitZ, 0.001, scene.NewMaterial().E(color.White, 1, true))},
	}

	// Each hit chooses glass, samples a height, scatters at a facet, then escapes.
	random := glassTestDraws(t,
		0, 0.5, 0, 0, 0.99, 0.99,
		0, 0.5, 0.25, 0.25, 0.99, 0.99,
	)
	ray := &scene.Ray{Origin: &vec3.T{0, 0, 2}, Heading: &vec3.T{0, 0, -1}}
	camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 2}
	got := tracePathSampled(ray, camera, scn, 0, []*scene.Material{scene.NewMaterial().T(1, true, 1)}, random)

	// The two radiance conversions cancel on escape into air.
	assertTransmissionColor(t, got, color.White)
}

// TestTracePathTransmissionRadiance checks radiance-mode transmission weighting
// by (n_incident/n_transmitted)^2, as described in PBRT 4e, section 9.5.2:
// https://pbr-book.org/4ed/Reflection_Models/Dielectric_BSDF
func TestTracePathTransmissionRadiance(test *testing.T) {
	testCases := []struct {
		name                  string
		inside                bool
		incidentCosine        float64
		outsideIndex          float64
		glassIndex            float64
		thin                  bool
		mirror                bool
		localEmission         bool
		expectedRadianceScale float64
	}{
		{name: "enter", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 4.0 / 9},
		{name: "exit", inside: true, incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 2.25},
		{name: "enter oblique", incidentCosine: 0.5, outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 4.0 / 9},
		{name: "exit oblique", inside: true, incidentCosine: math.Sqrt(0.75), outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 2.25},
		{name: "enter from water", incidentCosine: 1, outsideIndex: 1.33, glassIndex: 1.5, expectedRadianceScale: math.Pow(1.33/1.5, 2)},
		{name: "exit into water", inside: true, incidentCosine: 1, outsideIndex: 1.33, glassIndex: 1.5, expectedRadianceScale: math.Pow(1.5/1.33, 2)},
		{name: "matched indices", incidentCosine: 0.5, outsideIndex: 1.5, glassIndex: 1.5, expectedRadianceScale: 1},
		{name: "thin front", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, thin: true, expectedRadianceScale: 1},
		{name: "thin back", inside: true, incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, thin: true, expectedRadianceScale: 1},
		{name: "reflection", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, mirror: true, expectedRadianceScale: 1},
		{name: "total internal reflection", inside: true, incidentCosine: 0.5, outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 1},
		{name: "local emission", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, localEmission: true, expectedRadianceScale: 4.0 / 9},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(1, !testCase.thin, testCase.glassIndex)
			if testCase.mirror {
				glass.T(0, true, testCase.glassIndex).M(1, 0)
			}

			localEmission := color.NewColor(0, 0, 0)
			if testCase.localEmission {
				localEmission = color.NewColor(0.1, 0.2, 0.3)
				glass.E(localEmission, 1, false)
			}

			mediumStack := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
			if testCase.outsideIndex != 1 {
				mediumStack = append(mediumStack, scene.NewMaterial().T(1, true, testCase.outsideIndex))
			}

			rayOrigin := vec3.T{0, 0, 1}
			rayHeading := vec3.T{math.Sqrt(1 - testCase.incidentCosine*testCase.incidentCosine), 0, -testCase.incidentCosine}
			if testCase.inside {
				rayOrigin[2], rayHeading[2] = -1, testCase.incidentCosine
				if !testCase.thin {
					mediumStack = append(mediumStack, glass)
				}
			}

			environmentRadiance := color.NewColor(0.25, 0.5, 2)
			sceneNode := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
				Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 100,
					scene.NewMaterial().E(environmentRadiance, 1, true))},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
			actualRadiance := traceTransmissionPath(&scene.Ray{Origin: &rayOrigin, Heading: &rayHeading}, camera, sceneNode, 0, mediumStack)

			// PBRT 9.5.2 scales transmitted RGB; add local emission afterward and retain opacity.
			expectedRadiance := color.NewColor(
				float64(environmentRadiance.R)*testCase.expectedRadianceScale+float64(localEmission.R),
				float64(environmentRadiance.G)*testCase.expectedRadianceScale+float64(localEmission.G),
				float64(environmentRadiance.B)*testCase.expectedRadianceScale+float64(localEmission.B))
			assertTransmissionColor(test, actualRadiance, expectedRadiance)
		})
	}
}

// TestTracePathTransmissionRoundTrip checks that the radiance-conversion factors
// from TestTracePathTransmissionRadiance cancel through closed and nested boundaries.
func TestTracePathTransmissionRoundTrip(test *testing.T) {
	for _, nested := range []bool{false, true} {
		name := "glass"
		if nested {
			name = "nested glass and water"
		}
		test.Run(name, func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(1, true, 1.5)
			surfaces := []*scene.Sphere{scene.NewSphere(&vec3.T{}, 2, glass)}
			recursionDepth := 2
			if nested {
				water := scene.NewMaterial().M(0, 0).T(1, true, 1.33)
				surfaces = append(surfaces, scene.NewSphere(&vec3.T{}, 1, water))
				recursionDepth = 4
			}

			environmentRadiance := color.NewColor(0.25, 0.5, 2)
			surfaces = append(surfaces, scene.NewSphere(&vec3.T{}, 100, scene.NewMaterial().E(environmentRadiance, 1, true)))
			sceneNode := &scene.SceneNode{Spheres: surfaces}
			ray := &scene.Ray{Origin: &vec3.T{0, 0, 3}, Heading: &vec3.T{0, 0, -1}}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: recursionDepth}

			mediumStack := []*scene.Material{scene.NewMaterial().T(1, true, 1)}

			// Every index factor must cancel when the path returns to its initial medium.
			assertTransmissionColor(test, traceTransmissionPath(ray, camera, sceneNode, 0, mediumStack), environmentRadiance)
		})
	}
}

// assertTransmissionColor compares actualColor with expectedColor, requiring
// RGB differences below 1e-6 and equal opacity. Both colors must be non-nil.
func assertTransmissionColor(test *testing.T, actualColor, expectedColor *color.Color) {
	test.Helper()

	for channelIndex, channelPair := range [][2]float32{{actualColor.R, expectedColor.R}, {actualColor.G, expectedColor.G}, {actualColor.B, expectedColor.B}} {
		if !(math.Abs(float64(channelPair[0]-channelPair[1])) < 1e-6) {
			test.Fatalf("channel %d: radiance = %v, want %v", channelIndex, actualColor, expectedColor)
		}
	}

	if actualColor.A != expectedColor.A {
		test.Fatalf("opacity = %g, want %g", actualColor.A, expectedColor.A)
	}
}

// traceTransmissionPath selects the transmitted branch where it has nonzero
// probability. It leaves Fresnel and PDF weights intact; TIR still reflects.
func traceTransmissionPath(ray *scene.Ray, camera *scene.Camera, scn *scene.SceneNode, depth int, media []*scene.Material) *color.Color {
	return tracePathSampled(ray, camera, scn, depth, media, func() float64 { return math.Nextafter(1, 0) })
}
