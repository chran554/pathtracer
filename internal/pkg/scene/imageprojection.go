package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"
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
	Parallel    ProjectionType = "Parallel"
	Cylindrical ProjectionType = "Cylindrical"
	Spherical   ProjectionType = "Spherical"
)

type ImageProjection struct {
	ProjectionType                  ProjectionType `json:"ProjectionType"`
	ImageFilename                   string         `json:"ImageFilename"`
	Origin                          *vec3.T        `json:"Origin"`
	U                               *vec3.T        `json:"U"`
	V                               *vec3.T        `json:"V"`
	RepeatU                         bool           `json:"RepeatU,omitempty"`
	RepeatV                         bool           `json:"RepeatV,omitempty"`
	FlipU                           bool           `json:"FlipU,omitempty"`
	FlipV                           bool           `json:"FlipV,omitempty"`
	Gamma                           float64        `json:"Gamma,omitempty"`
	_image                          *image.FloatImage
	_invertedCoordinateSystemMatrix *mat3.T
}

func NewParallelImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(Parallel, textureFilename, origin, u, v, true, true, false, false)
}

func NewCylindricalImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(Cylindrical, textureFilename, origin, u, v, false, true, false, false)
}

func NewSphericalImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(Spherical, textureFilename, origin, u, v, false, true, false, false)
}

func NewImageProjection(projectionType ProjectionType, textureFilename string, origin vec3.T, u vec3.T, v vec3.T, repeatU bool, repeatV bool, flipU bool, flipV bool) ImageProjection {
	return ImageProjection{
		ProjectionType: projectionType,
		ImageFilename:  textureFilename,
		Origin:         &origin,
		U:              &u,
		V:              &v,
		RepeatU:        repeatU,
		RepeatV:        repeatV,
		FlipU:          flipU,
		FlipV:          flipV,
	}
}

func (imageProjection *ImageProjection) GetColor(point *vec3.T) *color.Color {
	if imageProjection.ProjectionType == Parallel {
		return imageProjection.getParallelColor(point)
	}
	if imageProjection.ProjectionType == Cylindrical {
		return imageProjection.getCylindricalColor(point)
	}
	if imageProjection.ProjectionType == Spherical {
		return imageProjection.getSphericalColor(point)
	}

	return &color.White
}

func (imageProjection *ImageProjection) getSphericalColor2(point *vec3.T) *color.Color {
	translatedPoint := *point
	translatedPoint.Sub(imageProjection.Origin)

	p := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)

	theta := math.Acos(p[1] / p.Length())

	var phi float64
	if p[0] > 0 {
		phi = math.Atan(p[2] / p[0])

	} else if (p[0] < 0) && (p[2] >= 0) {
		phi = math.Atan(p[2]/p[0]) + pi

	} else if (p[0] < 0) && (p[2] < 0) {
		phi = math.Atan(p[2]/p[0]) - pi

	} else if (p[0] == 0) && (p[2] >= 0) {
		phi = pi05

	} else if (p[0] == 0) && (p[2] < 0) {
		phi = -pi05
	}

	textureX := int((phi * pi2Inv) * float64(imageProjection._image.Width))
	textureY := int((theta * piInv) * float64(imageProjection._image.Height))

	return imageProjection._image.GetPixel(textureX, textureY)
}

func (imageProjection *ImageProjection) getSphericalColor(point *vec3.T) *color.Color {
	normalizedTextureCoordinate := imageProjection.getSphericalXY(point)

	textureX := int(normalizedTextureCoordinate[0] * float64(imageProjection._image.Width))
	textureY := int(normalizedTextureCoordinate[1] * float64(imageProjection._image.Height))

	return imageProjection._image.GetPixel(textureX, textureY)
}

func (imageProjection *ImageProjection) getSphericalXY(point *vec3.T) vec2.T {
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

	return vec2.T{theta * pi2Inv, phi * piInv}
}

func (imageProjection *ImageProjection) getCylindricalColor(point *vec3.T) *color.Color {
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
	// (Amount repeats along the equator/latitude can be of use though, see "textureLatitudeRepetitions".)

	if !imageProjection.RepeatV && ((v >= 1.0) || (v < 0.0)) {
		return &color.White
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

	textureX := int(math.Abs(fracU) * float64(imageProjection._image.Width))
	textureY := int(math.Abs(fracV) * float64(imageProjection._image.Height))
	textureY = (imageProjection._image.Height - 1) - textureY // The pixel at UV-origin should be the pixel at bottom left in image

	return imageProjection._image.GetPixel(textureX, textureY)
}

func (imageProjection *ImageProjection) getParallelColor(point *vec3.T) *color.Color {
	translatedPoint := *point
	translatedPoint.Sub(imageProjection.Origin)

	pointInUV := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)
	u := pointInUV[0]
	v := pointInUV[1]
	_, fracU := math.Modf(u)
	_, fracV := math.Modf(v)

	if !imageProjection.RepeatU && ((u >= 1.0) || (u < 0.0)) {
		return &color.White
	}

	if !imageProjection.RepeatV && ((v >= 1.0) || (v < 0.0)) {
		return &color.White
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

	textureX := int(math.Abs(fracU) * float64(imageProjection._image.Width))
	textureY := int(math.Abs(fracV) * float64(imageProjection._image.Height))
	textureY = (imageProjection._image.Height - 1) - textureY // The pixel at UV-origin should be the pixel at bottom left in image

	return imageProjection._image.GetPixel(textureX, textureY)
}

func (imageProjection *ImageProjection) ClearProjection() {
	imageProjection._image = nil
	imageProjection._invertedCoordinateSystemMatrix = nil
}

func (imageProjection *ImageProjection) Initialize() {
	if imageProjection._image == nil || !imageProjection._image.ContainImageData() {
		if imageProjection.Gamma == 0.0 {
			imageProjection.Gamma = 1.0 // No gamma correction
		}
		imageProjection._image = image.GetCachedImage(imageProjection.ImageFilename, imageProjection.Gamma)
	}

	switch imageProjection.ProjectionType {
	case Spherical:
		imageProjection.initializeSphericalProjection()

	case Cylindrical:
		imageProjection.initializeCylindricalProjection()

	case Parallel:
		imageProjection.initializeParallellProjection()

	default:
		fmt.Printf("can not initialize unknown projection type \"%s\"\n", imageProjection.ProjectionType)
	}
}

func (imageProjection *ImageProjection) initializeParallellProjection() {
	if imageProjection._invertedCoordinateSystemMatrix == nil || *imageProjection._invertedCoordinateSystemMatrix == mat3.Zero {
		W := vec3.Cross(imageProjection.U, imageProjection.V)
		imageProjection._invertedCoordinateSystemMatrix = &mat3.T{*imageProjection.U, *imageProjection.V, W}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			fmt.Println("could not initialize parallel projection as invert matrix for uv system could not be created.", imageProjection.ImageFilename, imageProjection.U, imageProjection.V, imageProjection.Origin)
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
			fmt.Println("could not initialize cylindrical projection as inverted matrix for uv system could not be created.", imageProjection.ImageFilename, imageProjection.U, imageProjection.V, imageProjection.Origin)
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
			fmt.Println("could not initialize spherical projection as inverted matrix for uv system could not be created.", imageProjection.ImageFilename, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}
}
