package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

// sampleReflection returns a mirror or GGX reflection sample with Fresnel fixed at one.
// geometricNormal and shadingNormal must be unit normals facing the incident medium;
// incomingDirection must be a unit direction toward the surface. roughness is
// perceptual roughness in [0,1]; values <= 0 select a mirror. radialSample and
// azimuthSample are independent uniform values in [0,1) for disk radius and azimuth,
// ignored for a mirror.
//
// The result contains a unit heading and its BRDF*cosine/PDF weight. For GGX,
// invalid incidence or a rejected direction returns a nil heading and zero weight.
// Mirror samples have weight one and fall back to geometric reflection, then
// geometricNormal, if a heading is invalid.
//
// Mirror reflection: PBRT 4e, section 9.3.3:
// https://pbr-book.org/4ed/Reflection_Models/Specular_Reflection_and_Transmission
//
// GGX sampling: Eric Heitz (2018), "Sampling the GGX Distribution of Visible Normals",
// listing 1, with equal widths on both tangent axes:
// https://jcgt.org/published/0007/04/01/paper.pdf
//
// The BRDF uses height-correlated Smith masking-shadowing; see evaluateRoughReflection.
func sampleReflection(geometricNormal, shadingNormal, incomingDirection *vec3.T, roughness, radialSample, azimuthSample float64) scatteringSample {
	if roughness <= 0 {
		// Ideal specular reflection follows the law of reflection (PBRT 9.3.3).
		heading := getReflectionVector(shadingNormal, incomingDirection)
		heading.Normalize()
		if !validSurfaceHeading(geometricNormal, heading) {
			// A delta reflection has no continuous PDF; retain the geometric mirror fallback.
			heading = getReflectionVector(geometricNormal, incomingDirection)
			heading.Normalize()
			if !validSurfaceHeading(geometricNormal, heading) {
				fallbackHeading := *geometricNormal
				heading = &fallbackHeading
			}
		}

		return scatteringSample{heading, 1}
	}

	viewDirection := incomingDirection.Inverted()
	if vec3.Dot(geometricNormal, &viewDirection) <= 0 || vec3.Dot(shadingNormal, &viewDirection) <= 0 {
		return scatteringSample{}
	}

	// Heitz 3.2: stretch the view in the shading frame to reduce the ellipsoid to a hemisphere.
	tangent, bitangent := shadingBasis(shadingNormal)
	microfacetWidth := reflectionAlpha(roughness)
	stretchedView := vec3.T{
		microfacetWidth * vec3.Dot(&viewDirection, &tangent),
		microfacetWidth * vec3.Dot(&viewDirection, &bitangent),
		vec3.Dot(&viewDirection, shadingNormal),
	}
	stretchedView.Normalize()

	// Heitz 4.1: align the sampling disk with the stretched view; handle normal incidence.
	viewTangent := vec3.T{1, 0, 0}
	if stretchedView[0]*stretchedView[0]+stretchedView[1]*stretchedView[1] > 0 {
		viewTangent = vec3.T{-stretchedView[1], stretchedView[0], 0}
		viewTangent.Normalize()
	}
	viewBitangent := vec3.Cross(&stretchedView, &viewTangent)

	// Heitz 4.2: warp a uniform disk sample to the hemisphere's projected visible area.
	diskRadius, azimuth := math.Sqrt(radialSample), 2*math.Pi*azimuthSample
	diskX, diskY := diskRadius*math.Cos(azimuth), diskRadius*math.Sin(azimuth)
	projectedAreaBlend := (1 + stretchedView[2]) / 2
	diskY = (1-projectedAreaBlend)*math.Sqrt(max(0, 1-diskX*diskX)) + projectedAreaBlend*diskY

	// Heitz 4.3 and 3.4: lift to the hemisphere, then transform the normal back to GGX.
	hemisphereZ := math.Sqrt(max(0, 1-diskX*diskX-diskY*diskY))
	localMicrofacetNormal := vec3.T{
		microfacetWidth * (diskX*viewTangent[0] + diskY*viewBitangent[0] + hemisphereZ*stretchedView[0]),
		microfacetWidth * (diskX*viewTangent[1] + diskY*viewBitangent[1] + hemisphereZ*stretchedView[1]),
		max(0, diskX*viewTangent[2]+diskY*viewBitangent[2]+hemisphereZ*stretchedView[2]),
	}
	localMicrofacetNormal.Normalize()

	// Reflect about the sampled microfacet in world space (Heitz, appendix B).
	microfacetNormal := vec3.T{}
	for axis := range microfacetNormal {
		microfacetNormal[axis] = localMicrofacetNormal[0]*tangent[axis] +
			localMicrofacetNormal[1]*bitangent[axis] + localMicrofacetNormal[2]*shadingNormal[axis]
	}
	heading := getReflectionVector(&microfacetNormal, incomingDirection)
	heading.Normalize()
	if !validSurfaceHeading(geometricNormal, heading) {
		return scatteringSample{}
	}

	brdf, pdf := evaluateRoughReflection(geometricNormal, shadingNormal, &viewDirection, heading, roughness)
	if !(pdf > 0 && brdf > 0) {
		// Rejected directions retain zero contribution; retries would condition and bias the PDF.
		return scatteringSample{}
	}

	// BRDF*cosine/PDF is the path weight for visible-normal sampling (Heitz, eq. 18).
	return scatteringSample{heading, brdf * vec3.Dot(shadingNormal, heading) / pdf}
}

