package scene

import (
	"fmt"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

func NewParallelImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(Parallel, textureFilename, origin, u, v, true, true, false, false)
}

func NewCylindricalImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ImageProjection {
	return NewImageProjection(Cylindrical, textureFilename, origin, u, v, false, true, false, false)
}

func NewImageProjection(projectionType ProjectionType, textureFilename string, origin vec3.T, u vec3.T, v vec3.T, repeatU bool, repeatV bool, flipU bool, flipV bool) ImageProjection {
	return ImageProjection{
		ProjectionType: projectionType,
		ImageFilename:  textureFilename,
		Origin:         origin,
		U:              u,
		V:              v,
		RepeatU:        repeatU,
		RepeatV:        repeatV,
		FlipU:          flipU,
		FlipV:          flipV,
	}
}

func (imageProjection *ImageProjection) GetUV(point *vec3.T) *color.Color {
	if imageProjection.ProjectionType == Parallel {
		return imageProjection.getParallelUV(point)
	}
	if imageProjection.ProjectionType == Cylindrical {
		return imageProjection.getCylindricalUV(point)
	}

	return &color.White
}

func (imageProjection *ImageProjection) getCylindricalUV(point *vec3.T) *color.Color {
	imageProjection.InitializeProjection()

	translatedPoint := *point
	translatedPoint.Sub(&imageProjection.Origin)

	uvPoint := imageProjection._invertedCoordinateSystemMatrix.MulVec3(&translatedPoint)

	invLength := 1 / math.Sqrt(uvPoint[0]*uvPoint[0]+uvPoint[2]*uvPoint[2])
	cosineOfAngle := uvPoint[0] * invLength
	sineOfAngle := uvPoint[2] * invLength

	var radAngle float64
	if sineOfAngle >= 0.0 {
		radAngle = math.Acos(cosineOfAngle)
	} else {
		radAngle = 2.0*math.Pi - math.Acos(cosineOfAngle)
	}

	textureLatitudeRepetitions := 1.0

	u, fracU := math.Modf((radAngle * textureLatitudeRepetitions) / (2.0 * math.Pi))
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

func (imageProjection *ImageProjection) getParallelUV(point *vec3.T) *color.Color {
	imageProjection.InitializeProjection()

	translatedPoint := *point
	translatedPoint.Sub(&imageProjection.Origin)

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
	imageProjection._image.Clear()
	imageProjection._invertedCoordinateSystemMatrix = mat3.Zero
}

func (imageProjection *ImageProjection) InitializeProjection() {
	if !imageProjection._image.ContainImage() {
		floatImage := image.LoadImageData(imageProjection.ImageFilename)
		imageProjection._image = *floatImage
	}

	if (imageProjection.ProjectionType == Cylindrical) && (imageProjection._invertedCoordinateSystemMatrix == mat3.Zero) {
		Un := imageProjection.U.Normalized()
		Vn := imageProjection.V.Normalized()
		W := vec3.Cross(&Un, &Vn)
		U := vec3.Cross(&Vn, &W)
		imageProjection._invertedCoordinateSystemMatrix = mat3.T{U, imageProjection.V, W}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			fmt.Println("Ouupps, could not initialize cylindrical projection as inverted matrix for uv system could not be created.", imageProjection.ImageFilename, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}

	if (imageProjection.ProjectionType == Parallel) && (imageProjection._invertedCoordinateSystemMatrix == mat3.Zero) {
		W := vec3.Cross(&imageProjection.U, &imageProjection.V)
		imageProjection._invertedCoordinateSystemMatrix = mat3.T{imageProjection.U, imageProjection.V, W}
		if _, err := imageProjection._invertedCoordinateSystemMatrix.Invert(); err != nil {
			fmt.Println("Ouupps, could not initialize parallel projection as invert matrixed for uv system could not be created.", imageProjection.ImageFilename, imageProjection.U, imageProjection.V, imageProjection.Origin)
			os.Exit(1)
		}
	}
}

func (camera *Camera) GetCameraCoordinateSystem() mat3.T {
	if camera._coordinateSystem == (mat3.T{}) {
		heading := camera.Heading.Normalized()

		cameraX := vec3.Cross(&camera.ViewUp, &heading)
		cameraX.Normalize()
		cameraY := vec3.Cross(&heading, &cameraX)
		cameraY.Normalize()

		camera._coordinateSystem = mat3.T{cameraX, cameraY, heading}
	}
	return camera._coordinateSystem
}
