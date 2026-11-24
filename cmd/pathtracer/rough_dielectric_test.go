package main

import (
	"math"
	"math/rand/v2"
	"testing"

	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestRoughGlassSecondFacet follows a reflection that points back into the
// microsurface, then transmits through a second facet at normal incidence.
func TestRoughGlassSecondFacet(t *testing.T) {
	incoming := vec3.UnitZ.Inverted()
	// The first normal tilts 60 degrees toward +X. The reflected ray travels
	// downward, where the second visible normal is exactly opposite that ray.
	random := glassTestDraws(t,
		0.5, 0.75, 0.25, 0,
		0.5, 1.0/9, 0.75, 0.99,
		0.999,
	)
	sample := sampleRoughDielectric(&vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1, 1.5, random)

	wantDirection := vec3.T{math.Sqrt(0.75), 0, -0.5}
	if sample.bounce != transmissionBounce || sample.heading == nil || vec3.Distance(sample.heading, &wantDirection) > 1e-12 {
		t.Fatalf("two-facet walk returned %+v, want transmission along %v", sample, wantDirection)
	}
	for _, weight := range sample.weight {
		if !(math.Abs(weight-4.0/9) < 1e-12) {
			t.Fatalf("transmission weight %g, want 4/9", weight)
		}
	}
}

// TestRoughGlassWhiteFurnace checks a clear sphere in uniform unit radiance.
// Both crossings and all internal reflections are sampled by the scene tracer.
func TestRoughGlassWhiteFurnace(t *testing.T) {
	for _, roughness := range []float64{0.25, 0.6, 1} {
		glass := scene.NewMaterial().M(0, roughness).T(1, true, 1.5)
		environment := scene.NewMaterial().E(color.White, 1, true)
		testScene := scene.NewSceneNode().S(scene.NewSphere(&vec3.T{}, 1, glass), scene.NewSphere(&vec3.T{}, 100, environment))
		media := []*scene.Material{scene.NewMaterial().T(1, true, 1)}

		for _, offset := range []float64{0, 0.9} {
			ray := &scene.Ray{Origin: &vec3.T{offset, 0, 2}, Heading: &vec3.T{0, 0, -1}}
			mean := meanGlassTestRadiance(ray, testScene, media, 96, 20000)
			t.Logf("roughness %.2f, ray offset %.1f: linear radiance %.5f", roughness, offset, mean[0])
			for _, channel := range mean {
				if !(math.Abs(channel-1) < 0.006) {
					t.Fatalf("roughness %g, ray offset %g: mean RGB %v, want white", roughness, offset, mean)
				}
			}
		}
	}
}

// TestRoughGlassTwoSidedLighting measures reflection and transmission separately
// using red light on the air side and blue light on the glass side of one boundary.
func TestRoughGlassTwoSidedLighting(t *testing.T) {
	glass := scene.NewMaterial().M(0, 1).T(1, true, 1.5)
	air := scene.NewMaterial().T(1, true, 1)
	testScene := &scene.SceneNode{
		FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
		Discs: []*scene.Disc{
			scene.NewDisc(&vec3.T{0, 0, 2}, &vec3.UnitZ, 1e9, scene.NewMaterial().E(color.NewColor(1, 0, 0), 1, true)),
			scene.NewDisc(&vec3.T{0, 0, -2}, &vec3.UnitZ, 1e9, scene.NewMaterial().E(color.NewColor(0, 0, 1), 1, true)),
		},
	}

	// Independent reference, two million walks per configuration. At alpha=1
	// D=1/pi, so a cosine-weighted hemisphere about the view, clipped to upward
	// normals by rejection, samples visible normals directly. Heights used
	// C(next)=C(current)/U^(1/Lambda) with Lambda=(1/cosine-1)/2.
	for _, testCase := range []struct {
		name            string
		fromGlass       bool
		incidenceCosine float64
		reflectance     float64
	}{
		{"air normal", false, 1, 0.01546},
		{"air oblique", false, 0.5, 0.02762},
		{"glass normal", true, 1, 0.40585},
		{"glass oblique", true, 0.5, 0.65705},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			direction := vec3.T{math.Sqrt(1 - testCase.incidenceCosine*testCase.incidenceCosine), 0, -testCase.incidenceCosine}
			media := []*scene.Material{air}
			indexRatio := 1 / glass.RefractionIndex
			if testCase.fromGlass {
				direction.Invert()
				media = append(media, glass)
				indexRatio = glass.RefractionIndex
			}

			origin := direction.Inverted() // All rays meet the plane at the origin.
			ray := &scene.Ray{Origin: &origin, Heading: &direction}
			const samples = 60000
			mean := meanGlassTestRadiance(ray, testScene, media, 1, samples)
			reflected, transmitted := mean[0], mean[2]
			if testCase.fromGlass {
				reflected, transmitted = mean[2], mean[0]
			}

			// Undo radiance conversion to compare both colors in units of incident power.
			transmitted /= indexRatio * indexRatio
			t.Logf("reflected %.5f, transmitted %.5f", reflected, transmitted)
			want := testCase.reflectance
			tolerance := 6*math.Sqrt(want*(1-want)/samples) + 0.001
			if !(math.Abs(reflected-want) < tolerance && math.Abs(transmitted-(1-want)) < tolerance && mean[1] == 0) {
				t.Fatalf("reflected/transmitted %g/%g, want %g/%g ± %g", reflected, transmitted, want, 1-want, tolerance)
			}
		})
	}
}

// meanGlassTestRadiance returns the mean linear RGB for count deterministic paths
// along ray in testScene, starting in media and allowing depth scene bounces.
func meanGlassTestRadiance(ray *scene.Ray, testScene *scene.SceneNode, media []*scene.Material, depth, count int) [3]float64 {
	camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: depth}
	random := rand.New(rand.NewPCG(42, 17))
	var sum [3]float64
	for range count {
		radiance := tracePathSampled(ray, camera, testScene, 0, media, random.Float64)
		sum[0] += float64(radiance.R)
		sum[1] += float64(radiance.G)
		sum[2] += float64(radiance.B)
	}

	for channel := range sum {
		sum[channel] /= float64(count)
	}
	return sum
}

// glassTestDraws returns the given uniform draws in order. It fails t if the
// tested path consumes a different number of draws than the scripted walk.
func glassTestDraws(t *testing.T, draws ...float64) func() float64 {
	t.Helper()
	t.Cleanup(func() {
		if len(draws) > 0 {
			t.Errorf("walk left %d unused draws", len(draws))
		}
	})

	return func() float64 {
		t.Helper()
		if len(draws) == 0 {
			t.Fatal("walk exceeded its scripted draws")
		}

		next := draws[0]
		draws = draws[1:]
		return next
	}
}
