package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// surfaceSample includes physical RGB attenuation and all selection probabilities.
// A complete microsurface walk contributes one scene bounce.
type surfaceSample struct {
	heading *vec3.T
	weight  [3]float64
	bounce  bounceType
}

// surfaceUnit bounds a scalar material parameter for sampling.
// value is the requested parameter value.
// The result lies in [0, 1], with NaN mapped to zero.
func surfaceUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(1, max(0, value))
}

// metalFresnel evaluates RGB Schlick Fresnel at a conductor facet.
// base contains the RGB reflectance at normal incidence.
// cosine is the cosine between the facet normal and the direction toward the previous vertex.
// The result is the RGB reflectance, approaching white at grazing incidence.
//
// https://google.github.io/filament/main/filament.html
func metalFresnel(base [3]float64, cosine float64) [3]float64 {
	grazing := math.Pow(1-surfaceUnit(cosine), 5)
	for i := range base {
		base[i] += (1 - base[i]) * grazing
	}
	return base
}

// opaqueReflectionProbability chooses how often to sample the exterior glossy component.
// fresnel is the dielectric Fresnel reflectance at the incoming angle.
// roughness is the perceptual interface roughness.
// incident is the refractive index in front of the interface.
// transmitted is the refractive index behind the interface.
// The result is a sampling probability with support for rough reflections at other facet angles.
func opaqueReflectionProbability(fresnel, roughness, incident, transmitted float64) float64 {
	if incident == transmitted {
		return 0
	}
	if roughness > 0 {
		return min(0.95, max(0.05, fresnel))
	}
	return fresnel
}

// sampleSurface chooses one bounce and returns its direction, RGB weight, and type.
// material supplies the dielectric transmission fraction and boundary type.
// roughness is the perceptual roughness after texture modulation.
// metallicness is the conductor fraction after texture modulation.
// base is the linear RGB surface color after texture modulation.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit shading normal facing the incident ray.
// incoming is the unit direction toward the surface.
// incident is the refractive index of the ray's current medium.
// transmitted is the refractive index on the other side or inside a thin wall.
// random supplies independent uniform draws in [0, 1).
// The result includes RGB throughput with component sampling probabilities accounted for.
func sampleSurface(material *scene.Material, roughness float64, metallicness float64, base [3]float64, geometric, shading, incoming *vec3.T, incident, transmitted float64, random func() float64) surfaceSample {
	metallicness = surfaceUnit(metallicness)
	transparency := surfaceUnit(material.Transparency)
	roughness = surfaceUnit(roughness)

	for channel := range base {
		base[channel] = surfaceUnit(base[channel])
	}

	choice := random()
	if (metallicness > 0) && (choice < metallicness) {
		return sampleMetal(roughness, base, geometric, shading, incoming, random)
	}

	// Remap the remaining interval to [0, 1) to reuse the draw for the dielectric.
	if metallicness == 1 {
		return surfaceSample{}
	}

	choice = (choice - metallicness) / (1 - metallicness)
	if (transparency > 0) && (choice < transparency) {
		return sampleTransmissiveDielectric(material, roughness, geometric, shading, incoming, incident, transmitted, random)
	}

	return sampleOpaqueDielectric(roughness, base, geometric, shading, incoming, incident, transmitted, random)
}

// sampleMetal samples the color-weighted reflection from a conductor.
// roughness is the effective perceptual roughness at the hit.
// base contains the RGB reflectance at normal incidence.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit optical normal facing the incident ray.
// incoming is the unit direction toward the surface.
// random supplies independent uniform draws in [0, 1).
// The result contains the reflected direction and accumulated RGB Fresnel weight.
func sampleMetal(roughness float64, base [3]float64, geometric, shading, incoming *vec3.T, random func() float64) surfaceSample {
	reflectance := func(cosine float64) [3]float64 { return metalFresnel(base, cosine) }
	return sampleGlossyReflection(geometric, shading, incoming, roughness, reflectance, random)
}

// sampleGlossyReflection integrates exterior facet reflections at one surface hit.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit shading normal facing the incident ray.
// incoming is the unit direction toward the surface.
// roughness is the effective perceptual roughness, with zero selecting a mirror.
// reflectance supplies RGB Fresnel reflectance for a facet incidence cosine.
// random supplies independent uniform draws in [0, 1).
// The result contains the escape direction and product of the facet weights.
func sampleGlossyReflection(geometric, shading, incoming *vec3.T, roughness float64, reflectance func(float64) [3]float64, random func() float64) surfaceSample {
	if roughness == 0 {
		heading := mirrorDirection(geometric, shading, incoming)
		return surfaceSample{heading, reflectance(reflectionCosine(incoming, heading)), reflectionBounce}
	}

	if !(vec3.Dot(geometric, incoming) < 0) {
		return surfaceSample{}
	}

	scatter := func(direction, facet vec3.T, _ bool) (vec3.T, [3]float64, bool) {
		weight := reflectance(-vec3.Dot(&direction, &facet))
		return getReflectionVector(&facet, &direction).Normalized(), weight, false
	}
	return walkGGX(shading, incoming, roughness, scatter, random).onGeometricSide(geometric)
}

