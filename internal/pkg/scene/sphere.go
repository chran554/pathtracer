package scene

import (
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type Sphere struct {
	Name     string
	Origin   *vec3.T
	Radius   float64
	Material *Material
}

func NewSphere(origin *vec3.T, radius float64, material *Material) *Sphere {
	return &Sphere{
		Origin:   origin,
		Radius:   radius,
		Material: material,
	}
}

func (s *Sphere) Initialize() {
	if s.Material != nil {
		for _, projection := range s.Material.Projections {
			projection.Initialize()
		}
	}
}

func (s *Sphere) Bounds() *Bounds {
	return &Bounds{
		Xmin: s.Origin[0] - s.Radius,
		Xmax: s.Origin[0] + s.Radius,
		Ymin: s.Origin[1] - s.Radius,
		Ymax: s.Origin[1] + s.Radius,
		Zmin: s.Origin[2] - s.Radius,
		Zmax: s.Origin[2] + s.Radius,
	}
}

func (s *Sphere) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	s.translate(translation, translatedPoints, translatedImageProjections)
}

func (s *Sphere) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	// Translate image projection (i.e. image projection origin)
	if s.Material != nil {
		for _, projection := range s.Material.Projections {
			origin := projection.Origin
			projectionOriginAlreadyTranslated := translatedPoints[origin]
			projectionAlreadyTranslated := translatedImageProjections[projection]

			if !projectionOriginAlreadyTranslated && !projectionAlreadyTranslated {
				origin.Add(translation)
				translatedPoints[origin] = true
				translatedImageProjections[projection] = true
			}
		}
	}

	// Translate sphere origin
	originAlreadyTranslated := translatedPoints[s.Origin]

	if !originAlreadyTranslated {
		s.Origin.Add(translation)
		translatedPoints[s.Origin] = true
	}
}

func (s *Sphere) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)

	s.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
}

func (s *Sphere) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	panic("Scale of sphere is not yet implemented")
}

func (s *Sphere) RotateX(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	s.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

func (s *Sphere) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	s.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

func (s *Sphere) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedImageProjections := make(map[*ImageProjection]bool)
	rotatedPoints := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	s.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
}

// rotate "rotates" a sphere. It does not rotate the sphere per se but rather rotates any projection associated with the sphere.
func (s *Sphere) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedImageProjections map[*ImageProjection]bool) {
	originAlreadyRotated := rotatedPoints[s.Origin]

	// Rotate sphere origin
	if !originAlreadyRotated {
		newVertex := s.Origin.Subed(rotationOrigin)
		newVertex[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
		rotatedOrigin := rotationMatrix.MulVec3(&newVertex)
		rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
		rotatedOrigin.Add(rotationOrigin)

		s.Origin[0] = rotatedOrigin[0]
		s.Origin[1] = rotatedOrigin[1]
		s.Origin[2] = rotatedOrigin[2]

		rotatedPoints[s.Origin] = true
	}

	// Rotate sphere image projection
	if s.Material != nil {
		for _, projection := range s.Material.Projections {
			projectionType := projection.Type

			imageProjectionAlreadyRotated := rotatedImageProjections[projection]
			imageProjectionOriginAlreadyRotated := rotatedPoints[projection.Origin]
			imageProjectionUAlreadyRotated := rotatedPoints[projection.U]
			imageProjectionVAlreadyRotated := rotatedPoints[projection.V]

			if !imageProjectionAlreadyRotated {
				// If a projection type is parallel or spherical, then rotate projection origin, u, and v vectors.
				// No need to translate u, and v vectors by rotation origin as they are pure vectors and not points without a specific location.
				if !imageProjectionOriginAlreadyRotated {
					projectionOrigin := *projection.Origin
					projectionOrigin.Sub(rotationOrigin)
					projectionOrigin[2] *= -1 // Change to right-hand coordinate system from left-hand coordinate system
					rotatedProjectionOrigin := rotationMatrix.MulVec3(&projectionOrigin)
					rotatedProjectionOrigin[2] *= -1 // Change back from right-hand coordinate system to left-hand coordinate system
					rotatedProjectionOrigin.Add(rotationOrigin)

					projection.Origin[0] = rotatedProjectionOrigin[0]
					projection.Origin[1] = rotatedProjectionOrigin[1]
					projection.Origin[2] = rotatedProjectionOrigin[2]

					rotatedPoints[projection.Origin] = true
				}

				if !imageProjectionUAlreadyRotated {
					projectionU := *projection.U
					projectionU[2] *= -1 // Convert to right-hand coordinate system before rotation matrix
					rotatedProjectionU := rotationMatrix.MulVec3(&projectionU)
					rotatedProjectionU[2] *= -1 // Convert back to left-hand coordinate system after rotation matrix

					projection.U[0] = rotatedProjectionU[0]
					projection.U[1] = rotatedProjectionU[1]
					projection.U[2] = rotatedProjectionU[2]

					rotatedPoints[projection.U] = true
				}

				if !imageProjectionVAlreadyRotated {
					projectionV := *projection.V
					projectionV[2] *= -1 // Convert to right-hand coordinate system before rotation matrix
					rotatedProjectionV := rotationMatrix.MulVec3(&projectionV)
					rotatedProjectionV[2] *= -1 // Convert back to left-hand coordinate system after rotation matrix

					projection.V[0] = rotatedProjectionV[0]
					projection.V[1] = rotatedProjectionV[1]
					projection.V[2] = rotatedProjectionV[2]

					rotatedPoints[projection.V] = true
				}

				if projectionType == ProjectionTypeParallel {
					projection._invertedCoordinateSystemMatrix = nil
					projection.initializeParallelProjection()

				} else if projectionType == ProjectionTypeSpherical {
					projection._invertedCoordinateSystemMatrix = nil
					projection.initializeSphericalProjection()

				} else if projectionType == ProjectionTypeCylindrical {
					projection._invertedCoordinateSystemMatrix = nil
					projection.initializeCylindricalProjection()
				}

				rotatedImageProjections[projection] = true
			}
		}
	}
}

func (s *Sphere) Normal(point *vec3.T) *vec3.T {
	normal := point.Subed(s.Origin)
	normal.Normalize()

	return &normal
}

func (s *Sphere) N(name string) *Sphere {
	s.Name = name
	return s
}
