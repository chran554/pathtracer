package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestCloserIntersectionClearsPerturbedNormal checks that untextured hits discard the previous normal map.
func TestCloserIntersectionClearsPerturbedNormal(t *testing.T) {
	for _, shape := range []string{"sphere", "disc", "facet"} {
		for _, missingMaterial := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/missingMaterial=%t", shape, missingMaterial), func(t *testing.T) {
				ray, ii := normalMappedTestHit(t, 5)
				material := scene.NewMaterial()
				if missingMaterial {
					material = nil
				}

				// Visit a closer object after the mapped sphere to reproduce traversal order.
				intersectTestShape(t, shape, ray, 1, material, ii)
				if ii.shortestDistance != 1 {
					t.Fatalf("closer hit was not accepted: distance = %g", ii.shortestDistance)
				}
				if ii.perturbedNormalAtIntersection != nil {
					t.Fatalf("closer %s inherited a perturbed normal: %v", shape, ii.perturbedNormalAtIntersection)
				}
			})
		}
	}
}

// TestCloserIntersectionUsesOwnNormalMap checks that clearing stale state preserves the new hit's map.
func TestCloserIntersectionUsesOwnNormalMap(t *testing.T) {
	for _, shape := range []string{"sphere", "facet"} {
		t.Run(shape, func(t *testing.T) {
			ray, ii := normalMappedTestHit(t, 5)
			material := normalMappedTestMaterial(vec3.T{-0.6, 0, 0.8})

			// The closer map tilts in the opposite direction to the old map.
			intersectTestShape(t, shape, ray, 1, material, ii)
			if ii.shortestDistance != 1 || ii.material != material {
				t.Fatal("closer mapped hit was not accepted")
			}
			want := vec3.T{-0.6, 0, -0.8}
			got := ii.perturbedNormalAtIntersection
			if got == nil || vec3.Distance(got, &want) > 1e-6 {
				t.Fatalf("perturbed normal = %v, want %v", got, want)
			}
		})
	}
}

// TestRejectedIntersectionPreservesPerturbedNormal checks that farther and tied hits leave shading intact.
func TestRejectedIntersectionPreservesPerturbedNormal(t *testing.T) {
	for _, shape := range []string{"sphere", "disc", "facet"} {
		for _, distance := range []float64{1, 3} {
			t.Run(fmt.Sprintf("%s/distance=%g", shape, distance), func(t *testing.T) {
				ray, ii := normalMappedTestHit(t, 1)
				before := *ii

				// Reset shading only when the candidate becomes the closest hit.
				intersectTestShape(t, shape, ray, distance, scene.NewMaterial(), ii)
				if *ii != before {
					t.Fatal("rejected intersection changed the closest hit")
				}
			})
		}
	}
}

// normalMappedTestHit creates a real sphere hit with a tilted shading normal.
func normalMappedTestHit(t *testing.T, distance float64) (*scene.Ray, *IntersectionInformation) {
	t.Helper()
	ray := &scene.Ray{Origin: &vec3.T{}, Heading: &vec3.T{0, 0, 1}}
	ii := NewIntersectionInformation()
	material := normalMappedTestMaterial(vec3.T{0.6, 0, 0.8})
	intersectTestShape(t, "sphere", ray, distance, material, ii)
	want := vec3.T{0.6, 0, -0.8}
	got := ii.perturbedNormalAtIntersection
	if !ii.intersection || ii.shortestDistance != distance || got == nil || vec3.Distance(got, &want) > 1e-6 {
		t.Fatalf("mapped fixture has distance %g and normal %v", ii.shortestDistance, got)
	}
	return ray, ii
}

// normalMappedTestMaterial supplies a constant tangent-space normal without loading files.
func normalMappedTestMaterial(normal vec3.T) *scene.Material {
	image := floatimage.NewFloatImage("test normal map", 1, 1)
	// GetNormal expects vectors already decoded from image colors.
	image.Fill(color.NewColor(normal[0], normal[1], normal[2]))
	projection := scene.NewParallelImageProjection(image, scene.TextureTypeNormalMap, &vec3.T{}, vec3.UnitX, vec3.UnitY)
	projection.Initialize()
	return scene.NewMaterial().P(&projection)
}

// intersectTestShape places a surface at z on the test ray, with its normal facing -Z.
func intersectTestShape(t *testing.T, shape string, ray *scene.Ray, z float64, material *scene.Material, ii *IntersectionInformation) {
	t.Helper()
	switch shape {
	case "sphere":
		sphere := scene.NewSphere(&vec3.T{0, 0, z + 1}, 1, material)
		processSphereIntersection(ray, sphere, ii)
	case "disc":
		disc := scene.NewDisc(&vec3.T{0, 0, z}, &vec3.T{0, 0, -1}, 2, material)
		processDiscIntersection(ray, disc, ii)
	case "facet":
		facet := &scene.Facet{Vertices: []*vec3.T{{-2, -2, z}, {0, 2, z}, {2, -2, z}}}
		structure := &scene.FacetStructure{Material: material, Facets: []*scene.Facet{facet}}
		structure.Initialize()
		processFacetStructureIntersection(ray, structure, ii)
	default:
		t.Fatalf("unknown test shape %q", shape)
	}
}
