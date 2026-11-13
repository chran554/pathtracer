package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestSurfaceBounceProbabilities checks the renderer-specific material mixture
// in surfaceBounceProbabilities, including Fresnel limits and scale invariance.
func TestSurfaceBounceProbabilities(test *testing.T) {
	testCases := []struct {
		name                    string
		componentWeights        [3]float64 // Diffuse, reflection, transmission.
		fresnelReflectance      float64
		maxFresnelReflectance   float64
		totalInternalReflection bool
		expectedProbabilities   bounceProbabilities
	}{
		{"equal base weights", [3]float64{0.5, 0.5, 0}, 0, 1, false, bounceProbabilities{0.5, 0.5, 0}},
		{"equal weights with Fresnel", [3]float64{0.5, 0.5, 0}, 0.04, 1, false, bounceProbabilities{0.48, 0.52, 0}},
		{"three components", [3]float64{0.25, 0.25, 0.5}, 0.04, 1, false, bounceProbabilities{0.24, 0.28, 0.48}},
		{"diffuse only", [3]float64{1, 0, 0}, 0, 1, false, bounceProbabilities{1, 0, 0}},
		{"reflection only", [3]float64{0, 1, 0}, 0.5, 0, false, bounceProbabilities{0, 1, 0}},
		{"transmission only", [3]float64{0, 0, 1}, 0, 1, false, bounceProbabilities{0, 0, 1}},
		{"glass", [3]float64{0, 0, 1}, 0.04, 1, false, bounceProbabilities{0, 0.04, 0.96}},
		{"zero Fresnel limit", [3]float64{0.5, 0.5, 0}, 0.5, 0, false, bounceProbabilities{0.75, 0.25, 0}},
		{"artistic Fresnel limit", [3]float64{0.5, 0.5, 0}, 0.4, 0.2, false, bounceProbabilities{0.62, 0.38, 0}},
		{"grazing", [3]float64{1, 1, 2}, 1, 1, false, bounceProbabilities{0, 1, 0}},
		{"grazing with limit", [3]float64{1, 1, 2}, 1, 0.25, false, bounceProbabilities{0.25, 0.25, 0.5}},
		{"total internal reflection", [3]float64{0.2, 0.3, 0.5}, 1, 0, true, bounceProbabilities{0, 1, 0}},
		{"no components", [3]float64{}, 0.04, 1, false, bounceProbabilities{}},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			// Equal scaling must not change the mixture or introduce absorption.
			for _, weightScale := range []float64{1e-300, 0.1, 1, 10, 1e300} {
				material := scene.NewMaterial()
				material.Diffuse = testCase.componentWeights[0] * weightScale
				material.Glossiness = testCase.componentWeights[1] * weightScale
				material.Transparency = testCase.componentWeights[2] * weightScale
				material.FresnelMaxGlossiness = testCase.maxFresnelReflectance

				actualProbabilities := surfaceBounceProbabilities(material, testCase.fresnelReflectance, testCase.totalInternalReflection)
				actualComponents := [3]float64{actualProbabilities.diffuse, actualProbabilities.reflection, actualProbabilities.transmission}
				expectedComponents := [3]float64{testCase.expectedProbabilities.diffuse, testCase.expectedProbabilities.reflection, testCase.expectedProbabilities.transmission}
				for componentIndex, probability := range actualComponents {
					if probability < 0 || probability > 1 || !(math.Abs(probability-expectedComponents[componentIndex]) < 1e-14) {
						test.Fatalf("scale %g: probabilities = %v, want %v", weightScale, actualProbabilities, testCase.expectedProbabilities)
					}
				}
				if testCase.componentWeights != [3]float64{} && math.Abs(actualProbabilities.diffuse+actualProbabilities.reflection+actualProbabilities.transmission-1) > 1e-14 {
					test.Fatalf("probabilities do not sum to one: %v", actualProbabilities)
				}
			}
		})
	}
}

