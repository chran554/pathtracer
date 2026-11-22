package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// surfaceSample includes physical RGB attenuation and all selection probabilities.
// The underlying directional samplers return conditional, unit-Fresnel weights.
type surfaceSample struct {
	heading *vec3.T
	weight  [3]float64
	bounce  bounceType
}

func surfaceUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(1, max(0, value))
}

// metalFresnel uses base color as normal-incidence RGB reflectance. Schlick's
// approximation approaches white at grazing incidence (Filament, section 4.8).
// https://google.github.io/filament/main/filament.html
func metalFresnel(base [3]float64, cosine float64) [3]float64 {
	grazing := math.Pow(1-surfaceUnit(cosine), 5)
	for i := range base {
		base[i] += (1 - base[i]) * grazing
	}
	return base
}

// opaqueReflectionProbability is a sampling heuristic, not physical reflectance.
// Rough interfaces need reflection support even when the macro Fresnel is tiny;
// equal indices are the exception because Fresnel then vanishes at every angle.
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
// Metallicness sets the probability of choosing metal. Given a dielectric,
// transparency sets the probability of choosing its transmissive component.
// Each component's selection probability cancels its mixture weight.
//
// Opaque reflection evaluates exact dielectric Fresnel or RGB conductor Schlick
// at the sampled half-vector. Diffuse uses a reciprocal single-pass coating
// approximation: (1-F(view))*(1-F(light))*base/pi. Internal diffuse returns and
// GGX multiple scattering are omitted, so rough surfaces may lose energy.
//
// The transmissive branch intentionally retains the legacy hybrid: smooth
// refraction and macro-normal Fresnel, with GGX reflection (mirror under TIR).
// It is not a rough dielectric BTDF. Only transmitted samples cross media.
func sampleSurface(material *scene.Material, base [3]float64, geometric, shading, incoming *vec3.T, incident, transmitted float64, random func() float64) surfaceSample {
	metallicness := surfaceUnit(material.Metallicness)
	transparency := surfaceUnit(material.Transparency)
	roughness := surfaceUnit(material.Roughness)
	for channel := range base {
		base[channel] = surfaceUnit(base[channel])
	}

	// A metal bounce always reflects, with base RGB as its normal-incidence reflectance.
	choice := random()
	if metallicness > 0 && choice < metallicness {
		sample := sampleReflection(geometric, shading, incoming, roughness, random(), random())
		if sample.heading == nil {
			return surfaceSample{}
		}

		// Use the actual reflection direction so Fresnel also follows a mirror fallback.
		cosine := reflectionCosine(incoming, sample.heading)
		reflectionWeight := metalFresnel(base, cosine)
		for channel := range reflectionWeight {
			reflectionWeight[channel] *= sample.weight
		}
		return surfaceSample{sample.heading, reflectionWeight, reflectionBounce}
	}

	// Remap the remaining interval to [0, 1) to reuse the draw for the dielectric.
	if metallicness == 1 {
		return surfaceSample{}
	}
	choice = (choice - metallicness) / (1 - metallicness)

	// For a transmissive dielectric, Fresnel chooses reflection or transmission.
	if transparency > 0 && choice < transparency {
		boundary := dielectricBoundary{normal: *shading}
		solid := material.SolidObject && material.RefractionIndex > 0
		if solid {
			boundary = evaluateDielectricBoundary(geometric, shading, incoming, incident, transmitted)
		} else {
			boundary.fresnel = FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
		}

		if random() < boundary.fresnel || boundary.totalInternalReflection {
			// Total internal reflection uses the mirror direction from the boundary.
			if boundary.totalInternalReflection {
				return surfaceSample{boundary.heading, [3]float64{1, 1, 1}, reflectionBounce}
			}

			sample := sampleReflection(geometric, &boundary.normal, incoming, roughness, random(), random())
			return surfaceSample{sample.heading, [3]float64{sample.weight, sample.weight, sample.weight}, reflectionBounce}
		}

		// Thin surfaces transmit straight through. Solids refract and scale radiance.
		heading, weight := incoming, 1.0
		if solid {
			heading, weight = boundary.heading, (incident/transmitted)*(incident/transmitted)
		}
		return surfaceSample{heading, [3]float64{weight, weight, weight}, transmissionBounce}
	}

	// Use Fresnel at the incoming angle to choose between specular and diffuse reflection.
	incomingFresnel := FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
	specularProbability := opaqueReflectionProbability(incomingFresnel, roughness, incident, transmitted)
	if random() < specularProbability {
		sample := sampleReflection(geometric, shading, incoming, roughness, random(), random())
		if sample.heading == nil {
			return surfaceSample{}
		}

		// Weight by Fresnel at the sampled angle and compensate for choosing reflection.
		sampledFresnel := FresnelDielectric(reflectionCosine(incoming, sample.heading), incident, transmitted)
		weight := sample.weight * sampledFresnel / specularProbability
		return surfaceSample{sample.heading, [3]float64{weight, weight, weight}, reflectionBounce}
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

func reflectionCosine(incoming, outgoing *vec3.T) float64 {
	half := outgoing.Subed(incoming)
	half.Normalize()
	return surfaceUnit(-vec3.Dot(incoming, &half))
}

func surfaceBaseColor(material *scene.Material, texture *color.Color) [3]float64 {
	if material.Color == nil {
		return [3]float64{float64(texture.R), float64(texture.G), float64(texture.B)}
	}
	r := 1.0 - (1.0-texture.R)*texture.A
	g := 1.0 - (1.0-texture.G)*texture.A
	b := 1.0 - (1.0-texture.B)*texture.A
	return [3]float64{float64(material.Color.R * r), float64(material.Color.G * g), float64(material.Color.B * b)}
}
