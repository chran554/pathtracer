package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

const maxDirectionSamples = 100

// orientSurfaceNormals faces local normals toward the incident side without changing the hit record.
func orientSurfaceNormals(ii *IntersectionInformation, incoming *vec3.T) (geometricNormal, shadingNormal vec3.T, isIngoingRay bool) {
	geometricNormal = *ii.normalAtIntersection
	if ii.intersectedFacet != nil {
		geometricNormal = *ii.intersectedFacet.Normal
	} else if ii.intersectedDisc != nil {
		geometricNormal = *ii.intersectedDisc.Normal
	}
	geometricNormal.Normalize()

	// Capture entry/exit before orienting; shading normals do not define a medium boundary.
	isIngoingRay = vec3.Dot(&geometricNormal, incoming) < 0
	if !isIngoingRay {
		geometricNormal.Invert()
	}

	shadingNormal = *ii.normalAtIntersection
	if ii.perturbedNormalAtIntersection != nil {
		shadingNormal = *ii.perturbedNormalAtIntersection
	}
	shadingNormal.Normalize()
	if vec3.Dot(&shadingNormal, &geometricNormal) < 0 {
		shadingNormal.Invert()
	}
	// Strong perturbations may still face away from the ray or be degenerate.
	if !validSurfaceHeading(&geometricNormal, &shadingNormal) || !(vec3.Dot(&shadingNormal, incoming) < 0) {
		shadingNormal = geometricNormal
	}
	return
}

// sampleDiffuseHeading samples the shading hemisphere while respecting the geometric surface.
func sampleDiffuseHeading(geometricNormal, shadingNormal *vec3.T) *vec3.T {
	for range maxDirectionSamples {
		heading := getRandomCosineWeightedHemisphereVector(shadingNormal)
		heading.Normalize()
		if validSurfaceHeading(geometricNormal, heading) {
			return heading
		}
	}
	// Exhaustion must not turn the last rejected direction into a transmitted ray.
	fallback := *geometricNormal
	return &fallback
}

// sampleReflectionHeading samples rough reflection with a geometric reflection fallback.
func sampleReflectionHeading(geometricNormal, shadingNormal, incoming *vec3.T, roughness float64) *vec3.T {
	perfectReflection := getReflectionVector(shadingNormal, incoming)
	weight := roughness * roughness
	for range maxDirectionSamples {
		diffuseHeading := getRandomCosineWeightedHemisphereVector(shadingNormal)
		heading := vec3.Interpolate(perfectReflection, diffuseHeading, weight)
		heading.Normalize()
		if validSurfaceHeading(geometricNormal, &heading) {
			return &heading
		}
	}

	// Ignore the perturbation if every candidate crosses the geometric surface.
	fallback := getReflectionVector(geometricNormal, incoming)
	fallback.Normalize()
	if validSurfaceHeading(geometricNormal, fallback) {
		return fallback
	}
	// An exactly grazing reflection is tangent, so use a strictly outward direction.
	normal := *geometricNormal
	return &normal
}

// validSurfaceHeading accepts finite unit directions strictly on the incident side.
func validSurfaceHeading(geometricNormal, heading *vec3.T) bool {
	return math.Abs(heading.LengthSqr()-1) < 1e-6 && vec3.Dot(geometricNormal, heading) > 0
}