// TestBounceSelection checks cumulative-probability boundaries and disabled
// components in bounceProbabilities.sample.
func TestBounceSelection(test *testing.T) {
	mixedProbabilities := bounceProbabilities{0.25, 0.5, 0.25}
	testCases := []struct {
		name            string
		probabilities   bounceProbabilities
		selectionSample float64
		expectedBounce  bounceType
	}{
		{"start", mixedProbabilities, 0, diffuseBounce},
		{"before reflection", mixedProbabilities, math.Nextafter(0.25, 0), diffuseBounce},
		{"reflection boundary", mixedProbabilities, 0.25, reflectionBounce},
		{"before transmission", mixedProbabilities, math.Nextafter(0.75, 0), reflectionBounce},
		{"transmission boundary", mixedProbabilities, 0.75, transmissionBounce},
		{"end", mixedProbabilities, math.Nextafter(1, 0), transmissionBounce},
		{"no reflection", bounceProbabilities{0.25, 0, 0.75}, 0.25, transmissionBounce},
		{"no diffuse", bounceProbabilities{0, 0.5, 0.5}, 0, reflectionBounce},
		// Deliberate rounding gaps must not enable transmission or stop scattering.
		{"rounded reflection end", bounceProbabilities{0.5, 0.5 - 1e-15, 0}, math.Nextafter(1, 0), reflectionBounce},
		{"rounded diffuse end", bounceProbabilities{1 - 1e-15, 0, 0}, math.Nextafter(1, 0), diffuseBounce},
		{"empty start", bounceProbabilities{}, 0, noBounce},
		{"empty end", bounceProbabilities{}, math.Nextafter(1, 0), noBounce},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			if actualBounce := testCase.probabilities.sample(testCase.selectionSample); actualBounce != testCase.expectedBounce {
				test.Fatalf("sample(%g) = %v, want %v", testCase.selectionSample, actualBounce, testCase.expectedBounce)
			}
		})
	}
}

// TestTracePathWhiteMixtures checks brightness for white mixtures with matched refractive indices.
func TestTracePathWhiteMixtures(test *testing.T) {
	testCases := []struct {
		name     string
		material *scene.Material
	}{
		{"diffuse", scene.NewMaterial()},
		{"reflection", scene.NewMaterial().M(1, 0)},
		{"glass", scene.NewMaterial().T(1, true, 1.5)},
		{"half reflection", scene.NewMaterial().M(0.5, 0)},
		{"three components", scene.NewMaterial().M(0.25, 0).T(0.5, true, 1.5)},
		{"overlapping weights", scene.NewMaterial().M(0.8, 0).T(0.8, true, 1.5)},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, weightScale := range []float64{0.1, 1, 10} {
				material := testCase.material.Copy()
				material.RefractionIndex = 1 // Isolate mixture weights from refractive radiance conversion.
				material.Diffuse *= weightScale
				material.Glossiness *= weightScale
				material.Transparency *= weightScale

				sceneNode := bounceTestScene(material)
				for range 100 {
					actualRadiance := bounceTestTrace(sceneNode)
					if actualRadiance.R != 1 || actualRadiance.G != 1 || actualRadiance.B != 1 || actualRadiance.A != 1 {
						test.Fatalf("scale %g: white mixture changed brightness: %v", weightScale, actualRadiance)
					}
				}
			}
		})
	}
}

