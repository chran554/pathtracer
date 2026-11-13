package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

// sampleDiffuse returns a cosine-weighted sample of a unit-albedo Lambertian BRDF.
//
// geometricNormal and shadingNormal must be unit normals facing the incident medium.
// incomingDirection must be a unit direction toward the surface.
// radialSample and azimuthSample are independent uniform values in [0,1) for disk radius and azimuth.
// The result contains a unit heading and its BRDF*cosine/PDF weight (one when accepted).
// Invalid incidence or a heading outside the geometric hemisphere returns a nil
// heading and zero weight, without resampling.
//
// Sampling uses Malley's method with a polar disk map (PBRT 4e, sections A.5.1 and A.5.3):
// https://pbr-book.org/4ed/Sampling_Algorithms/Sampling_Multidimensional_Functions
func sampleDiffuse(geometricNormal, shadingNormal, incomingDirection *vec3.T, radialSample, azimuthSample float64) scatteringSample {
	viewDirection := incomingDirection.Inverted()
	if !(vec3.Dot(geometricNormal, &viewDirection) > 0 && vec3.Dot(shadingNormal, &viewDirection) > 0) {
		return scatteringSample{}
	}

	// Malley's method lifts a uniform disk sample onto the hemisphere (PBRT A.5.3).
	diskRadius, azimuth := math.Sqrt(radialSample), 2*math.Pi*azimuthSample
	localX, localY, localZ := diskRadius*math.Cos(azimuth), diskRadius*math.Sin(azimuth), math.Sqrt(1-radialSample)

	// Rotate from the shading frame to world space before checking the geometric surface.
	tangent, bitangent := shadingBasis(shadingNormal)
	heading := vec3.T{}
	for axis := range heading {
		heading[axis] = localX*tangent[axis] + localY*bitangent[axis] + localZ*shadingNormal[axis]
	}
	heading.Normalize()
	if !validSurfaceHeading(geometricNormal, &heading) {
		// Clipped samples contribute zero; retries or a forced fallback would change the PDF.
		return scatteringSample{}
	}

	brdf, pdf := evaluateDiffuse(geometricNormal, shadingNormal, &viewDirection, &heading)
	if !(pdf > 0) {
		return scatteringSample{}
	}

	// Lambertian BRDF and Malley PDF cancel: (1/pi)*cos / (cos/pi) = 1; see evaluateDiffuse.
	return scatteringSample{&heading, brdf * vec3.Dot(shadingNormal, &heading) / pdf}
}

// evaluateDiffuse returns the unit-albedo Lambertian BRDF and cosine-weighted PDF
// per unit solid angle for outgoingDirection.
// geometricNormal and shadingNormal are unit normals facing the incident medium.
// viewDirection and outgoingDirection are unit directions pointing away from the surface.
// Both results are zero if either direction lies on or outside either normal's positive hemisphere.
// The PDF is not renormalized after clipping to the geometric hemisphere.
//
// Model: PBRT 4e, section 9.2, Lambertian reflection:
// https://pbr-book.org/4ed/Reflection_Models/Diffuse_Reflection
// PDF: PBRT 4e, section A.5.3, Malley's method (see sampleDiffuse).
func evaluateDiffuse(geometricNormal, shadingNormal, viewDirection, outgoingDirection *vec3.T) (brdf, pdf float64) {
	outgoingCosine := vec3.Dot(shadingNormal, outgoingDirection)
	if !(outgoingCosine > 0 && vec3.Dot(geometricNormal, outgoingDirection) > 0 &&
		vec3.Dot(shadingNormal, viewDirection) > 0 && vec3.Dot(geometricNormal, viewDirection) > 0) {
		return 0, 0
	}

	// Unit-albedo Lambertian reflection and its cosine PDF (PBRT 9.2 and A.5.3).
	return 1 / math.Pi, outgoingCosine / math.Pi
}
