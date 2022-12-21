package scene

import (
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"strings"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

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
		(-float64(width/2.0) + float64(x) + 0.5 + aliasOffset[0]) / magnification,
		(float64(height/2.0) - float64(y) - 0.5 + aliasOffset[1]) / magnification,
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
		Origin:          &rayOrigin,
		Heading:         &headingInSceneCoordinateSystem,
		RefractionIndex: 1.000273, // Refraction index of air (at 20 degrees Celsius, STP)
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
		apertureShapeImage := img.GetCachedImage(apertureShapeImageFilepath, 1.0)
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
	return sunflower(amountSamples, 1.0, sample, true)
}

func getCameraRayIntersectionWithFocalPlane(camera *Camera, perfectHeading *vec3.T) *vec3.T {
	ray := &Ray{
		Origin:  &vec3.Zero,
		Heading: perfectHeading,
	}

	focalPlane := &Plane{
		Origin: &vec3.T{0, 0, camera.FocusDistance},
		Normal: &vec3.T{0, 0, 1},
	}

	pointInFocalPlaneInCameraCoordinateSystem, _ := GetLinePlaneIntersectionPoint2(ray, focalPlane)

	return pointInFocalPlaneInCameraCoordinateSystem
}

// Distributes n points evenly within a circle with sunflowerRadius 1
// alpha controls point distribution on the edge. Typical values 1-2, higher values more points on the edge.
// i is the index of a point. It is in the range [1,n] .
// https://stackoverflow.com/questions/28567166/uniformly-distribute-x-points-inside-a-circle
func sunflower(amountPoints int, alpha float64, pointNumber int, randomize bool) (float64, float64) { // example: amountPoints=500, alpha=2, pointNumber=[1..amountPoints]
	pointIndex := float64(pointNumber)
	if randomize {
		pointIndex += rand.Float64() - 0.5
	}

	b := math.Round(alpha * math.Sqrt(float64(amountPoints))) // number of boundary points
	phi := (math.Sqrt(5.0) + 1.0) / 2.0                       // golden ratio
	r := sunflowerRadius(pointIndex, float64(amountPoints), b)
	theta := 2.0 * math.Pi * float64(pointIndex) / (phi * phi)

	return r * math.Cos(theta), r * math.Sin(theta)
}

func sunflowerRadius(i float64, n float64, b float64) float64 {
	r := float64(1) // put on the boundary
	if i <= (n - b) {
		r = math.Sqrt(i-0.5) / math.Sqrt(n-(b+1.0)/2.0) // apply square root
	}
	return r
}
