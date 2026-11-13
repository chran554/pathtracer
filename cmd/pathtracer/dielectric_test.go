package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestTracePathSmoothGlass checks smooth refraction, total internal reflection,
// and geometric fallback for entry and exit, using the Snell/Fresnel reference
// in dielectricTestReference and the radiance conversion in TestTracePathTransmissionRadiance.
func TestTracePathSmoothGlass(test *testing.T) {
	testCases := []struct {
		name                       string
		incidentAngleDegrees       float64
		normalAngleDegrees         float64
		incidentIndex              float64
		transmittedIndex           float64
		expectedNormalAngleDegrees float64
		inside                     bool
		totalInternalReflection    bool
	}{
		{"enter", 0, 30, 1, 1.5, 30, false, false},
		{"enter from water", 20, 10, 1.33, 1.5, 10, false, false},
		{"exit", 0, 30, 1.5, 1, 30, true, false},
		{"exit into water", 20, 10, 1.5, 1.33, 10, true, false},
		{"smooth TIR", 20, 30, 1.5, 1, 30, true, true},
		{"smooth transmission", 60, -30, 1.5, 1, -30, true, false},
		{"transmission falls back to TIR", 60, -20, 1.5, 1, 0, true, true},
		{"TIR falls back to transmission", 0, 60, 1.5, 1, 0, true, false},
	}

	for _, testCase := range testCases {
		for _, normalMapped := range []bool{false, true} {
			test.Run(fmt.Sprintf("%s/mapped=%t", testCase.name, normalMapped), func(test *testing.T) {
				incomingDirection, shadingNormal := dielectricTestDirections(testCase.incidentAngleDegrees, testCase.normalAngleDegrees)
				glass := scene.NewMaterial()
				if normalMapped {
					glass = normalMappedTestMaterial(shadingNormal)
				}
				glass.M(0, 0).T(1, true, 1.5)
				glass.FresnelMaxGlossiness = 0 // Isolate the selected transmission or forced TIR.

				// Exercise both interpolated vertex normals and texture-derived normals.
				plane := orientationTestPlane(glass)
				if !normalMapped {
					_, firstVertexNormal := dielectricTestDirections(0, testCase.normalAngleDegrees+5)
					_, secondVertexNormal := dielectricTestDirections(0, testCase.normalAngleDegrees-5)
					plane.Facets[0].VertexNormals = []*vec3.T{&firstVertexNormal, &secondVertexNormal, &shadingNormal}
				}

				// Seed the incident medium so entry and exit use the intended index pair.
				outsideIndex := testCase.incidentIndex
				if testCase.inside {
					outsideIndex = testCase.transmittedIndex
				}
				mediumStack := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
				if outsideIndex != 1 {
					mediumStack = append(mediumStack, scene.NewMaterial().T(1, true, outsideIndex))
				}
				if testCase.inside {
					mediumStack = append(mediumStack, glass)
				}

				// Reserve a sentinel to detect erroneous pushes on reflection.
				stackSentinel := scene.NewMaterial()
				stackStorage := append(mediumStack, stackSentinel)
				mediumStack = stackStorage[:len(stackStorage)-1]

				// A small emitter at the Snell-law reference direction detects angular errors.
				expectedHeading, _ := dielectricTestReference(testCase.incidentAngleDegrees, testCase.expectedNormalAngleDegrees, testCase.incidentIndex, testCase.transmittedIndex)
				if testCase.inside {
					incomingDirection.Invert()
					expectedHeading.Invert()
				}
				rayOrigin := incomingDirection.Inverted() // Each ray hits the world origin.
				lightPosition := expectedHeading.Scaled(2)

				sceneNode := &scene.SceneNode{
					FacetStructures: []*scene.FacetStructure{plane},
					Discs:           []*scene.Disc{scene.NewDisc(&lightPosition, &expectedHeading, 0.002, scene.NewMaterial().E(color.White, 1, true))},
				}

				recursionDepth := 1
				if testCase.totalInternalReflection {
					// Exit normally after TIR. The exit's index conversion detects
					// an incorrect pop of the glass medium during reflection.
					exitPoint := expectedHeading.Scaled(1)
					sceneNode.Discs = append(sceneNode.Discs, scene.NewDisc(&exitPoint, &expectedHeading, 0.2, glass))
					recursionDepth++
				}

				// Radiance conversion (PBRT 9.5.2; see TestTracePathTransmissionRadiance).
				expectedRadiance := math.Pow(testCase.incidentIndex/testCase.transmittedIndex, 2)
				for range 4 {
					actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &incomingDirection}, &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: recursionDepth}, sceneNode, 0, mediumStack)
					assertTransmissionColor(test, actualRadiance, color.NewColor(expectedRadiance, expectedRadiance, expectedRadiance))
				}

				if testCase.totalInternalReflection && stackStorage[len(stackStorage)-1] != stackSentinel {
					test.Fatal("TIR changed the medium stack")
				}
			})
		}
	}
}

