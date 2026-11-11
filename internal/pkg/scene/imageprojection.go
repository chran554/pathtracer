package scene

import (
	"fmt"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

const (
	pi     = math.Pi
	pi05   = pi / 2.0
	pi2    = 2.0 * pi
	piInv  = 1.0 / pi
	pi2Inv = 1.0 / pi2
)

type ProjectionType string

const (
	ProjectionTypeParallel    ProjectionType = "Parallel"
	ProjectionTypeCylindrical ProjectionType = "Cylindrical"
	ProjectionTypeSpherical   ProjectionType = "Spherical"
)

type ImageProjection struct {
	Type                            ProjectionType
	Texture                         *Texture
	Origin                          *vec3.T
	U                               *vec3.T
	V                               *vec3.T
	RepeatU                         bool
	RepeatV                         bool
	FlipU                           bool
	FlipV                           bool
	_invertedCoordinateSystemMatrix *mat3.T
}

func NewImageProjection(projectionType ProjectionType, image *floatimage.FloatImage, imageType TextureType, origin *vec3.T, u vec3.T, v vec3.T, repeatU bool, repeatV bool, flipU bool, flipV bool) ImageProjection {
	texture := &Texture{
		Type:          imageType,
		Strength:      1.0,
		Interpolation: floatimage.InterpolationNearestNeighbor,
		Image:         image,
	}

	return ImageProjection{
		Type:    projectionType,
		Texture: texture,
		Origin:  origin,
		U:       &u,
		V:       &v,
		RepeatU: repeatU,
		RepeatV: repeatV,
		FlipU:   flipU,
		FlipV:   flipV,
	}
}

func NewParallelImageProjection(texture *floatimage.FloatImage, textureType TextureType, origin *vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(ProjectionTypeParallel, texture, textureType, origin, u, v, true, true, false, false)
}

func NewCylindricalImageProjection(texture *floatimage.FloatImage, textureType TextureType, origin *vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(ProjectionTypeCylindrical, texture, textureType, origin, u, v, false, true, false, false)
}

func NewSphericalImageProjection(texture *floatimage.FloatImage, textureType TextureType, origin *vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(ProjectionTypeSpherical, texture, textureType, origin, u, v, false, true, false, false)
}

func (imageProjection *ImageProjection) GetColor(point *vec3.T) *color.Color {
	if imageProjection.Texture == nil || imageProjection.Texture.Image == nil {
		return nil
	}

	var normalizedTextureCoordinate *vec2.T

	if imageProjection.Type == ProjectionTypeParallel {
		normalizedTextureCoordinate = imageProjection.getParallelXY(point)
	} else if imageProjection.Type == ProjectionTypeCylindrical {
		normalizedTextureCoordinate = imageProjection.getCylindricalXY(point)
	} else if imageProjection.Type == ProjectionTypeSpherical {
		normalizedTextureCoordinate = imageProjection.getSphericalXY(point)
	}

	if normalizedTextureCoordinate == nil {
		return nil
	}

	textureX := normalizedTextureCoordinate[0] * float64(imageProjection.Texture.Image.Width)
	textureY := normalizedTextureCoordinate[1] * float64(imageProjection.Texture.Image.Height)

	return imageProjection.Texture.Image.GetInterpolatedPixel(textureX, textureY, imageProjection.Texture.Interpolation)
}

func (imageProjection *ImageProjection) GetNormal(point *vec3.T) *vec3.T {
	c := imageProjection.GetColor(point)

	if c == nil {
		return nil
	}

	// Texture values are already converted/decoded from texture color values to normal values
	n := &vec3.T{float64(c.R), float64(c.G), float64(c.B)}
	n.Normalize()

	return n
}

func (imageProjection *ImageProjection) getCylindricalXY(point *vec3.T) *vec2.T {
	translatedPoint := point.Subed(imageProjection.Origin)

	uvPoint := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)

	invLength := 1 / math.Sqrt(uvPoint[0]*uvPoint[0]+uvPoint[2]*uvPoint[2])
	cosineOfAngle := uvPoint[0] * invLength
	sineOfAngle := uvPoint[2] * invLength

	var radAngle float64
	if sineOfAngle >= 0.0 {
		radAngle = math.Acos(cosineOfAngle)
	} else {
		radAngle = pi2 - math.Acos(cosineOfAngle)
	}

	textureLatitudeRepetitions := 1.0

	u, fracU := math.Modf((radAngle * textureLatitudeRepetitions) * pi2Inv)
	v := uvPoint[1]
	_, fracV := math.Modf(v)

	// imageProjection.RepeatU:
	// Repeat (as true/false) along U or equator/latitude is not applicable for cylindrical projection.
	// TODO: Amount repeats along the equator/latitude can be of use though, see "textureLatitudeRepetitions".

	// If not repeat along the cylinder axis and point is above or below texture projection
	// Then no texture coordinate is returned.
	if !imageProjection.RepeatV && ((v >= 1.0) || (v < 0.0)) {
		return nil
	}

	if fracU < 0.0 {
		fracU = fracU + 1.0
	}
	if fracV < 0.0 {
		fracV = fracV + 1.0
	}

	if imageProjection.FlipU && (int(math.Abs(math.Floor(u)))%2 == 1) {
		fracU = 1.0 - fracU
	}
	if imageProjection.FlipV && (int(math.Abs(math.Floor(v)))%2 == 1) {
		fracV = 1.0 - fracV
	}

	// The pixel at UV-origin should be the pixel at bottom left in image
	return &vec2.T{math.Abs(fracU), 1 - math.Abs(fracV)}
}

