package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestTracePathDiffuseNormals checks clipped Lambertian radiance, albedo, emission,
// and both surface sides. See evaluateDiffuse for the Lambertian model and PDF.
func TestTracePathDiffuseNormals(test *testing.T) {
	testCases := []struct {
		name          string
		shadingNormal vec3.T
		backFacing    bool
		coloredAlbedo bool
		emissive      bool
		mirrorMixture bool
	}{
		{"flat", vec3.T{0, 0, 1}, false, false, false, false},
		{"tilted", vec3.T{0.6, 0, 0.8}, false, false, false, false},
		{"strong tilt", vec3.T{math.Sqrt(0.99), 0, 0.1}, false, false, false, false},
		{"back facing", vec3.T{0.6, 0, 0.8}, true, false, false, false},
		{"colored", vec3.T{0.6, 0, 0.8}, false, true, false, false},
		{"emissive", vec3.T{0.6, 0, 0.8}, false, true, true, false},
		{"mirror mixture", vec3.T{0.6, 0, 0.8}, false, false, false, true},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			material := normalMappedTestMaterial(testCase.shadingNormal)
			material.FresnelMaxGlossiness = 0 // Isolate diffuse transport from reflection selection.
			if testCase.mirrorMixture {
				material.M(0.5, 0)
			}
			channelAlbedo := [3]float64{1, 1, 1}
			if testCase.coloredAlbedo {
				material.C(color.NewColor(0.5, 0.25, 0.75))
				texture := floatimage.NewFloatImage("diffuse test albedo", 1, 1)
				texture.Fill(color.NewColor(0.5, 0.5, 0.25))
				projection := scene.NewParallelImageProjection(texture, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitY)
				projection.Initialize()
				material.P(&projection)
				channelAlbedo = [3]float64{0.25, 0.125, 0.1875}
			}
			if testCase.emissive {
				material.E(color.NewColor(2, 1, 0.5), 1, false)
			}

			plane := orientationTestPlane(material)
			// Only the incident side is lit; leaked rays would see black on the other side.
			lightPositionZ := 2.0
			rayOrigin, rayHeading := vec3.T{0, 0, 1}, vec3.T{0, 0, -1}
			if testCase.backFacing {
				lightPositionZ = -2
				rayOrigin.Invert()
				rayHeading.Invert()
			}
			sceneNode := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{plane},
				Discs: []*scene.Disc{scene.NewDisc(&vec3.T{0, 0, lightPositionZ}, &vec3.UnitZ, 1e12,
					scene.NewMaterial().E(color.White, 1, true))},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
			mediumStack := []*scene.Material{scene.NewMaterial()}

			const sampleCount = 30000
			var radianceSum [3]float64
			for range sampleCount {
				actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &rayHeading}, camera, sceneNode, 0, mediumStack)
				if actualRadiance.A != 1 {
					test.Fatalf("rejected diffuse sample changed opacity: %v", actualRadiance)
				}
				for channelIndex, channelRadiance := range [3]float32{actualRadiance.R, actualRadiance.G, actualRadiance.B} {
					radianceSum[channelIndex] += float64(channelRadiance)
				}
			}

			// Integrating max(0, Ns·w)/pi over Ng·w>0 gives acceptance (1 + Ng·Ns)/2.
			// This is the fraction of Malley's disk covered by the hemisphere overlap.
			acceptanceProbability := (1 + testCase.shadingNormal[2]) / 2
			if testCase.mirrorMixture {
				acceptanceProbability = 0.5 + 0.5*acceptanceProbability // The mirror component reaches the light with unit weight.
			}
			for channelIndex, reflectance := range channelAlbedo {
				expectedMean := acceptanceProbability * reflectance
				if testCase.emissive {
					expectedMean += [3]float64{2, 1, 0.5}[channelIndex] * reflectance
				}
				// Eight standard errors for accepted/rejected paths, scaled by channel albedo.
				tolerance := 8*reflectance*math.Sqrt(acceptanceProbability*(1-acceptanceProbability)/sampleCount) + 1e-6
				if actualMean := radianceSum[channelIndex] / sampleCount; !(math.Abs(actualMean-expectedMean) < tolerance) {
					test.Fatalf("channel %d: mean radiance = %g, want %g ± %g", channelIndex, actualMean, expectedMean, tolerance)
				}
			}
		})
	}
}

