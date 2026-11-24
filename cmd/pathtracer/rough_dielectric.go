package main

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

// sampleRoughDielectric samples a lossless Smith GGX interface by walking between
// microfacets.
//
// geometric and shading are unit normals facing the incident medium.
// incoming is the unit ray toward the surface, and roughness is in (0,1].
// incident and transmitted are the two positive refractive indices.
// random supplies independent uniform draws in [0,1).
// The result gives the escape direction, net reflection or transmission, and
// radiance weight. A nil heading contributes zero.
//
// https://eheitzresearch.wordpress.com/240-2/
func sampleRoughDielectric(geometric, shading, incoming *vec3.T, roughness, incident, transmitted float64, random func() float64) surfaceSample {
	if !(vec3.Dot(incoming, shading) < 0 && vec3.Dot(incoming, geometric) < 0) {
		return surfaceSample{}
	}

	// Measure height and orient facet normals from the ray's current medium.
	width := reflectionAlpha(roughness)
	direction, mediumNormal := *incoming, *shading
	currentIndex, otherIndex := incident, transmitted
	logHeightCDF := 0.0 // Above every microfacet on arrival: C(height) = 1.
	weight := 1.0

	for collisions := 0; ; collisions++ {
		var escaped bool
		logHeightCDF, escaped = sampleGGXHeight(logHeightCDF, vec3.Dot(&direction, &mediumNormal), width, random())
		if escaped {
			break
		}

		// Russian roulette bounds expected work on long walks.
		// Surviving paths carry the omitted paths' weight, independently of scene recursion depth.
		if collisions > 0 && collisions%32 == 0 {
			if random() < 0.5 {
				return surfaceSample{}
			}
			weight *= 2
		}

		view := direction.Inverted()
		facet := sampleVisibleGGXNormal(&mediumNormal, &view, width, random(), random())
		if !(vec3.Dot(&facet, &view) > 0) {
			return surfaceSample{}
		}
		var crossed bool
		direction, crossed = scatterAtGlassFacet(&direction, &facet, currentIndex, otherIndex, random())
		if crossed {
			currentIndex, otherIndex = otherIndex, currentIndex
			mediumNormal.Invert()
			logHeightCDF = oppositeLogHeightCDF(logHeightCDF)
		}
	}

	// Only the medium at escape determines the scene bounce.
	// Intermediate refractions cancel in pairs, including their radiance conversion factors.
	bounce := reflectionBounce
	escapeNormal := *geometric
	if currentIndex != incident {
		bounce = transmissionBounce
		escapeNormal.Invert()
		indexRatio := incident / transmitted
		weight *= indexRatio * indexRatio
	}
	if !validSurfaceHeading(&escapeNormal, &direction) {
		return surfaceSample{}
	}

	return surfaceSample{&direction, [3]float64{weight, weight, weight}, bounce}
}

// sampleGGXHeight returns the next log height CDF and whether the ray escaped.
//
// logHeightCDF is log(C(height)) on the current side.
// cosine is the signed ray cosine relative to that side's normal.
// width is GGX alpha, and draw is uniform in [0,1).
// Sampling in log space avoids constructing or tracing microgeometry.
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
// Both branches preserve the small tail probability near either height extreme.
func oppositeLogHeightCDF(logHeightCDF float64) float64 {
	if logHeightCDF < -math.Ln2 {
		return math.Log1p(-math.Exp(logHeightCDF))
	}
	return math.Log(-math.Expm1(logHeightCDF))
}

// scatterAtGlassFacet returns a unit direction and whether it crosses the facet.
// incoming points toward the facet (unit length).
// normal faces the current medium (unit length).
// currentIndex, otherIndex ate the current and adjacent IORs.
// draw is a uniform Fresnel branch sample in [0,1).
func scatterAtGlassFacet(incoming, normal *vec3.T, currentIndex, otherIndex, draw float64) (direction vec3.T, crossed bool) {
	fresnel := FresnelDielectric(-vec3.Dot(incoming, normal), currentIndex, otherIndex)
	if draw < fresnel {
		return getReflectionVector(normal, incoming).Normalized(), false
	}

	heading, totalInternalReflection := getRefractionVector(normal, incoming, currentIndex, otherIndex)
	return heading.Normalized(), !totalInternalReflection
}