// dielectricTestDirections returns unit incomingDirection and surfaceNormal
// vectors in the XZ plane. incidentAngleDegrees is measured from -Z and
// normalAngleDegrees from +Z; positive angles tilt each vector toward +X.
func dielectricTestDirections(incidentAngleDegrees, normalAngleDegrees float64) (incomingDirection, surfaceNormal vec3.T) {
	incidentAngle, normalAngle := incidentAngleDegrees*math.Pi/180, normalAngleDegrees*math.Pi/180
	return vec3.T{math.Sin(incidentAngle), 0, -math.Cos(incidentAngle)}, vec3.T{math.Sin(normalAngle), 0, math.Cos(normalAngle)}
}

// dielectricTestReference returns a unit outgoingDirection and unpolarized
// Fresnel reflectance for a dielectric interface in the XZ plane.
// incidentAngleDegrees is measured from -Z and normalAngleDegrees from +Z,
// with positive angles toward +X; the normal must oppose the incoming direction.
// incidentIndex and transmittedIndex are the positive refractive indices of
// the incident and transmitted media. Total internal reflection returns the
// reflected direction and reflectance one; otherwise the direction is refracted.
//
// This angular reference is independent of the production vector implementation.
// Snell's law and Fresnel equations: PBRT 4e, sections 9.3.4 and 9.3.5:
// https://pbr-book.org/4ed/Reflection_Models/Specular_Reflection_and_Transmission
func dielectricTestReference(incidentAngleDegrees, normalAngleDegrees, incidentIndex, transmittedIndex float64) (outgoingDirection vec3.T, reflectance float64) {
	incidentAngle, normalAngle := incidentAngleDegrees*math.Pi/180, normalAngleDegrees*math.Pi/180

	// Snell's law uses the angle to the tilted normal (PBRT 9.3.4).
	transmittedSine := incidentIndex / transmittedIndex * math.Sin(incidentAngle+normalAngle)
	if math.Abs(transmittedSine) > 1 {
		return vec3.T{math.Sin(incidentAngle + 2*normalAngle), 0, math.Cos(incidentAngle + 2*normalAngle)}, 1
	}

	// Convert the transmitted angle from the normal frame back to the world XZ plane.
	transmittedAngle := math.Asin(transmittedSine)
	outgoingDirection = vec3.T{math.Sin(transmittedAngle - normalAngle), 0, -math.Cos(transmittedAngle - normalAngle)}
	if incidentIndex == transmittedIndex {
		return outgoingDirection, 0
	}

	// Average squared s/p Fresnel amplitudes for unpolarized light (PBRT 9.3.5).
	incidentCosine, transmittedCosine := math.Cos(incidentAngle+normalAngle), math.Cos(transmittedAngle)
	perpendicularAmplitude := (incidentIndex*incidentCosine - transmittedIndex*transmittedCosine) / (incidentIndex*incidentCosine + transmittedIndex*transmittedCosine)
	parallelAmplitude := (transmittedIndex*incidentCosine - incidentIndex*transmittedCosine) / (transmittedIndex*incidentCosine + incidentIndex*transmittedCosine)
	return outgoingDirection, (perpendicularAmplitude*perpendicularAmplitude + parallelAmplitude*parallelAmplitude) / 2
}