// TestDiffuseBRDF checks the Lambertian BRDF and cosine PDF from evaluateDiffuse,
// including reciprocity and geometric/shading hemisphere support.
func TestDiffuseBRDF(test *testing.T) {
	geometricNormal, shadingNormal := vec3.T{0, 0, 1}, vec3.T{0.6, 0, 0.8}
	for _, outgoingDirection := range []vec3.T{shadingNormal, geometricNormal, {0, 0.6, 0.8}} {
		brdf, pdf := evaluateDiffuse(&geometricNormal, &shadingNormal, &geometricNormal, &outgoingDirection)
		outgoingCosine := vec3.Dot(&shadingNormal, &outgoingDirection)
		if math.Abs(brdf-1/math.Pi) > 1e-14 || math.Abs(pdf-outgoingCosine/math.Pi) > 1e-14 {
			test.Fatalf("direction %v: f/pdf = %g/%g", outgoingDirection, brdf, pdf)
		}
		if math.Abs(brdf*outgoingCosine/pdf-1) > 1e-14 {
			test.Fatalf("Lambertian BRDF*cos/PDF did not cancel: %g", brdf*outgoingCosine/pdf)
		}
		reverseBRDF, _ := evaluateDiffuse(&geometricNormal, &shadingNormal, &outgoingDirection, &geometricNormal)
		if reverseBRDF != brdf {
			test.Fatalf("nonreciprocal diffuse BRDF: %g versus %g", brdf, reverseBRDF)
		}
	}

	for _, excludedDirection := range []vec3.T{{0.96, 0, -0.28}, {-0.96, 0, 0.28}, {1, 0, 0}} {
		// Reject either direction when it lies outside either required hemisphere.
		for _, directionPair := range [][2]vec3.T{{geometricNormal, excludedDirection}, {excludedDirection, geometricNormal}} {
			brdf, pdf := evaluateDiffuse(&geometricNormal, &shadingNormal, &directionPair[0], &directionPair[1])
			if brdf != 0 || pdf != 0 {
				test.Fatalf("unsupported view/heading %v: f/pdf = %g/%g", directionPair, brdf, pdf)
			}
		}
	}
}

// TestDiffuseSamples checks Malley samples for valid directions, unit weights,
// and rejection without fallback. See sampleDiffuse for the algorithm reference.
func TestDiffuseSamples(test *testing.T) {
	geometricNormal, shadingNormal, incomingDirection := vec3.T{0, 0, 1}, vec3.T{0.6, 0, 0.8}, vec3.T{0, 0, -1}
	rejectedSample := sampleDiffuse(&geometricNormal, &shadingNormal, &incomingDirection, 0.99, 0.25)
	if rejectedSample.heading != nil || rejectedSample.weight != 0 {
		test.Fatalf("clipped sample was replaced instead of rejected: %v", rejectedSample)
	}

	for _, normalPair := range [][2]vec3.T{
		{geometricNormal, geometricNormal},
		{geometricNormal, shadingNormal},
		{{0, 0, -1}, {0.6, 0, -0.8}},
		{{1, 0, 0}, {0.8, 0.6, 0}},
	} {
		incomingDirection := normalPair[0].Inverted()
		originalNormals := normalPair
		for sampleIndex := range 1024 {
			// Include the center and near-rim samples; inverse-golden-ratio multiples spread azimuths.
			sample := sampleDiffuse(&normalPair[0], &normalPair[1], &incomingDirection, float64(sampleIndex)/1024, math.Mod(float64(sampleIndex)*0.6180339887498949, 1))
			if sample.heading == nil {
				if sample.weight != 0 {
					test.Fatalf("empty sample has nonzero weight: %v", sample)
				}
				continue
			}
			if !validSurfaceHeading(&normalPair[0], sample.heading) || !(vec3.Dot(&normalPair[1], sample.heading) > 0) || !(math.Abs(sample.weight-1) < 1e-12) {
				test.Fatalf("invalid diffuse direction or weight: %v", sample)
			}
		}

		if normalPair != originalNormals {
			test.Fatal("sampling mutated surface normals")
		}
	}

	// Invalid incidence must not manufacture a fallback direction.
	for _, incomingDirection := range []vec3.T{{0, 0, 1}, {-1, 0, 0}} {
		sample := sampleDiffuse(&geometricNormal, &shadingNormal, &incomingDirection, 0, 0)
		if sample.heading != nil || sample.weight != 0 {
			test.Fatalf("invalid incidence produced a sample: %v", sample)
		}
	}
}

