package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/ungerik/go3d/float64/vec3"
)

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
				if !(w >= 0 && w <= 1+1e-12) {
					t.Fatalf("metal weight %g", w)
				}
			}
		}
	}
}

// Integrate the model independently in solid angle, then compare against the
// complete stochastic mixture (including its discrete component probabilities).
func TestSurfaceSamplingIntegral(t *testing.T) {
	for _, tc := range []struct{ m, r, tr, cos float64 }{
		{0, 0.5, 0, 1}, {0, 1, 0, 0.2}, {1, 0.5, 0, 0.6}, {0.4, 0.7, 0.6, 0.8},
	} {
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			base := [3]float64{0.8, 0.3, 0.1}
			material := scene.NewMaterial().M(tc.m, tc.r).T(tc.tr, false, 1.5)
			incoming := vec3.T{math.Sqrt(1 - tc.cos*tc.cos), 0, -tc.cos}
			want := surfaceReferenceIntegral(base, tc.m, tc.r, tc.tr, tc.cos, 1.5)
			rng := rand.New(rand.NewPCG(37, 99))
			const count = 150000
			var sum, squares [3]float64
			for range count {
				s := sampleSurface(material, material.Roughness, material.Metallicness, base, &vec3.UnitZ, &vec3.UnitZ, &incoming, 1, 1.5, rng.Float64)
				if s.heading == nil {
					continue
				}
				for c, w := range s.weight {
					if !(w >= 0 && !math.IsInf(w, 0)) {
						t.Fatalf("invalid sample %+v", s)
					}
					sum[c] += w
					squares[c] += w * w
				}
			}
			for c := range sum {
				mean := sum[c] / count
				tolerance := 8*math.Sqrt(max(0, squares[c]/count-mean*mean)/count) + 0.001
				if math.Abs(mean-want[c]) > tolerance {
					t.Fatalf("channel %d mean %g, integral %g ± %g", c, mean, want[c], tolerance)
				}
			}
		})
	}
}

// White-furnace upper bounds; single-scattering GGX is allowed to lose energy.
func TestSurfaceEnergyBounds(t *testing.T) {
	for _, ior := range []float64{1, 1.1, 1.5, 2.5} {
		for _, m := range []float64{0, 0.5, 1} {
			for _, r := range []float64{0.35, 0.7, 1} {
				for _, cos := range []float64{0.05, 0.5, 1} {
					energy := surfaceReferenceIntegral([3]float64{1, 1, 1}, m, r, 0, cos, ior)[0]
					if !(energy >= 0 && energy <= 1.002) {
						t.Fatalf("ior/m/r/cos %g/%g/%g/%g: energy %g", ior, m, r, cos, energy)
					}
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

func surfaceReferenceIntegral(base [3]float64, m, r, tr, viewCos, ior float64) [3]float64 {
	const steps = 128
	v := vec3.T{-math.Sqrt(1 - viewCos*viewCos), 0, viewCos}
	fv := surfaceReferenceFresnel(viewCos, ior)
	a2 := math.Pow(r, 4)
	var result [3]float64
	for j := range steps {
		lc := (float64(j) + 0.5) / steps
		fl := surfaceReferenceFresnel(lc, ior)
		for k := range steps {
			phi := 2 * math.Pi * (float64(k) + 0.5) / steps
			l := vec3.T{math.Sqrt(1-lc*lc) * math.Cos(phi), math.Sqrt(1-lc*lc) * math.Sin(phi), lc}
			h := v.Added(&l)
			h.Normalize()
			vh := vec3.Dot(&v, &h)
			d := a2 / (math.Pi * math.Pow(h[2]*h[2]*(a2-1)+1, 2))
			lambdaV := (math.Sqrt(1+a2*(1-viewCos*viewCos)/(viewCos*viewCos)) - 1) / 2
			lambdaL := (math.Sqrt(1+a2*(1-lc*lc)/(lc*lc)) - 1) / 2
			g := 1 / (1 + lambdaV + lambdaL)
			fd := surfaceReferenceFresnel(vh, ior)
			for c := range result {
				fm := base[c] + (1-base[c])*math.Pow(1-vh, 5)
				spec := (m*fm + (1-m)*((1-tr)*fd+tr*fv)) * d * g / (4 * viewCos * lc)
				diffuse := (1 - m) * (1 - tr) * base[c] * (1 - fv) * (1 - fl) / math.Pi
				result[c] += (spec + diffuse) * lc * 2 * math.Pi / (steps * steps)
			}
		}
	}
	for c := range result {
		result[c] += (1 - m) * tr * (1 - fv)
	} // thin transmission has no eta factor
	return result
}

func TestTraceMetalBaseColorAndCoverage(t *testing.T) {
	for _, coverage := range []float64{0, 1} {
		material := scene.NewMaterial().M(1, 0).T(1, true, 1.5).C(color.NewColorRGBA(0.8, 0.4, 0.2, coverage))
		fresnelTestAlbedo(material, 1)
		material.Projections[0].Texture.Image.Fill(color.NewColor(0.5, 0.25, 1))
		want := color.NewColor(0.4, 0.1, 0.2)
		if coverage == 0 {
			want = color.White
		}
		assertTransmissionColor(t, bounceTestTrace(bounceTestScene(material)), want)
	}
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