// TestDielectricBoundary checks optical directions, Fresnel reflectance, and
// normal fallback without input mutation, against dielectricTestReference.
func TestDielectricBoundary(test *testing.T) {
	testCases := []struct {
		name                            string
		incidentAngleDegrees            float64
		normalAngleDegrees              float64
		incidentIndex                   float64
		transmittedIndex                float64
		expectedNormalAngleDegrees      float64
		expectedTotalInternalReflection bool
	}{
		{"smooth entry", 0, 30, 1, 1.5, 30, false},
		{"smooth exit", 0, 30, 1.5, 1, 30, false},
		{"smooth TIR despite geometric transmission", 20, 30, 1.5, 1, 30, true},
		{"smooth transmission despite geometric TIR", 60, -30, 1.5, 1, -30, false},
		{"wrong-side transmission becomes geometric TIR", 60, -20, 1.5, 1, 0, true},
		{"wrong-side TIR becomes geometric transmission", 0, 60, 1.5, 1, 0, false},
		{"matched indices", 80, -20, 1.5, 1.5, -20, false},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			incomingDirection, shadingNormal := dielectricTestDirections(testCase.incidentAngleDegrees, testCase.normalAngleDegrees)
			geometricNormal := vec3.T{0, 0, 1}
			originalDirections := [3]vec3.T{geometricNormal, shadingNormal, incomingDirection}

			boundary := evaluateDielectricBoundary(&geometricNormal, &shadingNormal, &incomingDirection, testCase.incidentIndex, testCase.transmittedIndex)
			expectedHeading, expectedReflectance := dielectricTestReference(testCase.incidentAngleDegrees, testCase.expectedNormalAngleDegrees, testCase.incidentIndex, testCase.transmittedIndex)
			_, expectedNormal := dielectricTestDirections(0, testCase.expectedNormalAngleDegrees)
			if boundary.heading == nil || vec3.Distance(boundary.heading, &expectedHeading) > 1e-12 ||
				vec3.Distance(&boundary.normal, &expectedNormal) > 1e-12 || math.Abs(boundary.fresnel-expectedReflectance) > 1e-12 ||
				boundary.totalInternalReflection != testCase.expectedTotalInternalReflection {
				test.Fatalf("boundary = %+v, direction %v; want normal %v, direction %v, F %g, TIR %t",
					boundary, boundary.heading, expectedNormal, expectedHeading, expectedReflectance, testCase.expectedTotalInternalReflection)
			}

			if [3]vec3.T{geometricNormal, shadingNormal, incomingDirection} != originalDirections {
				test.Fatal("modified input normals or incoming ray")
			}
		})
	}
}

// TestDielectricBoundaryLimits checks critical and grazing angles, matched
// indices, and invalid directions. See dielectricTestReference for Snell's law.
func TestDielectricBoundaryLimits(test *testing.T) {
	geometricNormal := vec3.T{0, 0, 1}

	// Snell's law gives a critical 3-4-5 ray at index ratio 5/4 (see dielectricTestReference).
	// Fused arithmetic must still leave a finite Fresnel value and tangent transmission.
	criticalIncomingDirection := vec3.T{0.8, 0, -0.6}
	criticalBoundary := evaluateDielectricBoundary(&geometricNormal, &geometricNormal, &criticalIncomingDirection, 1.25, 1)
	if !(math.Abs(criticalBoundary.fresnel-1) < 1e-12) || criticalBoundary.heading != nil || criticalBoundary.totalInternalReflection {
		test.Fatalf("critical boundary = %+v, want finite unit Fresnel and rejected tangent transmission", criticalBoundary)
	}

	// Sweep around the Snell critical angle, including tilted shading normals.
	criticalAngleDegrees := math.Asin(1/1.5) * 180 / math.Pi
	for _, incidentAngleDegrees := range []float64{0, criticalAngleDegrees - 1e-7, criticalAngleDegrees, criticalAngleDegrees + 1e-7, 89.999999} {
		for _, normalAngleDegrees := range []float64{0, -20, 20} {
			incomingDirection, shadingNormal := dielectricTestDirections(incidentAngleDegrees, normalAngleDegrees)
			if vec3.Dot(&incomingDirection, &shadingNormal) >= 0 {
				continue
			}
			boundary := evaluateDielectricBoundary(&geometricNormal, &shadingNormal, &incomingDirection, 1.5, 1)
			if !(boundary.fresnel >= 0 && boundary.fresnel <= 1) {
				test.Fatalf("nonfinite Fresnel: %+v", boundary)
			}
			if boundary.heading != nil {
				outgoingHemisphereNormal := geometricNormal
				if !boundary.totalInternalReflection {
					outgoingHemisphereNormal.Invert()
				}
				if !validSurfaceHeading(&outgoingHemisphereNormal, boundary.heading) {
					test.Fatalf("invalid direction: %+v", boundary)
				}
			}
		}
	}

	// Equal indices must preserve even extremely shallow nonzero geometric crossings.
	for _, incidentCosine := range []float64{1e-10, 1e-100, 1e-200} {
		incomingDirection := vec3.T{1, 0, -incidentCosine}
		boundary := evaluateDielectricBoundary(&geometricNormal, &geometricNormal, &incomingDirection, 1.5, 1.5)
		if boundary.heading == nil || *boundary.heading != incomingDirection || boundary.fresnel != 0 || boundary.totalInternalReflection {
			test.Fatalf("matched-index grazing result: %+v", boundary)
		}
	}

	// Tangent and nonfinite directions cannot cross a geometric boundary.
	for _, incomingDirection := range []vec3.T{{1, 0, 0}, {0, 0, 0}, {math.NaN(), 0, -1}, {math.Inf(1), 0, -1}} {
		boundary := evaluateDielectricBoundary(&geometricNormal, &geometricNormal, &incomingDirection, 1, 1)
		if boundary.heading != nil {
			test.Fatalf("accepted invalid or tangent direction: %v", boundary.heading)
		}
	}
}

