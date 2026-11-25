package main

import (
	"math"
	"math/rand/v2"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestRoughReflectionSamples checks hemisphere support down to grazing incidence.
// test reports invalid weights and escape directions.
func TestRoughReflectionSamples(test *testing.T) {
	random := rand.New(rand.NewPCG(12, 24))
	for _, roughness := range []float64{1e-8, 0.1, 0.5, 1} {
		for _, cosine := range []float64{1, 0.5, 1e-8, 1e-200} {
			incoming := vec3.T{-math.Sqrt(1 - cosine*cosine), 0, -cosine}
			for range 1024 {
				sample := sampleMetal(roughness, [3]float64{1, 1, 1}, &vec3.UnitZ, &vec3.UnitZ, &incoming, random.Float64)
				if !(sample.weight[0] >= 0 && !math.IsInf(sample.weight[0], 0)) {
					test.Fatalf("roughness %g, cosine %g: invalid weight %v", roughness, cosine, sample.weight)
				}
				if sample.heading != nil && !validSurfaceHeading(&vec3.UnitZ, sample.heading) {
					test.Fatalf("invalid reflected direction %v", sample.heading)
				}
			}
		}
	}
}

// TestRoughReflectionIntegration checks the one-facet contribution against GGX/Smith quadrature.
// test reports differences between the sampled and integrated radiance.
func TestRoughReflectionIntegration(test *testing.T) {
	testCases := []struct {
		name          string
		shadingNormal vec3.T
		viewDirection vec3.T
		roughness     float64
	}{
		{"normal", vec3.T{0, 0, 1}, vec3.T{0, 0, 1}, 0.5},
		{"rough normal", vec3.T{0, 0, 1}, vec3.T{0, 0, 1}, 1},
		{"oblique", vec3.T{0, 0, 1}, vec3.T{math.Sqrt(0.75), 0, 0.5}, 0.5},
		{"rough oblique", vec3.T{0, 0, 1}, vec3.T{math.Sqrt(0.75), 0, 0.5}, 1},
		{"normal mapped", vec3.T{0.6, 0, 0.8}, vec3.T{0, 0, 1}, 0.7},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			expectedRadiance := reflectionReferenceIntegral(&testCase.shadingNormal, &testCase.viewDirection, testCase.roughness)
			incoming := testCase.viewDirection.Inverted()
			random := rand.New(rand.NewPCG(32, 91))

			// The full walk must reproduce the analytic GGX term for paths escaping after one hit.
			const samples = 150000
			var sum float64
			for range samples {
				encounters := 0
				reflectance := func(float64) [3]float64 {
					encounters++
					return [3]float64{1, 1, 1}
				}
				sample := sampleGlossyReflection(&vec3.UnitZ, &testCase.shadingNormal, &incoming, testCase.roughness, reflectance, random.Float64)
				if encounters == 1 && sample.heading != nil {
					sum += sample.weight[0] * reflectionTestIllumination(sample.heading)
				}
			}
			if mean := sum / samples; !(math.Abs(mean-expectedRadiance) < 0.004) {
				test.Fatalf("one-facet radiance %g, want %g", mean, expectedRadiance)
			}
		})
	}
}

// TestTracePathRoughReflection follows a two-facet reflection through the scene tracer.
// test reports incorrect Fresnel products, emission, opacity, or mixture weighting.
func TestTracePathRoughReflection(test *testing.T) {
	base := [3]float64{0.25, 0.5, 0.75}
	material := scene.NewMaterial().M(0.4, 1).C(color.NewColor(base[0], base[1], base[2]))
	material.E(color.NewColor(2, 1, 0.5), 1, false)
	media := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
	ray := &scene.Ray{Origin: &vec3.UnitZ, Heading: &vec3.T{0, 0, -1}}
	camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}

	// Choose metal, reflect down at 60 degrees, and return at normal facet incidence.
	draw := scriptedScatterDraws(test, 0.2, 0.5, 0.75, 0.25, 0.5, 1.0/9, 0.75, 0.999)
	got := tracePathSampled(ray, camera, bounceTestScene(material), 0, media, draw)
	var expected [3]float64
	for channel, albedo := range base {
		firstFresnel := albedo + (1-albedo)*math.Pow(0.5, 5)
		expected[channel] = albedo * (firstFresnel + [3]float64{2, 1, 0.5}[channel])
	}
	assertTransmissionColor(test, got, color.NewColor(expected[0], expected[1], expected[2]))
}

