package main

import (
	"math"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

type bounceType uint8

const (
	noBounce bounceType = iota
	diffuseBounce
	reflectionBounce
	transmissionBounce
)

// bounceProbabilities describes the normalized material mixture after Fresnel adjustment.
type bounceProbabilities struct {
	diffuse, reflection, transmission float64
}

// surfaceBounceProbabilities mixes finite, nonnegative material weights without adding absorption.
func surfaceBounceProbabilities(material *scene.Material, fresnel float64, totalInternalReflection bool) bounceProbabilities {
	weightScale := max(material.Diffuse, material.Glossiness, material.Transparency)
	if weightScale == 0 {
		return bounceProbabilities{}
	}
	if totalInternalReflection {
		return bounceProbabilities{reflection: 1}
	}

	// Scaling first keeps normalization finite even for very large relative weights.
	diffuse := material.Diffuse / weightScale
	reflection := material.Glossiness / weightScale
	transmission := material.Transparency / weightScale
	nonReflection := diffuse + transmission
	if nonReflection == 0 {
		return bounceProbabilities{reflection: 1}
	}

	baseReflection := reflection / (reflection + nonReflection)
	reflectionProbability := (1-fresnel)*baseReflection + fresnel*material.FresnelMaxGlossiness
	remaining := 1 - reflectionProbability
	return bounceProbabilities{
		diffuse:      remaining * (diffuse / nonReflection),
		reflection:   reflectionProbability,
		transmission: remaining * (transmission / nonReflection),
	}
}

// sample selects a bounce for a uniform sample in [0,1), or no bounce for an empty mixture.
func (p bounceProbabilities) sample(value float64) bounceType {
	if p.diffuse > 0 && value < p.diffuse {
		return diffuseBounce
	}
	if p.reflection > 0 && value < p.diffuse+p.reflection {
		return reflectionBounce
	}
	if p.transmission > 0 {
		return transmissionBounce
	}
	// Rounding at the upper boundary must still choose an enabled component.
	if p.reflection > 0 {
		return reflectionBounce
	}
	if p.diffuse > 0 {
		return diffuseBounce
	}
	return noBounce
}

// boundaryRefractionIndices selects the media on either side of the geometric boundary.
func boundaryRefractionIndices(material *scene.Material, rayContexts []*scene.Material, isIngoingRay bool) (incident, transmitted float64) {
	incident = rayContexts[len(rayContexts)-1].RefractionIndex
	transmitted = material.RefractionIndex
	if material.SolidObject && !isIngoingRay {
		// Exiting returns to the enclosing medium, which need not be air.
		transmitted = incident
		if len(rayContexts) > 1 {
			transmitted = rayContexts[len(rayContexts)-2].RefractionIndex
		}
	}
	return
}

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

// validSurfaceHeading accepts finite unit directions strictly on the incident side.
func validSurfaceHeading(geometricNormal, heading *vec3.T) bool {
	return math.Abs(heading.LengthSqr()-1) < 1e-6 && vec3.Dot(geometricNormal, heading) > 0
}

// scatteringSample carries a direction and its conditional BRDF*cosine/PDF weight.
type scatteringSample struct {
	heading *vec3.T
	weight  float64
}

// shadingBasis constructs an orthonormal frame around a unit shading normal.
func shadingBasis(normal *vec3.T) (tangent, bitangent vec3.T) {
	axis := vec3.UnitX
	if math.Abs(normal[0]) > 0.9 {
		axis = vec3.UnitY
	}
	tangent = vec3.Cross(&axis, normal)
	tangent.Normalize()
	bitangent = vec3.Cross(normal, &tangent)
	return
}