// TestSmoothGlassSharedEdge checks continuity of refraction and Fresnel
// reflectance across a shared mesh edge, using barycentric normal interpolation
// from scene.FacetIntersection2 and optics from evaluateDielectricBoundary.
func TestSmoothGlassSharedEdge(test *testing.T) {
	lowerEdgeVertex, upperEdgeVertex := &vec3.T{0, -1, 0}, &vec3.T{0, 1, 0}
	lowerEdgeNormal, upperEdgeNormal := &vec3.T{0.2, 0, math.Sqrt(0.96)}, &vec3.T{0.4, 0, math.Sqrt(0.84)}
	leftFacet := &scene.Facet{Vertices: []*vec3.T{{-1, 0, -0.2}, lowerEdgeVertex, upperEdgeVertex}, VertexNormals: []*vec3.T{{0, 0, 1}, lowerEdgeNormal, upperEdgeNormal}}
	rightFacet := &scene.Facet{Vertices: []*vec3.T{lowerEdgeVertex, {1, 0, -0.2}, upperEdgeVertex}, VertexNormals: []*vec3.T{lowerEdgeNormal, {0.6, 0, 0.8}, upperEdgeNormal}}
	leftFacet.UpdateNormal()
	rightFacet.UpdateNormal()

	// Approach the shared edge from both triangles; the optical solutions should converge.
	previousOpticalGap := math.Inf(1)
	for _, edgeDistance := range []float64{1e-2, 1e-4, 1e-6} {
		var boundaries [2]dielectricBoundary
		for facetIndex, facet := range []*scene.Facet{leftFacet, rightFacet} {
			rayOriginX := edgeDistance
			if facetIndex == 0 {
				rayOriginX = -rayOriginX
			}
			ray := &scene.Ray{Origin: &vec3.T{rayOriginX, 0, 1}, Heading: &vec3.T{0, 0, -1}}
			hit, _, interpolatedNormal, vertexWeights := scene.FacetIntersection2(ray, facet)
			if !hit {
				test.Fatal("missed shared-edge fixture")
			}

			// Barycentric interpolation: the outer weight is |x| and the edge weights are (1-|x|)/2.
			// This fixture supplies independent weights for scene.FacetIntersection2.
			expectedNormal := lowerEdgeNormal.Scaled((1 - edgeDistance) / 2)
			upperNormalContribution := upperEdgeNormal.Scaled((1 - edgeDistance) / 2)
			expectedNormal.Add(&upperNormalContribution)
			outerNormalContribution := facet.VertexNormals[facetIndex].Scaled(edgeDistance)
			expectedNormal.Add(&outerNormalContribution).Normalize()
			if vec3.Distance(interpolatedNormal, &expectedNormal) > 1e-12 {
				test.Fatalf("wrong interpolation: normal %v, weights %v", interpolatedNormal, vertexWeights)
			}

			intersection := IntersectionInformation{intersectedFacet: facet, normalAtIntersection: interpolatedNormal}
			geometricNormal, shadingNormal, isEntering := orientSurfaceNormals(&intersection, ray.Heading)
			if !isEntering {
				test.Fatal("normal interpolation changed entry classification")
			}
			boundaries[facetIndex] = evaluateDielectricBoundary(&geometricNormal, &shadingNormal, ray.Heading, 1, 1.5)
			if boundaries[facetIndex].heading == nil || vec3.Distance(&boundaries[facetIndex].normal, &shadingNormal) > 1e-12 {
				test.Fatal("valid smooth boundary unexpectedly fell back")
			}
		}

		opticalGap := vec3.Distance(boundaries[0].heading, boundaries[1].heading) + math.Abs(boundaries[0].fresnel-boundaries[1].fresnel)
		if !(opticalGap < previousOpticalGap/10) {
			test.Fatalf("optics did not converge across shared edge: %g after %g", opticalGap, previousOpticalGap)
		}
		previousOpticalGap = opticalGap
	}

	if previousOpticalGap > 1e-6 {
		test.Fatalf("residual shared-edge discontinuity: %g", previousOpticalGap)
	}
}

