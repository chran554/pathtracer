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

func (disc *Disc) Initialize() {
	disc.Normal.Normalize()

	projection := disc.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (disc *Disc) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	disc.translate(translation, translatedPoints, translatedImageProjections)
}

func (disc *Disc) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	if disc.Material != nil && disc.Material.Projection != nil {
		panic(fmt.Sprintf("No disc translation implementation for projection type %s", disc.Material.Projection.ProjectionType))
	}

	originAlreadyTranslated := translatedPoints[disc.Origin]

	if !originAlreadyTranslated {
		disc.Origin.Add(translation)
		translatedPoints[disc.Origin] = true
	}
}

func (disc *Disc) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)

	disc.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
}

func (disc *Disc) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	if disc.Material != nil && disc.Material.Projection != nil {
		panic(fmt.Sprintf("No disc translation implementation for projection type %s", disc.Material.Projection.ProjectionType))
	}

	panic("Scale of disc is not yet implemented")
}

func (disc *Disc) RotateX(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}

func (disc *Disc) RotateY(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}

func (disc *Disc) RotateZ(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}

func (disc *Disc) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T) {
	normal := *disc.Normal
	normal[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedNormal := rotationMatrix.MulVec3(&normal)
	rotatedNormal[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	disc.Normal[0] = rotatedNormal[0]
	disc.Normal[1] = rotatedNormal[1]
	disc.Normal[2] = rotatedNormal[2]

	origin := *disc.Origin
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	disc.Origin[0] = rotatedOrigin[0]
	disc.Origin[1] = rotatedOrigin[1]
	disc.Origin[2] = rotatedOrigin[2]

	if disc.Material != nil && disc.Material.Projection != nil {
		panic(fmt.Sprintf("No disc rotation implementation for projection type %s", disc.Material.Projection.ProjectionType))
	}
}

func (disc *Disc) Bounds() *Bounds {
	return &Bounds{
		Xmin: disc.Origin[0] - disc.Radius,
		Xmax: disc.Origin[0] + disc.Radius,
		Ymin: disc.Origin[1] - disc.Radius,
		Ymax: disc.Origin[1] + disc.Radius,
		Zmin: disc.Origin[2] - disc.Radius,
		Zmax: disc.Origin[2] + disc.Radius,
	}
}
