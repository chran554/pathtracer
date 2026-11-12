package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestOrientSurfaceNormals checks incident-facing normals and the original geometric entry/exit sign.
func TestOrientSurfaceNormals(t *testing.T) {
	tests := []struct {
		name          string
		hit           IntersectionInformation
		incoming      vec3.T
		wantGeometric vec3.T
		wantShading   vec3.T
		wantIngoing   bool
	}{
		{
			"front mapped sphere",
			IntersectionInformation{normalAtIntersection: &vec3.T{0, 0, 1}, perturbedNormalAtIntersection: &vec3.T{0.6, 0, 0.8}},
			vec3.T{0, 0, -1}, vec3.T{0, 0, 1}, vec3.T{0.6, 0, 0.8}, true,
		},
		{
			"inside mapped sphere",
			IntersectionInformation{normalAtIntersection: &vec3.T{0, 0, 1}, perturbedNormalAtIntersection: &vec3.T{0.6, 0, 0.8}},
			vec3.T{0, 0, 1}, vec3.T{0, 0, -1}, vec3.T{-0.6, 0, -0.8}, false,
		},
		{
			"smooth facet facing away",
			IntersectionInformation{normalAtIntersection: &vec3.T{0.8, 0, 0.6}, intersectedFacet: &scene.Facet{Normal: &vec3.T{0, 0, 1}}},
			vec3.T{0.8, 0, -0.6}, vec3.T{0, 0, 1}, vec3.T{0, 0, 1}, true,
		},
		{
			"back disc with preflipped normal",
			IntersectionInformation{normalAtIntersection: &vec3.T{0, 0, -1}, intersectedDisc: &scene.Disc{Normal: &vec3.T{0, 0, 1}}},
			vec3.T{0, 0, 1}, vec3.T{0, 0, -1}, vec3.T{0, 0, -1}, false,
		},
		{
			"perturbed normal facing away",
			IntersectionInformation{normalAtIntersection: &vec3.T{0, 0, 1}, perturbedNormalAtIntersection: &vec3.T{0.8, 0, 0.6}},
			vec3.T{0.8, 0, -0.6}, vec3.T{0, 0, 1}, vec3.T{0, 0, 1}, true,
		},
		{
			"zero perturbed normal",
			IntersectionInformation{normalAtIntersection: &vec3.T{0, 0, 1}, perturbedNormalAtIntersection: &vec3.T{}},
			vec3.T{0, 0, -1}, vec3.T{0, 0, 1}, vec3.T{0, 0, 1}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := *tt.hit.normalAtIntersection
			var beforePerturbed vec3.T
			if tt.hit.perturbedNormalAtIntersection != nil {
				beforePerturbed = *tt.hit.perturbedNormalAtIntersection
			}
			geometric, shading, ingoing := orientSurfaceNormals(&tt.hit, &tt.incoming)
			if ingoing != tt.wantIngoing || !(vec3.Distance(&geometric, &tt.wantGeometric) < 1e-6) || !(vec3.Distance(&shading, &tt.wantShading) < 1e-6) {
				t.Fatalf("got normals %v/%v, ingoing=%t; want %v/%v, ingoing=%t", geometric, shading, ingoing, tt.wantGeometric, tt.wantShading, tt.wantIngoing)
			}
			// Orientation is local; intersection and scene normals remain reusable.
			if *tt.hit.normalAtIntersection != before || (tt.hit.perturbedNormalAtIntersection != nil && *tt.hit.perturbedNormalAtIntersection != beforePerturbed) {
				t.Fatal("orientation mutated the hit normals")
			}
		})
	}
}

