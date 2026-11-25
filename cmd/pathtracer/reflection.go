package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

// mirrorDirection reflects about the shading normal with a geometric fallback.
// geometric is the unit normal of the scene surface facing the incident medium.
// shading is the unit optical normal facing the incident ray.
// incoming is the unit direction toward the surface.
// The result is a unit direction on the incident side.
func mirrorDirection(geometric, shading, incoming *vec3.T) *vec3.T {
	heading := getReflectionVector(shading, incoming)
	heading.Normalize()
	if validSurfaceHeading(geometric, heading) {
		return heading
	}

	// A perturbed shading normal may reflect through the geometry.
	heading = getReflectionVector(geometric, incoming)
	heading.Normalize()
	if validSurfaceHeading(geometric, heading) {
		return heading
	}
	fallback := *geometric
	return &fallback
}

// sampleVisibleGGXNormal samples a visible facet of an isotropic GGX microsurface.
// shadingNormal is the unit normal pointing into the current medium.
// viewDirection is the unit direction toward the previous vertex, possibly below the shading horizon.
// microfacetWidth is the positive GGX width.
// radialSample is the independent uniform disk-radius draw in [0, 1).
// azimuthSample is the independent uniform azimuth draw in [0, 1).
// The result is a world-space unit facet normal visible from viewDirection.
//
// Heitz (2018), listing 1: https://jcgt.org/published/0007/04/01/paper.pdf
func sampleVisibleGGXNormal(shadingNormal, viewDirection *vec3.T, microfacetWidth, radialSample, azimuthSample float64) vec3.T {
	// Heitz 3.2: stretch the view in the shading frame to reduce the ellipsoid to a hemisphere.
	tangent, bitangent := shadingBasis(shadingNormal)
	stretchedView := vec3.T{
		microfacetWidth * vec3.Dot(viewDirection, &tangent),
		microfacetWidth * vec3.Dot(viewDirection, &bitangent),
		vec3.Dot(viewDirection, shadingNormal),
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

	// Transform the sampled microfacet normal back to world space.
	microfacetNormal := vec3.T{}
	for axis := range microfacetNormal {
		microfacetNormal[axis] = localMicrofacetNormal[0]*tangent[axis] +
			localMicrofacetNormal[1]*bitangent[axis] + localMicrofacetNormal[2]*shadingNormal[axis]
	}
	return microfacetNormal
}

// reflectionAlpha converts perceptual roughness to GGX width.
// roughness is the perceptual roughness in [0, 1].
// The result is roughness squared with a numerical floor of 1e-4.
//
// https://google.github.io/filament/main/filament.html
func reflectionAlpha(roughness float64) float64 {
	return max(1e-4, roughness*roughness)
}

// ggxProjectedRoot computes the projected-area root in the Smith GGX height distribution.
// directionCosine is the absolute cosine relative to the microsurface normal.
// microfacetWidth is the nonnegative GGX width.
// The result is sqrt(cosine² + width² * (1 - cosine²)).
func ggxProjectedRoot(directionCosine, microfacetWidth float64) float64 {
	directionCosine = min(1.0, directionCosine)
	return math.Hypot(directionCosine, microfacetWidth*math.Sqrt(max(0, 1-directionCosine*directionCosine)))
}