// evaluateRoughReflection returns the unit-Fresnel GGX BRDF and visible-normal
// sampling PDF per unit solid angle for outgoingDirection. geometricNormal and
// shadingNormal are unit normals facing the incident medium; viewDirection and
// outgoingDirection are unit directions pointing away from the surface. roughness
// is perceptual roughness in (0,1]. Both results are zero for roughness <= 0 or if
// either direction lies on or outside either normal's positive hemisphere.
// The PDF is not renormalized after clipping to the geometric hemisphere.
//
// Model: Trowbridge-Reitz (GGX) with height-correlated Smith masking-shadowing,
//
// PBRT 4e, sections 9.6.1-9.6.4:
// https://pbr-book.org/4ed/Reflection_Models/Roughness_Using_Microfacet_Theory
//
// Reflection PDF: Heitz (2018), appendix B, eq. 17 (see sampleReflection).
func evaluateRoughReflection(geometricNormal, shadingNormal, viewDirection, outgoingDirection *vec3.T, roughness float64) (brdf, pdf float64) {
	viewCosine, outgoingCosine := vec3.Dot(shadingNormal, viewDirection), vec3.Dot(shadingNormal, outgoingDirection)
	if roughness <= 0 || viewCosine <= 0 || outgoingCosine <= 0 ||
		vec3.Dot(geometricNormal, viewDirection) <= 0 || vec3.Dot(geometricNormal, outgoingDirection) <= 0 {
		return 0, 0
	}

	halfVector := viewDirection.Added(outgoingDirection)
	if halfVector.LengthSqr() == 0 {
		return 0, 0
	}
	halfVector.Normalize()

	// Isotropic GGX D (PBRT 9.6.1), rearranged to avoid cancellation for narrow lobes.
	microfacetWidth := reflectionAlpha(roughness)
	halfCosine := min(1.0, max(0.0, vec3.Dot(shadingNormal, &halfVector)))
	distributionDenominator := (1-halfCosine)*(1+halfCosine) + microfacetWidth*microfacetWidth*halfCosine*halfCosine
	normalDensity := microfacetWidth * microfacetWidth / (math.Pi * distributionDenominator * distributionDenominator)

	// Height-correlated Smith G2 (PBRT 9.6.3): cancel cosines before grazing-angle division.
	viewRoot, outgoingRoot := ggxProjectedRoot(viewCosine, microfacetWidth), ggxProjectedRoot(outgoingCosine, microfacetWidth)
	brdf = normalDensity / (2 * (outgoingCosine*viewRoot + viewCosine*outgoingRoot))

	// Heitz eq. 17: the visible-normal PDF and reflection Jacobian cancel the half-vector dots.
	pdf = normalDensity / (2 * (viewCosine + viewRoot))
	return
}

// reflectionAlpha returns the GGX microfacet width for perceptual roughness in [0,1].
// The result is roughness squared with a minimum of 1e-4, including at zero.
// The squared mapping follows Filament, section 4.8.3; the floor is a local stability limit:
// https://google.github.io/filament/main/filament.html
func reflectionAlpha(roughness float64) float64 {
	return max(1e-4, roughness*roughness)
}

// ggxProjectedRoot returns the Smith masking-shadowing root
// sqrt(directionCosine² + microfacetWidth²*(1-directionCosine²)). directionCosine
// is the cosine between a unit direction and the shading normal, in [0,1];
// microfacetWidth is the nonnegative GGX alpha.
// Values of directionCosine above one are clamped to one to tolerate roundoff.
// This is the cosine-scaled root in Smith's GGX Lambda (PBRT 4e, section 9.6.2;
// see evaluateRoughReflection), evaluated with Hypot to avoid squaring tiny cosines.
func ggxProjectedRoot(directionCosine, microfacetWidth float64) float64 {
	directionCosine = min(1.0, directionCosine)
	return math.Hypot(directionCosine, microfacetWidth*math.Sqrt(max(0, 1-directionCosine*directionCosine)))
}
