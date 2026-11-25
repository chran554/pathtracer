package main

import (
	"math"
	"math/rand/v2"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

// TestMetalFresnel checks RGB reflectance and independence from dielectric parameters.
// t records invalid conductor responses.
func TestMetalFresnel(t *testing.T) {
	base := [3]float64{0.8, 0.4, 0.1}
	for _, cosine := range []float64{1, 0.5, 0.01, 0} {
		f := metalFresnel(base, cosine)
		for c := range f {
			want := base[c] + (1-base[c])*math.Pow(1-cosine, 5)
			if math.Abs(f[c]-want) > 1e-14 {
				t.Fatalf("F(%g)=%v", cosine, f)
			}
		}
	}
	// Pure metals ignore transmission and dielectric IOR, including dielectric TIR.
	for _, roughness := range []float64{0, 0.3, 1} {
		material := scene.NewMaterial().M(1, roughness).T(1, true, 1)
		rng := rand.New(rand.NewPCG(1, 2))
		incoming := vec3.T{math.Sqrt(0.75), 0, -0.5}
		for range 1000 {
			s := sampleSurface(material, material.Roughness, material.Metallicness, base, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1.5, 1, rng.Float64)
			if s.heading != nil && s.bounce != reflectionBounce {
				t.Fatalf("metal emitted nonreflection: %+v", s)
			}
			for _, w := range s.weight {
				if !(w >= 0 && !math.IsInf(w, 0)) {
					t.Fatalf("metal weight %g", w)
				}
			}
		}
	}
}

// TestSurfaceSamplingIntegral checks mixed material weights against the separate materials.
// t reports discrepancies under nonuniform illumination.
func TestSurfaceSamplingIntegral(t *testing.T) {
	inputColor := [3]float64{0.8, 0.3, 0.1}
	incoming := vec3.T{0.6, 0, -0.8}
	const roughness, metallicness, transparency = 0.7, 0.4, 0.6
	const samples = 50000
	var expected [3]float64
	for _, component := range []struct {
		metal, transparent, fraction float64
	}{
		{1, 0, metallicness}, {0, 0, (1 - metallicness) * (1 - transparency)},
		{0, 1, (1 - metallicness) * transparency}, {metallicness, transparency, 1},
	} {
		material := scene.NewMaterial().M(component.metal, roughness).T(component.transparent, false, 1.5)
		random := rand.New(rand.NewPCG(37, 99))
		var sum [3]float64
		for range samples {
			sample := sampleSurface(material, roughness, component.metal, inputColor, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, random.Float64)
			if sample.heading != nil {
				lighting := 0.55 + 0.45*sample.heading[2]
				for channel, weight := range sample.weight {
					sum[channel] += weight * lighting
				}
			}
		}
		for channel, total := range sum {
			mean := total / samples
			if component.metal == metallicness {
				if !(math.Abs(mean-expected[channel]) < 0.006) {
					t.Fatalf("channel %d mean %g, want %g", channel, mean, expected[channel])
				}
			} else {
				expected[channel] += component.fraction * mean
			}
		}
	}
}

// TestSurfaceEnergyBounds checks sampled unit-white reflection over material mixtures.
// t records white-furnace energy outside the physical bounds.
func TestSurfaceEnergyBounds(t *testing.T) {
	const samples = 12000
	for _, index := range []float64{1, 1.1, 1.5, 2.5} {
		for _, metal := range []float64{0, 0.5, 1} {
			for _, roughness := range []float64{0.35, 0.7, 1} {
				material := scene.NewMaterial().M(metal, roughness).T(0, false, index)
				for _, cosine := range []float64{0.05, 0.5, 1} {
					incoming := vec3.T{math.Sqrt(1 - cosine*cosine), 0, -cosine}
					random := rand.New(rand.NewPCG(32, 45))
					var sum, squares float64
					for range samples {
						sample := sampleSurface(material, roughness, metal, [3]float64{1, 1, 1}, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, index, random.Float64)
						weight := sample.weight[0]
						sum += weight
						squares += weight * weight
					}
					mean := sum / samples
					tolerance := 6*math.Sqrt(max(0, squares/samples-mean*mean)/samples) + 0.001
					if !(mean >= 0 && mean <= 1+tolerance) {
						t.Fatalf("index/metal/roughness/cosine %g/%g/%g/%g: mean %g", index, metal, roughness, cosine, mean)
					}
				}
			}
		}
	}
}

// TestRoughMetalWhiteFurnace measures a white sphere with one available scene bounce.
// t records the mean linear RGB at the center and near the silhouette.
func TestRoughMetalWhiteFurnace(t *testing.T) {
	media := []*scene.Material{scene.NewMaterial().T(1, true, 1)}
	environment := scene.NewSphere(&vec3.T{}, 100, scene.NewMaterial().E(color.White, 1, true))
	for _, roughness := range []float64{0.25, 0.6, 1} {
		material := scene.NewMaterial().C(color.White).M(1, roughness)
		spheres := scene.NewSceneNode().S(scene.NewSphere(&vec3.T{}, 1, material), environment)
		for _, offset := range []float64{0, 0.99} {
			ray := &scene.Ray{Origin: &vec3.T{offset, 0, 2}, Heading: &vec3.T{0, 0, -1}}
			mean := meanTestPathRadiance(ray, spheres, media, 1, 20000)
			t.Logf("roughness %g, offset %g: linear RGB %v", roughness, offset, mean)
			for _, channel := range mean {
				if !(math.Abs(channel-1) < 0.006) {
					t.Fatalf("roughness %g, offset %g: mean RGB %v, want white", roughness, offset, mean)
				}
			}
		}
	}
}

// Independent scalar unpolarized Fresnel reference for air -> material.
func surfaceReferenceFresnel(cos, ior float64) float64 {
	if ior == 1 {
		return 0
	}
	ct := math.Sqrt(1 - (1-cos*cos)/(ior*ior))
	rs := (cos - ior*ct) / (cos + ior*ct)
	rp := (ior*cos - ct) / (ior*cos + ct)
	return (rs*rs + rp*rp) / 2
}

func TestTraceSurfaceBaseColor(t *testing.T) {
	metal := scene.NewMaterial().M(1, 0).T(1, true, 1.5).C(color.NewColor(0.8, 0.4, 0.2))
	setTestAlbedoTexture(metal, color.NewColor(0.5, 0.25, 1))
	assertTransmissionColor(t, bounceTestTrace(bounceTestScene(metal)), color.NewColor(0.4, 0.1, 0.2))

	// A zero-albedo, index-matched dielectric absorbs while preserving emission.
	material := scene.NewMaterial().C(color.Black).T(0, false, 1)
	assertTransmissionColor(t, bounceTestTrace(bounceTestScene(material)), color.Black)
}

func TestSurfaceMirrorColor(t *testing.T) {
	base := [3]float64{0.8, 0.4, 0.1}
	incoming := vec3.T{0, 0, -1}
	// Zero samples choose reflection in both models. Dielectric Fresnel cancels
	// its selection probability; metal RGB reflectance remains in throughput.
	for _, metallic := range []float64{0, 1} {
		material := scene.NewMaterial().M(metallic, 0)
		s := sampleSurface(material, material.Roughness, material.Metallicness, base, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, func() float64 { return 0 })
		want := [3]float64{1, 1, 1}
		if metallic == 1 {
			want = base
		}
		if s.heading == nil || *s.heading != vec3.UnitZ || s.bounce != reflectionBounce {
			t.Fatalf("invalid mirror: %+v", s)
		}
		for i := range want {
			if math.Abs(s.weight[i]-want[i]) > 1e-12 {
				t.Fatalf("m=%g weight %v, want %v", metallic, s.weight, want)
			}
		}
	}
}
