package main

import "github.com/ungerik/go3d/float64/vec3"

// dielectricBoundary keeps the optical normal, Fresnel weight, and direction together.
// heading is transmission unless totalInternalReflection is true; nil rejects that event.
type dielectricBoundary struct {
	normal                  vec3.T
	fresnel                 float64
	heading                 *vec3.T
	totalInternalReflection bool
}

// evaluateDielectricBoundary returns the optical solution for a dielectric interface.
// geometricNormal and shadingNormal must be unit normals facing the incident medium;
// incomingDirection must be a unit direction toward the surface. incidentIndex and
// transmittedIndex are the positive refractive indices of the two media, respectively.
//
// The result contains the normal used, Fresnel reflectance in [0,1], and a unit heading
// for transmission or, when totalInternalReflection is true, reflection. If shadingNormal
// produces a heading on the wrong geometric side, geometricNormal is tried instead.
// A nil heading means neither normal produced a valid, non-tangent direction.
//
// Optics: Snell's law and the unpolarized Fresnel equations, PBRT 4e, section 9.3:
// https://pbr-book.org/4ed/Reflection_Models/Specular_Reflection_and_Transmission
// The geometric-normal fallback is a renderer policy for perturbed normals.
func evaluateDielectricBoundary(geometricNormal, shadingNormal, incomingDirection *vec3.T, incidentIndex, transmittedIndex float64) dielectricBoundary {
	var boundary dielectricBoundary
	for _, opticalNormal := range [2]vec3.T{*shadingNormal, *geometricNormal} {
		// Evaluate Snell refraction and Fresnel reflectance with the same normal (PBRT 9.3).
		boundary.normal = opticalNormal
		boundary.fresnel = FresnelDielectric(-vec3.Dot(&opticalNormal, incomingDirection), incidentIndex, transmittedIndex)
		boundary.heading, boundary.totalInternalReflection = getRefractionVector(&opticalNormal, incomingDirection, incidentIndex, transmittedIndex)
		boundary.heading.Normalize()

		// Transmission crosses the geometric surface; TIR remains on the incident side.
		outgoingHemisphereNormal := *geometricNormal
		if !boundary.totalInternalReflection {
			outgoingHemisphereNormal.Invert()
		}
		if validSurfaceHeading(&outgoingHemisphereNormal, boundary.heading) {
			return boundary
		}

		// Discard an invalid heading before retrying with the geometric normal.
		boundary.heading = nil
	}

	return boundary
}
