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

// sampleSurface mixes a conductor, an opaque dielectric, and a transmissive
// dielectric with weights m, (1-m)*(1-t), and (1-m)*t. Physical mixture weights
// cancel their selection probabilities; they are never normalized into albedo.
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
	m, t := surfaceUnit(material.Metallicness), surfaceUnit(material.Transparency)
	roughness := surfaceUnit(material.Roughness)
	for i := range base {
		base[i] = surfaceUnit(base[i])
	}
	// Choose conductor versus dielectric first, then opaque versus transmissive.
	choice := random()
	if m > 0 && choice < m {
		sample := sampleReflection(geometric, shading, incoming, roughness, random(), random())
		if sample.heading == nil {
			return surfaceSample{}
		}
		cosine := reflectionCosine(incoming, sample.heading)
		// A mirror fallback may change the optical normal; the half-vector follows it.
		f := metalFresnel(base, cosine)
		for i := range f {
			f[i] *= sample.weight
		}
		return surfaceSample{sample.heading, f, reflectionBounce}
	}
	if m == 1 {
		return surfaceSample{}
	}
	choice = (choice - m) / (1 - m)
	if t > 0 && choice < t {
		boundary := dielectricBoundary{normal: *shading}
		solid := material.SolidObject && material.RefractionIndex > 0
		if solid {
			boundary = evaluateDielectricBoundary(geometric, shading, incoming, incident, transmitted)
		} else {
			boundary.fresnel = FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
		}
		if random() < boundary.fresnel || boundary.totalInternalReflection {
			if boundary.totalInternalReflection {
				return surfaceSample{boundary.heading, [3]float64{1, 1, 1}, reflectionBounce}
			}
			sample := sampleReflection(geometric, &boundary.normal, incoming, roughness, random(), random())
			return surfaceSample{sample.heading, [3]float64{sample.weight, sample.weight, sample.weight}, reflectionBounce}
		}
		heading, weight := incoming, 1.0
		if solid {
			heading, weight = boundary.heading, (incident/transmitted)*(incident/transmitted)
		}
		return surfaceSample{heading, [3]float64{weight, weight, weight}, transmissionBounce}
	}
	fv := FresnelDielectric(-vec3.Dot(shading, incoming), incident, transmitted)
	q := opaqueReflectionProbability(fv, roughness, incident, transmitted)
	if random() < q {
		sample := sampleReflection(geometric, shading, incoming, roughness, random(), random())
		if sample.heading == nil {
			return surfaceSample{}
		}
		f := FresnelDielectric(reflectionCosine(incoming, sample.heading), incident, transmitted)
		weight := sample.weight * f / q
		return surfaceSample{sample.heading, [3]float64{weight, weight, weight}, reflectionBounce}
	}
	sample := sampleDiffuse(geometric, shading, incoming, random(), random())
	if sample.heading == nil {
		return surfaceSample{}
	}
	fl := FresnelDielectric(vec3.Dot(shading, sample.heading), incident, transmitted)
	weight := sample.weight * (1 - fv) * (1 - fl) / (1 - q)
	for i := range base {
		base[i] *= weight
	}
	return surfaceSample{sample.heading, base, diffuseBounce}
}

func reflectionCosine(incoming, outgoing *vec3.T) float64 {
	half := outgoing.Subed(incoming)
	half.Normalize()
	return surfaceUnit(-vec3.Dot(incoming, &half))
}

func surfaceBaseColor(material *scene.Material, texture *color.Color) [3]float64 {
	return [3]float64{float64(material.Color.R * texture.R), float64(material.Color.G * texture.G), float64(material.Color.B * texture.B)}
}
