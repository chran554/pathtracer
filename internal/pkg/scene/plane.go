package scene

import "github.com/ungerik/go3d/float64/vec3"

type Plane struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Material *Material
}

// NewPlane creates a plane from three vertices.
func NewPlane(v1, v2, v3 *vec3.T, name string, material *Material) *Plane {
	a := v2.Subed(v1)
	b := v3.Subed(v1)
	n := vec3.Cross(&a, &b)
	n.Normalize()

	return &Plane{
		Name:     name,
		Origin:   v1,
		Normal:   &n,
		Material: material,
	}
}
