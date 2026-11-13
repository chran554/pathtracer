package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestFresnelReflectAmount checks unpolarized Fresnel reflectance on both sides
// of an interface, plus the renderer's artistic limits. Reference: PBRT 4e, section 9.3.5:
// https://pbr-book.org/4ed/Reflection_Models/Specular_Reflection_and_Transmission
func TestFresnelReflectAmount(test *testing.T) {
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
			actualReflectance := FresnelReflectAmount(testCase.incidentIndex, testCase.transmittedIndex, &normal, &incomingDirection, 0, 1)
			if math.IsNaN(actualReflectance) || math.Abs(actualReflectance-testCase.expectedReflectance) > 1e-12 {
				test.Fatalf("reflectance = %g, want %g", actualReflectance, testCase.expectedReflectance)
			}

			// Retain the material's artistic reflectance limits.
			actualReflectance = FresnelReflectAmount(testCase.incidentIndex, testCase.transmittedIndex, &normal, &incomingDirection, 0.2, 0.8)
			expectedReflectance := 0.2*(1-testCase.expectedReflectance) + 0.8*testCase.expectedReflectance
			if math.IsNaN(actualReflectance) || math.Abs(actualReflectance-expectedReflectance) > 1e-12 {
				test.Fatalf("adjusted reflectance = %g, want %g", actualReflectance, expectedReflectance)
			}
		})
	}
}

// TestTracePathGlassFresnel checks Fresnel event frequencies, transmission radiance,
// and alpha coverage. See TestFresnelReflectAmount and TestTracePathTransmissionRadiance
// for the optical references; alpha coverage follows the renderer's material model.
func TestTracePathGlassFresnel(test *testing.T) {
	testCases := []struct {
		name                      string
		inside                    bool
		incidentCosine            float64
		outsideIndex              float64
		transparency              float64
		materialCoverage          float64
		textureCoverage           float64
		expectedReflectedFraction float64
	}{
		{"enter", false, 1, 1, 1, 1, 1, 0.04},
		{"exit", true, 1, 1, 1, 1, 1, 0.04},
		{"enter oblique", false, 0.5, 1, 1, 1, 1, 0.08918671280221276},
		{"exit oblique", true, math.Sqrt(0.75), 1, 1, 1, 1, 0.05519016729537589},
		{"exit nested medium", true, 1, 1.33, 1, 1, 1, math.Pow((1.5-1.33)/(1.5+1.33), 2)},
		{"material coverage", false, 1, 1, 1, 0.5, 1, 0.02},
		{"texture coverage", false, 1, 1, 1, 1, 0.5, 0.02},
		{"combined coverage", false, 1, 1, 1, 0.5, 0.5, 0.01},
		// Half transmission leaves half of the non-Fresnel light for diffuse scattering.
		{"partial transmission", false, 1, 1, 0.5, 1, 1, 0.52},
		{"partial transmission and coverage", false, 1, 1, 0.5, 0.5, 0.5, 0.13},
		{"total internal reflection", true, 0.5, 1, 1, 1, 1, 1},
		{"total internal reflection and holes", true, 0.5, 1, 1, 0.5, 0.5, 0.25},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(testCase.transparency, true, 1.5)
			glass.C(color.NewColorRGBA(1, 1, 1, testCase.materialCoverage))
			fresnelTestAlbedo(glass, testCase.textureCoverage)

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
			coverage := testCase.materialCoverage * testCase.textureCoverage

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
				// Reflected-side paths (including diffuse) have unit radiance; refraction alone scales it.
				if reflectedRadiance == 1 && transmittedRadiance == 0 {
					reflectedCount++
				} else if reflectedRadiance != 0 || !(math.Abs(float64(transmittedRadiance)-radianceScale) < 1e-6 || (coverage < 1 && transmittedRadiance == 1)) {
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

			// Count holes separately so the index factor cannot accidentally scale alpha coverage.
			transmissionProbability := coverage - testCase.expectedReflectedFraction
			expectedRadiance := 1 - coverage + transmissionProbability*radianceScale
			variance := max(0, 1-coverage+transmissionProbability*radianceScale*radianceScale-expectedRadiance*expectedRadiance)
			tolerance = 8*math.Sqrt(variance/sampleCount) + 1e-6
			if actualRadiance := transmittedRadianceSum / sampleCount; math.Abs(actualRadiance-expectedRadiance) > tolerance {
				test.Fatalf("mean transmitted radiance = %g, want %g ± %g", actualRadiance, expectedRadiance, tolerance)
			}
		})
	}
}

// TestTracePathAlphaHole checks that uncovered glass preserves direction and the
// medium stack. Its emitter position uses Snell's law (see dielectricTestReference).
func TestTracePathAlphaHole(test *testing.T) {
	testCases := []struct {
		name             string
		materialCoverage float64
		textureCoverage  float64
		terminator       bool
	}{
		{"material", 0, 1, false},
		{"texture", 1, 0, false},
		{"terminator", 0, 1, true},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			holeMaterial := scene.NewMaterial().T(1, true, 1.5).C(color.NewColorRGBA(1, 1, 1, testCase.materialCoverage))
			holeMaterial.RayTerminator = testCase.terminator
			fresnelTestAlbedo(holeMaterial, testCase.textureCoverage)

			glass := scene.NewMaterial().T(1, true, 1.5)
			// Force transmission at the second interface to isolate the hole's effect.
			glass.FresnelMaxGlossiness = 0
			secondPlane := orientationTestPlane(glass)
			secondPlane.Translate(&vec3.T{0, 0, -1})

			rayOrigin, rayHeading := vec3.T{0, 0, 1}, vec3.T{0.5, 0, -math.Sqrt(0.75)}

			// Snell's law: travel two Z units in air, then one in glass with sin(thetaT)=1/3.
			lightPosition := vec3.T{2*0.5/math.Sqrt(0.75) + (1.0/3)/math.Sqrt(8.0/9), 0, -2}

			sceneNode := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(holeMaterial), secondPlane},
				Discs: []*scene.Disc{scene.NewDisc(&lightPosition, &vec3.UnitZ, 0.005,
					scene.NewMaterial().E(color.White, 1, true))},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 2}
			mediumStack := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
			actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &rayHeading}, camera, sceneNode, 0, mediumStack)

			// Only the second, covered interface changes the refractive medium and radiance.
			assertTransmissionColor(test, actualRadiance, color.NewColor(4.0/9, 4.0/9, 4.0/9))
		})
	}
}

// fresnelTestAlbedo appends an initialized, constant white albedo texture to
// material. textureAlpha is its coverage in [0,1].
func fresnelTestAlbedo(material *scene.Material, textureAlpha float64) {
	texture := floatimage.NewFloatImage("test coverage", 1, 1)
	texture.Fill(color.NewColorRGBA(1, 1, 1, textureAlpha))

	projection := scene.NewParallelImageProjection(texture, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitY)
	projection.Initialize()
	material.P(&projection)
}