// TestSurfaceHeadingFallback checks valid directions after every shading sample is rejected.
func TestSurfaceHeadingFallback(t *testing.T) {
	geometric := vec3.T{0, 0, 1}
	t.Run("diffuse", func(t *testing.T) {
		// An opposite hemisphere makes all 100 candidates invalid.
		got := sampleDiffuseHeading(&geometric, &vec3.T{0, 0, -1})
		if !validSurfaceHeading(&geometric, got) || !(vec3.Distance(got, &geometric) < 1e-6) {
			t.Fatalf("invalid diffuse fallback: %v", got)
		}
	})
	for _, roughness := range []float64{0, 0.1} {
		t.Run(fmt.Sprintf("reflection/roughness=%g", roughness), func(t *testing.T) {
			incoming := vec3.T{-0.6, 0, -0.8}
			// This tilt reflects below the surface even with the small roughness blend.
			got := sampleReflectionHeading(&geometric, &vec3.T{0.96, 0, 0.28}, &incoming, roughness)
			want := vec3.T{-0.6, 0, 0.8}
			if !validSurfaceHeading(&geometric, got) || !(vec3.Distance(got, &want) < 1e-6) {
				t.Fatalf("reflection fallback = %v, want %v", got, want)
			}
		})
	}
	t.Run("grazing reflection", func(t *testing.T) {
		got := sampleReflectionHeading(&geometric, &geometric, &vec3.T{1, 0, 0}, 0)
		if !validSurfaceHeading(&geometric, got) {
			t.Fatalf("grazing fallback must leave the surface: %v", got)
		}
	})
}

// TestTracePathReflectionFallback checks that an invalid mapped reflection reaches the geometric fallback light.
func TestTracePathReflectionFallback(t *testing.T) {
	for _, roughness := range []float64{0, 0.1} {
		t.Run(fmt.Sprintf("roughness=%g", roughness), func(t *testing.T) {
			material := normalMappedTestMaterial(vec3.T{0.96, 0, 0.28}).M(1, roughness)
			scn := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(material)},
				Discs:           []*scene.Disc{scene.NewDisc(&vec3.T{-1.5, 0, 1}, &vec3.T{0, 0, 1}, 0.005, scene.NewMaterial().E(color.White, 1, true))},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
			ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{-0.6, 0, -0.8}}
			// All mapped candidates point down; the geometric reflection points at the light.
			got := tracePath(ray, camera, scn, 0, []*scene.Material{scene.NewMaterial()})
			if got.R != 1 || got.G != 1 || got.B != 1 {
				t.Fatalf("fallback failed to reflect toward the light: %v", got)
			}
		})
	}
}

// TestTracePathInsideOpaqueSphere checks that interior reflections cannot see an outside light.
func TestTracePathInsideOpaqueSphere(t *testing.T) {
	for _, mode := range []string{"diffuse", "mirror", "rough reflection"} {
		for _, mapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/mapped=%t", mode, mapped), func(t *testing.T) {
				material := orientationTestMaterial(mode, mapped)
				material.SolidObject = true
				scn := &scene.SceneNode{Spheres: []*scene.Sphere{
					scene.NewSphere(&vec3.T{}, 1, material),
					scene.NewSphere(&vec3.T{}, 10, scene.NewMaterial().E(color.White, 1, true)),
				}}
				camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
				ray := &scene.Ray{Origin: &vec3.T{}, Heading: &vec3.T{0, 0, 1}}

				// Every bounce must stay inside the unlit, opaque shell.
				for range 8 {
					got := tracePath(ray, camera, scn, 0, []*scene.Material{scene.NewMaterial(), material})
					if got.R != 0 || got.G != 0 || got.B != 0 {
						t.Fatalf("outside light leaked through the sphere: %v", got)
					}
				}
			})
		}
	}
}

