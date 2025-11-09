package util

import (
	"math"
	"math/rand"

	"github.com/ungerik/go3d/float64/vec3"
)

// NOTE: AI generated

// -------- Uniform ON the unit sphere (surface) --------

// UniformOnSphereGaussian uses gaussian normalization (x,y,z ~ N(0,1))
func UniformOnSphereGaussian() *vec3.T {
	for {
		x := rand.NormFloat64()
		y := rand.NormFloat64()
		z := rand.NormFloat64()
		n2 := x*x + y*y + z*z
		if n2 > 0 {
			inv := 1.0 / math.Sqrt(n2)
			return &vec3.T{x * inv, y * inv, z * inv}
		}
	}
}

// UniformOnSphereAngleZ uses (θ, z) trick — θ ~ U[0,2π), z ~ U[-1,1]
func UniformOnSphereAngleZ() *vec3.T {
	theta := 2 * math.Pi * rand.Float64()
	z := 2*rand.Float64() - 1 // [-1,1]
	xy := math.Sqrt(1 - z*z)
	return &vec3.T{xy * math.Cos(theta), xy * math.Sin(theta), z}
}

// -------- Uniform ON hemisphere aligned with unit normal n --------

func UniformOnHemisphereGaussian(n *vec3.T) *vec3.T {
	v := UniformOnSphereGaussian()
	if vec3.Dot(v, n) < 0 {
		v.Invert()
	}
	return v
}

func UniformOnHemisphereAngleZ(n *vec3.T) *vec3.T {
	v := UniformOnSphereAngleZ()
	if vec3.Dot(v, n) < 0 {
		v.Invert()
	}
	return v
}

// -------- Cosine-weighted hemisphere (Lambertian) --------
// Useful for your diffuse `diffuseHeading`: no extra cos factor needed.

func CosineWeightedHemisphere(n *vec3.T) *vec3.T {
	// Concentric disk or (r,phi) method; here we use (r,phi):
	u1 := rand.Float64()
	u2 := rand.Float64()
	rad := math.Sqrt(u1)
	phi := 2 * math.Pi * u2
	x := rad * math.Cos(phi)
	y := rad * math.Sin(phi)
	z := math.Sqrt(math.Max(0, 1-u1)) // lifts the disk into a hemisphere (cosine-weighted)

	// Build an orthonormal basis (u,v,w) with w=n
	w := n.Normalized()
	// Choose the most orthogonal axis to avoid degeneracy
	var a vec3.T
	if math.Abs(w[0]) > 0.9 {
		a = vec3.T{0, 1, 0}
	} else {
		a = vec3.T{1, 0, 0}
	}
	u := vec3.Cross(&a, &w)
	v := vec3.Cross(&w, &u)
	u.Normalize()
	v.Normalize()

	// Transform local (x,y,z) to world
	out := vec3.T{
		u[0]*x + v[0]*y + w[0]*z,
		u[1]*x + v[1]*y + w[1]*z,
		u[2]*x + v[2]*y + w[2]*z,
	}
	return &out
}

// -------- Uniform IN a ball of radius R (volume) --------

func UniformInBallGaussian(R float64) *vec3.T {
	dir := UniformOnSphereGaussian()
	u := rand.Float64()       // [0,1)
	scale := R * math.Cbrt(u) // cube-root for volume uniformity
	return &vec3.T{dir[0] * scale, dir[1] * scale, dir[2] * scale}
}

func UniformInBallAngleZ(R float64) *vec3.T {
	dir := UniformOnSphereAngleZ()
	u := rand.Float64()
	scale := R * math.Cbrt(u)
	return &vec3.T{dir[0] * scale, dir[1] * scale, dir[2] * scale}
}