// TestTracePathRejectedSolidTransmission checks that rejected transmission
// preserves local emission and the medium stack.
func TestTracePathRejectedSolidTransmission(test *testing.T) {
	for _, emissive := range []bool{false, true} {
		glass := scene.NewMaterial().M(0, 0).T(1, true, 1)
		glass.FresnelMaxGlossiness = 0 // Force the exactly critical, tangent transmission candidate.
		expectedRadiance := color.NewColor(0, 0, 0)
		if emissive {
			expectedRadiance = color.NewColor(0.2, 0.3, 0.4)
			glass.E(expectedRadiance, 1, false)
		}

		sceneNode := &scene.SceneNode{
			FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
			Spheres:         []*scene.Sphere{scene.NewSphere(&vec3.T{}, 100, scene.NewMaterial().E(color.White, 1, true))},
		}

		// Spare stack capacity exposes any push attempted by a rejected transmission.
		enclosingMedium := scene.NewMaterial().T(1, true, 1.25)
		stackSentinel := scene.NewMaterial()
		stackStorage := []*scene.Material{enclosingMedium, stackSentinel}
		mediumStack := stackStorage[:1]
		ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0.8, 0, -0.6}}

		// Verify a valid transmission selection before checking its tangent-direction rejection.
		boundary := evaluateDielectricBoundary(&vec3.UnitZ, &vec3.UnitZ, ray.Heading, 1.25, 1)
		probabilities := surfaceBounceProbabilities(glass, boundary.fresnel, boundary.totalInternalReflection)
		if boundary.heading != nil || probabilities.transmission != 1 {
			test.Fatalf("invalid critical-angle fixture: boundary %+v, probabilities %+v", boundary, probabilities)
		}

		actualRadiance := tracePath(ray, &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}, sceneNode, 0, mediumStack)
		assertTransmissionColor(test, actualRadiance, expectedRadiance)
		if stackStorage[1] != stackSentinel {
			test.Fatal("rejected transmission changed the medium stack")
		}
	}
}

