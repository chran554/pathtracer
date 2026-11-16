package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

func TestBeer(test *testing.T) {
	for _, tc := range []struct {
		name       string
		absorption *color.Color
		distance   float64
		want       *color.Color
	}{
		{"unset", nil, 10, color.White},
		{"vacuum at infinity", color.Black, math.Inf(1), color.White},
		{"zero length", color.NewColor(10, 20, 30), 0, color.White},
		{"channel half lives", color.NewColorRGBA(math.Ln2, 2*math.Ln2, 0, 0), 2, color.NewColor(0.25, 0.0625, 1)},
		{"infinite length", color.NewColor(1, 0, 2), math.Inf(1), color.NewColor(0, 1, 0)},
		{"underflow", color.NewColorGrey(1000), 1000, color.Black},
		{"negative coefficients cannot amplify", color.NewColor(-1, 0, -2), 2, color.White},
		{"negative length cannot amplify", color.NewColorGrey(1), -1, color.White},
	} {
		test.Run(tc.name, func(test *testing.T) {
			assertTransmissionColor(test, Beer(tc.absorption, tc.distance), tc.want)
		})
	}

	// Splitting the same medium at an arbitrary surface must not change extinction.
	absorption := color.NewColor(0.1, 0.2, 0.7)
	first, second := Beer(absorption, 2), Beer(absorption, 3)
	assertTransmissionColor(test, Beer(absorption, 5), color.NewColor(
		float64(first.R*second.R), float64(first.G*second.G), float64(first.B*second.B)))
}

// Analytic diameters isolate absorption from Fresnel roulette and verify that
// inner media replace (rather than add to) the enclosing medium.
func TestTracePathBeerVolumes(test *testing.T) {
	for _, tc := range []struct {
		name       string
		cameraZ    float64
		inner      bool
		innerColor *color.Color
	}{
		{name: "filled sphere", cameraZ: 3},
		{name: "same sphere farther from camera", cameraZ: 30},
		{name: "nested absorbing solid", cameraZ: 3, inner: true, innerColor: color.NewColor(0.7, 0.2, 0)},
		{name: "clear bubble", cameraZ: 3, inner: true},
	} {
		test.Run(tc.name, func(test *testing.T) {
			outer := scene.NewMaterial().C(color.NewColor(0.2, 0.4, 0.6)).
				T(1, true, 1.5).Absorb(color.NewColor(0.1, 0.3, 0.7))
			outer.FresnelMaxGlossiness = 0
			spheres := []*scene.Sphere{scene.NewSphere(&vec3.T{}, 2, outer)}
			outerDistance, innerDistance := 4-epsilonDistance, 0.0
			depth := 2
			innerAbsorption := color.Black
			if tc.inner {
				inner := scene.NewMaterial().T(1, true, 1.33).Absorb(tc.innerColor)
				inner.FresnelMaxGlossiness = 0
				spheres = append(spheres, scene.NewSphere(&vec3.T{}, 1, inner))
				outerDistance, innerDistance = 2-2*epsilonDistance, 2-epsilonDistance
				depth = 4
				if tc.innerColor != nil {
					innerAbsorption = tc.innerColor
				}
			}
			light := color.NewColor(0.25, 0.5, 2)
			spheres = append(spheres, scene.NewSphere(&vec3.T{}, 100, scene.NewMaterial().E(light, 1, true)))
			actual := tracePath(&scene.Ray{Origin: &vec3.T{0, 0, tc.cameraZ}, Heading: &vec3.T{0, 0, -1}},
				&scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: depth}, &scene.SceneNode{Spheres: spheres},
				0, []*scene.Material{scene.NewMaterial().T(1, true, 1)})
			want := color.NewColor(
				float64(light.R)*math.Exp(-float64(outer.Absorption.R)*outerDistance-float64(innerAbsorption.R)*innerDistance),
				float64(light.G)*math.Exp(-float64(outer.Absorption.G)*outerDistance-float64(innerAbsorption.G)*innerDistance),
				float64(light.B)*math.Exp(-float64(outer.Absorption.B)*outerDistance-float64(innerAbsorption.B)*innerDistance))
			assertTransmissionColor(test, actual, want)
		})
	}
}

func TestTracePathBeerSurfaceEmission(test *testing.T) {
	// The light's material must not absorb its own emission at the boundary.
	// Even a ray terminating at depth zero travels through the incident medium.
	light := scene.NewMaterial().E(color.NewColor(2, 1, 0.5), 1, true).Absorb(color.NewColorGrey(100))
	scn := &scene.SceneNode{FacetStructures: []*scene.FacetStructure{orientationTestPlane(light)}}
	medium := scene.NewMaterial().Absorb(color.NewColor(math.Ln2, 0, 2*math.Ln2))
	actual := tracePath(&scene.Ray{Origin: &vec3.T{0, 0, 2}, Heading: &vec3.T{0, 0, -1}},
		&scene.Camera{RenderType: scene.Pathtracing}, scn, 0, []*scene.Material{medium})
	assertTransmissionColor(test, actual, color.NewColor(0.5, 1, 0.03125))
}

func TestTracePathBeerUncrossedBoundaries(test *testing.T) {
	for _, name := range []string{"thin transmission", "alpha hole", "exterior reflection"} {
		test.Run(name, func(test *testing.T) {
			material := scene.NewMaterial().T(1, true, 1.5).Absorb(color.NewColorGrey(100))
			material.FresnelMaxGlossiness = 0
			switch name {
			case "thin transmission":
				material.SolidObject = false
			case "alpha hole":
				material.C(color.WhiteTransparent)
			case "exterior reflection":
				material.T(0, true, 1.5).M(1, 0)
			}
			scn := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(material)},
				Spheres:         []*scene.Sphere{scene.NewSphere(&vec3.T{}, 10, scene.NewMaterial().E(color.White, 1, true))},
			}
			actual := tracePath(&scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0, 0, -1}},
				&scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}, scn, 0, []*scene.Material{scene.NewMaterial()})
			assertTransmissionColor(test, actual, color.White)
		})
	}
}

func TestTracePathBeerInternalReflection(test *testing.T) {
	for _, tir := range []bool{false, true} {
		name := "mirror inside medium"
		if tir {
			name = "total internal reflection"
		}
		test.Run(name, func(test *testing.T) {
			medium := scene.NewMaterial().T(1, true, 1.5).Absorb(color.NewColor(0.1, 0.2, 0.3))
			boundary := medium
			if !tir {
				boundary = scene.NewMaterial().M(1, 0)
			}
			// cos(theta)=0.5 forces TIR from glass to air. Each leg spans one
			// unit along z, so the total geometric length is four units.
			scn := &scene.SceneNode{
				FacetStructures: []*scene.FacetStructure{orientationTestPlane(boundary)},
				Discs: []*scene.Disc{scene.NewDisc(&vec3.T{2 * math.Sqrt(3), 0, -1}, &vec3.UnitZ, 0.01,
					scene.NewMaterial().E(color.White, 1, true))},
			}
			stack := []*scene.Material{scene.NewMaterial().T(1, true, 1), medium}
			actual := tracePath(&scene.Ray{Origin: &vec3.T{0, 0, -1}, Heading: &vec3.T{math.Sqrt(0.75), 0, 0.5}},
				&scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}, scn, 0, stack)
			distance := 4 - 2*epsilonDistance
			assertTransmissionColor(test, actual, color.NewColor(
				math.Exp(-float64(medium.Absorption.R)*distance),
				math.Exp(-float64(medium.Absorption.G)*distance),
				math.Exp(-float64(medium.Absorption.B)*distance)))
		})
	}
}
