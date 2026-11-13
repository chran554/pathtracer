package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestRoughReflectionBRDF checks analytic GGX values, reciprocity, and hemisphere
// support. See evaluateRoughReflection for the GGX/Smith model references.
func TestRoughReflectionBRDF(test *testing.T) {
	normal := vec3.T{0, 0, 1}
	for _, outgoingCosine := range []float64{1, 0.5, 0.01} {
		outgoingDirection := vec3.T{math.Sqrt(1 - outgoingCosine*outgoingCosine), 0, outgoingCosine}
		brdf, pdf := evaluateRoughReflection(&normal, &normal, &normal, &outgoingDirection, 1)
		// GGX with correlated Smith G2 simplifies at alpha=1 and normal incidence
		// to f=1/(2*pi*(1+cos)) and p=1/(4*pi); see evaluateRoughReflection.
		expectedBRDF, expectedPDF := 1/(2*math.Pi*(1+outgoingCosine)), 1/(4*math.Pi)
		if math.Abs(brdf-expectedBRDF) > 1e-12 || math.Abs(pdf-expectedPDF) > 1e-12 {
			test.Fatalf("cosine %g: f/pdf = %g/%g, want %g/%g", outgoingCosine, brdf, pdf, expectedBRDF, expectedPDF)
		}
	}

	// Reciprocity must hold when the view and outgoing directions are exchanged.
	viewDirection, outgoingDirection := vec3.T{0.6, 0, 0.8}, vec3.T{0.3, 0.4, math.Sqrt(0.75)}
	forwardBRDF, _ := evaluateRoughReflection(&normal, &normal, &viewDirection, &outgoingDirection, 0.7)
	reverseBRDF, _ := evaluateRoughReflection(&normal, &normal, &outgoingDirection, &viewDirection, 0.7)
	if !(math.Abs(forwardBRDF-reverseBRDF) < 1e-12) {
		test.Fatalf("nonreciprocal BRDF: %g versus %g", forwardBRDF, reverseBRDF)
	}

	shadingNormal := vec3.T{0.8, 0, 0.6}
	for _, direction := range []vec3.T{{0.8, 0, -0.6}, {-0.8, 0, 0.6}, {1, 0, 0}} {
		// Either the geometric or shading hemisphere excludes these directions.
		brdf, pdf := evaluateRoughReflection(&normal, &shadingNormal, &normal, &direction, 0.7)
		if brdf != 0 || pdf != 0 {
			test.Fatalf("unsupported direction %v has f/pdf %g/%g", direction, brdf, pdf)
		}
	}
}

// TestRoughReflectionSamples checks Heitz visible-normal samples for finite
// weights and rejection without retries. See sampleReflection for the source.
func TestRoughReflectionSamples(test *testing.T) {
	normal, incomingDirection := vec3.T{0, 0, 1}, vec3.T{0, 0, -1}
	sample := sampleReflection(&normal, &normal, &incomingDirection, 1, 0.25, 0.37)
	if sample.heading == nil || math.Abs((*sample.heading)[2]-0.5) > 1e-12 || math.Abs(sample.weight-2.0/3) > 1e-12 {
		test.Fatalf("analytic GGX sample = %v, want cosine 0.5 and weight 2/3", sample)
	}

	// This microfacet reflects below the surface; it must not be replaced by another sample.
	rejectedSample := sampleReflection(&normal, &normal, &incomingDirection, 1, 0.75, 0.37)
	if rejectedSample.heading != nil || rejectedSample.weight != 0 {
		test.Fatalf("rejected sample acquired a contribution: %v", rejectedSample)
	}

	// Probe narrow lobes and grazing incidence with deterministic disk samples.
	for _, roughness := range []float64{1e-8, 0.1, 0.5, 1} {
		for _, incidentCosine := range []float64{1, 0.5, 1e-8, 1e-200} {
			grazingIncomingDirection := vec3.T{-math.Sqrt(1 - incidentCosine*incidentCosine), 0, -incidentCosine}

			// Fractional inverse-golden-ratio multiples spread the sample azimuths.
			for sampleIndex := range 1024 {
				reflectionSample := sampleReflection(&normal, &normal, &grazingIncomingDirection, roughness, (float64(sampleIndex)+0.5)/1024, math.Mod(float64(sampleIndex)*0.6180339887498949, 1))
				if !(reflectionSample.weight >= 0 && reflectionSample.weight <= 1+1e-12) {
					test.Fatalf("roughness %g, cosine %g: invalid weight %g", roughness, incidentCosine, reflectionSample.weight)
				}
				if reflectionSample.heading != nil && !validSurfaceHeading(&normal, reflectionSample.heading) {
					test.Fatalf("invalid reflected direction: %v", reflectionSample.heading)
				}
			}
		}
	}
}