// TestTracePathSmoothGlassNestedMedia checks transmission directions and
// radiance through nested smooth boundaries, using dielectricTestReference for Snell refraction.
func TestTracePathSmoothGlassNestedMedia(test *testing.T) {
	for _, nested := range []bool{false, true} {
		test.Run(fmt.Sprintf("nested=%t", nested), func(test *testing.T) {
			glass := scene.NewMaterial().M(0, 0).T(1, true, 1.5)
			water := scene.NewMaterial().M(0, 0).T(1, true, 1.33)
			glass.FresnelMaxGlossiness = 0
			water.FresnelMaxGlossiness = 0

			// Record each boundary and the medium on both sides in traversal order.
			boundaryMaterials := []*scene.Material{glass, glass}
			refractiveIndices := []float64{1, 1.5, 1}
			exitingBoundaries := []bool{false, true}
			if nested {
				boundaryMaterials = []*scene.Material{glass, water, water, glass}
				refractiveIndices = []float64{1, 1.5, 1.33, 1.5, 1}
				exitingBoundaries = []bool{false, false, true, true}
			}

			incomingDirection, shadingNormal := dielectricTestDirections(15, 20)
			rayOrigin := incomingDirection.Inverted()

			sceneNode := &scene.SceneNode{}
			lightPosition := vec3.T{}
			incidentAngleDegrees := 15.
			for boundaryIndex, material := range boundaryMaterials {
				plane := orientationTestPlane(material)
				vertexNormal := shadingNormal
				if exitingBoundaries[boundaryIndex] {
					plane.Facets[0].ChangeWindingOrder()
					plane.UpdateNormals()
					vertexNormal.Invert()
				}
				plane.Facets[0].VertexNormals = []*vec3.T{&vertexNormal, &vertexNormal, &vertexNormal}
				plane.Translate(&vec3.T{0, 0, -float64(boundaryIndex)})
				sceneNode.FacetStructures = append(sceneNode.FacetStructures, plane)

				// Advance one unit in -Z along the independent Snell-law solution.
				transmittedDirection, _ := dielectricTestReference(incidentAngleDegrees, 20, refractiveIndices[boundaryIndex], refractiveIndices[boundaryIndex+1])
				lightPosition.Add(&vec3.T{-transmittedDirection[0] / transmittedDirection[2], 0, -1})
				incidentAngleDegrees = math.Atan2(transmittedDirection[0], -transmittedDirection[2]) * 180 / math.Pi
			}

			// Returning to air cancels every index-ratio factor along this path.
			sceneNode.Discs = []*scene.Disc{scene.NewDisc(&lightPosition, &vec3.UnitZ, 0.002, scene.NewMaterial().E(color.White, 1, true))}
			actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &incomingDirection},
				&scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: len(boundaryMaterials)}, sceneNode, 0,
				[]*scene.Material{scene.NewMaterial().T(1, true, 1)})
			assertTransmissionColor(test, actualRadiance, color.White)
		})
	}
}

// TestTracePathShadingFresnel checks reflection probabilities from shading
// normals and geometric fallback, against the Fresnel equations in dielectricTestReference.
func TestTracePathShadingFresnel(test *testing.T) {
	testCases := []struct {
		name                       string
		normalAngleDegrees         float64
		expectedNormalAngleDegrees float64
		inside                     bool
		thin                       bool
	}{
		{"smooth entry", 60, 60, false, false},
		{"smooth exit", 40, 40, true, false},
		{"TIR fallback recomputes Fresnel", 60, 0, true, false},
		{"thin surface", 60, 60, false, true},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			incomingDirection, shadingNormal := dielectricTestDirections(0, testCase.normalAngleDegrees)
			glass := normalMappedTestMaterial(shadingNormal).M(0, 0).T(1, !testCase.thin, 1.5)
			mediumStack := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
			incidentIndex, transmittedIndex := 1., 1.5
			if testCase.inside {
				incomingDirection.Invert()
				mediumStack = append(mediumStack, glass)
				incidentIndex, transmittedIndex = transmittedIndex, incidentIndex
			}
			rayOrigin := incomingDirection.Inverted()

			// Red and green emitters identify the side reached by each scattered ray.
			sceneNode := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
				Discs: []*scene.Disc{
					scene.NewDisc(&vec3.T{0, 0, 2}, &vec3.UnitZ, 1e6, scene.NewMaterial().E(color.NewColor(1, 0, 0), 1, true)),
					scene.NewDisc(&vec3.T{0, 0, -2}, &vec3.UnitZ, 1e6, scene.NewMaterial().E(color.NewColor(0, 1, 0), 1, true)),
				},
			}

			_, expectedReflectance := dielectricTestReference(0, testCase.expectedNormalAngleDegrees, incidentIndex, transmittedIndex)
			const sampleCount = 30000
			reflectedCount := 0.
			for range sampleCount {
				actualRadiance := tracePath(&scene.Ray{Origin: &rayOrigin, Heading: &incomingDirection},
					&scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}, sceneNode, 0, mediumStack)
				reflectionChannel := actualRadiance.R
				if testCase.inside {
					reflectionChannel = actualRadiance.G
				}
				reflectedCount += float64(reflectionChannel)
				if actualRadiance.A != 1 {
					test.Fatalf("unexpected opacity: %v", actualRadiance)
				}
			}

			// Reflection counts are Bernoulli trials; allow eight standard errors.
			tolerance := 8*math.Sqrt(expectedReflectance*(1-expectedReflectance)/sampleCount) + 1e-4
			if math.Abs(reflectedCount/sampleCount-expectedReflectance) > tolerance {
				test.Fatalf("reflection fraction %g, want shading Fresnel %g ± %g", reflectedCount/sampleCount, expectedReflectance, tolerance)
			}
		})
	}
}
