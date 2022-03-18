package scene

import (
	"math"
	"math/rand"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

func CreateCameraRay(x int, y int, width int, height int, camera *Camera, sampleIndex int) *Ray {
	rayOrigin := camera.Origin

	cameraCoordinateSystem := camera.GetCameraCoordinateSystem()

	magnification := camera.Magnification
	if magnification == 0.0 {
		magnification = 1.0
	}

	perfectHeadingInCameraCoordinateSystem := vec3.T{
		(-float64(width/2.0) + float64(x) + 0.5) / magnification,
		(float64(height/2.0) - float64(y) - 0.5) / magnification,
		camera.ViewPlaneDistance,
	}

	if camera.AntiAlias && (camera.Samples > 1) {
		// Anti aliasing rays (random offsets within the pixel square)
		xOffset := rand.Float64() - 0.5
		yOffset := rand.Float64() - 0.5
		aliasOffset := vec3.T{xOffset, yOffset, 0}

		perfectHeadingInCameraCoordinateSystem.Add(&aliasOffset)
	}

	var headingInCameraCoordinateSystem vec3.T

	if camera.LensRadius > 0 && camera.Samples > 0 {
		cameraPointOffset := getCameraLensPoint(camera.LensRadius, camera.Samples, sampleIndex+1)
		focalPointInCameraCoordinateSystem := getCameraRayIntersectionWithFocalPlane(camera, perfectHeadingInCameraCoordinateSystem)

		headingInCameraCoordinateSystem = focalPointInCameraCoordinateSystem
		headingInCameraCoordinateSystem.Sub(&cameraPointOffset)

		cameraOffsetInSceneCoordinateSystem := cameraCoordinateSystem.MulVec3(&cameraPointOffset)
		rayOrigin.Add(&cameraOffsetInSceneCoordinateSystem)
	} else {
		headingInCameraCoordinateSystem = perfectHeadingInCameraCoordinateSystem
	}

	headingInSceneCoordinateSystem := cameraCoordinateSystem.MulVec3(&headingInCameraCoordinateSystem)

	return &Ray{
		Origin:  rayOrigin,
		Heading: headingInSceneCoordinateSystem,
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

func getCameraLensPoint(radius float64, amountSamples int, sample int) vec3.T {
	xOffset, yOffset := sunflower(amountSamples, 1.0, sample, true)
	return vec3.T{radius * xOffset, radius * yOffset, 0}
}

func getCameraRayIntersectionWithFocalPlane(camera *Camera, perfectHeading vec3.T) vec3.T {
	ray := &Ray{
		Origin:  vec3.Zero,
		Heading: perfectHeading,
	}

	focalPlane := &Plane{
		Origin: vec3.T{0, 0, camera.FocalDistance},
		Normal: vec3.T{0, 0, 1},
	}

	pointInFocalPlaneInCameraCoordinateSystem, _ := GetLinePlaneIntersectionPoint(ray, focalPlane)

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

	b := math.Round(float64(alpha) * math.Sqrt(float64(amountPoints))) // number of boundary points
	phi := (math.Sqrt(5.0) + 1.0) / 2.0                                // golden ratio
	r := sunflowerRadius(float64(pointIndex), float64(amountPoints), b)
	theta := 2.0 * math.Pi * float64(pointIndex) / (phi * phi)

	return float64(r * math.Cos(theta)), float64(r * math.Sin(theta))
}

func sunflowerRadius(i float64, n float64, b float64) float64 {
	r := float64(1) // put on the boundary
	if i <= (n - b) {
		r = math.Sqrt(i-0.5) / math.Sqrt(n-(b+1.0)/2.0) // apply square root
	}
	return r
}