// TestRoughReflectionIntegration compares Heitz visible-normal sampling with
// independent GGX/Smith quadrature in reflectionReferenceIntegral.
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
			expectedRadiance, expectedAcceptanceProbability := reflectionReferenceIntegral(&testCase.shadingNormal, &testCase.viewDirection, testCase.roughness)
			geometricNormal, incomingDirection := vec3.T{0, 0, 1}, testCase.viewDirection.Inverted()

			// Sample-space midpoint grids keep this integration check reproducible.
			const samplesPerAxis = 512
			var meanRadiance, acceptanceProbability float64
			for radialIndex := range samplesPerAxis {
				for azimuthIndex := range samplesPerAxis {
					sample := sampleReflection(&geometricNormal, &testCase.shadingNormal, &incomingDirection, testCase.roughness, (float64(radialIndex)+0.5)/samplesPerAxis, (float64(azimuthIndex)+0.5)/samplesPerAxis)
					if sample.heading == nil {
						continue
					}
					acceptanceProbability++
					meanRadiance += sample.weight * reflectionTestIllumination(sample.heading)
				}
			}

			// Normalize by all trials, including rejected samples with zero contribution.
			meanRadiance /= samplesPerAxis * samplesPerAxis
			acceptanceProbability /= samplesPerAxis * samplesPerAxis
			if math.Abs(meanRadiance-expectedRadiance) > 0.003 || math.Abs(acceptanceProbability-expectedAcceptanceProbability) > 0.003 {
				test.Fatalf("radiance/accepted = %g/%g, want %g/%g", meanRadiance, acceptanceProbability, expectedRadiance, expectedAcceptanceProbability)
			}
		})
	}
}

// TestTracePathRoughReflection checks GGX/Smith weighting, mixture selection, tint,
// and local emission. See evaluateRoughReflection for the model references.
func TestTracePathRoughReflection(test *testing.T) {
	testCases := []struct {
		name           string
		glossiness     float64
		tintReflection bool
		emissive       bool
	}{
		{"reflection", 1, false, false},
		{"tinted reflection", 1, true, false},
		{"local emission", 1, true, true},
		{"diffuse mixture", 0.5, false, false},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			material := scene.NewMaterial().M(testCase.glossiness, 1)
			material.RefractionIndex = 1 // Keep the selected mixture independent of Fresnel.
			material.ColorizeReflection = testCase.tintReflection
			if testCase.tintReflection {
				material.C(color.NewColor(0.25, 0.5, 0.75))
			}
			if testCase.emissive {
				material.E(color.NewColor(2, 1, 0.5), 1, false)
			}

			sceneNode := bounceTestScene(material)
			const sampleCount = 30000
			var radianceSum [3]float64
			for range sampleCount {
				actualRadiance := bounceTestTrace(sceneNode)
				if actualRadiance.A != 1 {
					test.Fatalf("reflection changed opacity: %v", actualRadiance)
				}
				for channelIndex, channelRadiance := range [3]float32{actualRadiance.R, actualRadiance.G, actualRadiance.B} {
					radianceSum[channelIndex] += float64(channelRadiance)
				}
			}

			// At alpha=1 and normal incidence, GGX/Smith reflectance is
			// integral_0^1 cos/(1+cos) dcos = 1-ln(2); see TestRoughReflectionBRDF.
			integratedReflectance := 1 - math.Log(2)
			for channelIndex, albedo := range [3]float32{material.Color.R, material.Color.G, material.Color.B} {
				expectedMean := (1-testCase.glossiness)*float64(albedo) + testCase.glossiness*integratedReflectance
				if testCase.tintReflection {
					expectedMean = integratedReflectance * float64(albedo)
				}
				if testCase.emissive {
					expectedMean += [3]float64{2, 1, 0.5}[channelIndex] * float64(albedo)
				}
				// The random scattering term lies in [0,1]; eight standard errors use variance <= 1/4.
				if actualMean := radianceSum[channelIndex] / sampleCount; !(math.Abs(actualMean-expectedMean) < 4/math.Sqrt(sampleCount)) {
					test.Fatalf("channel %d: radiance = %g, want %g", channelIndex, actualMean, expectedMean)
				}
			}
		})
	}
}

// reflectionTestIllumination returns scalar environment radiance for a unit
// outgoingDirection in the +Z hemisphere, varying with its Z and squared X components.
func reflectionTestIllumination(outgoingDirection *vec3.T) float64 {
	return 0.2 + 0.5*(*outgoingDirection)[2] + 0.3*(*outgoingDirection)[0]*(*outgoingDirection)[0]
}

// reflectionReferenceIntegral returns numerically integrated unit-Fresnel GGX
// radiance under reflectionTestIllumination and acceptedProbability, the sampling
// PDF mass in the +Z hemisphere. shadingNormal is a unit normal in the +Z hemisphere;
// viewDirection is a unit direction away from the surface, inside both the +Z and
// shadingNormal hemispheres. roughness is in [0.01,1], where the production
// microfacet-width floor is inactive.
//
// Midpoint quadrature uses cos(theta)/azimuth cells from PBRT 4e, section A.5.2:
// https://pbr-book.org/4ed/Sampling_Algorithms/Sampling_Multidimensional_Functions
// GGX/Smith and visible-normal PDF sources are documented in evaluateRoughReflection.
func reflectionReferenceIntegral(shadingNormal, viewDirection *vec3.T, roughness float64) (radiance, acceptedProbability float64) {
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

			// GGX/Smith f*cos and Heitz's reflection PDF, evaluated without production helpers
			// (PBRT 9.6 and Heitz eq. 17; see evaluateRoughReflection).
			weightedBRDF := normalDensity * outgoingCosine / (2 * (outgoingCosine*viewRoot + viewCosine*outgoingRoot))
			radiance += weightedBRDF * reflectionTestIllumination(&outgoingDirection)
			acceptedProbability += normalDensity / (2 * (viewCosine + viewRoot))
		}
	}

	cellSolidAngle := 2 * math.Pi / (cosineSteps * azimuthSteps)
	return radiance * cellSolidAngle, acceptedProbability * cellSolidAngle
}
