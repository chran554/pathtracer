package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type Disc struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Radius   float64
	Material *Material `json:"Material,omitempty"`
}

func NewDisc(origin *vec3.T, normal *vec3.T, radius float64, material *Material) *Disc {
	return &Disc{
		Origin:   origin,
		Normal:   normal,
		Radius:   radius,
		Material: material,
	}
}

func (d *Disc) Initialize() {
	d.Normal.Normalize()

	projection := d.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (d *Disc) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	d.translate(translation, translatedPoints, translatedImageProjections)
}

func (d *Disc) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	// Translate image projection (i.e. image projection origin)
	if d.Material != nil && d.Material.Projection != nil {
		origin := d.Material.Projection.Origin
		projectionOriginAlreadyTranslated := translatedPoints[origin]
		projectionAlreadyTranslated := translatedImageProjections[d.Material.Projection]

		if !projectionOriginAlreadyTranslated && !projectionAlreadyTranslated {
			origin.Add(translation)
			translatedPoints[origin] = true
			translatedImageProjections[d.Material.Projection] = true
		}
	}

	originAlreadyTranslated := translatedPoints[d.Origin]

	if !originAlreadyTranslated {
		d.Origin.Add(translation)
		translatedPoints[d.Origin] = true
	}
}

func (d *Disc) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)

	d.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
}

func (d *Disc) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	if d.Material != nil && d.Material.Projection != nil {
		panic(fmt.Sprintf("No disc translation implementation for projection type %s", d.Material.Projection.ProjectionType))
	}

	panic("Scale of disc is not yet implemented")
}

func (d *Disc) RotateX(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	d.rotate(rotationOrigin, rotationMatrix)
}

func (d *Disc) RotateY(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	d.rotate(rotationOrigin, rotationMatrix)
}

func (d *Disc) RotateZ(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	d.rotate(rotationOrigin, rotationMatrix)
}

func (d *Disc) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T) {
	normal := *d.Normal
	normal[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedNormal := rotationMatrix.MulVec3(&normal)
	rotatedNormal[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	d.Normal[0] = rotatedNormal[0]
	d.Normal[1] = rotatedNormal[1]
	d.Normal[2] = rotatedNormal[2]

	origin := *d.Origin
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	d.Origin[0] = rotatedOrigin[0]
	d.Origin[1] = rotatedOrigin[1]
	d.Origin[2] = rotatedOrigin[2]

	if d.Material != nil && d.Material.Projection != nil {
		panic(fmt.Sprintf("No disc rotation implementation for projection type %s", d.Material.Projection.ProjectionType))
	}
}

func (d *Disc) Bounds() *Bounds {
	return &Bounds{
		Xmin: d.Origin[0] - d.Radius,
		Xmax: d.Origin[0] + d.Radius,
		Ymin: d.Origin[1] - d.Radius,
		Ymax: d.Origin[1] + d.Radius,
		Zmin: d.Origin[2] - d.Radius,
		Zmax: d.Origin[2] + d.Radius,
	}
}

func (d *Disc) N(name string) *Disc {
	d.Name = name
	return d
}
