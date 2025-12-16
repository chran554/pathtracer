package main

import (
	"math"
	"math/rand"

	"github.com/ungerik/go3d/float64/vec3"
)

type Line struct {
	Start       vec3.T
	End         vec3.T
	StartRadius float64
	EndRadius   float64
	Level       int
}

type Params struct {
	Levels           int // recursion depth
	TrunkLength      float64
	TrunkRadius      float64
	LengthFalloff    float64 // 0..1
	RadiusFalloff    float64 // 0..1
	ChildrenMin      int     // e.g. 2
	ChildrenMax      int     // e.g. 4
	ConeAngleRadians float64 // branching spread, e.g. 0.7 (~40 deg)
	LengthJitter     float64 // e.g. 0.25 (±25%)
	AngleJitter      float64 // e.g. 0.25 (adds noise to cone angle)
	UpBias           float64 // 0..1, adds +Y bias to child directions
}

// GenerateTreeLines returns a tree model as tapered line segments (radius = StartRadius/EndRadius).
// seed makes it deterministic; change seed for different trees.
// This generator is Y-up (positive Y is “up”).
func GenerateTreeLines(seed int64, root vec3.T, p Params) []Line {
	rng := rand.New(rand.NewSource(seed))

	var out []Line
	dirUp := vec3.T{0, 1, 0}
	grow(&out, rng, root, &dirUp, p.TrunkLength, p.TrunkRadius, p.Levels, p)
	return out
}

func grow(out *[]Line, rng *rand.Rand, start vec3.T, dir *vec3.T, length, r float64, level int, p Params) {
	if level <= 0 || length <= 0.05 || r <= 0.001 {
		return
	}

	// Segment end
	dirN := dir.Normalized()
	end := start.Added(dirN.Scale(length))

	// Taper within the segment
	endR := r * 0.85
	if endR <= 0 {
		return
	}

	*out = append(*out, Line{
		Start:       start,
		End:         end,
		StartRadius: r,
		EndRadius:   endR,
		Level:       level,
	})

	// Decide number of children
	branchChildCount := p.ChildrenMin
	if p.ChildrenMax > p.ChildrenMin {
		branchChildCount += rng.Intn(p.ChildrenMax - p.ChildrenMin + 1)
	}

	childLenBase := length * p.LengthFalloff
	childR := endR * p.RadiusFalloff
	if childLenBase <= 0 || childR <= 0 {
		return
	}

	for branchChildIndex := 0; branchChildIndex < branchChildCount; branchChildIndex++ {
		// Random direction in a cone around parent direction
		cone := p.ConeAngleRadians * (1.0 + (rng.Float64()*2-1.0)*p.AngleJitter)
		childDir := randomDirInCone(rng, dirN, cone)

		// Y-up bias so it looks more “tree-ish”
		if p.UpBias != 0 {
			newDir := childDir.Added(&vec3.T{0, p.UpBias, 0})
			childDir = &newDir
			childDir.Normalize()
		}

		// Random length jitter
		j := 1.0 + (rng.Float64()*2-1.0)*p.LengthJitter
		childLen := childLenBase * j
		if childLen <= 0 {
			continue
		}

		grow(out, rng, end, childDir, childLen, childR, level-1, p)
	}
}

// randomDirInCone samples a direction within an angle `cone` around axis `axis`.
// Uses an orthonormal basis + spherical sampling.
func randomDirInCone(rng *rand.Rand, axis vec3.T, cone float64) *vec3.T {
	axisN := axis.Normalized()

	// Build basis (u,v,axis)
	var a vec3.T
	if math.Abs(axisN[1]) < 0.99 {
		a = vec3.T{0, 1, 0}
	} else {
		a = vec3.T{1, 0, 0}
	}

	u := vec3.Cross(&axisN, &a)
	u.Normalize()
	v := vec3.Cross(&axisN, &u)
	v.Normalize()

	// Uniform solid angle: cos(theta) uniform in [cos(cone), 1]
	cosMin := math.Cos(cone)
	cosT := cosMin + rng.Float64()*(1.0-cosMin)
	sinT := math.Sqrt(math.Max(0, 1.0-cosT*cosT))
	phi := rng.Float64() * 2 * math.Pi

	axisScaled := axisN.Scaled(cosT)
	uScaled := u.Scaled(math.Cos(phi) * sinT)
	vScaled := v.Scaled(math.Sin(phi) * sinT)

	d := axisScaled
	d.Add(&uScaled)
	d.Add(&vScaled)
	d.Normalize()

	return &d
}