func (imageProjection *ImageProjection) getSphericalXY(point *vec3.T) *vec2.T {
	translatedPoint := *point
	translatedPoint.Sub(imageProjection.Origin)

	p := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)
	pxzLength := math.Sqrt(p[0]*p[0] + p[2]*p[2])

	theta := math.Acos(p[0] / pxzLength)
	phi := math.Acos(p[1] / p.Length())

	if math.IsNaN(theta) {
		theta = 0.0
	}

	if math.IsNaN(phi) {
		phi = 0.0
	}

	if p[2] < 0 {
		theta = pi2 - theta
	}

	if phi >= math.Pi {
		phi = 0.0
	} else if phi < 0.0 {
		phi = 0.0
	}

	for theta >= pi2 {
		theta -= pi2
	}

	return &vec2.T{theta * pi2Inv, phi * piInv}
}

func (imageProjection *ImageProjection) getParallelXY(point *vec3.T) *vec2.T {
	translatedPoint := *point
	translatedPoint.Sub(imageProjection.Origin)

	pointInUV := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)
	u := pointInUV[0]
	v := pointInUV[1]
	_, fracU := math.Modf(u)
	_, fracV := math.Modf(v)

	if !imageProjection.RepeatU && ((u >= 1.0) || (u < 0.0)) {
		return nil
	}

	if !imageProjection.RepeatV && ((v >= 1.0) || (v < 0.0)) {
		return nil
	}

	if fracU < 0.0 {
		fracU += 1.0
	}
	if fracV < 0.0 {
		fracV += 1.0
	}

	if imageProjection.FlipU && (int(math.Abs(math.Floor(u)))%2 == 1) {
		fracU = 1.0 - fracU
	}
	if imageProjection.FlipV && (int(math.Abs(math.Floor(v)))%2 == 1) {
		fracV = 1.0 - fracV
	}

	return &vec2.T{math.Abs(fracU), 1 - math.Abs(fracV)}
}

func (imageProjection *ImageProjection) ClearProjection() {
	if imageProjection.Texture != nil {
		imageProjection.Texture.Image = nil
	}
	imageProjection._invertedCoordinateSystemMatrix = nil
}

func (imageProjection *ImageProjection) Initialize() {
	switch imageProjection.Type {
	case ProjectionTypeSpherical:
		imageProjection.initializeSphericalProjection()

	case ProjectionTypeCylindrical:
		imageProjection.initializeCylindricalProjection()

	case ProjectionTypeParallel:
		imageProjection.initializeParallelProjection()

	default:
		fmt.Printf("can not initialize unknown projection type \"%s\"\n", imageProjection.Type)
	}
}

func (imageProjection *ImageProjection) initializeTextureMapping() {
	// Nothing by intention
}

func (imageProjection *ImageProjection) initializeParallelProjection() {
	if imageProjection._invertedCoordinateSystemMatrix == nil || *imageProjection._invertedCoordinateSystemMatrix == mat3.Zero {
		W := vec3.Cross(imageProjection.U, imageProjection.V)
		imageProjection._invertedCoordinateSystemMatrix = &mat3.T{*imageProjection.U, *imageProjection.V, W}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			imageName := "nil"
			if imageProjection.Texture != nil && imageProjection.Texture.Image != nil {
				imageName = imageProjection.Texture.Image.Name()
			}
			fmt.Println("could not initialize parallel projection as invert matrix for uv system could not be created.", imageName, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}
}

func (imageProjection *ImageProjection) initializeCylindricalProjection() {
	if imageProjection._invertedCoordinateSystemMatrix == nil || *imageProjection._invertedCoordinateSystemMatrix == mat3.Zero {
		Un := imageProjection.U.Normalized()
		Vn := imageProjection.V.Normalized()
		W := vec3.Cross(&Un, &Vn)
		U := vec3.Cross(&Vn, &W)
		imageProjection._invertedCoordinateSystemMatrix = &mat3.T{U, *imageProjection.V, W}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			imageName := "nil"
			if imageProjection.Texture != nil && imageProjection.Texture.Image != nil {
				imageName = imageProjection.Texture.Image.Name()
			}
			fmt.Println("could not initialize cylindrical projection as inverted matrix for uv system could not be created.", imageName, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}
}

func (imageProjection *ImageProjection) initializeSphericalProjection() {
	if imageProjection._invertedCoordinateSystemMatrix == nil || *imageProjection._invertedCoordinateSystemMatrix == mat3.Zero {
		Un := imageProjection.U.Normalized()
		Vn := imageProjection.V.Normalized()
		Wn := vec3.Cross(&Un, &Vn)
		//U := vec3.Cross(&Vn, &W)
		imageProjection._invertedCoordinateSystemMatrix = &mat3.T{Un, Vn, Wn}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			imageName := "nil"
			if imageProjection.Texture != nil && imageProjection.Texture.Image != nil {
				imageName = imageProjection.Texture.Image.Name()
			}
			fmt.Println("could not initialize spherical projection as inverted matrix for uv system could not be created.", imageName, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}
}
