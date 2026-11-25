package main

import "github.com/ungerik/go3d/float64/vec3"

// sampleRoughDielectric samples reflection or refraction at a solid GGX interface.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit shading normal facing the incident medium.
// incoming is the unit ray toward the surface.
// roughness is the positive perceptual GGX roughness.
// incident is the refractive index before crossing the boundary.
// transmitted is the refractive index beyond the boundary.
// random supplies independent uniform draws in [0, 1).
// The result includes the net radiance conversion on transmission.
func sampleRoughDielectric(geometric, shading, incoming *vec3.T, roughness, incident, transmitted float64, random func() float64) surfaceSample {
	if !(vec3.Dot(incoming, geometric) < 0) {
		return surfaceSample{}
	}

	sample := walkGlassFace(shading, incoming, roughness, incident, transmitted, random).onGeometricSide(geometric)
	if sample.bounce == transmissionBounce {
		indexRatio := incident / transmitted
		for channel := range sample.weight {
			sample.weight[channel] *= indexRatio * indexRatio
		}
	}
	return sample
}

// walkGlassFace samples one lossless rough dielectric face in units of power.
// shading is the unit face normal pointing toward the incident medium.
// incoming is the unit direction approaching the face.
// roughness is the positive perceptual roughness of the face.
// incident is the index of the medium where this face walk starts.
// transmitted is the index on the opposite side.
// random supplies independent uniform draws in [0, 1).
// The result identifies the escape side and carries the roulette weight.
func walkGlassFace(shading, incoming *vec3.T, roughness, incident, transmitted float64, random func() float64) surfaceSample {
	scatter := func(direction, facet vec3.T, inside bool) (vec3.T, [3]float64, bool) {
		currentIndex, otherIndex := incident, transmitted
		if inside {
			currentIndex, otherIndex = otherIndex, currentIndex
		}

		outgoing, crossed := scatterAtGlassFacet(&direction, &facet, currentIndex, otherIndex, random())
		return outgoing, [3]float64{1, 1, 1}, crossed
	}
	return walkGGX(shading, incoming, roughness, scatter, random)
}

// sampleRoughThinWall follows a ray between two independently rough, parallel glass faces.
// geometric is the unit geometric normal pointing toward the incident medium.
// shading is the unit shading normal of the entry face.
// incoming is the unit direction approaching the wall.
// roughness is the positive perceptual roughness of both faces.
// surroundingIndex is the refractive index outside either face.
// wallIndex is the refractive index between the faces.
// random supplies independent uniform draws in [0, 1).
// The result gives the net escape side and RGB weight back in the surrounding medium.
func sampleRoughThinWall(geometric, shading, incoming *vec3.T, roughness, surroundingIndex, wallIndex float64, random func() float64) surfaceSample {
	if !(vec3.Dot(geometric, incoming) < 0) {
		return surfaceSample{}
	}

	entry := walkGlassFace(shading, incoming, roughness, surroundingIndex, wallIndex, random)
	if entry.bounce != transmissionBounce {
		return entry.onGeometricSide(geometric)
	}

	// Inside the wall the back face presents the same incident-facing normal
	// as the entry face. Each internal reflection sends the ray to the other face.
	direction, faceNormal := *entry.heading, *shading
	weight := entry.weight
	for face := 1; ; face++ {
		if !continueScattering(face, &weight, random) {
			return surfaceSample{}
		}
		sample := walkGlassFace(&faceNormal, &direction, roughness, wallIndex, surroundingIndex, random)
		if sample.heading == nil {
			return surfaceSample{}
		}

		for channel := range weight {
			weight[channel] *= sample.weight[channel]
		}
		if sample.bounce == transmissionBounce {
			// Both external media are identical, so the two index factors cancel.
			if face%2 == 0 {
				sample.bounce = reflectionBounce
			}
			sample.weight = weight
			return sample.onGeometricSide(geometric)
		}

		direction = *sample.heading
		faceNormal.Invert()
	}
}

// scatterAtGlassFacet samples reflection or refraction at one dielectric facet.
// incoming is the unit direction toward the facet.
// normal is the unit facet normal facing the current medium.
// currentIndex is the refractive index on the incident side.
// otherIndex is the refractive index on the far side.
// draw is a uniform sample in [0, 1) for the Fresnel choice.
// direction is the outgoing unit vector.
// crossed reports whether the ray refracted through the facet.
func scatterAtGlassFacet(incoming, normal *vec3.T, currentIndex, otherIndex, draw float64) (direction vec3.T, crossed bool) {
	fresnel := FresnelDielectric(-vec3.Dot(incoming, normal), currentIndex, otherIndex)
	if draw < fresnel {
		return getReflectionVector(normal, incoming).Normalized(), false
	}

	heading, totalInternalReflection := getRefractionVector(normal, incoming, currentIndex, otherIndex)
	return heading.Normalized(), !totalInternalReflection
}
