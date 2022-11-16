package scene

import (
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type Sphere struct {
	Name     string
	Origin   *vec3.T
	Radius   float64
	Material *Material `json:"material,omitempty"`
}

func (sphere *Sphere) Initialize() {
	projection := sphere.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (sphere *Sphere) Bounds() *Bounds {
	return &Bounds{
		Xmin: sphere.Origin[0] - sphere.Radius,
		Xmax: sphere.Origin[0] + sphere.Radius,
		Ymin: sphere.Origin[1] - sphere.Radius,
		Ymax: sphere.Origin[1] + sphere.Radius,
		Zmin: sphere.Origin[2] - sphere.Radius,
		Zmax: sphere.Origin[2] + sphere.Radius,
	}
}

func (sphere *Sphere) Translate(translation *vec3.T) {
	sphere.Origin.Add(translation)
}

func (sphere *Sphere) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	sphere.rotate(rotationOrigin, rotationMatrix)
}

func (sphere *Sphere) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T) {
	origin := *sphere.Origin
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	sphere.Origin[0] = rotatedOrigin[0]
	sphere.Origin[1] = rotatedOrigin[1]
	sphere.Origin[2] = rotatedOrigin[2]
}
