package scene

import (
	"fmt"
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
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	sphere.translate(translation, translatedPoints, translatedImageProjections)
}

func (sphere *Sphere) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	if sphere.Material != nil && sphere.Material.Projection != nil {
		panic(fmt.Sprintf("No sphere translation implementation for projection type %s", sphere.Material.Projection.ProjectionType))
	}

	originAlreadyTranslated := translatedPoints[sphere.Origin]

	if !originAlreadyTranslated {
		sphere.Origin.Add(translation)
		translatedPoints[sphere.Origin] = true
	}
}

func (sphere *Sphere) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)

	sphere.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
}

func (sphere *Sphere) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	panic("Scale of sphere is not yet implemented")
}

func (sphere *Sphere) RotateX(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	sphere.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

func (sphere *Sphere) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	sphere.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

func (sphere *Sphere) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	sphere.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

// rotate "rotates" a sphere. It does not rotate the sphere per se but rather rotates any projection associated with the sphere.
func (sphere *Sphere) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedProjections map[*ImageProjection]bool) {
	originAlreadyRotated := rotatedPoints[sphere.Origin]

	if !originAlreadyRotated {
		sphere.Origin.Sub(rotationOrigin)
		sphere.Origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
		rotatedProjectionOrigin := rotationMatrix.MulVec3(sphere.Origin)
		rotatedProjectionOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
		rotatedProjectionOrigin.Add(rotationOrigin)

		sphere.Origin[0] = rotatedProjectionOrigin[0]
		sphere.Origin[1] = rotatedProjectionOrigin[1]
		sphere.Origin[2] = rotatedProjectionOrigin[2]

		rotatedPoints[sphere.Origin] = true
	}

	if sphere.Material != nil && sphere.Material.Projection != nil {
		imageProjection := sphere.Material.Projection
		projectionType := imageProjection.ProjectionType

		imageProjectionAlreadyRotated := rotatedProjections[imageProjection]

		if !imageProjectionAlreadyRotated {

			// If projection type is parallel or spherical then rotate projection origin, u, and v vectors.
			// No need to translate u, and v vectors by rotation origin as they are pure vectors and not points without a specific location.
			if projectionType == ProjectionTypeParallel || projectionType == ProjectionTypeSpherical {
				projectionOrigin := *imageProjection.Origin

				projectionOrigin.Sub(rotationOrigin)
				projectionOrigin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
				rotatedProjectionOrigin := rotationMatrix.MulVec3(&projectionOrigin)
				rotatedProjectionOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
				rotatedProjectionOrigin.Add(rotationOrigin)

				imageProjection.Origin[0] = rotatedProjectionOrigin[0]
				imageProjection.Origin[1] = rotatedProjectionOrigin[1]
				imageProjection.Origin[2] = rotatedProjectionOrigin[2]

				projectionU := *imageProjection.U
				projectionU[2] *= -1 // Convert to right hand coordinate system before rotation matrix
				rotatedProjectionU := rotationMatrix.MulVec3(&projectionU)
				rotatedProjectionU[2] *= -1 // Convert back to left hand coordinate system after rotation matrix

				imageProjection.U[0] = rotatedProjectionU[0]
				imageProjection.U[1] = rotatedProjectionU[1]
				imageProjection.U[2] = rotatedProjectionU[2]

				projectionV := *imageProjection.V
				projectionV[2] *= -1 // Convert to right hand coordinate system before rotation matrix
				rotatedProjectionV := rotationMatrix.MulVec3(&projectionV)
				rotatedProjectionV[2] *= -1 // Convert back to left hand coordinate system after rotation matrix

				imageProjection.V[0] = rotatedProjectionV[0]
				imageProjection.V[1] = rotatedProjectionV[1]
				imageProjection.V[2] = rotatedProjectionV[2]

				if projectionType == ProjectionTypeParallel {
					imageProjection._invertedCoordinateSystemMatrix = nil
					imageProjection.initializeParallellProjection()

				} else if projectionType == ProjectionTypeSpherical {
					imageProjection._invertedCoordinateSystemMatrix = nil
					imageProjection.initializeSphericalProjection()
				}
			} else {
				panic(fmt.Sprintf("No sphere rotation implementation for projection type %s", projectionType))
			}
		}
	}
}

func (sphere *Sphere) Normal(point *vec3.T) *vec3.T {
	normal := point.Subed(sphere.Origin)
	normal.Normalize()

	return &normal
}
