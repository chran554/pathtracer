package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

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
		materialHole          bool
		textureHole           bool
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
		{name: "material hole", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, materialHole: true, expectedRadianceScale: 1},
		{name: "texture hole", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, textureHole: true, expectedRadianceScale: 1},
		{name: "reflection", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, mirror: true, expectedRadianceScale: 1},
		{name: "total internal reflection", inside: true, incidentCosine: 0.5, outsideIndex: 1, glassIndex: 1.5, expectedRadianceScale: 1},
		{name: "local emission", incidentCosine: 1, outsideIndex: 1, glassIndex: 1.5, localEmission: true, expectedRadianceScale: 4.0 / 9},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(1, !testCase.thin, testCase.glassIndex)
			glass.FresnelMaxGlossiness = 0 // Isolate transport scaling from Fresnel roulette.
			if testCase.materialHole {
				glass.C(color.NewColorRGBA(1, 1, 1, 0))
			}
			if testCase.textureHole {
				fresnelTestAlbedo(glass, 0)
			}
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
			actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &rayHeading}, camera, sceneNode, 0, mediumStack)

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
			glass := scene.NewMaterial().T(1, true, 1.5)
			glass.FresnelMaxGlossiness = 0
			surfaces := []*scene.Sphere{scene.NewSphere(&vec3.T{}, 2, glass)}
			recursionDepth := 2
			if nested {
				water := scene.NewMaterial().T(1, true, 1.33)
				water.FresnelMaxGlossiness = 0
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
			assertTransmissionColor(test, tracePath(ray, camera, sceneNode, 0, mediumStack), environmentRadiance)
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
