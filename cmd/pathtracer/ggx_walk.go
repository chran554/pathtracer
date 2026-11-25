package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

// facetScatter describes the optical event at one visible facet.
// incoming points toward the facet.
// normal faces the ray's current medium.
// inside reports whether the ray is on the far side of the interface.
// outgoing is the new unit direction.
// attenuation is the RGB throughput multiplier.
// crossed reports transmission through the facet.
type facetScatter func(incoming, normal vec3.T, inside bool) (outgoing vec3.T, attenuation [3]float64, crossed bool)

// walkGGX follows a ray through one Smith microsurface until escape.
// shading is the unit shading normal facing the incident medium.
// incoming is the unit ray directed into the interface.
// roughness is the positive perceptual GGX roughness.
// scatter supplies the optical law at each facet.
// random supplies independent uniform draws in [0, 1).
// The result contains the escape direction, RGB weight and side, or zero on termination.
//
// https://eheitzresearch.wordpress.com/240-2/
func walkGGX(shading, incoming *vec3.T, roughness float64, scatter facetScatter, random func() float64) surfaceSample {
	if !(vec3.Dot(incoming, shading) < 0) {
		return surfaceSample{}
	}

	width := reflectionAlpha(roughness)
	direction, mediumNormal := *incoming, *shading
	logHeightCDF := 0.0 // Start above the facets on the incident side.
	weight := [3]float64{1, 1, 1}
	inside := false

	for collisions := 0; ; collisions++ {
		var escaped bool
		logHeightCDF, escaped = sampleGGXHeight(logHeightCDF, vec3.Dot(&direction, &mediumNormal), width, random())
		if escaped {
			bounce := reflectionBounce
			if inside {
				bounce = transmissionBounce
			}
			return surfaceSample{&direction, weight, bounce}
		}

		if !continueScattering(collisions, &weight, random) {
			return surfaceSample{}
		}

		view := direction.Inverted()
		facet := sampleVisibleGGXNormal(&mediumNormal, &view, width, random(), random())
		if !(vec3.Dot(&facet, &view) > 0) {
			return surfaceSample{}
		}

		outgoing, attenuation, crossed := scatter(direction, facet, inside)
		direction = outgoing
		for channel := range weight {
			weight[channel] *= attenuation[channel]
		}

		// A transmitted ray measures all subsequent heights from the other side.
		if crossed {
			inside = !inside
			mediumNormal.Invert()
			logHeightCDF = oppositeLogHeightCDF(logHeightCDF)
		}
	}
}

// continueScattering applies compensated Russian roulette to long walks.
// encounters counts preceding facet or wall encounters in this walk.
// weight holds the RGB throughput of the path and is scaled on survival.
// random supplies an independent uniform draw in [0, 1) when needed.
// The result reports whether the path continues.
func continueScattering(encounters int, weight *[3]float64, random func() float64) bool {
	const interval = 32
	if encounters == 0 || encounters%interval != 0 {
		return true
	}

	// Polynomial survival bounds expected work even for trapped rays while
	// avoiding exponentially large weights after repeated internal reflections.
	block := float64(encounters / interval)
	ratio := (block + 3) / (block + 4)
	survival := ratio * ratio
	if random() >= survival {
		return false
	}
	for channel := range weight {
		weight[channel] /= survival
	}
	return true
}

// onGeometricSide validates a completed walk against the actual scene surface.
// geometric faces the incident ray's medium.
// The result is the sample, or zero if it escaped on the wrong geometric side.
func (sample surfaceSample) onGeometricSide(geometric *vec3.T) surfaceSample {
	escapeNormal := *geometric
	if sample.bounce == transmissionBounce {
		escapeNormal.Invert()
	}
	if sample.heading == nil || !validSurfaceHeading(&escapeNormal, sample.heading) {
		return surfaceSample{}
	}
	return sample
}

// sampleGGXHeight samples the optical distance to the next facet in height space.
// logHeightCDF is log(C(height)) on the current side.
// cosine is the signed ray cosine relative to that side's normal.
// width is the GGX microfacet width.
// draw is a uniform sample in [0, 1).
// nextLogCDF is the sampled log height CDF, reaching or exceeding zero on upward escape.
// escaped reports whether the ray left the microsurface.
func sampleGGXHeight(logHeightCDF, cosine, width, draw float64) (nextLogCDF float64, escaped bool) {
	cosine = min(1, max(-1, cosine))
	if cosine == 1 {
		return 0, true
	}
	if cosine == 0 {
		return logHeightCDF, false // Horizontal travel intersects at the same height.
	}

	// Smith's projected facet area facing the ray is (root - cosine)/2.
	// Rationalize the difference for upward travel on nearly smooth surfaces.
	root := ggxProjectedRoot(math.Abs(cosine), width)
	projectedArea := (root - cosine) / 2
	if cosine > 0 {
		projectedArea = width * width * (1 - cosine) * (1 + cosine) / (2 * (root + cosine))
	}

	// A unit exponential optical distance advances log(C) by distance / Lambda.
	// Move the zero endpoint half a random bin inwards to keep collisions interior.
	opticalDistance := -math.Log1p(-max(draw, 0x1p-54))
	nextLogCDF = logHeightCDF + opticalDistance*cosine/projectedArea
	return nextLogCDF, cosine > 0 && nextLogCDF >= 0
}

// oppositeLogHeightCDF converts log(C) to log(1-C) when a ray crosses a facet.
// logHeightCDF is the log height CDF on the incident side.
// The result is the log height CDF on the opposite side, retaining small tail probabilities.
func oppositeLogHeightCDF(logHeightCDF float64) float64 {
	if logHeightCDF < -math.Ln2 {
		return math.Log1p(-math.Exp(logHeightCDF))
	}
	return math.Log(-math.Expm1(logHeightCDF))
}
