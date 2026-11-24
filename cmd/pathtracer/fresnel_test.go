package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestFresnelDielectric checks unpolarized Fresnel reflectance on both sides
// of an interface, plus the renderer's artistic limits. Reference: PBRT 4e, section 9.3.5:
// https://pbr-book.org/4ed/Reflection_Models/Specular_Reflection_and_Transmission
func TestFresnelDielectric(test *testing.T) {
	testCases := []struct {
		name                string
		incidentIndex       float64
		transmittedIndex    float64
		incidentCosine      float64
		expectedReflectance float64
	}{
		{"enter normal", 1, 1.5, 1, 0.04},
		{"enter oblique", 1, 1.5, 0.5, 0.08918671280221276},
		{"exit normal", 1.5, 1, 1, 0.04},
		{"exit oblique", 1.5, 1, math.Sqrt(0.75), 0.05519016729537589},
		{"total internal reflection", 1.5, 1, 0.5, 1},
		{"equal indices", 1.5, 1.5, 0.5, 0},
		{"equal indices grazing", 1.5, 1.5, 0, 0},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			normal := vec3.T{0, 0, 1}
			incomingDirection := vec3.T{math.Sqrt(1 - testCase.incidentCosine*testCase.incidentCosine), 0, -testCase.incidentCosine}
			actualReflectance := FresnelDielectric(-vec3.Dot(&normal, &incomingDirection), testCase.incidentIndex, testCase.transmittedIndex)
			if math.IsNaN(actualReflectance) || math.Abs(actualReflectance-testCase.expectedReflectance) > 1e-12 {
				test.Fatalf("reflectance = %g, want %g", actualReflectance, testCase.expectedReflectance)
			}
		})
	}
}

// TestTracePathGlassFresnel checks reflected-side event frequencies and transmitted
// radiance for rays entering and leaving smooth glass.
func TestTracePathGlassFresnel(test *testing.T) {
	testCases := []struct {
		name                      string
		inside                    bool
		incidentCosine            float64
		outsideIndex              float64
		transparency              float64
		expectedReflectedFraction float64
	}{
		{"enter", false, 1, 1, 1, 0.04},
		{"exit", true, 1, 1, 1, 0.04},
		{"enter oblique", false, 0.5, 1, 1, 0.08918671280221276},
		{"exit oblique", true, math.Sqrt(0.75), 1, 1, 0.05519016729537589},
		{"exit nested medium", true, 1, 1.33, 1, math.Pow((1.5-1.33)/(1.5+1.33), 2)},
		// Half transmission leaves half of the non-Fresnel light for diffuse scattering.
		{"partial transmission", false, 1, 1, 0.5, 0.52},
		{"total internal reflection", true, 0.5, 1, 1, 1},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(testCase.transparency, true, 1.5)

			// Initialize the medium stack before choosing entry or exit.
			air := scene.NewMaterial().T(1, true, 1)
			mediumStack := []*scene.Material{air}
			if testCase.outsideIndex != 1 {
				mediumStack = append(mediumStack, scene.NewMaterial().T(1, true, testCase.outsideIndex))
			}
			rayOrigin := vec3.T{0, 0, 1}
			rayHeading := vec3.T{math.Sqrt(1 - testCase.incidentCosine*testCase.incidentCosine), 0, -testCase.incidentCosine}
			if testCase.inside {
				rayOrigin[2], rayHeading[2] = -1, testCase.incidentCosine
				mediumStack = append(mediumStack, glass)
			}

			sceneNode := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
				Discs: []*scene.Disc{
					scene.NewDisc(&vec3.T{0, 0, 2}, &vec3.UnitZ, 1e9, scene.NewMaterial().E(color.NewColor(1, 0, 0), 1, true)),
					scene.NewDisc(&vec3.T{0, 0, -2}, &vec3.UnitZ, 1e9, scene.NewMaterial().E(color.NewColor(0, 1, 0), 1, true)),
				},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}

			const sampleCount = 50000

			// Radiance conversion (PBRT 9.5.2; see TestTracePathTransmissionRadiance).
			radianceScale := math.Pow(testCase.outsideIndex/1.5, 2)
			if testCase.inside {
				radianceScale = 1 / radianceScale
			}

			var reflectedCount, transmittedRadianceSum float64
			for range sampleCount {
				actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &rayHeading}, camera, sceneNode, 0, mediumStack)
				reflectedRadiance, transmittedRadiance := actualRadiance.R, actualRadiance.G
				if testCase.inside {
					reflectedRadiance, transmittedRadiance = actualRadiance.G, actualRadiance.R
				}
				// Count reflected-side events, including attenuated diffuse paths.
				if reflectedRadiance > 0 && transmittedRadiance == 0 {
					reflectedCount++
				} else if reflectedRadiance != 0 || !(math.Abs(float64(transmittedRadiance)-radianceScale) < 1e-6) {
					test.Fatalf("unexpected reflected/transmitted radiance: %v", actualRadiance)
				}
				if actualRadiance.B != 0 || actualRadiance.A != 1 {
					test.Fatalf("unexpected color or opacity: %v", actualRadiance)
				}
				transmittedRadianceSum += float64(transmittedRadiance)
			}

			actualReflectedFraction := reflectedCount / sampleCount
			// Allow eight standard errors for Bernoulli reflected-side counts.
			tolerance := 8*math.Sqrt(testCase.expectedReflectedFraction*(1-testCase.expectedReflectedFraction)/sampleCount) + 1e-4
			if math.Abs(actualReflectedFraction-testCase.expectedReflectedFraction) > tolerance {
				test.Fatalf("reflected-side fraction = %g, want %g ± %g", actualReflectedFraction, testCase.expectedReflectedFraction, tolerance)
			}

			// A transmitted ray carries the index conversion factor in the transmitted-side channel.
			transmissionProbability := 1 - testCase.expectedReflectedFraction
			expectedRadiance := transmissionProbability * radianceScale
			variance := transmissionProbability * testCase.expectedReflectedFraction * radianceScale * radianceScale
			tolerance = 8*math.Sqrt(variance/sampleCount) + 1e-6
			if actualRadiance := transmittedRadianceSum / sampleCount; math.Abs(actualRadiance-expectedRadiance) > tolerance {
				test.Fatalf("mean transmitted radiance = %g, want %g ± %g", actualRadiance, expectedRadiance, tolerance)
			}
		})
	}
}