// TestExteriorDielectricReflection checks the two Fresnel factors for a glossy coating.
// test reports an incorrect escape direction or selection weight.
func TestExteriorDielectricReflection(test *testing.T) {
	incoming := vec3.UnitZ.Inverted()
	draw := scriptedScatterDraws(test, 0, 0.5, 0.75, 0.25, 0.5, 1.0/9, 0.75, 0.999)
	sample := sampleOpaqueDielectric(1, [3]float64{1, 1, 1}, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, draw)
	direction := vec3.T{-math.Sqrt(0.75), 0, 0.5}
	if sample.heading == nil || sample.bounce != reflectionBounce || vec3.Distance(sample.heading, &direction) > 1e-12 {
		test.Fatalf("two-facet escape %v, want %v", sample.heading, direction)
	}
	weight := surfaceReferenceFresnel(0.5, 1.5) * surfaceReferenceFresnel(1, 1.5) / 0.05
	for _, actual := range sample.weight {
		if !(math.Abs(actual-weight) < 1e-12) {
			test.Fatalf("weight %g, want %g", actual, weight)
		}
	}
}

// reflectionTestIllumination returns scalar environment radiance for a unit
// outgoingDirection in the +Z hemisphere, varying with its Z and squared X components.
func reflectionTestIllumination(outgoingDirection *vec3.T) float64 {
	return 0.2 + 0.5*(*outgoingDirection)[2] + 0.3*(*outgoingDirection)[0]*(*outgoingDirection)[0]
}

// reflectionReferenceIntegral integrates the single-reflection Smith GGX term in solid angle.
// shadingNormal is the unit shading normal in the upper geometric hemisphere.
// viewDirection is the unit direction toward the observer.
// roughness is the positive perceptual roughness.
// The result is the radiance under reflectionTestIllumination.
func reflectionReferenceIntegral(shadingNormal, viewDirection *vec3.T, roughness float64) float64 {
	var radiance float64
	const cosineSteps, azimuthSteps = 256, 512
	microfacetWidthSquared := math.Pow(roughness, 4)
	viewCosine := vec3.Dot(shadingNormal, viewDirection)
	viewRoot := math.Sqrt(viewCosine*viewCosine + microfacetWidthSquared*(1-viewCosine*viewCosine))

	// Equal cos(theta)/azimuth cells have equal solid angle (PBRT A.5.2).
	for cosineIndex := range cosineSteps {
		elevationCosine := (float64(cosineIndex) + 0.5) / cosineSteps
		for azimuthIndex := range azimuthSteps {
			azimuth := 2 * math.Pi * (float64(azimuthIndex) + 0.5) / azimuthSteps
			outgoingDirection := vec3.T{
				math.Sqrt(1-elevationCosine*elevationCosine) * math.Cos(azimuth),
				math.Sqrt(1-elevationCosine*elevationCosine) * math.Sin(azimuth),
				elevationCosine,
			}
			outgoingCosine := vec3.Dot(shadingNormal, &outgoingDirection)
			if outgoingCosine <= 0 {
				continue
			}

			// Evaluate isotropic GGX D and Smith roots directly (PBRT 9.6.1-9.6.3).
			halfVector := viewDirection.Added(&outgoingDirection)
			halfVector.Normalize()
			halfCosine := vec3.Dot(shadingNormal, &halfVector)
			normalDensity := microfacetWidthSquared / (math.Pi * math.Pow(1+(microfacetWidthSquared-1)*halfCosine*halfCosine, 2))
			outgoingRoot := math.Sqrt(outgoingCosine*outgoingCosine + microfacetWidthSquared*(1-outgoingCosine*outgoingCosine))

			// Integrate GGX/Smith BRDF times cosine independently of the sampler.
			weightedBRDF := normalDensity * outgoingCosine / (2 * (outgoingCosine*viewRoot + viewCosine*outgoingRoot))
			radiance += weightedBRDF * reflectionTestIllumination(&outgoingDirection)
		}
	}

	cellSolidAngle := 2 * math.Pi / (cosineSteps * azimuthSteps)
	return radiance * cellSolidAngle
}