// TestTracePathSurfaceSides checks front and back scattering, including flat normal maps.
func TestTracePathSurfaceSides(t *testing.T) {
	for _, mode := range []string{"diffuse", "mirror", "rough reflection"} {
		for _, solid := range []bool{false, true} {
			for _, mapped := range []bool{false, true} {
				for _, back := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/solid=%t/mapped=%t/back=%t", mode, solid, mapped, back), func(t *testing.T) {
						material := orientationTestMaterial(mode, mapped)
						material.SolidObject = solid
						red, green := color.NewColor(1, 0, 0), color.NewColor(0, 1, 0)
						scn := &scene.SceneNode{
							FacetStructures: []*scene.FacetStructure{orientationTestPlane(material)},
							Discs: []*scene.Disc{
								scene.NewDisc(&vec3.T{0, 0, 2}, &vec3.T{0, 0, 1}, 1e12, scene.NewMaterial().E(red, 1, true)),
								scene.NewDisc(&vec3.T{0, 0, -2}, &vec3.T{0, 0, 1}, 1e12, scene.NewMaterial().E(green, 1, true)),
							},
						}
						ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0, 0, -1}}
						want := red
						if back {
							ray.Origin.Invert()
							ray.Heading.Invert()
							want = green
						}
						camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}

						// Differently colored lights reveal which side the bounce reaches.
						for range 4 {
							got := tracePath(ray, camera, scn, 0, []*scene.Material{scene.NewMaterial()})
							if got.R != want.R || got.G != want.G || got.B != want.B {
								t.Fatalf("bounce reached the wrong side: got %v, want %v", got, want)
							}
						}
					})
				}
			}
		}
	}
}

// TestTracePathRefractionOrientation checks entry, exit, and total internal reflection after normal orientation.
func TestTracePathRefractionOrientation(t *testing.T) {
	tests := []struct {
		name    string
		origin  vec3.T
		heading vec3.T
		want    vec3.T
		inside  bool
		targetZ float64
	}{
		{"enter", vec3.T{0, 0, 1}, vec3.T{0.5, 0, -math.Sqrt(0.75)}, vec3.T{1.0 / 3, 0, -math.Sqrt(8.0 / 9)}, false, -1},
		{"exit", vec3.T{0, 0, -1}, vec3.T{0.5, 0, math.Sqrt(0.75)}, vec3.T{0.75, 0, math.Sqrt(7.0 / 16)}, true, 1},
		{"total internal reflection", vec3.T{0, 0, -1}, vec3.T{math.Sqrt(0.75), 0, 0.5}, vec3.T{math.Sqrt(0.75), 0, -0.5}, true, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glass := normalMappedTestMaterial(vec3.T{0.6, 0, 0.8}).T(1, true, 1.5)
			air := scene.NewMaterial().T(1, true, 1)
			contexts := []*scene.Material{air}
			if tt.inside {
				contexts = append(contexts, glass)
			}

			// A small light lies on the direction predicted by Snell's law.
			hitX := -tt.origin[2] / tt.heading[2] * tt.heading[0]
			target := vec3.T{hitX + tt.targetZ/tt.want[2]*tt.want[0], 0, tt.targetZ}
			scn := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
				Discs:           []*scene.Disc{scene.NewDisc(&target, &vec3.T{0, 0, 1}, 0.005, scene.NewMaterial().E(color.White, 1, true))},
			}
			camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
			got := tracePath(&scene.Ray{Origin: &tt.origin, Heading: &tt.heading}, camera, scn, 0, contexts)
			if got.R != 1 || got.G != 1 || got.B != 1 {
				t.Fatalf("ray missed the expected refracted/reflected direction: %v", got)
			}
		})
	}
}

// orientationTestMaterial forces one scattering mode, optionally with a flat normal map.
func orientationTestMaterial(mode string, mapped bool) *scene.Material {
	material := scene.NewMaterial()
	if mapped {
		material = normalMappedTestMaterial(vec3.T{0, 0, 1})
	}
	material.FresnelMaxGlossiness = 0
	if mode != "diffuse" {
		material.Glossiness = 1
		material.FresnelMaxGlossiness = 1
		material.Roughness = 0
		if mode == "rough reflection" {
			material.Roughness = 1
		}
	}
	return material
}

// orientationTestPlane creates a triangle in z=0 with an outward normal of +Z.
func orientationTestPlane(material *scene.Material) *scene.FacetStructure {
	facet := &scene.Facet{Vertices: []*vec3.T{{-10, -10, 0}, {10, -10, 0}, {0, 10, 0}}}
	structure := &scene.FacetStructure{Material: material, Facets: []*scene.Facet{facet}}
	structure.Initialize()
	return structure
}