// TestDiffuseIntegration compares Malley sampling against diffuseReferenceIntegral
// and the analytic cosine-weighted mass of the overlapping hemispheres.
func TestDiffuseIntegration(test *testing.T) {
	testCases := []struct {
		name          string
		shadingNormal vec3.T
		backFacing    bool
	}{
		{"aligned", vec3.T{0, 0, 1}, false},
		{"tilted", vec3.T{0.6, 0, 0.8}, false},
		{"strong tilt", vec3.T{math.Sqrt(0.9999), 0, 0.01}, false},
		{"back facing", vec3.T{0.6, 0, -0.8}, true},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			geometricNormal := vec3.T{0, 0, 1}
			if testCase.backFacing {
				geometricNormal.Invert()
			}
			incomingDirection := geometricNormal.Inverted()

			// A grid of sample-space midpoints makes the distribution check reproducible.
			const samplesPerAxis = 512
			var acceptanceProbability, meanRadiance float64
			for radialIndex := range samplesPerAxis {
				for azimuthIndex := range samplesPerAxis {
					sample := sampleDiffuse(&geometricNormal, &testCase.shadingNormal, &incomingDirection, (float64(radialIndex)+0.5)/samplesPerAxis, (float64(azimuthIndex)+0.5)/samplesPerAxis)
					if sample.heading != nil {
						acceptanceProbability++
						meanRadiance += sample.weight * diffuseTestIllumination(&geometricNormal, sample.heading)
					}
				}
			}

			// Keep rejected samples in the denominator; their contribution is zero.
			acceptanceProbability /= samplesPerAxis * samplesPerAxis
			meanRadiance /= samplesPerAxis * samplesPerAxis

			// The overlap integral is (1 + Ng·Ns)/2; radiance uses independent quadrature.
			expectedAcceptanceProbability := (1 + vec3.Dot(&geometricNormal, &testCase.shadingNormal)) / 2
			expectedRadiance := diffuseReferenceIntegral(&geometricNormal, &testCase.shadingNormal)
			if math.Abs(acceptanceProbability-expectedAcceptanceProbability) > 0.002 || math.Abs(meanRadiance-expectedRadiance) > 0.002 {
				test.Fatalf("accepted/radiance = %g/%g, want %g/%g", acceptanceProbability, meanRadiance, expectedAcceptanceProbability, expectedRadiance)
			}
		})
	}
}

// diffuseTestIllumination returns scalar environment radiance for the unit
// outgoingDirection in the positive hemisphere of the unit geometricNormal.
// The radiance varies with inclination to geometricNormal and squared X component.
func diffuseTestIllumination(geometricNormal, outgoingDirection *vec3.T) float64 {
	return 0.2 + 0.5*vec3.Dot(geometricNormal, outgoingDirection) + 0.3*outgoingDirection[0]*outgoingDirection[0]
}

// diffuseReferenceIntegral returns numerically integrated unit-albedo Lambertian
// radiance under diffuseTestIllumination. geometricNormal must be +Z or -Z;
// shadingNormal must be a unit normal facing the same hemisphere.
//
// Midpoint quadrature uses equal-solid-angle cells in cos(theta) and azimuth
// (PBRT 4e, section A.5.2), independent of the production Malley sampler:
// https://pbr-book.org/4ed/Sampling_Algorithms/Sampling_Multidimensional_Functions
// The Lambertian model is documented in evaluateDiffuse.
func diffuseReferenceIntegral(geometricNormal, shadingNormal *vec3.T) float64 {
	const cosineSteps, azimuthSteps = 256, 512
	var radiance float64

	// Equal steps in cos(theta) and azimuth give equal solid angles (PBRT A.5.2).
	for cosineIndex := range cosineSteps {
		elevationCosine := (float64(cosineIndex) + 0.5) / cosineSteps
		for azimuthIndex := range azimuthSteps {
			azimuth := 2 * math.Pi * (float64(azimuthIndex) + 0.5) / azimuthSteps
			outgoingDirection := vec3.T{
				math.Sqrt(1-elevationCosine*elevationCosine) * math.Cos(azimuth),
				math.Sqrt(1-elevationCosine*elevationCosine) * math.Sin(azimuth),
				elevationCosine * geometricNormal[2],
			}

			// Lambertian f*cos = max(0, Ns·w)/pi; see evaluateDiffuse for the model.
			weightedBRDF := max(0, vec3.Dot(shadingNormal, &outgoingDirection)) / math.Pi
			radiance += weightedBRDF * diffuseTestIllumination(geometricNormal, &outgoingDirection)
		}
	}

	// Multiply the midpoint sum by each cell's solid angle.
	return radiance * (2 * math.Pi / (cosineSteps * azimuthSteps))
}