// TestTracePathMixtureAlbedo checks mixture ratios and absorption from material and texture RGB.
func TestTracePathMixtureAlbedo(test *testing.T) {
	testCases := []struct {
		name                   string
		componentWeights       [3]float64
		tintReflection         bool
		materialCoverage       float64
		textureCoverage        float64
		expectedTintedFraction float64
	}{
		{"diffuse", [3]float64{1, 0, 0}, false, 1, 1, 1},
		{"untinted reflection", [3]float64{0, 1, 0}, false, 1, 1, 0},
		{"tinted reflection", [3]float64{0, 1, 0}, true, 1, 1, 1},
		{"transmission", [3]float64{0, 0, 1}, false, 1, 1, 0},
		{"equal weights", [3]float64{0.5, 0.5, 0}, false, 1, 1, 0.5},
		{"small total", [3]float64{0.05, 0.05, 0}, false, 1, 1, 0.5},
		{"large total", [3]float64{5, 5, 0}, false, 1, 1, 0.5},
		{"three components", [3]float64{0.2, 0.3, 0.5}, false, 1, 1, 0.2},
		{"three components tinted", [3]float64{0.2, 0.3, 0.5}, true, 1, 1, 0.5},
		{"material and texture holes", [3]float64{1, 0, 0}, false, 0.5, 0.5, 0.25},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			material := scene.NewMaterial().C(color.NewColorRGBA(0.5, 0.25, 0.75, testCase.materialCoverage))
			material.Diffuse, material.Glossiness, material.Transparency = testCase.componentWeights[0], testCase.componentWeights[1], testCase.componentWeights[2]
			material.ColorizeReflection = testCase.tintReflection
			material.Roughness = 0 // Isolate mixture and albedo from directional BRDF weighting.
			material.FresnelMaxGlossiness = 0
			material.RefractionIndex = 1 // Match the surrounding medium to isolate base mixture weights.

			// A constant texture makes the expected material-times-texture albedo explicit.
			texture := floatimage.NewFloatImage("test albedo", 1, 1)
			texture.Fill(color.NewColorRGBA(0.5, 0.5, 0.25, testCase.textureCoverage))
			projection := scene.NewParallelImageProjection(texture, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitY)
			projection.Initialize()
			material.P(&projection)

			sceneNode := bounceTestScene(material)
			const sampleCount = 30000
			var radianceSum [3]float64
			for range sampleCount {
				actualRadiance := bounceTestTrace(sceneNode)
				for channelIndex, channelRadiance := range [3]float32{actualRadiance.R, actualRadiance.G, actualRadiance.B} {
					if !(channelRadiance >= 0 && channelRadiance <= 1) {
						test.Fatalf("invalid passive radiance: %v", actualRadiance)
					}
					radianceSum[channelIndex] += float64(channelRadiance)
				}
			}

			for channelIndex, albedo := range [3]float64{0.25, 0.125, 0.1875} {
				// Uncolored reflection, transmission, and holes return the white environment.
				tintedFraction := testCase.expectedTintedFraction
				expectedMean := 1 - tintedFraction*(1-albedo)
				// Tinted paths are Bernoulli trials; allow eight standard errors plus roundoff.
				tolerance := 8*(1-albedo)*math.Sqrt(tintedFraction*(1-tintedFraction)/sampleCount) + 1e-6
				if actualMean := radianceSum[channelIndex] / sampleCount; math.Abs(actualMean-expectedMean) > tolerance {
					test.Fatalf("channel %d: mean = %g, want %g ± %g", channelIndex, actualMean, expectedMean, tolerance)
				}
			}
		})
	}
}

// TestTracePathNoScattering checks opaque termination, local emission, and alpha holes with zero weights.
func TestTracePathNoScattering(test *testing.T) {
	testCases := []struct {
		name             string
		coverage         float64
		emission         *color.Color
		expectedRadiance color.Color
	}{
		{"opaque black", 1, nil, color.Color{A: 1}},
		{"local emission", 1, color.NewColor(2, 1, 0.5), color.Color{R: 2, G: 1, B: 0.5, A: 1}},
		{"alpha hole", 0, nil, color.Color{R: 1, G: 1, B: 1, A: 1}},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			material := scene.NewMaterial().C(color.NewColorRGBA(1, 1, 1, testCase.coverage)).E(testCase.emission, 1, false)
			material.Diffuse, material.Glossiness, material.Transparency = 0, 0, 0

			// A white environment exposes accidental continuation through the covered surface.
			actualRadiance := bounceTestTrace(bounceTestScene(material))
			if *actualRadiance != testCase.expectedRadiance {
				test.Fatalf("zero-weight result = %v, want %v", actualRadiance, testCase.expectedRadiance)
			}
		})
	}
}

// bounceTestScene returns a new scene with a triangle at z=0 using material,
// facing +Z, and surrounded by a spherical emitter with unit white radiance.
func bounceTestScene(material *scene.Material) *scene.SceneNode {
	return &scene.SceneNode{
		FacetStructures: []*scene.FacetStructure{orientationTestPlane(material)},
		Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 100,
			scene.NewMaterial().E(color.White, 1, true))},
	}
}

// bounceTestTrace returns the radiance and opacity of one path through sceneNode,
// starting at (0,0,1) toward -Z with refractive index one and recursion depth one.
func bounceTestTrace(sceneNode *scene.SceneNode) *color.Color {
	ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0, 0, -1}}
	camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
	return tracePath(ray, camera, sceneNode, 0, []*scene.Material{scene.NewMaterial().T(1, true, 1)})
}
