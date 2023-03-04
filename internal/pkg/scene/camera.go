package scene

import (
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"pathtracer/internal/pkg/util"
	"strings"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

type Camera struct {
	Origin            *vec3.T
	Heading           *vec3.T
	ViewUp            *vec3.T
	ViewPlaneDistance float64 // ViewPlaneDistance determine the focal length, the view angle of the camera.
	_coordinateSystem *mat3.T
	ApertureSize      float64 // ApertureSize is the size of the aperture opening. The wider the aperture the less focus depth. Value 0.0 is infinite focus depth.
	ApertureShape     string  // ApertureShape is the file path to a black and white image where white define the aperture shape. Aperture size determine the size of the longest side (width or height) of the image. If empty string then a default round aperture shape is used.
	FocusDistance     float64
	Samples           int
	AntiAlias         bool
	Magnification     float64
	RenderType        RenderType
	RecursionDepth    int
}

func NewCamera(origin *vec3.T, viewPoint *vec3.T, amountSamples int, magnification float64) *Camera {
	heading := viewPoint.Subed(origin)
	focusDistance := heading.Length()
	heading.Normalize()

	return &Camera{
		Origin:            origin,
		Heading:           &heading,
		ViewUp:            &vec3.UnitY,
		ViewPlaneDistance: 800,
		ApertureSize:      0.0, // Use default aperture, with no "Depth of Field" (DOF).
		ApertureShape:     "",  // Use default, round aperture.
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        Pathtracing,
		RecursionDepth:    4,
	}
}

func (camera *Camera) R(origin *vec3.T, focusPoint *vec3.T, updateFocusDistance bool) *Camera {
	heading := focusPoint.Subed(origin)
	if updateFocusDistance {
		camera.FocusDistance = heading.Length()
	}
	heading.Normalize()

	camera.Origin = origin
	camera.Heading = &heading

	return camera
}

func (camera *Camera) S(amountSamples int) *Camera {
	camera.Samples = amountSamples
	return camera
}

func (camera *Camera) D(maxRayDepth int) *Camera {
	camera.RecursionDepth = maxRayDepth
	return camera
}

func (camera *Camera) A(apertureSize float64, apertureShape string) *Camera {
	camera.ApertureSize = apertureSize
	camera.ApertureShape = apertureShape
	return camera
}

func (camera *Camera) F(focusDistance float64) *Camera {
	camera.FocusDistance = focusDistance
	return camera
}

func (camera *Camera) V(viewPlaneDistance float64) *Camera {
	camera.ViewPlaneDistance = viewPlaneDistance
	return camera
}

func (camera *Camera) M(magnification float64) *Camera {
	camera.Magnification = magnification
	return camera
}

func CreateCameraRay(x int, y int, width int, height int, camera *Camera, sampleIndex int) *Ray {
	rayOrigin := *camera.Origin

	cameraCoordinateSystem := camera.GetCameraCoordinateSystem()

	magnification := camera.Magnification
	if magnification == 0.0 {
		magnification = 1.0
	}

	aliasOffset := vec2.T{0, 0}
	if camera.AntiAlias && (camera.Samples > 1) {
		// Anti aliasing rays (random offsets within the pixel square)
		xOffset := rand.Float64() - 0.5
		yOffset := rand.Float64() - 0.5
		aliasOffset = vec2.T{xOffset, yOffset}
	}

	perfectHeadingInCameraCoordinateSystem := &vec3.T{
		(-float64(width)/2.0 + float64(x) + 0.5 + aliasOffset[0]) / magnification,
		(float64(height)/2.0 - float64(y) - 0.5 + aliasOffset[1]) / magnification,
		camera.ViewPlaneDistance,
	}

	var headingInCameraCoordinateSystem *vec3.T

	if camera.ApertureSize > 0 && camera.Samples > 0 {
		cameraPointOffset := getCameraLensPoint(camera.ApertureSize, camera.ApertureShape, camera.Samples, sampleIndex+1)
		focalPointInCameraCoordinateSystem := getCameraRayIntersectionWithFocalPlane(camera, perfectHeadingInCameraCoordinateSystem)

		headingInCameraCoordinateSystem = focalPointInCameraCoordinateSystem
		headingInCameraCoordinateSystem.Sub(&cameraPointOffset)

		cameraOffsetInSceneCoordinateSystem := cameraCoordinateSystem.MulVec3(&cameraPointOffset)
		rayOrigin.Add(&cameraOffsetInSceneCoordinateSystem)
	} else {
		headingInCameraCoordinateSystem = perfectHeadingInCameraCoordinateSystem
	}

	headingInSceneCoordinateSystem := cameraCoordinateSystem.MulVec3(headingInCameraCoordinateSystem)
	headingInSceneCoordinateSystem.Normalize()

	return &Ray{
		Origin:  &rayOrigin,
		Heading: &headingInSceneCoordinateSystem,
	}
}

func (camera *Camera) GetCameraCoordinateSystem() *mat3.T {
	if camera._coordinateSystem == nil {
		heading := camera.Heading.Normalized()

		cameraX := vec3.Cross(camera.ViewUp, &heading)
		cameraX.Normalize()
		cameraY := vec3.Cross(&heading, &cameraX)
		cameraY.Normalize()

		camera._coordinateSystem = &mat3.T{cameraX, cameraY, heading}
	}
	return camera._coordinateSystem
}

func getCameraLensPoint(radius float64, apertureShapeImageFilepath string, amountSamples int, sample int) vec3.T {
	xOffset := 0.0
	yOffset := 0.0

	if strings.TrimSpace(apertureShapeImageFilepath) != "" {
		apertureShapeImage := img.GetCachedImage(apertureShapeImageFilepath)
		xOffset, yOffset = shapedApertureOffset(apertureShapeImage)
	} else {
		xOffset, yOffset = roundApertureOffset(amountSamples, sample)
	}

	return vec3.T{radius * xOffset, radius * yOffset, 0}
}

// shapedApertureOffset gives a xy-offset, where both x and y are in the range [-1,1]
// https://blog.demofox.org/2018/07/04/pathtraced-depth-of-field-bokeh/
func shapedApertureOffset(image *img.FloatImage) (float64, float64) {
	maxSize := math.Max(float64(image.Width), float64(image.Height))

	offsetX := 0.0
	offsetY := 0.0

	for c := color.Black; c != color.White; { // TODO be smarter than re-iterating until we randomly hit a white pixel...
		x := rand.Intn(image.Width)
		y := rand.Intn(image.Height)

		offsetX = (float64(x)/(maxSize-1))*2 - (float64(image.Width) / maxSize)
		offsetY = (float64(y)/(maxSize-1))*2 - (float64(image.Height) / maxSize)

		c = *image.GetPixel(x, (image.Height-1)-y)
	}

	return offsetX, offsetY
}

func roundApertureOffset(amountSamples int, sample int) (float64, float64) {
	return util.Sunflower(amountSamples, 0.0, sample, true)
}

func getCameraRayIntersectionWithFocalPlane(camera *Camera, perfectHeading *vec3.T) *vec3.T {
	ray := &Ray{
		Origin:  &vec3.T{0, 0, 0},
		Heading: perfectHeading,
	}

	focalPlane := &Plane{
		Origin: &vec3.T{0, 0, camera.FocusDistance},
		Normal: &vec3.T{0, 0, 1},
	}

	pointInFocalPlaneInCameraCoordinateSystem, _ := GetLinePlaneIntersectionPoint2(ray, focalPlane)

	return pointInFocalPlaneInCameraCoordinateSystem
}