// sampleTransmissiveDielectric selects the solid interface or thin-wall model.
// material specifies whether the object encloses a volume.
// roughness is the effective perceptual roughness at the hit.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit optical normal facing the incident ray.
// incoming is the unit direction toward the surface.
// incident is the refractive index of the surrounding or incident medium.
// transmitted is the refractive index across the interface or inside the wall.
// random supplies independent uniform draws in [0, 1).
// The result contains the outgoing direction, net bounce type and radiance weight.
func sampleTransmissiveDielectric(material *scene.Material, roughness float64, geometric, shading, incoming *vec3.T, incident, transmitted float64, random func() float64) surfaceSample {
	unitWeight := [3]float64{1, 1, 1}
	straightThrough := *incoming
	if incident == transmitted {
		return surfaceSample{&straightThrough, unitWeight, transmissionBounce}
	}

	solid := material.SolidObject && material.RefractionIndex > 0
	if roughness > 0 && incident > 0 && transmitted > 0 {
		if solid {
			return sampleRoughDielectric(geometric, shading, incoming, roughness, incident, transmitted, random)
		}
		return sampleRoughThinWall(geometric, shading, incoming, roughness, incident, transmitted, random)
	}

	if !solid {
		// Sum all internal reflections between the smooth, parallel faces.
		fresnel := FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
		if random() < 2*fresnel/(1+fresnel) {
			return surfaceSample{mirrorDirection(geometric, shading, incoming), unitWeight, reflectionBounce}
		}
		return surfaceSample{&straightThrough, unitWeight, transmissionBounce}
	}

	boundary := evaluateDielectricBoundary(geometric, shading, incoming, incident, transmitted)
	if boundary.totalInternalReflection {
		return surfaceSample{boundary.heading, unitWeight, reflectionBounce}
	}
	if random() < boundary.fresnel {
		return surfaceSample{mirrorDirection(geometric, &boundary.normal, incoming), unitWeight, reflectionBounce}
	}

	indexRatio := incident / transmitted
	weight := indexRatio * indexRatio
	return surfaceSample{boundary.heading, [3]float64{weight, weight, weight}, transmissionBounce}
}

// sampleOpaqueDielectric mixes exterior reflection with a single-pass diffuse coating.
// roughness is the effective perceptual roughness of the exterior interface.
// base is the RGB albedo of the diffuse substrate.
// geometric is the unit geometric normal facing the incident medium.
// shading is the unit shading normal facing the incident ray.
// incoming is the unit direction toward the surface.
// incident is the refractive index on the incident side.
// transmitted is the refractive index on the substrate side.
// random supplies independent uniform draws in [0, 1).
// The result contains a diffuse or glossy bounce compensated for lobe selection.
func sampleOpaqueDielectric(roughness float64, base [3]float64, geometric, shading, incoming *vec3.T, incident, transmitted float64, random func() float64) surfaceSample {
	// Use Fresnel at the incoming angle to choose between specular and diffuse reflection.
	incomingFresnel := FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
	specularProbability := opaqueReflectionProbability(incomingFresnel, roughness, incident, transmitted)
	if random() < specularProbability {
		reflectance := func(cosine float64) [3]float64 {
			fresnel := FresnelDielectric(cosine, incident, transmitted)
			return [3]float64{fresnel, fresnel, fresnel}
		}
		sample := sampleGlossyReflection(geometric, shading, incoming, roughness, reflectance, random)
		for channel := range sample.weight {
			sample.weight[channel] /= specularProbability
		}
		return sample
	}

	// The remaining probability selects diffuse scattering.
	sample := sampleDiffuse(geometric, shading, incoming, random(), random())
	if sample.heading == nil {
		return surfaceSample{}
	}

	// Diffuse light passes through the coating on the way in and again on the way out.
	outgoingFresnel := FresnelDielectric(vec3.Dot(shading, sample.heading), incident, transmitted)
	weight := sample.weight * (1 - incomingFresnel) * (1 - outgoingFresnel) / (1 - specularProbability)
	for channel := range base {
		base[channel] *= weight
	}
	return surfaceSample{sample.heading, base, diffuseBounce}
}

// reflectionCosine reconstructs the facet incidence cosine for a mirror reflection.
// incoming is the unit ray toward the surface.
// outgoing is the reflected unit direction, including any geometric fallback.
// The result is the clamped dot product with the inferred facet normal.
func reflectionCosine(incoming, outgoing *vec3.T) float64 {
	half := outgoing.Subed(incoming)
	half.Normalize()
	return surfaceUnit(-vec3.Dot(incoming, &half))
}

// surfaceBaseColor converts a scene color to the scattering representation.
// c is the linear surface color after texture evaluation.
// The result contains its red, green, and blue components.
func surfaceBaseColor(c *color.Color) [3]float64 {
	return [3]float64{float64(c.R), float64(c.G), float64(c.B)}
}
