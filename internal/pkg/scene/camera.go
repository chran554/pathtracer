package scene

import (
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"pathtracer/internal/pkg/sunflower"
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
	return sunflower.Sunflower(amountSamples, 0.0, sample, true)
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

type FrameFormat struct {
	Name   string
	Width  float64 // Width is the width of the camera sensor or camera film frame in mm.
	Height float64 // Height is the height of the camera sensor or camera film frame in mm.
}

var (
	FrameFormat35mm       = &FrameFormat{Name: "35mm (135 film for analog cameras)", Width: 36, Height: 24} // "35mm" is the most common analog film format for cameras and sensor size for "full format" ("FX" for Nikon). https://en.wikipedia.org/wiki/135_film
	FrameFormatFullFormat = FrameFormat35mm                                                                 // "35mm" is the most common analog film format for cameras and sensor size for "full format" ("FX" for Nikon). https://en.wikipedia.org/wiki/135_film
)

func NewFrameFormat(name string, width float64, height float64) FrameFormat {
	return FrameFormat{Name: name, Width: width, Height: height}
}

// AngleOfViewScreen gives the (diagonal) angle of view of screen viewed at a certain distance.
// The angle is in radians.
//
// distance is the distance from the camera/eye to the screen measured in the unit "pixels".
func AngleOfViewScreen(resolution ScreenResolution, distanceInPixels float64) float64 {
	diagonal := math.Sqrt(float64(resolution.width*resolution.width + resolution.height*resolution.height))
	return 2.0 * math.Atan((diagonal/2.0)/distanceInPixels)
}

// AngleOfViewScreenWithPixelDensity gives the (diagonal) angle of view of screen viewed at a certain distance.
// The angle is in radians.
//
// screenPixelDensityDPI is the pixel density DPI ("Dots Per Inch") of the screen. Basically it is the size measurement of the screen pixels.
// distanceInMeters is the distance from the camera/eye to the screen measured in the unit meters.
//
// Sorry about the mashup of unit systems (meters vs incomprehensible British imperial units),
// but unfortunately the unit DPI is widely used for screen pixel density.
// To make things easy for you to convert, an inch is 0.0254 meters or 1/36 of a yard or 1/12 of a foot.
// So go on, run and measure your feet now...
func AngleOfViewScreenWithPixelDensity(resolution ScreenResolution, screenPixelDensityDPI float64, distanceInMeters float64) float64 {
	distanceInPixels := (distanceInMeters / 0.0254) * screenPixelDensityDPI
	return AngleOfViewScreen(resolution, distanceInPixels)
}

func (ff *FrameFormat) AspectRatio() float64 {
	return ff.Width / ff.Height
}

// angleOfView gets the angle of view for a frame format in radians.
//
// https://www.omnicalculator.com/other/camera-field-of-view
//
// fov = 2 * atan(sensorSize/(2*focalLength))
//
// sensorSize is given by camera manufacturer specs and can be the film or sensor width, height, or diagonal width (in mm).
//
// focalLength is the lens value of "zoom". It is specified in the unit of mm and that number is almost always imprinted on camera lenses.
// For common camera film (35mm film) or "full format" digital camera a lens with imprinted 28mm is "wide angle" lenses, 50mm lenses are "normal" and above 50mm are "zoom" lenses.
//
// Verified against https://www.nikonians.org/reviews/fov-tables
func angleOfView(size float64, focalLength float64) float64 {
	return 2.0 * math.Atan(size/(2.0*focalLength))
}

// AngleOfView gets the diagonal angle of view for a frame format in radians.
//
// focalLength is the lens value of "zoom". It is specified in the unit of mm and that number is almost always imprinted on camera lenses.
// For common camera film (35mm film) or "full format" digital camera a lens with imprinted 28mm is "wide angle" lenses, 50mm lenses are "normal" and above 50mm are "zoom" lenses.
func (ff *FrameFormat) AngleOfView(focalLength float64) float64 {
	return angleOfView(ff.Diagonal(), focalLength)
}

// AngleOfViewWidth gets the width/horizontal angle of view for a frame format in radians.
//
// focalLength is the lens value of "zoom". It is specified in the unit of mm and that number is almost always imprinted on camera lenses.
// For common camera film (35mm film) or "full format" digital camera a lens with imprinted 28mm is "wide angle" lenses, 50mm lenses are "normal" and above 50mm are "zoom" lenses.
func (ff *FrameFormat) AngleOfViewWidth(focalLength float64) float64 {
	return angleOfView(ff.Width, focalLength)
}

// AngleOfViewHeight gets the height/vertical angle of view for a frame format in radians.
//
// focalLength is the lens value of "zoom". It is specified in the unit of mm and that number is almost always imprinted on camera lenses.
// For common camera film (35mm film) or "full format" digital camera a lens with imprinted 28mm is "wide angle" lenses, 50mm lenses are "normal" and above 50mm are "zoom" lenses.
func (ff *FrameFormat) AngleOfViewHeight(focalLength float64) float64 {
	return angleOfView(ff.Height, focalLength)
}

// Diagonal gets the diagonal length of the camera sensor or camera film frame in mm.
func (ff *FrameFormat) Diagonal() float64 {
	return math.Sqrt(ff.Width*ff.Width + ff.Height*ff.Height)
}

func fieldOfView(angleOfView float64, distance float64) float64 {
	return 2.0 * math.Tan(angleOfView/2.0) * distance
}

// FieldOfViewWidth gets the depicted length horizontally of the "real world" at a certain distance from a lens with a certain focal length (angle).
// The unit of the horizontal length is the same as used for the distance.
func (ff *FrameFormat) FieldOfViewWidth(focalLength float64, distance float64) float64 {
	return fieldOfView(ff.AngleOfViewWidth(focalLength), distance)
}

// FieldOfViewHeight gets the depicted height vertically of the "real world" at a certain distance from a lens with a certain focal length (angle).
// The unit of the vertical length is the same as used for the distance.
func (ff *FrameFormat) FieldOfViewHeight(focalLength float64, distance float64) float64 {
	return fieldOfView(ff.AngleOfViewHeight(focalLength), distance)
}

// LensFocalLength
//
// Verified against https://www.nikonians.org/reviews/fov-tables
func (ff *FrameFormat) LensFocalLength(diagonalAngleOfView float64) float64 {
	return ff.Diagonal() / (2.0 * math.Tan(diagonalAngleOfView/2.0))
}
